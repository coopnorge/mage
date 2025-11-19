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
	"strings"
)

func main() {
	baseURL := os.Getenv("POLICY_BOT_BASE_URL")

	if baseURL != "" {
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}

		log.Printf("POLICY_BOT_BASE_URL=%s detected", baseURL)
	} else {
		log.Printf("POLICY_BOT_BASE_URL not configured; starting local server")
		err := startPolicyBotServer()
		if err != nil {
			log.Fatal(err)
		}
		baseURL = "http://127.0.0.1:8080/"
	}

	log.Println("Using policy base URL", baseURL)
	ctx := context.Background()
	err := validate(ctx, baseURL)
	if err != nil {
		log.Fatal(err)
	}
}

// -----------------------------------------------------------------------------
// validation, using the <policy-bot>/api/validate endpoint
// -----------------------------------------------------------------------------

func validate(ctx context.Context, baseURL string) error {
	validateURL := baseURL + "api/validate"
	log.Printf("Validation endpoint: %s\n", validateURL)

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file does not exist: %s, skipping validation", configPath)
		return nil
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
// read config path from location; this is to be mounted to this path when
// running the application via a dockerfile
// -----------------------------------------------------------------------------

func getConfigPath() (string, error) {
	path := "/.policy.yml"
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("validation config not found at %s", path)
}
