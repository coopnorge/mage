package devtool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/hashicorp/go-version"
	"github.com/magefile/mage/sh"
)

// Tofu holds the devtool for OpenTofu
type Tofu struct{}

// Run runs the OpenTofu devtool
func (t Tofu) Run(env map[string]string, workdir string, args ...string) error {
	if !isCommandAvailable("tofu") {
		fmt.Println("tofu binary not found. Falling back to running the docker version")
		return t.runInDocker(env, workdir, args...)
	}

	err := t.versionOK()
	if err != nil {
		fmt.Printf("tofu does not meet version constraints. Falling back to docker version\n error: %s\n", err)
		return t.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native tofu")
	return t.runNative(env, workdir, args...)
}

func (t Tofu) versionOK() error {
	out, err := sh.Output("tofu", "version")
	if err != nil {
		return err
	}
	current, err := version.NewVersion(strings.Fields(out)[1])
	if err != nil {
		return err
	}

	constraintString := ">= 1.6.3"
	constraint, err := version.NewConstraint(constraintString)
	if err != nil {
		return err
	}
	if !constraint.Check(current) {
		return fmt.Errorf("version found %s does not match constraint %s", current.Original(), constraint.String())
	}
	return nil
}

func (t Tofu) runNative(env map[string]string, workdir string, args ...string) error {
	if env == nil {
		env = map[string]string{}
	}

	if core.Verbose() {
		return core.RunAtWith(env, core.GetAbsWorkDir(workdir), "tofu", args...)
	}
	out, err := core.OutputAtWith(env, core.GetAbsWorkDir(workdir), "tofu", args...)
	if err != nil {
		fmt.Println(out)
		return err
	}
	return err
}

func (t Tofu) runInDocker(env map[string]string, workdir string, args ...string) error {
	devtool, err := getTool(ToolsDockerfile, "tofu")
	if err != nil {
		return err
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/app", path),
		"--workdir", filepath.Join("/app", workdir),
		"--volume", "$HOME/.terraform.d:/root/.terraform.d",
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
