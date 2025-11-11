package javascript

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
)

// Install fetches all Node.js dependencies.
func Install() error {
	if HasPackageConfig() {
		tokenName := "GITHUB_TOKEN"
		tokenValue := os.Getenv(tokenName)

		env := map[string]string{}

		if tokenValue != "" && IsNpmrcConfiguredForPrivateRepo() {
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

	if skip {
		return nil
	}

	if HasBiomeConfig() {
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

	if skip {
		return nil
	}

	if HasBiomeConfig() {
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

	if skip {
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

	if skip {
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

	if skip {
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

// NpmPublish publishes npm repository into a github packages
func NpmPublish() error {
	tokenName := "GITHUB_TOKEN"
	tokenValue := os.Getenv(tokenName)
	privateEnv := os.Getenv("PRIVATE")
	isPrivate := strings.ToLower(privateEnv) == "true" || privateEnv == "1"
	githubTagname := os.Getenv("GITHUB_TAGNAME")
	newVersion := strings.TrimPrefix(githubTagname, "v")

	env := map[string]string{}

	if tokenValue != "" && IsNpmrcConfiguredForPrivateRepo() {
		env[tokenName] = tokenValue
	}

	access := "public"

	if isPrivate {
		access = "restricted"
	}

	if newVersion == "" {
		return errors.New("no new package version set. Set GITHUB_TAGNAME env variable")
	}

	if !core.FileExists(".npmrc") {
		return errors.New(".npmrc file missing")
	}

	if !IsNpmrcConfiguredForPrivateRepo() {
		return errors.New(".npmrc has no auth configuration")
	}

	if err := sh.RunV("npm", "run", "version", newVersion); err != nil {
		return fmt.Errorf("bumping version failed: %w", err)
	}

	if err := sh.RunWithV(env, "npm", "publish", "--access", access); err != nil {
		return fmt.Errorf("bumping version failed: %w", err)
	}

	return nil
}

// IsNpmrcConfiguredForPrivateRepo checks if the project is setup for private
// repositories
func IsNpmrcConfiguredForPrivateRepo() bool {
	directory := "."

	registryURL := "npm.pkg.github.com"
	scope := "@coopnorge"
	tokenIndicator := "_authToken="

	npmrcContent, err := os.ReadFile(fmt.Sprintf("%s/.npmrc", directory))

	if err != nil {
		return false
	}

	contentStr := string(npmrcContent)

	if !strings.Contains(contentStr, registryURL) && !strings.Contains(contentStr, scope) && !strings.Contains(contentStr, tokenIndicator) {
		return false
	}

	return true
}

// HasBiomeConfig checks if project has biome setup
func HasBiomeConfig() bool {
	return core.FileExists("biome.json")
}

// HasPackageConfig checks if project has package.json file
func HasPackageConfig() bool {
	return core.FileExists("package.json")
}
