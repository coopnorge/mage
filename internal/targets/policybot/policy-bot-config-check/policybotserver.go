package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// -----------------------------------------------------------------------------
// Start a local bare-minimum policy-bot server to make api/validate available
// -----------------------------------------------------------------------------

func startPolicyBotServer() error {
	// Start mock-server
	log.Println("Starting embedded mock-server...")
	mockSrv := startMockServer()
	defer func(mockSrv *http.Server, ctx context.Context) {
		err := mockSrv.Shutdown(ctx)
		if err != nil {
			log.Printf("Error shutting down mock server: %v", err)
		}
	}(mockSrv, context.Background())

	// Start local policy-bot
	policyBotPath := filepath.Join("bin", "linux-amd64", "policy-bot")

	log.Println("Starting local policy-bot...")
	bot := exec.Command(policyBotPath, "server", "--config", "/secrets/policy-bot.yml")
	bot.Stdout = os.Stdout
	bot.Stderr = os.Stderr
	if err := bot.Start(); err != nil {
		return fmt.Errorf("failed to start policy-bot: %w", err)
	}
	log.Printf("policy-bot started (pid %d)", bot.Process.Pid)

	// Wait for policy-bot readiness
	localBase := "http://127.0.0.1:8080/"
	healthURL := localBase + "api/health"

	log.Printf("Waiting for local policy-bot to become ready at %s", healthURL)
	if err := waitForReady(healthURL, 60*time.Second); err != nil {
		return err
	}

	return nil
}

// -----------------------------------------------------------------------------
// waiting for startup
// -----------------------------------------------------------------------------

func waitForReady(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for policy-bot readiness at %s", url)
		}

		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			if err := resp.Body.Close(); err != nil {
				log.Printf("warning: failed to close response body: %v", err)
			}
			log.Println("policy-bot is ready")
			return nil
		}

		if resp != nil {
			if err := resp.Body.Close(); err != nil {
				log.Printf("warning: failed to close response body: %v", err)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}
