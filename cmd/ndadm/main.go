package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

const (
	configFile       = ".needy.conf"
	defaultPort      = 4222
	registrationSubj = "needy.register"
	messageStream    = "MESSAGES"
	messageSubj      = "needy.messages"
)

func readConfig() map[string]string {
	cfg := map[string]string{}
	f, err := os.Open(configFile)
	if err != nil {
		return cfg
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			cfg[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return cfg
}

func getPort() int {
	cfg := readConfig()
	port := defaultPort
	if p, ok := cfg["port"]; ok {
		_, _ = fmt.Sscanf(p, "%d", &port)
	}
	return port
}

type RegistrationRequest struct {
	AgentName string `json:"agent_name"`
	ClientID  string `json:"client_id"`
}

type RegistrationResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	IsReregister bool   `json:"is_reregister"`
}

// Global registry instance
var registry = NewRegistry()

func main() {
	natsPort := getPort()

	// Start embedded NATS server with JetStream
	opts := &server.Options{
		Host:      "127.0.0.1",
		Port:      natsPort,
		JetStream: true,
		StoreDir:  "./.nats-data",
	}

	ns, err := server.NewServer(opts)
	if err != nil {
		log.Fatalf("Failed to create NATS server: %v", err)
	}

	go ns.Start()

	if !ns.ReadyForConnections(5000000000) { // 5 second timeout
		log.Fatal("NATS server failed to start")
	}

	fmt.Printf("ndadm: NATS server started on port %d\n", natsPort)

	// Connect to our own NATS server
	nc, err := nats.Connect(fmt.Sprintf("nats://127.0.0.1:%d", natsPort))
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Setup JetStream
	if err := setupJetStream(nc); err != nil {
		log.Fatalf("Failed to setup JetStream: %v", err)
	}

	// Subscribe to registration requests
	_, err = nc.Subscribe(registrationSubj, handleRegistration)
	if err != nil {
		log.Fatalf("Failed to subscribe to registrations: %v", err)
	}

	// Subscribe to send requests
	_, err = nc.Subscribe("needy.send", func(msg *nats.Msg) {
		handleSend(nc, msg)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to send: %v", err)
	}

	// Subscribe to read requests
	_, err = nc.Subscribe("needy.read", func(msg *nats.Msg) {
		handleRead(nc, msg)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to read: %v", err)
	}

	// Subscribe to get requests
	_, err = nc.Subscribe("needy.get", func(msg *nats.Msg) {
		handleGet(nc, msg)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to get: %v", err)
	}

	fmt.Println("ndadm: Listening for agent registrations...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nndadm: Shutting down...")
	_ = nc.Drain()
	ns.Shutdown()
	ns.WaitForShutdown()
}

func handleRegistration(msg *nats.Msg) {
	var req RegistrationRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		log.Printf("Invalid registration request: %v", err)
		return
	}

	success, message, isReregister := registry.RegisterAgent(req.AgentName, req.ClientID)

	if success && !isReregister {
		fmt.Printf("ndadm: Registered agent '%s' with client ID %s\n", req.AgentName, req.ClientID)
	}

	resp := RegistrationResponse{
		Success:      success,
		Message:      message,
		IsReregister: isReregister,
	}

	// Send response
	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)
}

func handleSend(nc *nats.Conn, msg *nats.Msg) {
	var req map[string]interface{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		_ = msg.Respond([]byte(`{"success": false, "message": "Invalid payload"}`))
		return
	}

	clientID, _ := req["client_id"].(string)
	agentName := registry.GetAgentName(clientID)
	if agentName == "" {
		_ = msg.Respond([]byte(`{"success": false, "message": "Not registered. Please register first: nd register --name <your-name>"}`))
		return
	}

	js, _ := nc.JetStream()

	// Validate Intent logic
	msgType, _ := req["type"].(string)
	needID, _ := req["need_id"].(string)

	if msgType == "intent" {
		if needID != "" {
			registry.RecordIntent(agentName, needID)
		}
	}

	if msgType == "solution" {
		if needID == "" {
			_ = msg.Respond([]byte(`{"success": false, "message": "Solution must provide need_id"}`))
			return
		}
		if !registry.HasIntent(agentName, needID) {
			_ = msg.Respond([]byte(`{"success": false, "message": "You must first announce intent to respond"}`))
			return
		}
	}

	newMsg := Message{
		Type:      msgType,
		Sender:    agentName,
		Text:      req["text"].(string),
		NeedID:    needID,
		Timestamp: makeTimestamp(),
	}
	if d, ok := req["data"].(string); ok {
		newMsg.Data = d
	}

	msgData, _ := json.Marshal(newMsg)

	// Publish to stream
	// We publish to the subject tracked by the stream
	_, err := js.Publish(messageSubj, msgData)
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		_ = msg.Respond([]byte(`{"success": false, "message": "Internal error storing message"}`))
		return
	}

	_ = msg.Respond([]byte(fmt.Sprintf(`{"success": true, "message": "Sent %s successfully"}`, msgType)))
	fmt.Printf("ndadm: Agent '%s' sent %s\n", agentName, msgType)
}

func handleRead(nc *nats.Conn, msg *nats.Msg) {
	var req map[string]interface{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		_ = msg.Respond([]byte(`{"success": false, "message": "Invalid payload"}`))
		return
	}

	clientID, _ := req["client_id"].(string)
	agentName := registry.GetAgentName(clientID)
	if agentName == "" {
		_ = msg.Respond([]byte(`{"success": false, "message": "Not registered. Please register first: nd register --name <your-name>"}`))
		return
	}

	js, _ := nc.JetStream()

	// Create durable consumer name based on agent name
	consumerName := fmt.Sprintf("AGENT_%s", agentName)

	// Pull subscription
	// We bind to the stream and use the durable name
	// If it doesn't exist, we might need to create it or just subscribing creates it if we use Durable?
	// With PullSubscribe, we should specify durable.
	sub, err := js.PullSubscribe(messageSubj, consumerName, nats.BindStream(messageStream))
	if err != nil {
		log.Printf("Subscribe failed: %v", err)
		_ = msg.Respond([]byte(`{"success": false, "message": "Mailbox error"}`))
		return
	}

	// Fetch messages, using timeout from request if provided
	waitDuration := 100 * time.Millisecond
	if timeoutMs, ok := req["timeout_ms"].(float64); ok && timeoutMs > 0 {
		waitDuration = time.Duration(timeoutMs) * time.Millisecond
	}
	msgs, _ := sub.Fetch(10, nats.MaxWait(waitDuration))

	responseMsgs := []map[string]interface{}{}

	for _, m := range msgs {
		var payload Message
		_ = json.Unmarshal(m.Data, &payload)

		// Map for response
		meta, _ := m.Metadata()
		seq := uint64(0)
		if meta != nil {
			seq = meta.Sequence.Stream
		}

		rMsg := map[string]interface{}{
			"id":        fmt.Sprintf("%d", seq), // Use JetStream sequence as ID
			"type":      payload.Type,
			"sender":    payload.Sender,
			"text":      payload.Text,
			"data":      payload.Data,
			"timestamp": payload.Timestamp,
		}
		responseMsgs = append(responseMsgs, rMsg)
		_ = m.Ack()
	}

	resp := map[string]interface{}{
		"success":  true,
		"messages": responseMsgs,
	}
	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)
}

