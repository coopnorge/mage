package devtool

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/hashicorp/go-version"
	"github.com/magefile/mage/sh"
)

// Dyff holds the devtool for policy-bot
type Dyff struct{}

// DyffDocker the content of dyff.Dockerfile
//
//go:embed policy-bot/policy-bot.Dockerfile
var DyffDocker string

const dyffVersion = "v1.10.5"

// Run runs the policy-bot devtool
func (dyff Dyff) Run(env map[string]string, args ...string) (string, string, error) {
	if !isCommandAvailable("dyff") {
		fmt.Println("Dyff binary not found. Use 'brew install dyff' to install. Falling back to running the docker version")
		return "", "", dyff.runInDocker(env, args...)
	}

	err := dyff.versionOK()
	if err != nil {
		fmt.Printf("Dyff does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return "", "", dyff.runInDocker(env, args...)
	}

	fmt.Println("Using native dyff")
	return dyff.runNative(env, args...)
	// for now only support running in Docker
}

func (dyff Dyff) versionOK() error {
	// example v3.17.1+g980d8ac
	out, err := sh.Output("dyff", "version")
	if err != nil {
		return err
	}
	current, err := version.NewVersion(strings.Split(out, " ")[2])
	if err != nil {
		return err
	}
	devtool, err := version.NewVersion(dyffVersion)
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

func (dyff Dyff) runNative(env map[string]string, args ...string) (string, string, error) {
	outs := setupStdOutErr(true)
	_, err := sh.Exec(env, outs.StdOut, outs.StdErr, "dyff", args...)

	return strings.TrimSuffix((outs.BufOut).String(), "\n"), strings.TrimSuffix((outs.BufErr).String(), "\n"), err
}

func (dyff Dyff) runInDocker(env map[string]string, args ...string) error {
	image, err := dyff.buildImage()
	if err != nil {
		return err
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	// workdir is dependant on the version of dependabot
	origWorkDir, err := sh.Output(
		"docker", "inspect",
		"--format={{.Config.WorkingDir}}",
		image,
	)
	if err != nil {
		return err
	}

	// the binary is in the original working directory
	entryPoint := filepath.Join(origWorkDir, fmt.Sprintf("bin/linux-%s/policy-bot", runtime.GOARCH))

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/app", path), // Mount the source code
		"--workdir", "/app", // set workdir to where we want to run
		"--entrypoint", entryPoint,
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
	runArgs = append(runArgs, image)
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

func (dyff Dyff) buildImage() (string, error) {
	imagename := fmt.Sprintf("%s:%s", "dyff", dyffVersion)

	_, cleanup, err := core.WriteTempFile(core.OutputDir, fmt.Sprintf("%s.dockerfile", "dyff"), DyffDocker)
	if err != nil {
		return "", err
	}
	defer cleanup()

	_, cleanup, err = core.MkdirTemp()
	if err != nil {
		return "", nil
	}
	defer cleanup()

	return imagename, nil
}
