package features

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

// State to track IDs across steps in the tutorial
var capturedNeedID string
var capturedSolutionID string

func InitializeTutorialSteps(ctx *godog.ScenarioContext) {
	// Reset state before scenario
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		capturedNeedID = ""
		capturedSolutionID = ""
		return ctx, nil
	})

	ctx.Step(`^Agent(Alice|Bob|Charlie) registers with "([^"]*)"$`, agentRegistersWith)
	ctx.Step(`^Agent(Alice|Bob|Charlie) (?:also )?runs "([^"]*)"$`, agentRunsWithIdReplacement)

	ctx.Step(`^all three agents should be registered successfully$`, allThreeAgentsShouldBeRegisteredSuccessfully)
	ctx.Step(`^the message should be broadcast to all agents$`, theMessageShouldBeBroadcastToAllAgents)
	ctx.Step(`^AgentAlice should see "([^"]*)"$`, agentAliceShouldSee)

	ctx.Step(`^AgentBob should see a need from AgentAlice with text "([^"]*)"$`, agentBobShouldSeeANeedFromAgentAliceWithText)
	ctx.Step(`^the need should have an ID$`, theNeedShouldHaveAnID)

	ctx.Step(`^AgentBob should see "([^"]*)"$`, agentBobShouldSee)
	ctx.Step(`^the intent should be linked to the need$`, theIntentShouldBeLinkedToTheNeed)

	ctx.Step(`^the solution should be linked to the need$`, theSolutionShouldBeLinkedToTheNeed)

	ctx.Step(`^AgentAlice should see an intent from AgentBob$`, agentAliceShouldSeeAnIntentFromAgentBob)
	ctx.Step(`^AgentAlice should see a solution from AgentBob$`, agentAliceShouldSeeASolutionFromAgentBob)
	ctx.Step(`^AgentAlice can retrieve the payload with "([^"]*)"$`, agentAliceCanRetrieveThePayloadWith)

	// AgentCharlie steps are now covered by generic agentRunsWithIdReplacement
	ctx.Step(`^AgentCharlie should see the same need, intent, and solution$`, agentCharlieShouldSeeTheSameNeedIntentAndSolution)
	ctx.Step(`^all agents have a consistent view of the conversation$`, allAgentsHaveAConsistentViewOfTheConversation)
}

