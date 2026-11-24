package devtool

import (
	"fmt"
	"os"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/hashicorp/go-version"
	"github.com/magefile/mage/sh"
)

// Helm holds the devtool for helm
type Helm struct{}

// Run runs the helm devtool. It returns stdout, stderr and error. If verbose
// is enable on mage it will also stream stdout to the console
func (helm Helm) Run(env map[string]string, workdir string, args ...string) (string, string, error) {
	if val, found := os.LookupEnv("HELM_IN_DOCKER"); found && val == "1" {
		return helm.runInDocker(env, workdir, args...)
	}
	if !isCommandAvailable("helm") {
		fmt.Println("helm binary not found. Use 'brew install helm' to install. Falling back to running the docker version")
		return helm.runInDocker(env, workdir, args...)
	}

	err := helm.versionOK()
	if err != nil {
		fmt.Printf("helm does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return helm.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native helm")
	return helm.runNative(env, workdir, args...)
}

func (helm Helm) versionOK() error {
	devtoolData, err := getTool(ToolsDockerfile, "helm")
	if err != nil {
		return err
	}
	// example v3.17.1+g980d8ac
	out, err := sh.Output("helm", "version", "--short")
	if err != nil {
		return err
	}
	current, err := version.NewVersion(strings.Split(out, "+")[0])
	if err != nil {
		return err
	}
	devtool, err := version.NewVersion(devtoolData.version)
	if err != nil {
		return err
	}
	return versionWithinMajor(devtool, current, 5)
}

func (helm Helm) runNative(env map[string]string, workdir string, args ...string) (string, string, error) {
	outs := setupStdOutErr(false)
	_, err := core.ExecAt(env, outs.StdOut, outs.StdErr, workdir, "helm", helm.addDefautsArgs(args...)...)

	return outs.printOut(), outs.printErr(), err
}

func (helm Helm) runInDocker(env map[string]string, workdir string, args ...string) (string, string, error) {
	devtool, err := getTool(ToolsDockerfile, "helm")
	if err != nil {
		return "", "", err
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/app", workdir), // Mount the source code
		"--workdir", "/app", // set workdir to where we want to run
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
	runArgs = append(runArgs, helm.addDefautsArgs(args...)...)

	outs := setupStdOutErr(false)
	_, err = sh.Exec(env, outs.StdOut, outs.StdErr, "docker", runArgs...)

	return outs.printOut(), outs.printErr(), err
}

func (helm Helm) addDefautsArgs(args ...string) []string {
	return args
}
