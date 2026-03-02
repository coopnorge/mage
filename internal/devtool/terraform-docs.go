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

// TerraformDocs holds the devtool for terraform-docs
type TerraformDocs struct{}

// Run runs the terraform-docs devtool
func (tfdocs TerraformDocs) Run(env map[string]string, workdir string, args ...string) error {
	if !isCommandAvailable("terraform-docs") {
		fmt.Println("tfdocs binary not found. Install using 'brew install terraform-docs' Falling back to running the docker version")
		return tfdocs.runInDocker(env, workdir, args...)
	}

	err := tfdocs.versionOK()
	if err != nil {
		fmt.Printf("terraform-docs does not meet version constraints. Falling back to docker version\n error: %s\n", err)
		return tfdocs.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native terraform-docs")
	return tfdocs.runNative(env, workdir, args...)
}

func (tfdocs TerraformDocs) versionOK() error {
	devtoolData, err := getTool(ToolsDockerfile, "terraform-docs")
	if err != nil {
		return err
	}

	out, err := sh.Output("terraform-docs", "--version")
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
		return fmt.Errorf("version found %s does not match constraint %s", current.Original(), constraint.String())
	}
	return nil
}

func (tfdocs TerraformDocs) runNative(env map[string]string, workdir string, args ...string) error {
	if env == nil {
		env = map[string]string{}
	}
	// set cache
	// skip for now
	// env["TF_PLUGIN_CACHE_DIR"] = "$HOME/.tfdocs.d/plugin-cache"

	if core.Verbose() {
		return core.RunAtWith(env, core.GetAbsWorkDir(workdir), "terraform-docs", args...)
	}
	out, err := core.OutputAtWith(env, core.GetAbsWorkDir(workdir), "terraform-docs", args...)
	if err != nil {
		fmt.Println(out)
		return err
	}
	return err
}

func (tfdocs TerraformDocs) runInDocker(env map[string]string, workdir string, args ...string) error {
	devtool, err := getTool(ToolsDockerfile, "terraform-docs")
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
