package devtool

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	golangcilint "github.com/coopnorge/mage/internal/devtool/golangci-lint"
	"github.com/hashicorp/go-version"
	"github.com/magefile/mage/sh"
)

// golangciLintFile is the name of the configuration
const golangciLintFile = ".golangci-lint.yaml"

// GoLangCILint holds the devtool for golnagci lint
type GoLangCILint struct{}

// Run runs the Go devtool
func (gl GoLangCILint) Run(env map[string]string, workdir string, args ...string) error {
	if !isCommandAvailable("golangci-lint") {
		fmt.Println("Golangci-lint binary not found. Use 'brew install golangci-lint' to install. Falling back to running the docker version")
		return gl.runInDocker(env, workdir, args...)
	}

	err := gl.versionOK()
	if err != nil {
		fmt.Printf("Golangci-lint does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return gl.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native golangci-lint")
	return gl.runNative(env, workdir, args...)
}

func (gl GoLangCILint) versionOK() error {
	devtoolData, err := getTool(ToolsDockerfile, "golangci-lint")
	if err != nil {
		return err
	}
	out, err := sh.Output("golangci-lint", "--version")
	if err != nil {
		return err
	}
	current, err := version.NewVersion(strings.Split(out, " ")[3])
	if err != nil {
		return err
	}
	devtool, err := version.NewVersion(devtoolData.version)
	if err != nil {
		return err
	}
	// set constraint that minor minus 2 version should be minimum
	constraintString := fmt.Sprintf(">= %s.%s", strconv.Itoa(devtool.Segments()[0]), strconv.Itoa(devtool.Segments()[1]-2))
	constraint, err := version.NewConstraint(constraintString)
	if err != nil {
		return err
	}
	if !constraint.Check(current) {
		return fmt.Errorf("version found %s does not match constrant %s", current.Original(), constraint.String())
	}
	return nil
}

func (gl GoLangCILint) runNative(env map[string]string, workdir string, args ...string) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if os.Chdir(workdir) != nil {
		return err
	}
	defer func() {
		err = os.Chdir(originalDir)
	}()
	if err != nil {
		return fmt.Errorf("failed to return to original dir: %s, error: %s", originalDir, err)
	}

	if core.Verbose() {
		return sh.RunWith(env, "golangci-lint", args...)
	}
	out, err := sh.OutputWith(env, "golangci-lint", args...)
	if err != nil {
		fmt.Println(out)
		return err
	}
	return err
}

// DevtoolGo runs the devtool for Go
func (gl GoLangCILint) runInDocker(env map[string]string, workdir string, args ...string) error {
	devtool, err := getTool(ToolsDockerfile, "golangci-lint")
	if err != nil {
		return err
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	goModCache, err := sh.Output("go", "env", "GOMODCACHE")
	if err != nil {
		goModCache = "$HOME/go/pkg/mod"
	}
	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/app", path), // Mount the source code
		"--volume", fmt.Sprintf("%s:/go/pkg/mod", goModCache), // Mount downloaded go modules
		"--env", "GOMODCACHE=/go/pkg/mod", // Ensure that the GOMODCACHE env is set correctly
		"--volume", "$HOME/.cache:/root/.cache", // Mount caches, such as linter cache, Go build cache, etc.
		"--workdir", filepath.Join("/app", workdir), // set workdir to where we want to run
	}

	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	runArgs := []string{
		"run",
		"--rm",
	}
	runArgs = append(runArgs, dockerArgs...)
	runArgs = append(runArgs, devtool.image)
	runArgs = append(runArgs, "golangci-lint")
	runArgs = append(runArgs, args...)

	if core.Verbose() {
		return sh.RunWith(env, "docker", runArgs...)
	}
	out, err := sh.OutputWith(env, "docker", runArgs...)
	if err != nil {
		fmt.Println(out)
		return err
	}
	return err
}

// FetchGolangCILintConfig fetches and writes the golangci-lint configuration file
// to the specified directory relative to the repository root.
// The config file will be named .golangci-lint.yaml.
//
// The where parameter specifies the directory path relative to the repository root.
// Use "." or "" to write to the repository root directory.
func FetchGolangCILintConfig(where string) error {
	golangCILintCfg := golangcilint.Cfg()
	// Get the repository root directory
	repoRoot, err := core.GetRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	dirs := path.Join(repoRoot, where)
	filePath := path.Join(dirs, golangciLintFile)
	if core.FileExists(filePath) {
		log.Printf("Config file already exists at %s", filePath)
		b, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		if bytes.Equal([]byte(golangCILintCfg), b) {
			log.Println("golangci-lint config exists and it's the latest")
			return nil
		}

		// file exists but it's different. Refresh.
	}

	log.Printf("Writing golangci-lint config to %s", filePath)
	err = os.MkdirAll(dirs, 0755)
	if err != nil {
		return fmt.Errorf("unable to create directory %s: %w", dirs, err)
	}
	return os.WriteFile(filePath, []byte(golangCILintCfg), 0644)
}