func agentRegistersWith(agent, command string) error {
	agentName := "Agent" + agent

	// Backup existing config if present
	if _, err := os.Stat(".needy.conf"); err == nil {
		_ = os.Rename(".needy.conf", ".needy.conf.bak")
		defer func() { _ = os.Rename(".needy.conf.bak", ".needy.conf") }()
	}

	// Clean slate for new registration (fresh config with port only)
	writeTestConfig()

	// Run command via runCmd (assuming no tricky quotes in register command)
	fullCmd := strings.Replace(command, "nd ", "../bin/nd ", 1)
	parts := strings.Fields(fullCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	if err := runCmd(parts[0], parts[1:]...); err != nil {
		return err
	}

	// Save config (with client-id) for this agent
	if _, err := os.Stat(".needy.conf"); err == nil {
		_ = os.Rename(".needy.conf", fmt.Sprintf(".needy.conf.%s", agentName))
	} else {
		return fmt.Errorf("registration seemed to fail, no .needy.conf file created")
	}

	return nil
}

func agentRunsWithIdReplacement(agent, command string) error {
	agentName := "Agent" + agent

	// Replace <need-id> with captured ID
	if strings.Contains(command, "<need-id>") {
		if capturedNeedID == "" {
			return fmt.Errorf("no Need ID captured to replace <need-id>")
		}
		command = strings.ReplaceAll(command, "<need-id>", capturedNeedID)
	}

	return agentRunsCommand(agentName, command)
}

func allThreeAgentsShouldBeRegisteredSuccessfully() error {
	for _, agent := range []string{"AgentAlice", "AgentBob", "AgentCharlie"} {
		confFile := fmt.Sprintf(".needy.conf.%s", agent)
		if _, err := os.Stat(confFile); err != nil {
			return fmt.Errorf("expected config file for %s to exist, but it doesn't", agent)
		}
	}
	return nil
}

func theMessageShouldBeBroadcastToAllAgents() error {
	if !strings.Contains(lastOutput, "Sent need successfully") {
		return fmt.Errorf("expected send to succeed, but got: %s", lastOutput)
	}
	return nil
}

func agentAliceShouldSee(expectedText string) error {
	if !strings.Contains(lastOutput, expectedText) {
		return fmt.Errorf("expected Alice to see %q, got: %s", expectedText, lastOutput)
	}
	return nil
}

func agentBobShouldSeeANeedFromAgentAliceWithText(expectedText string) error {
	// Check output for Need logic
	// e.g. [1] NEED from AgentAlice: fix the authentication bug
	if !strings.Contains(lastOutput, expectedText) {
		return fmt.Errorf("expected Bob to see need text %q, got: %s", expectedText, lastOutput)
	}
	if !strings.Contains(lastOutput, "NEED from AgentAlice") {
		return fmt.Errorf("expected Bob to see sender AgentAlice, got: %s", lastOutput)
	}
	return nil
}

func theNeedShouldHaveAnID() error {
	// Parse ID from last output
	// Example: "[1] NEED ..."
	re := regexp.MustCompile(`\[(\d+)\] NEED`)
	matches := re.FindStringSubmatch(lastOutput)
	if len(matches) < 2 {
		return fmt.Errorf("could not find Need ID in output: %s", lastOutput)
	}
	capturedNeedID = matches[1]
	return nil
}

func agentBobShouldSee(expectedText string) error {
	if !strings.Contains(lastOutput, expectedText) {
		return fmt.Errorf("expected Bob to see %q, got: %s", expectedText, lastOutput)
	}
	return nil
}

func theIntentShouldBeLinkedToTheNeed() error {
	if !strings.Contains(lastOutput, "Sent intent successfully") {
		return fmt.Errorf("expected intent to be sent successfully, but got: %s", lastOutput)
	}
	return nil
}

func theSolutionShouldBeLinkedToTheNeed() error {
	if !strings.Contains(lastOutput, "Sent solution successfully") {
		return fmt.Errorf("expected solution to be sent successfully, but got: %s", lastOutput)
	}
	return nil
}

func agentAliceShouldSeeAnIntentFromAgentBob() error {
	// Output should contain INTENT from AgentBob
	if !strings.Contains(lastOutput, "INTENT from AgentBob") {
		return fmt.Errorf("expected INTENT from AgentBob, got: %s", lastOutput)
	}
	return nil
}

func agentAliceShouldSeeASolutionFromAgentBob() error {
	// Output should contain SOLUTION from AgentBob
	if !strings.Contains(lastOutput, "SOLUTION from AgentBob") {
		return fmt.Errorf("expected SOLUTION from AgentBob, got: %s", lastOutput)
	}
	return nil
}

func agentAliceCanRetrieveThePayloadWith(commandTemplate string) error {
	// We need to capture the Solution ID implies we parse it in previous step?
	// Receive output: [3] SOLUTION from AgentBob: ...

	// Let's parse all IDs from output
	// We assume the solution is the last one or we find it by type
	// The step is "AgentAlice can retrieve ... with 'nd get <solution-id>'"

	// Parse Solution ID
	re := regexp.MustCompile(`\[(\d+)\] SOLUTION`)
	matches := re.FindStringSubmatch(lastOutput)
	if len(matches) < 2 {
		return fmt.Errorf("could not find Solution ID in output: %s", lastOutput)
	}
	capturedSolutionID = matches[1]

	cmd := strings.ReplaceAll(commandTemplate, "<solution-id>", capturedSolutionID)

	// Run it
	err := agentRunsCommand("AgentAlice", cmd)
	if err != nil {
		return err
	}

	// Verify payload
	if !strings.Contains(lastOutput, "Updated auth.go with fix") {
		return fmt.Errorf("expected payload, got: %s", lastOutput)
	}

	return nil
}

func agentCharlieShouldSeeTheSameNeedIntentAndSolution() error {
	if !strings.Contains(lastOutput, "NEED from AgentAlice") {
		return fmt.Errorf("Charlie missing need")
	}
	if !strings.Contains(lastOutput, "INTENT from AgentBob") {
		return fmt.Errorf("Charlie missing intent")
	}
	if !strings.Contains(lastOutput, "SOLUTION from AgentBob") {
		return fmt.Errorf("Charlie missing solution")
	}
	return nil
}

func allAgentsHaveAConsistentViewOfTheConversation() error {
	if !strings.Contains(lastOutput, "NEED from AgentAlice") {
		return fmt.Errorf("expected consistent view with need, but got: %s", lastOutput)
	}
	if !strings.Contains(lastOutput, "INTENT from AgentBob") {
		return fmt.Errorf("expected consistent view with intent, but got: %s", lastOutput)
	}
	if !strings.Contains(lastOutput, "SOLUTION from AgentBob") {
		return fmt.Errorf("expected consistent view with solution, but got: %s", lastOutput)
	}
	return nil
}
