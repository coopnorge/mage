package policybot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
)

// Validate submits policy file to policy-bot docker app to validate it
func Validate() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Check for config file
	configPath := getConfigPath()
	if _, err := os.Stat(filepath.Join(cwd, configPath)); err != nil {
		if os.IsNotExist(err) {
			// No config file → do nothing
			return nil
		}
		return err
	}

	absConfigPath, err := filepath.Abs(filepath.Join(cwd, configPath))
	if err != nil {
		return err
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/policy-bot-1.39.3/.policy.yml", absConfigPath),
	}

	policyBotBaseURL := os.Getenv("POLICY_BOT_BASE_URL")

	if policyBotBaseURL != "" {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", "POLICY_BOT_BASE_URL", policyBotBaseURL))
	}

	return devtool.Run("policy-bot", dockerArgs, "validate", "/.policy.yml")
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
