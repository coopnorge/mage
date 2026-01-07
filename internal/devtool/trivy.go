package devtool

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/hashicorp/go-version"
	"github.com/magefile/mage/sh"
)

// Trivy holds the devtool for trivy
type Trivy struct{}

// Run runs the trivy devtool
func (trivy Trivy) Run(env map[string]string, workdir string, args ...string) error {
	if !isCommandAvailable("trivy") {
		fmt.Println("trivy binary not found. Install using 'brew install trivy' Falling back to running the docker version")
		return trivy.runInDocker(env, workdir, args...)
	}

	err := trivy.versionOK()
	if err != nil {
		fmt.Printf("trivy does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return trivy.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native trivy")
	return trivy.runNative(env, workdir, args...)
}

func (trivy Trivy) versionOK() error {
	devtoolData, err := getTool(ToolsDockerfile, "trivy")
	if err != nil {
		return err
	}

	out, err := sh.Output("trivy", "--version")
	if err != nil {
		return err
	}
	current, err := version.NewVersion(strings.Fields(out)[1])
	if err != nil {
		return err
	}
	devtool, err := version.NewVersion(devtoolData.version)
	if err != nil {
		return err
	}
	// set constraint that minor minus 2 version should be minimum
	constraintString := fmt.Sprintf(">= %s.%s", strconv.Itoa(devtool.Segments()[0]), strconv.Itoa(devtool.Segments()[1]-0))
	constraint, err := version.NewConstraint(constraintString)
	if err != nil {
		return err
	}
	if !constraint.Check(current) {
		return fmt.Errorf("version found %s does not match constrant %s", current.Original(), constraint.String())
	}
	return nil
}

func (trivy Trivy) runNative(env map[string]string, workdir string, args ...string) error {
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

	if env == nil {
		env = map[string]string{}
	}
	// set cache
	// skip for now
	// env["TF_PLUGIN_CACHE_DIR"] = "$HOME/.trivy.d/plugin-cache"

	if core.Verbose() {
		return sh.RunWith(env, "trivy", args...)
	}
	out, err := sh.OutputWith(env, "trivy", args...)
	if err != nil {
		fmt.Println(out)
		return err
	}
	return err
}

func (trivy Trivy) runInDocker(env map[string]string, workdir string, args ...string) error {
	devtool, err := getTool(ToolsDockerfile, "trivy")
	if err != nil {
		return err
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/app", path), // Mount the source code
		"--workdir", filepath.Join("/app", workdir), // set workdir to where we want to run
		"--volume", "$HOME/.cache/trivy:/root/.cache/trivy",
	}

	if env == nil {
		env = map[string]string{}
	}
	if _, exists := env["GITHUB_TOKEN"]; !exists {
		if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
			env["GITHUB_TOKEN"] = token
		}
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
