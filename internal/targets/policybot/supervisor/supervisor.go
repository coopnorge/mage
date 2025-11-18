package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Supervisor failed: %v", err)
		os.Exit(1)
	}
	log.Println("Supervisor finished successfully")
}

func run() error {
	baseURL := os.Getenv("POLICY_BOT_BASE_URL")

	// --------------------------------------------------------------------------
	// MODE 1: External policy-bot (CI mode)
	// No mock-server. No local policy-bot.
	// --------------------------------------------------------------------------
	if baseURL != "" {
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}

		log.Printf("POLICY_BOT_BASE_URL=%s detected", baseURL)
		log.Println("Skipping mock-server and local policy-bot")

		ctx := context.Background()
		return validate(ctx, baseURL)
	}

	// --------------------------------------------------------------------------
	// MODE 2: Local policy-bot (developer / container mode)
	// Start mock-server, then local policy-bot.
	// --------------------------------------------------------------------------

	// Start mock-server
	log.Println("Starting mock-server...")
	mock := exec.Command("/usr/local/bin/mock-server")
	mock.Stdout = os.Stdout
	mock.Stderr = os.Stderr
	if err := mock.Start(); err != nil {
		return fmt.Errorf("failed to start mock-server: %w", err)
	}
	log.Printf("mock-server started (pid %d)", mock.Process.Pid)

	// Discover local policy-bot path based on WORKDIR
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	policyBotPath := filepath.Join(wd, "bin", "linux-amd64", "policy-bot")

	// Start local policy-bot
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

	// Validate using local base URL
	log.Println("Starting policy validation (local mode)")
	ctx := context.Background()
	return validate(ctx, localBase)
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

// -----------------------------------------------------------------------------
// validation logic (your code, adapted)
// -----------------------------------------------------------------------------

func validate(ctx context.Context, baseURL string) error {
	validateURL := baseURL + "api/validate"
	log.Printf("Validation endpoint: %s\n", validateURL)

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Config file not found, skipping validation")
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, validateURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	log.Println("Sending request to policy-bot...")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("warning: failed to close response body: %v", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	msg := result.Message
	if msg == "" {
		msg = string(body)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Policy bot returned error: %s\n", msg)
		return fmt.Errorf("%s", msg)
	}

	log.Printf("Policy validation successful: %s\n", msg)
	return nil
}

// -----------------------------------------------------------------------------
// reads config path from env or default locations
// -----------------------------------------------------------------------------

func getConfigPath() (string, error) {
	path := "/data/.policy.yml"
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("validation config not found at %s", path)
}
