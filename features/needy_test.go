package features

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

var lastOutput string
var lastError error

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"."},
			TestingT: t,
			Strict:   true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	// Learning about registration
	sc.Step(`^the `+"`"+`nd`+"`"+` CLI is available$`, theNdCLIIsAvailable)
	sc.Step(`^the client in not registered$`, theClientInNotRegistered)
	sc.Step(`^I run the `+"`"+`nd`+"`"+` command$`, iRunTheNdCommand)
	sc.Step(`^the output should explain that registration is required$`, theOutputShouldExplainThatRegistrationIsRequired)
	sc.Step(`^show the usage for the "([^"]*)" command$`, showTheUsageForTheCommand)

	// Learning about sending
	sc.Step(`^the client is registered$`, theClientIsRegistered)
	sc.Step(`^I run the `+"`"+`nd register`+"`"+` command$`, iRunTheNdRegisterCommand)
	sc.Step(`^registration is successful$`, registrationIsSuccessful)
	sc.Step(`^the output should explain the mechanics of sending and receiving messages$`, theOutputShouldExplainTheMechanicsOfSendingAndReceiving)

	// Learning about getting payloads
	sc.Step(`^there is a need message to receive$`, thereIsANeedMessageToReceive)
	sc.Step(`^I run the `+"`"+`nd receive`+"`"+` command$`, iRunTheNdReceiveCommand)
	sc.Step(`^the output should explain the usage of the "get" command and include the id of the message$`, theOutputShouldExplainTheUsageOfTheGetCommand)

	// Learning about the workflow of intents
	sc.Step(`^the output should explain that the agent must issue a "intent" message referring to the id of the need if it wants to offer a solution$`, theOutputShouldExplainTheWorkflowOfIntents)

	// Learning about the workflow of solutions
	sc.Step(`^there is a need$`, thereIsANeed)
	sc.Step(`^I have issued a "intent" message$`, iHaveIssuedAIntentMessage)
	sc.Step(`^the output should explain how to offer a solution by issuing a "solution" message referring to the id of the need$`, theOutputShouldExplainHowToOfferASolution)

	// Learning about the limited space for messages (Need)
	sc.Step(`^I use a to long need message with a payload$`, iUseAToLongNeedMessageWithAPayload)
	sc.Step(`^the output should explain that the message is too long and that it must contain a message and a payload$`, theOutputShouldExplainThatTheMessageIsTooLong)

	// Learning about the length of intent messages
	sc.Step(`^I use a to intent message that is too long$`, iUseAToLongIntentMessage)
	sc.Step(`^the output should explain that intents should be short and not have payloads$`, theOutputShouldExplainIntentPolicy)

	// Learning about the limited space for solution messages
	sc.Step(`^I use a to long solution message$`, iUseAToLongSolutionMessage)
	sc.Step(`^the output should explain that the solution message is too long and that it must contain a message and a payload$`, theOutputShouldExplainThatTheMessageIsTooLong)

	// Learning about the listening with timeout
	sc.Step(`^I run the `+"`"+`nd receive`+"`"+` command and there is nothing to receive$`, iRunTheNdReceiveCommandWithNothing)
	sc.Step(`^the output should explain the mechanics of listening with a timeout$`, theOutputShouldExplainTimeoutMechanics)
}

func theNdCLIIsAvailable() error {
	_, err := os.Stat("../bin/nd")
	if os.IsNotExist(err) {
		return fmt.Errorf("nd binary not found in bin/")
	}
	return nil
}

func theClientInNotRegistered() error { return nil }

func iRunTheNdCommand() error {
	return runCmd("../bin/nd")
}

func theOutputShouldExplainThatRegistrationIsRequired() error {
	if !strings.Contains(lastOutput, "Registration is required") {
		return fmt.Errorf("expected output to contain 'Registration is required', but got: %s", lastOutput)
	}
	return nil
}

func showTheUsageForTheCommand(cmdName string) error {
	if !strings.Contains(lastOutput, "Usage: nd "+cmdName) {
		return fmt.Errorf("expected output to show usage for %s, but got: %s", cmdName, lastOutput)
	}
	return nil
}

func theClientIsRegistered() error { return nil }

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

func thereIsANeedMessageToReceive() error { return nil }

func iRunTheNdReceiveCommand() error {
	return runCmd("../bin/nd", "receive")
}

func theOutputShouldExplainTheUsageOfTheGetCommand() error {
	if !strings.Contains(lastOutput, "get") || !strings.Contains(lastOutput, "123") {
		return fmt.Errorf("expected output to explain 'get' command and include an ID, but got: %s", lastOutput)
	}
	return nil
}

func theOutputShouldExplainTheWorkflowOfIntents() error {
	if !strings.Contains(lastOutput, "intent") || !strings.Contains(lastOutput, "id") {
		return fmt.Errorf("expected output to explain 'intent' workflow with ID reference, but got: %s", lastOutput)
	}
	return nil
}

func thereIsANeed() error { return nil }

func iHaveIssuedAIntentMessage() error {
	return runCmd("../bin/nd", "send", "intent", "fix the bug")
}

func theOutputShouldExplainHowToOfferASolution() error {
	if !strings.Contains(lastOutput, "solution") || !strings.Contains(lastOutput, "id") {
		return fmt.Errorf("expected output to explain 'solution' workflow with ID reference, but got: %s", lastOutput)
	}
	return nil
}

func iUseAToLongNeedMessageWithAPayload() error {
	longMsg := strings.Repeat("a", 150)
	return runCmd("../bin/nd", "send", "need", longMsg, "--data", "payload")
}

func theOutputShouldExplainThatTheMessageIsTooLong() error {
	if lastError == nil {
		return fmt.Errorf("expected command to fail due to length, but it succeeded. Output: %s", lastOutput)
	}
	if !strings.Contains(lastOutput, "too long") || !strings.Contains(lastOutput, "payload") {
		return fmt.Errorf("expected output to explain length limit and payload usage, but got: %s", lastOutput)
	}
	return nil
}

func iUseAToLongIntentMessage() error {
	longMsg := strings.Repeat("a", 60)
	return runCmd("../bin/nd", "send", "intent", longMsg)
}

func theOutputShouldExplainIntentPolicy() error {
	if lastError == nil {
		return fmt.Errorf("expected intent to fail due to length, but it succeeded. Output: %s", lastOutput)
	}
	if !strings.Contains(lastOutput, "short") || !strings.Contains(lastOutput, "payloads") {
		return fmt.Errorf("expected output to explain intent policy, but got: %s", lastOutput)
	}
	return nil
}

func iUseAToLongSolutionMessage() error {
	longMsg := strings.Repeat("a", 150)
	return runCmd("../bin/nd", "send", "solution", longMsg)
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

func runCmd(path string, args ...string) error {
	absPath, _ := filepath.Abs(path)
	cmd := exec.Command(absPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	lastError = err
	lastOutput = out.String()
	return nil
}
