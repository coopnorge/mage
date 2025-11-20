package policybot

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
)

// Validate submits policy file to policy-bot docker app to validate it
func Validate(args ...string) error {
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
		"--volume", fmt.Sprintf("%s:/.policy.yml", absConfigPath),
	}

	policyBotBaseURL := os.Getenv("POLICY_BOT_BASE_URL")

	if policyBotBaseURL != "" {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", "POLICY_BOT_BASE_URL", policyBotBaseURL))
	}

	// policy-bot needs an RSA key to start
	rsaPEMKey := generateRSAPrivateKeyPEM(4096)
	dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", "GITHUB_APP_PRIVATE_KEY", rsaPEMKey))

	return devtool.Run("policy-bot-config-check", dockerArgs, "", args...)
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

func generateRSAPrivateKeyPEM(bits int) string {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Fatalf("failed to generate RSA key: %v", err)
	}

	der := x509.MarshalPKCS1PrivateKey(key)

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: der,
	}

	return string(pem.EncodeToMemory(block))
}
