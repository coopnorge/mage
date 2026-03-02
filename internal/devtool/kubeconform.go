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

// KubeConform holds the devtool for kubeconform
type KubeConform struct{}

// Run runs the kubeconform devtool
func (kf KubeConform) Run(env map[string]string, workdir string, args ...string) (string, string, error) {
	if val, found := os.LookupEnv("KUBECONFORM_IN_DOCKER"); found && val == "1" {
		return kf.runInDocker(env, workdir, args...)
	}
	if !isCommandAvailable("kubeconform") {
		fmt.Println("kubeconform binary not found. Use 'brew install kubeconform' to install. Falling back to running the docker version")
		return kf.runInDocker(env, workdir, args...)
	}

	err := kf.versionOK()
	if err != nil {
		fmt.Printf("kubeconform does not meet version constraints. Falling back to docker version\n error: %s\n", err)
		return kf.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native kubeconform")
	return kf.runNative(env, workdir, args...)
}

func (kf KubeConform) versionOK() error {
	devtoolData, err := getTool(ToolsDockerfile, "kubeconform")
	if err != nil {
		return err
	}
	out, err := sh.Output("kubeconform", "-v")
	if err != nil {
		return err
	}
	current, err := version.NewVersion(out)
	if err != nil {
		return err
	}
	devtool, err := version.NewVersion(devtoolData.version)
	if err != nil {
		return err
	}
	// set constraint that minor minus 1 version should be minimum
	constraintString := fmt.Sprintf(">= %s.%s", strconv.Itoa(devtool.Segments()[0]), strconv.Itoa(devtool.Segments()[1]-1))
	constraint, err := version.NewConstraint(constraintString)
	if err != nil {
		return err
	}
	if !constraint.Check(current) {
		return fmt.Errorf("version found %s does not match constraint %s", current.Original(), constraint.String())
	}
	return nil
}

func (kf KubeConform) runNative(env map[string]string, workdir string, args ...string) (string, string, error) {
	outs := setupStdOutErr(true)
	_, err := core.ExecAt(env, outs.StdOut, outs.StdErr, workdir, "kubeconform", kf.addDefautsArgs(args...)...)

	return strings.TrimSuffix((outs.BufOut).String(), "\n"), strings.TrimSuffix((outs.BufErr).String(), "\n"), err
}

// DevtoolGo runs the devtool for Go
func (kf KubeConform) runInDocker(env map[string]string, workdir string, args ...string) (string, string, error) {
	devtool, err := getTool(ToolsDockerfile, "kubeconform")
	if err != nil {
		return "", "", err
	}

	//  kubeconform --strict -verbose  -schema-location "https://raw.githubusercontent.com/coopnorge/kubernetes-schemas/main/pallets/{{ .ResourceKind }}{{ .KindSuffix }}.json" .pallet/gitconfig.yaml

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
	runArgs = append(runArgs, kf.addDefautsArgs(args...)...)

	outs := setupStdOutErr(true)
	_, err = core.Exec(env, outs.StdOut, outs.StdErr, "docker", runArgs...)

	return strings.TrimSuffix((outs.BufOut).String(), "\n"), strings.TrimSuffix((outs.BufErr).String(), "\n"), err
}

func (kf KubeConform) addDefautsArgs(args ...string) []string {
	return append([]string{"--output", "pretty", "--strict", "--verbose"}, args...)
}
