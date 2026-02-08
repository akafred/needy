package features

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
)

func InitializeCommunicationSteps(ctx *godog.ScenarioContext) {
	// Learning about sending
	ctx.Step(`^the client is registered$`, theClientIsRegistered)
	ctx.Step(`^I run the `+"`"+`nd register`+"`"+` command$`, iRunTheNdRegisterCommand)
	ctx.Step(`^registration is successful$`, registrationIsSuccessful)
	ctx.Step(`^the output should explain the mechanics of sending and receiving messages$`, theOutputShouldExplainTheMechanicsOfSendingAndReceiving)

	// Learning about getting payloads
	ctx.Step(`^there is a need message to receive$`, thereIsANeedMessageToReceive)
	ctx.Step(`^I run the `+"`"+`nd receive`+"`"+` command$`, iRunTheNdReceiveCommand)
	ctx.Step(`^the output should explain the usage of the "get" command and include the id of the message$`, theOutputShouldExplainTheUsageOfTheGetCommand)

	// Learning about the workflow of intents
	ctx.Step(`^the output should explain that the agent must issue a "intent" message referring to the id of the need if it wants to offer a solution$`, theOutputShouldExplainTheWorkflowOfIntents)

	// Learning about the workflow of solutions
	ctx.Step(`^there is a need$`, thereIsANeed)
	ctx.Step(`^I have issued a "intent" message$`, iHaveIssuedAIntentMessage)
	ctx.Step(`^the output should explain how to offer a solution by issuing a "solution" message referring to the id of the need$`, theOutputShouldExplainHowToOfferASolution)

	// Learning about the limited space for messages (Need)
	ctx.Step(`^I use a to long need message with a payload$`, iUseAToLongNeedMessageWithAPayload)
	ctx.Step(`^the output should explain that the message is too long and that it must contain a message and a payload$`, theOutputShouldExplainThatTheMessageIsTooLong)

	// Learning about the length of intent messages
	ctx.Step(`^I use a to intent message that is too long$`, iUseAToLongIntentMessage)
	ctx.Step(`^the output should explain that intents should be short and not have payloads$`, theOutputShouldExplainIntentPolicy)

	// Learning about the limited space for solution messages
	ctx.Step(`^I use a to long solution message$`, iUseAToLongSolutionMessage)
	ctx.Step(`^the output should explain that the solution message is too long and that it must contain a message and a payload$`, theOutputShouldExplainThatTheMessageIsTooLong)

	// Learning about the listening with timeout
	ctx.Step(`^I run the `+"`"+`nd receive`+"`"+` command and there is nothing to receive$`, iRunTheNdReceiveCommandWithNothing)
	ctx.Step(`^the output should explain the mechanics of listening with a timeout$`, theOutputShouldExplainTimeoutMechanics)

	// Core scenarios
	ctx.Step(`^agent "([^"]*)" runs "([^"]*)"$`, agentRunsCommand)
	ctx.Step(`^agent "([^"]*)" has sent a need "([^"]*)"$`, agentHasSentANeed)
	ctx.Step(`^agent "([^"]*)" should receive a message with text "([^"]*)"$`, agentShouldReceiveAMessageWithText)
	ctx.Step(`^the command should fail with "([^"]*)"$`, theCommandShouldFailWith)
	ctx.Step(`^the command should succeed$`, theCommandShouldSucceed)

	ctx.Step(`^I am registered as "([^"]*)"$`, iAmRegisteredAs)
	ctx.Step(`^another agent "([^"]*)" is registered$`, anotherAgentIsRegistered)
	ctx.Step(`^a registered agent "([^"]*)"$`, aRegisteredAgent)
}

// Learning steps implementation

func theClientIsRegistered() error {
	return iRunTheNdRegisterCommand()
}

func iRunTheNdRegisterCommand() error {
	return runCmd("../bin/nd", "register", "--name", "DiscoveryAgent")
}

func registrationIsSuccessful() error {
	if lastError != nil {
		return fmt.Errorf("registration failed: %v", lastError)
	}
	return nil
}

