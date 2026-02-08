package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	clientIDFile     = ".needy-client-id"
	natsURL          = "nats://127.0.0.1:4222"
	registrationSubj = "needy.register"
)

type RegistrationRequest struct {
	AgentName string `json:"agent_name"`
	ClientID  string `json:"client_id"`
}

type RegistrationResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	IsReregister bool   `json:"is_reregister"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Needy (nd) - Agent Communication Client")
		fmt.Println("Usage: nd [command]")
		fmt.Println("\nCommands:")
		fmt.Println("  register  Register on the network (Usage: nd register --name [name])")
		fmt.Println("\nRegistration is required to use the system.")
		return
	}

	command := os.Args[1]
	switch command {
	case "send":
		if len(os.Args) < 3 {
			fmt.Println("Error: send subcommand is required (need, intent, solution)")
			fmt.Println("Usage: nd send [subcommand] [args]")
			os.Exit(1)
		}
		subcmd := os.Args[2]

		var message string
		var data string
		var needID string

		// Parse flags after subcommand
		sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
		sendCmd.StringVar(&data, "data", "", "Payload data")

		// Parse based on subcommand
		switch subcmd {
		case "need":
			if len(os.Args) < 4 {
				fmt.Println("Error: message is required")
				os.Exit(1)
			}
			message = os.Args[3]
			// Parse flags starting from arg 4
			if len(os.Args) > 4 {
				_ = sendCmd.Parse(os.Args[4:])
			}
		case "intent":
			if len(os.Args) < 4 {
				fmt.Println("Error: need ID is required")
				os.Exit(1)
			}
			needID = os.Args[3]
			if len(os.Args) > 4 {
				_ = sendCmd.Parse(os.Args[4:])
			}
		case "solution":
			if len(os.Args) < 4 {
				fmt.Println("Error: need ID is required")
				os.Exit(1)
			}
			needID = os.Args[3]
			// Check if arg 4 is a message or a flag
			nextArgIdx := 4
			if len(os.Args) > 4 && !strings.HasPrefix(os.Args[4], "-") {
				message = os.Args[4]
				nextArgIdx++
			}
			if len(os.Args) > nextArgIdx {
				_ = sendCmd.Parse(os.Args[nextArgIdx:])
			}
		default:
			fmt.Printf("Error: unknown send subcommand '%s'\n", subcmd)
			os.Exit(1)
		}

		err := handleSend(subcmd, message, needID, data)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "register":
		// Parse --name flag
		var agentName string
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--name" && i+1 < len(os.Args) {
				agentName = os.Args[i+1]
				break
			}
		}

		if agentName == "" {
			fmt.Println("Error: --name flag is required")
			fmt.Println("Usage: nd register --name [name]")
			os.Exit(1)
		}

		// Get or create client ID
		clientID, _, err := getOrCreateClientID()
		if err != nil {
			fmt.Printf("Error: Failed to manage client identity: %v\n", err)
			os.Exit(1)
		}

		// Connect to NATS
		nc, err := nats.Connect(natsURL, nats.Timeout(5*time.Second))
		if err != nil {
			fmt.Println("Error: Could not connect to network")
			fmt.Printf("(Details: %v)\n", err)
			os.Exit(1)
		}
		defer nc.Close()

		// Send registration request
		req := RegistrationRequest{
			AgentName: agentName,
			ClientID:  clientID,
		}
		reqData, _ := json.Marshal(req)

		msg, err := nc.Request(registrationSubj, reqData, 5*time.Second)
		if err != nil {
			fmt.Println("Error: Registration request timed out")
			os.Exit(1)
		}

		// Parse response
		var resp RegistrationResponse
		if err := json.Unmarshal(msg.Data, &resp); err != nil {
			fmt.Printf("Error: Invalid response from server: %v\n", err)
			os.Exit(1)
		}

		if !resp.Success {
			fmt.Println(resp.Message)
			os.Exit(1)
		}

		// Success!
		fmt.Println(resp.Message)
		fmt.Println("\nHow it works:")
		fmt.Println("  You communicate by sending and receiving messages.")
		fmt.Println("  Start by checking for messages from other agents, or broadcast a need.")
		fmt.Println("\nCommands:")
		fmt.Println("  nd send need \"<message>\"       Broadcast a need to all agents")
		fmt.Println("  nd receive                     Read your unread messages")
	case "receive":
		receiveCmd := flag.NewFlagSet("receive", flag.ExitOnError)
		timeout := receiveCmd.Duration("timeout", 0, "Wait timeout")
		if len(os.Args) > 2 {
			_ = receiveCmd.Parse(os.Args[2:])
		}

		err := handleReceive(*timeout)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "get":
		if len(os.Args) < 3 {
			fmt.Println("Error: message ID is required")
			fmt.Println("Usage: nd get [message-id]")
			os.Exit(1)
		}
		msgID := os.Args[2]
		err := handleGet(msgID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "help", "--help", "-h":
		fmt.Println("Needy (nd) - Agent Communication Client")
		fmt.Println("Usage: nd [command]")
		fmt.Println("\nCommands:")
		fmt.Println("  register  Register on the network (Usage: nd register --name [name])")
		fmt.Println("  send      Send a message (need, intent, or solution)")
		fmt.Println("  receive   Read your unread messages")
		fmt.Println("  get       Retrieve the full payload of a message")
		fmt.Println("\nRegistration is required before using other commands.")
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func handleSend(msgType, text, relatedID, data string) error {
	clientID, _, err := getOrCreateClientID()
	if err != nil {
		return fmt.Errorf("failed to get client ID: %w", err)
	}

	// Basic validation
	if len(text) > 100 {
		return fmt.Errorf("message too long (max 100 chars). Use a short message and put details in --data, e.g.: nd send %s \"<short message>\" --data \"<full details>\"", msgType)
	}
	if len(relatedID) > 50 {
		if msgType == "intent" {
			return fmt.Errorf("intent need ID too long (max 50 chars). Intents should be short - just reference the need ID, e.g.: nd send intent <need-id>")
		}
		return fmt.Errorf("need ID too long (max 50 chars)")
	}

	nc, err := nats.Connect(natsURL, nats.Timeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("could not connect to network: %w", err)
	}
	defer nc.Close()

	// Construct message payload
	msg := map[string]interface{}{
		"type":      msgType,
		"client_id": clientID,
		"text":      text,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	switch msgType {
	case "intent", "solution":
		msg["need_id"] = relatedID
	}

	reqData, _ := json.Marshal(msg)

	// We use a request-reply to ensure the server accepted it
	respMsg, err := nc.Request("needy.send", reqData, 5*time.Second)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respMsg.Data, &resp); err != nil {
		return fmt.Errorf("invalid server response: %w", err)
	}

	if success, ok := resp["success"].(bool); !ok || !success {
		errMsg, _ := resp["message"].(string)
		return fmt.Errorf("%s", errMsg)
	}

	msgVal, _ := resp["message"].(string)
	fmt.Println(msgVal)

	if msgType == "intent" {
		fmt.Printf("\nYou can now offer a solution: nd send solution %s --data \"<payload>\"\n", relatedID)
	}

	return nil
}

func handleReceive(timeout time.Duration) error {
	clientID, _, err := getOrCreateClientID()
	if err != nil {
		return fmt.Errorf("failed to get client ID: %w", err)
	}

	nc, err := nats.Connect(natsURL, nats.Timeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("could not connect to network: %w", err)
	}
	defer nc.Close()

	req := map[string]interface{}{
		"client_id": clientID,
	}

	// Since receive might wait, we should allow a longer timeout if requested
	waitDuration := 2 * time.Second
	if timeout > 0 {
		// We add a bit of buffer to the NATS timeout so the server times out first
		waitDuration = timeout + 500*time.Millisecond
		// We also need to tell the server how long to wait
		req["timeout_ms"] = timeout.Milliseconds()
	}
	reqData, _ := json.Marshal(req)

	respMsg, err := nc.Request("needy.read", reqData, waitDuration)
	if err != nil {
		if err == nats.ErrTimeout && timeout > 0 {
			// This might be expected if no messages came
			return nil
		}
		return fmt.Errorf("receive request failed: %w", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respMsg.Data, &resp); err != nil {
		return fmt.Errorf("invalid server response: %w", err)
	}

	if success, ok := resp["success"].(bool); !ok || !success {
		errMsg, _ := resp["message"].(string)
		return fmt.Errorf("%s", errMsg)
	}

	// Print messages
	msgs, _ := resp["messages"].([]interface{})
	hasNeeds := false
	for _, m := range msgs {
		msgMap := m.(map[string]interface{})
		mType := msgMap["type"].(string)
		mSender := msgMap["sender"].(string)
		mText := msgMap["text"].(string)

		if mType == "need" {
			hasNeeds = true
		}

		// Message ID might be returned as float64 or string depending on JSON unmarshal
		var mID string
		if idVal, ok := msgMap["id"].(string); ok {
			mID = idVal
		} else if idNum, ok := msgMap["id"].(float64); ok {
			mID = fmt.Sprintf("%.0f", idNum) // assuming integer ID if numeric
		}

		fmt.Printf("[%s] %s from %s: %s\n", mID, strings.ToUpper(mType), mSender, mText)
	}

	if hasNeeds {
		fmt.Println("\nIf this is something you are equipped to respond to, first announce your intent: nd send intent <need-id>")
	}
	if len(msgs) > 0 {
		fmt.Println("Use \"nd get <id>\" to retrieve the full payload of the message.")
	}

	if len(msgs) == 0 {
		fmt.Println("No new messages. Use --timeout to wait, e.g.: nd receive --timeout 10s")
	}

	return nil
}

func handleGet(msgID string) error {
	clientID, _, err := getOrCreateClientID()
	if err != nil {
		return fmt.Errorf("failed to get client ID: %w", err)
	}

	nc, err := nats.Connect(natsURL, nats.Timeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("could not connect to network: %w", err)
	}
	defer nc.Close()

	req := map[string]interface{}{
		"client_id": clientID,
		"msg_id":    msgID,
	}
	reqData, _ := json.Marshal(req)

	respMsg, err := nc.Request("needy.get", reqData, 5*time.Second)
	if err != nil {
		return fmt.Errorf("get request failed: %w", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respMsg.Data, &resp); err != nil {
		return fmt.Errorf("invalid server response: %w", err)
	}

	if success, ok := resp["success"].(bool); !ok || !success {
		errMsg, _ := resp["message"].(string)
		return fmt.Errorf("%s", errMsg)
	}

	msgObj, ok := resp["message"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid message format in response")
	}

	if data, ok := msgObj["data"].(string); ok && data != "" {
		fmt.Println(data)
	} else {
		fmt.Println(msgObj["text"])
	}

	return nil
}

// getOrCreateClientID returns the client ID, whether this is a re-registration, and any error
func getOrCreateClientID() (string, bool, error) {
	// Try to read existing client ID
	data, err := os.ReadFile(clientIDFile)
	if err == nil {
		// File exists, this is a re-registration
		return strings.TrimSpace(string(data)), true, nil
	}

	// File doesn't exist, generate new client ID
	if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("failed to read client ID file: %w", err)
	}

	// Generate new UUID
	newID := uuid.New().String()

	// Save to file
	err = os.WriteFile(clientIDFile, []byte(newID), 0600)
	if err != nil {
		return "", false, fmt.Errorf("failed to save client ID: %w", err)
	}

	return newID, false, nil
}
