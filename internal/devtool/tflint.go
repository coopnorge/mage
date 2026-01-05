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

// TFLint holds the devtool for tflint
type TFLint struct{}

// Run runs the tflint devtool
func (tfl TFLint) Run(env map[string]string, workdir string, args ...string) error {
	if !isCommandAvailable("tflint") {
		fmt.Println("tflint binary not found. Falling back to running the docker version")
		return tfl.runInDocker(env, workdir, args...)
	}

	err := tfl.versionOK()
	if err != nil {
		fmt.Printf("tflint does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return tfl.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native tflint")
	return tfl.runNative(env, workdir, args...)
}

func (tfl TFLint) versionOK() error {
	devtoolData, err := getTool(ToolsDockerfile, "tflint")
	if err != nil {
		return err
	}

	out, err := sh.Output("tflint", "--version")
	if err != nil {
		return err
	}
	current, err := version.NewVersion(strings.Fields(out)[2])
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

func (tfl TFLint) runNative(env map[string]string, workdir string, args ...string) error {
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
	// env["TF_PLUGIN_CACHE_DIR"] = "$HOME/.tflint.d/plugin-cache"

	if core.Verbose() {
		return sh.RunWith(env, "tflint", args...)
	}
	out, err := sh.OutputWith(env, "tflint", args...)
	if err != nil {
		fmt.Println(out)
		return err
	}
	return err
}

func (tfl TFLint) runInDocker(env map[string]string, workdir string, args ...string) error {
	devtool, err := getTool(ToolsDockerfile, "tflint")
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
		"--volume", "$HOME/.tflint.docker.d:/root/.tflint.d", // do not share with the os cache because of binary issues
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
