package devtool

import (
	"fmt"
	"os"
	"strconv"

	"github.com/coopnorge/mage/internal/core"
	"github.com/hashicorp/go-version"
	"github.com/magefile/mage/sh"
)

// KubeConform holds the devtool for kubeconform
type KubeConform struct{}

// Run runs the kubeconform devtool
func (kf KubeConform) Run(env map[string]string, args ...string) error {
	if !isCommandAvailable("kubeconform") {
		fmt.Println("kubeconform binary not found. Use 'brew install kubeconform' to install. Falling back to running the docker version")
		return kf.runInDocker(env, args...)
	}

	err := kf.versionOK()
	if err != nil {
		fmt.Printf("kubeconform does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return kf.runInDocker(env, args...)
	}

	fmt.Println("Using native kubeconform")
	return kf.runNative(env, args...)
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
		return fmt.Errorf("version found %s does not match constrant %s", current.Original(), constraint.String())
	}
	return nil
}

func (kf KubeConform) runNative(env map[string]string, args ...string) error {
	if core.Verbose() {
		return sh.RunWith(env, "kubeconform", kf.addDefautsArgs(args...)...)
	}
	out, err := sh.OutputWith(env, "kubeconform", kf.addDefautsArgs(args...)...)
	if err != nil {
		fmt.Println(out)
		return err
	}
	return err
}

// DevtoolGo runs the devtool for Go
func (kf KubeConform) runInDocker(env map[string]string, args ...string) error {
	devtool, err := getTool(ToolsDockerfile, "kubeconform")
	if err != nil {
		return err
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	//  kubeconform --strict -verbose  -schema-location "https://raw.githubusercontent.com/coopnorge/kubernetes-schemas/main/pallets/{{ .ResourceKind }}{{ .KindSuffix }}.json" .pallet/gitconfig.yaml

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/app", path), // Mount the source code
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

func (kf KubeConform) addDefautsArgs(args ...string) []string {
	return append([]string{"--output", "pretty", "--strict", "--verbose"}, args...)
}
