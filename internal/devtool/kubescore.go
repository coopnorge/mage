package devtool

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/hashicorp/go-version"
	"github.com/magefile/mage/sh"
)

// KubeScore holds the devtool for kubescore
type KubeScore struct{}

// Run runs the kubescore devtool
func (kubescore KubeScore) Run(env map[string]string, workdir string, args ...string) (string, string, error) {
	if val, found := os.LookupEnv("KUBESCORE_IN_DOCKER"); found && val == "1" {
		return kubescore.runInDocker(env, workdir, args...)
	}
	if !isCommandAvailable("kube-score") {
		fmt.Println("kube-score binary not found. Use 'brew install kube-score' to install. Falling back to running the docker version")
		return kubescore.runInDocker(env, workdir, args...)
	}

	err := kubescore.versionOK()
	if err != nil {
		fmt.Printf("kube-score does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return kubescore.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native kube-score")
	return kubescore.runNative(env, workdir, args...)
}

func (kubescore KubeScore) versionOK() error {
	devtoolData, err := getTool(ToolsDockerfile, "kube-score")
	if err != nil {
		return err
	}
	out, err := sh.Output("kube-score", "version")
	if err != nil {
		return err
	}
	// kube-score version: 1.18.0, commit: 0fb5f668e153c22696aa75ec769b080c41b5dd3d, built: 2024-02-05T14:08:35Z

	versionString := strings.Split(strings.Split(out, ",")[0], ":")[1]
	current, err := version.NewVersion(strings.TrimSpace(versionString))
	if err != nil {
		return err
	}
	devtool, err := version.NewVersion(devtoolData.version)
	if err != nil {
		return err
	}
	// set constraint that minor minus 5 version should be minimum
	constraintString := fmt.Sprintf(">= %s.%s", strconv.Itoa(devtool.Segments()[0]), strconv.Itoa(devtool.Segments()[1]-2))
	constraint, err := version.NewConstraint(constraintString)
	if err != nil {
		return err
	}
	if !constraint.Check(current) {
		return fmt.Errorf("version found %s does not match constraint %s", current.Original(), constraint.String())
	}
	return nil
}

func (kubescore KubeScore) runNative(env map[string]string, workdir string, args ...string) (string, string, error) {
	outs := setupStdOutErr(true)
	_, err := core.ExecAt(env, outs.StdOut, outs.StdErr, workdir, "kube-score", kubescore.addDefautsArgs(args...)...)

	return outs.printOut(), outs.printErr(), err
}

func (kubescore KubeScore) runInDocker(env map[string]string, workdir string, args ...string) (string, string, error) {
	devtool, err := getTool(ToolsDockerfile, "kube-score")
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
	runArgs = append(runArgs, kubescore.addDefautsArgs(args...)...)

	outs := setupStdOutErr(true)
	_, err = core.Exec(env, outs.StdOut, outs.StdErr, "docker", kubescore.addDefautsArgs(runArgs...)...)

	return outs.printOut(), outs.printErr(), err
}

func (kubescore KubeScore) addDefautsArgs(args ...string) []string {
	return args
}
