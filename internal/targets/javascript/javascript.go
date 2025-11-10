package javascript

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/coopnorge/mage/internal/javascript"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Install fetches all Node.js dependencies.
func Install() error {
	if javascript.HasPackageConfig() {
		tokenName := "GITHUB_TOKEN"
		tokenValue := os.Getenv(tokenName)


		env := map[string]string{}

		if githubToken != "" && IsNpmrcConfiguredForPrivateRepo() {
			env[tokenName] = tokenValue
		}

		if err := sh.RunWithV(env, "npm", "install"); err != nil {
			return fmt.Errorf("dependency installation failed: %w", err)
		}
	}

	return nil
}

// Lint runs the standard linting script defined in package.json.
func Lint() error {
	envData := os.Getenv("SKIP_LINT")
	skip := strings.ToLower(envData) == "true" || envData == "1"

	if (skip) {
		return nil
	}

	if javascript.HasBiomeConfig() {
		return errors.New("biome not setup in your project. Install @coopnorge/web-devtools")
	}

	if err := sh.RunV("npm", "run", "lint"); err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	return nil
}

// Format runs the standard formatting check script defined in package.json.
func Format() error {
	envData := os.Getenv("SKIP_FORMAT")
	skip := strings.ToLower(envData) == "true" || envData == "1"

	if (skip) {
		return nil
	}

	if javascript.HasBiomeConfig() {
		return errors.New("biome not setup in your project. Install @coopnorge/web-devtools")
	}

	if err := sh.RunV("npm", "run", "format:code"); err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	return nil
}

// UnitTest runs unit tests using the package.json script.
func UnitTest() error {
	envData := os.Getenv("SKIP_UNIT_TEST")
	skip := strings.ToLower(envData) == "true" || envData == "1"

	if (skip) {
		return nil
	}

	if err := sh.RunV("npm", "run", "test:unit"); err != nil {
		return fmt.Errorf("unit tests failed: %w", err)
	}

	return nil
}

// E2ETest runs End-to-End tests (often separate and slower).
func E2ETest() error {
	envData := os.Getenv("SKIP_E2E_TEST")
	skip := strings.ToLower(envData) == "true" || envData == "1"

	if (skip) {
		return nil
	}

	if err := sh.RunV("npm", "run", "test:e2e"); err != nil {
		return fmt.Errorf("E2E tests failed: %w", err)
	}

	return nil
}

// Build compiles the JavaScript/TypeScript into distribution files.
func Build(buildCommand string) error {
	envData := os.Getenv("SKIP_BUILD")
	skip := strings.ToLower(envData) == "true" || envData == "1"

	if (skip) {
		return nil
	}

	if buildCommand == "" {
		buildCommand = "build"
	}

	if err := sh.RunV("npm", "run", buildCommand); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	return nil
}