func handleGet(nc *nats.Conn, msg *nats.Msg) {
	var req map[string]interface{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		_ = msg.Respond([]byte(`{"success": false, "message": "Invalid payload"}`))
		return
	}

	clientID, _ := req["client_id"].(string)
	agentName := registry.GetAgentName(clientID)
	if agentName == "" {
		_ = msg.Respond([]byte(`{"success": false, "message": "Not registered. Please register first: nd register --name <your-name>"}`))
		return
	}

	msgIDStr, _ := req["msg_id"].(string)
	if msgIDStr == "" {
		_ = msg.Respond([]byte(`{"success": false, "message": "msg_id is required"}`))
		return
	}

	js, _ := nc.JetStream()

	// Convert msgID to sequence
	var seq uint64
	_, _ = fmt.Sscanf(msgIDStr, "%d", &seq)

	// Using GetMsg on stream requires stream name
	m, err := js.GetMsg(messageStream, seq)
	if err != nil {
		log.Printf("GetMsg failed: %v", err)
		_ = msg.Respond([]byte(`{"success": false, "message": "Message not found"}`))
		return
	}

	var payload Message
	_ = json.Unmarshal(m.Data, &payload)

	resp := map[string]interface{}{
		"success": true,
		"message": payload,
	}
	respData, _ := json.Marshal(resp)
	_ = msg.Respond(respData)
}

func makeTimestamp() int64 {
	return time.Now().Unix()
}

// Message types
type Message struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "need", "intent", "solution"
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Data      string `json:"data,omitempty"`
	NeedID    string `json:"need_id,omitempty"`   // For intent/solution
	IntentID  string `json:"intent_id,omitempty"` // For solution
	Timestamp int64  `json:"timestamp"`
}

func setupJetStream(nc *nats.Conn) error {
	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("failed to get JetStream context: %w", err)
	}

	// Create or update the message stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     messageStream,
		Subjects: []string{messageSubj},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	fmt.Println("ndadm: JetStream message stream ready")
	return nil
}
