package features

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

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
	// Initialize grouped steps
	InitializeCommonSteps(sc)
	InitializeRegistrationSteps(sc)
	InitializeCommunicationSteps(sc)
	InitializeTutorialSteps(sc)

	// Cleanup before each scenario
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Clean up config files (and legacy client ID files)
		for _, pattern := range []string{".needy.conf*", ".needy-client-id*"} {
			files, _ := filepath.Glob(pattern)
			for _, f := range files {
				_ = os.Remove(f)
			}
		}
		// Clean up .nats-data with retries
		for i := 0; i < 10; i++ {
			if err := os.RemoveAll(".nats-data"); err == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		// Reset test state
		networkDown = false
		lastOutput = ""
		lastError = nil
		registrationResults = nil

		// Start ndadm server if network should be up
		if !networkDown {
			startNdadmServer()
		}
		return ctx, nil
	})

	// Cleanup after each scenario
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		stopNdadmServer()
		return ctx, nil
	})
}