func theOutputShouldExplainTheMechanicsOfSendingAndReceiving() error {
	if !strings.Contains(lastOutput, "send") || !strings.Contains(lastOutput, "receive") {
		return fmt.Errorf("expected output to explain send/receive, but got: %s", lastOutput)
	}
	return nil
}

func iRunTheNdReceiveCommand() error {
	return runCmd("../bin/nd", "receive")
}

func theOutputShouldExplainTheUsageOfTheGetCommand() error {
	if !strings.Contains(lastOutput, "nd get") {
		return fmt.Errorf("expected output to explain 'nd get' usage, but got: %s", lastOutput)
	}
	return nil
}

func theOutputShouldExplainTheWorkflowOfIntents() error {
	if !strings.Contains(lastOutput, "intent") {
		return fmt.Errorf("expected output to explain intent workflow, but got: %s", lastOutput)
	}
	return nil
}

var discoveryNeedID string

func thereIsANeed() error {
	// Ensure registered
	if err := theClientIsRegistered(); err != nil {
		return err
	}

	// Send need
	_ = runCmd("../bin/nd", "send", "need", "discovery need")
	if lastError != nil {
		return fmt.Errorf("failed to send need: %s", lastOutput)
	}

	// Receive to get ID
	_ = runCmd("../bin/nd", "receive")
	if lastError != nil {
		return fmt.Errorf("failed to receive need: %s", lastOutput)
	}

	// Parse ID from output: [123] NEED ...
	// Matches last output which is from receive
	start := strings.Index(lastOutput, "[")
	end := strings.Index(lastOutput, "]")
	if start != -1 && end != -1 && end > start {
		discoveryNeedID = lastOutput[start+1 : end]
		return nil
	}
	return fmt.Errorf("could not parse need ID from output: %s", lastOutput)
}

func thereIsANeedMessageToReceive() error {
	// Register the discovery agent
	if err := theClientIsRegistered(); err != nil {
		return err
	}

	// Register a helper agent to send the need
	_ = os.Rename(".needy.conf", ".needy.conf.DiscoveryAgent")
	writeTestConfig() // fresh config with port only (no client-id)
	_ = runCmd("../bin/nd", "register", "--name", "HelperAgent")
	if lastError != nil {
		return fmt.Errorf("failed to register helper: %s", lastOutput)
	}

	// Send the need as HelperAgent
	_ = runCmd("../bin/nd", "send", "need", "discovery need")
	if lastError != nil {
		return fmt.Errorf("failed to send need: %s", lastOutput)
	}

	// Switch back to DiscoveryAgent so nd receive will see the unread need
	_ = os.Remove(".needy.conf")
	return os.Rename(".needy.conf.DiscoveryAgent", ".needy.conf")
}

func iHaveIssuedAIntentMessage() error {
	if discoveryNeedID == "" {
		return fmt.Errorf("no need created/captured")
	}
	return runCmd("../bin/nd", "send", "intent", discoveryNeedID)
}

func theOutputShouldExplainHowToOfferASolution() error {
	if !strings.Contains(lastOutput, "solution") {
		return fmt.Errorf("expected output to explain how to offer a solution, but got: %s", lastOutput)
	}
	return nil
}

func iUseAToLongNeedMessageWithAPayload() error {
	longMsg := strings.Repeat("a", 150)
	return runCmd("../bin/nd", "send", "need", longMsg, "--data", "payload")
}

func theOutputShouldExplainThatTheMessageIsTooLong() error {
	if lastError == nil {
		return fmt.Errorf("expected command to fail due to length, but it succeeded")
	}
	if !strings.Contains(lastOutput, "too long") {
		return fmt.Errorf("expected output to explain length limit, but got: %s", lastOutput)
	}
	return nil
}

func iUseAToLongIntentMessage() error {
	longMsg := strings.Repeat("a", 150)
	return runCmd("../bin/nd", "send", "intent", longMsg)
}

func theOutputShouldExplainIntentPolicy() error {
	if lastError == nil {
		return fmt.Errorf("expected intent to fail")
	}
	if !strings.Contains(lastOutput, "short") || !strings.Contains(lastOutput, "intent") {
		return fmt.Errorf("expected output to explain intents should be short, but got: %s", lastOutput)
	}
	return nil
}

