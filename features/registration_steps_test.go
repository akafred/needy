package features

import (
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"
)

var registrationResults []error

func InitializeRegistrationSteps(ctx *godog.ScenarioContext) {
	// Learning about registration
	ctx.Step(`^the `+"`"+`nd`+"`"+` CLI is available$`, theNdCLIIsAvailable)
	ctx.Step(`^the client in not registered$`, theClientInNotRegistered)
	ctx.Step(`^I run the `+"`"+`nd`+"`"+` command$`, iRunTheNdCommand)
	ctx.Step(`^the output should explain that registration is required$`, theOutputShouldExplainThatRegistrationIsRequired)
	ctx.Step(`^show the usage for the "([^"]*)" command$`, showTheUsageForTheCommand)

	// Additional registration scenarios
	ctx.Step(`^the network is up$`, theNetworkIsUp)
	ctx.Step(`^the network is not running$`, theNetworkIsNotRunning)
	ctx.Step(`^a mailbox for "([^"]*)" should be created$`, aMailboxForShouldBeCreated)
	ctx.Step(`^"([^"]*)" is already registered from a different client$`, isAlreadyRegisteredFromDifferentClient)
	ctx.Step(`^I previously registered as "([^"]*)"$`, iPreviouslyRegisteredAs)
	ctx.Step(`^the mailbox for "([^"]*)" should be reconnected$`, theMailboxForShouldBeReconnected)
	ctx.Step(`^all registrations should succeed$`, allRegistrationsShouldSucceed)
	ctx.Step(`^mailboxes for "([^"]*)", "([^"]*)", and "([^"]*)" should exist$`, mailboxesForShouldExist)
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

func theNetworkIsUp() error {
	// Start the server if it's not already running
	if ndadmCmd == nil || ndadmCmd.Process == nil {
		startNdadmServer()
	}
	return nil
}

func theNetworkIsNotRunning() error {
	networkDown = true
	stopNdadmServer()
	return nil
}

func aMailboxForShouldBeCreated(agentName string) error {
	if lastError != nil {
		return fmt.Errorf("registration failed: %v", lastError)
	}
	return nil
}

func isAlreadyRegisteredFromDifferentClient(agentName string) error {
	// First, register the agent with the current client ID
	err := iRun(fmt.Sprintf("nd register --name %s", agentName))
	if err != nil {
		return fmt.Errorf("failed to register agent initially: %v", err)
	}

	differentClientID := "00000000-0000-0000-0000-000000000000"
	_ = os.WriteFile(".needy-client-id", []byte(differentClientID), 0600)

	return nil
}

func iPreviouslyRegisteredAs(agentName string) error {
	// Re-run registration to create .needy-client-id
	return iRun(fmt.Sprintf("nd register --name %s", agentName))
}

func theMailboxForShouldBeReconnected(agentName string) error {
	// Verify re-registration message
	if !strings.Contains(lastOutput, "Re-registered") {
		return fmt.Errorf("expected re-registration message, got: %s", lastOutput)
	}
	return nil
}

func allRegistrationsShouldSucceed() error {
	for i, err := range registrationResults {
		if err != nil {
			return fmt.Errorf("registration %d failed: %v", i+1, err)
		}
	}
	return nil
}

func mailboxesForShouldExist(agent1, agent2, agent3 string) error {
	return allRegistrationsShouldSucceed()
}
