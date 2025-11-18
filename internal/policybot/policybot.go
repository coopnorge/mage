package policybot

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

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/git"
)

// Validate submits policy file to policy-bot server to validate it
func Validate(ctx context.Context) error {
	log.Println("Starting policy validation")

	baseURL := os.Getenv("POLICY_BOT_BASE_URL")
	if baseURL == "" {
		return fmt.Errorf("POLICY_BOT_BASE_URL not set")
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	validateURL := baseURL + "api/validate"
	log.Printf("Validation endpoint: %s\n", validateURL)

	configPath := getConfigPath()

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

	log.Println("Sending request to policy bot...")
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

	log.Printf("Policy validation successful: %s\n\n", msg)
	return nil
}

// HasChanges checks if the current branch has policy bot config file changes
// from the main branch
func HasChanges() (bool, error) {
	changedFiles, err := git.DiffToMain()
	if err != nil {
		return false, err
	}
	// always trigger on go.mod/sum and workflows because of changes in ci.
	additionalGlobs := []string{"go.mod", "go.sum", ".github/workflows/*"}
	return core.CompareChangesToPaths(changedFiles, []string{getConfigPath()}, additionalGlobs)
}

func getConfigPath() string {
	configPath := os.Getenv("POLICY_CONFIG_FILE_PATH")
	if configPath == "" {
		configPath = ".policy.yml"
	}
	log.Printf("Using config: %s\n", configPath)
	return configPath
}