func iUseAToLongSolutionMessage() error {
	longMsg := strings.Repeat("a", 150)
	id := discoveryNeedID
	if id == "" {
		id = "1" // fallback if step skipped
	}
	return runCmd("../bin/nd", "send", "solution", id, longMsg)
}

func iRunTheNdReceiveCommandWithNothing() error {
	return runCmd("../bin/nd", "receive")
}

func theOutputShouldExplainTimeoutMechanics() error {
	if !strings.Contains(lastOutput, "timeout") {
		return fmt.Errorf("expected output to explain timeout mechanics, but got: %s", lastOutput)
	}
	return nil
}

// Core scenario implementation

func agentRunsCommand(agentName, command string) error {
	// Swap identity
	targetConfFile := fmt.Sprintf(".needy.conf.%s", agentName)
	if _, err := os.Stat(targetConfFile); err != nil {
		return fmt.Errorf("identity for agent %s not found (did you register them?)", agentName)
	}

	// Backup current config
	if _, err := os.Stat(".needy.conf"); err == nil {
		_ = os.Rename(".needy.conf", ".needy.conf.bak")
		defer func() { _ = os.Rename(".needy.conf.bak", ".needy.conf") }()
	}

	// Copy target config
	input, _ := os.ReadFile(targetConfFile)
	_ = os.WriteFile(".needy.conf", input, 0600)
	defer func() { _ = os.Remove(".needy.conf") }()

	// Execute
	fullCmd := strings.Replace(command, "nd ", "../bin/nd ", 1)

	cmd := exec.Command("bash", "-c", fullCmd)
	out, err := cmd.CombinedOutput()
	lastOutput = string(out)
	lastError = err

	return nil
}

func agentHasSentANeed(agentName, needText string) error {
	return agentRunsCommand(agentName, fmt.Sprintf("nd send need '%s'", needText))
}

func agentShouldReceiveAMessageWithText(agentName, expectedText string) error {
	if err := agentRunsCommand(agentName, "nd receive"); err != nil {
		return err
	}
	if !strings.Contains(lastOutput, expectedText) {
		return fmt.Errorf("expected agent %s to see '%s', output was: %s", agentName, expectedText, lastOutput)
	}
	return nil
}

func theCommandShouldFailWith(expectedError string) error {
	if lastError == nil {
		return fmt.Errorf("expected command to fail, but it succeeded. Output: %s", lastOutput)
	}
	if !strings.Contains(lastOutput, expectedError) {
		return fmt.Errorf("expected error containing '%s', got: %s", expectedError, lastOutput)
	}
	return nil
}

func theCommandShouldSucceed() error {
	if lastError != nil {
		// Output might contain helpful info
		return fmt.Errorf("expected command to succeed, but failed: %v. Output: %s", lastError, lastOutput)
	}
	return nil
}

func iAmRegisteredAs(name string) error {
	return runCmd("../bin/nd", "register", "--name", name)
}

func anotherAgentIsRegistered(name string) error {
	if _, err := os.Stat(".needy.conf"); err == nil {
		_ = os.Rename(".needy.conf", ".needy.conf.primary")
	}
	defer func() {
		if _, err := os.Stat(".needy.conf.primary"); err == nil {
			_ = os.Rename(".needy.conf.primary", ".needy.conf")
		}
	}()

	writeTestConfig() // fresh config with port only (no client-id)

	cmd := exec.Command("../bin/nd", "register", "--name", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to register other agent: %s: %w", output, err)
	}

	if err := os.Rename(".needy.conf", fmt.Sprintf(".needy.conf.%s", name)); err != nil {
		return fmt.Errorf("failed to save other agent config: %w", err)
	}

	return nil
}

func aRegisteredAgent(name string) error {
	// Check if already registered (config file exists for this name)?
	targetConfFile := fmt.Sprintf(".needy.conf.%s", name)
	if _, err := os.Stat(targetConfFile); err == nil {
		// Already registered
		return nil
	}

	// Register new
	return anotherAgentIsRegistered(name)
}
