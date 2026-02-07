package features

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/cucumber/godog"
	"github.com/nats-io/nats.go"
)

var lastOutput string
var lastError error
var ndadmCmd *exec.Cmd
var networkDown bool

// Shared helper functions

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

func startNdadmServer() {
	// Aggressively kill any existing ndadm or process on 4222
	exec.Command("pkill", "ndadm").Run()
	exec.Command("sh", "-c", "lsof -ti :4222 | xargs kill -9").Run()

	// Wait for port to be free
	waitForPortFree(4222)

	cmd := exec.Command("../bin/ndadm")

	outfile, err := os.Create("ndadm.log")
	if err != nil {
		log.Printf("Failed to create ndadm.log: %v", err)
		// Fallback to stdout/stderr
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = outfile
		cmd.Stderr = outfile
	}

	ndadmCmd = cmd
	err = ndadmCmd.Start()
	if err != nil {
		log.Printf("Failed to start ndadm: %v", err)
		return
	}

	// Wait for server to be ready
	waitForNATS()
}

func waitForNATS() {
	url := "nats://127.0.0.1:4222"
	timeout := 5 * time.Second
	start := time.Now()

	for time.Since(start) < timeout {
		nc, err := nats.Connect(url)
		if err == nil {
			nc.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Printf("Warning: NATS server did not become ready within 5s")
}

func waitForPortFree(port int) {
	timeout := 5 * time.Second
	start := time.Now()
	for time.Since(start) < timeout {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err != nil {
			// Connection refused means port is free
			return
		}
		conn.Close()
		time.Sleep(100 * time.Millisecond)
	}
	log.Printf("Warning: Port %d did not become free within 5s", port)
}

func stopNdadmServer() {
	if ndadmCmd != nil && ndadmCmd.Process != nil {
		ndadmCmd.Process.Signal(syscall.SIGTERM)
		ndadmCmd.Wait()
		ndadmCmd = nil
	}
}

// Common steps

func InitializeCommonSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I run "([^"]*)"$`, iRun)
	ctx.Step(`^the output should contain "([^"]*)"$`, theOutputShouldContain)
	ctx.Step(`^the command should fail$`, theCommandShouldFail)
}

func iRun(cmdLine string) error {
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Replace "nd" with the actual binary path
	if parts[0] == "nd" {
		parts[0] = "../bin/nd"
	}

	return runCmd(parts[0], parts[1:]...)
}

func theOutputShouldContain(expected string) error {
	if !strings.Contains(lastOutput, expected) {
		return fmt.Errorf("expected output to contain %q, but got: %s", expected, lastOutput)
	}
	return nil
}

func theCommandShouldFail() error {
	if lastError == nil {
		return fmt.Errorf("expected command to fail, but it succeeded")
	}
	return nil
}
