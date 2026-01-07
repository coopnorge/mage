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

// Terraform holds the devtool for terraform
type Terraform struct{}

// Run runs the terraform devtool
func (tf Terraform) Run(env map[string]string, workdir string, args ...string) error {
	if !isCommandAvailable("terraform") {
		fmt.Println("terraform binary not found. Falling back to running the docker version")
		return tf.runInDocker(env, workdir, args...)
	}

	err := tf.versionOK()
	if err != nil {
		fmt.Printf("terraform does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return tf.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native terraform")
	return tf.runNative(env, workdir, args...)
}

func (tf Terraform) versionOK() error {
	out, err := sh.Output("terraform", "version")
	if err != nil {
		return err
	}
	current, err := version.NewVersion(strings.Fields(out)[1])
	if err != nil {
		return err
	}

	// set constraints in the current supported terraform versions in coop
	constraintString := ">= 1.3.6, < 1.6.0"
	constraint, err := version.NewConstraint(constraintString)
	if err != nil {
		return err
	}
	if !constraint.Check(current) {
		return fmt.Errorf("version found %s does not match constrant %s", current.Original(), constraint.String())
	}
	return nil
}

func (tf Terraform) runNative(env map[string]string, workdir string, args ...string) error {
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
	// env["TF_PLUGIN_CACHE_DIR"] = "$HOME/.terraform.d/plugin-cache"

	if core.Verbose() {
		return sh.RunWith(env, "terraform", args...)
	}
	out, err := sh.OutputWith(env, "terraform", args...)
	if err != nil {
		fmt.Println(out)
		return err
	}
	return err
}

func (tf Terraform) runInDocker(env map[string]string, workdir string, args ...string) error {
	devtool, err := getTool(ToolsDockerfile, "terraform")
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
		"--volume", "$HOME/.terraform.d:/root/.terraform.d", // mount credentials and cache
	}

	if env == nil {
		env = map[string]string{}
	}
	// set cache
	// skip for now
	// env["TF_PLUGIN_CACHE_DIR"] = "$HOME/.terraform.d/plugin-cache"
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
