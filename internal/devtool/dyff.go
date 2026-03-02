package devtool

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/hashicorp/go-version"
	"github.com/magefile/mage/sh"
)

// Dyff holds the devtool for policy-bot
type Dyff struct{}

// DyffDockerfile the content of dyff.Dockerfile
//
//go:embed dyff/dyff.Dockerfile
var DyffDockerfile string

const dyffVersion = "1.10.5"

// Run runs the policy-bot devtool
func (dyff Dyff) Run(env map[string]string, workdir string, args ...string) (string, string, error) {
	if val, found := os.LookupEnv("DYFF_IN_DOCKER"); found && val == "1" {
		return dyff.runInDocker(env, workdir, args...)
	}

	if !isCommandAvailable("dyff") {
		fmt.Println("Dyff binary not found. Use 'brew install dyff' to install. Falling back to running the docker version")
		return dyff.runInDocker(env, workdir, args...)
	}

	err := dyff.versionOK()
	if err != nil {
		fmt.Printf("Dyff does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return dyff.runInDocker(env, workdir, args...)
	}

	fmt.Println("Using native dyff")
	return dyff.runNative(env, workdir, args...)
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

func (dyff Dyff) runNative(env map[string]string, workdir string, args ...string) (string, string, error) {
	outs := setupStdOutErr(true)
	_, err := core.ExecAt(env, outs.StdOut, outs.StdErr, workdir, "dyff", args...)

	return strings.TrimSuffix((outs.BufOut).String(), "\n"), strings.TrimSuffix((outs.BufErr).String(), "\n"), err
}

func (dyff Dyff) runInDocker(env map[string]string, workdir string, args ...string) (string, string, error) {
	image, err := dyff.buildImage()
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
	runArgs = append(runArgs, image)
	runArgs = append(runArgs, args...)

	outs := setupStdOutErr(true)
	_, err = core.Exec(env, outs.StdOut, outs.StdErr, "docker", runArgs...)

	return strings.TrimSuffix((outs.BufOut).String(), "\n"), strings.TrimSuffix((outs.BufErr).String(), "\n"), err
}

func (dyff Dyff) buildImage() (string, error) {
	// Entity valiator does not really seem to be maintained. We should look
	// into alternatives in the future.
	//
	imageName := fmt.Sprintf("%s:%s", "dyff", dyffVersion)

	file, cleanup, err := core.WriteTempFile(core.OutputDir, fmt.Sprintf("%s.Dockerfile", "dyff"), DyffDockerfile)
	if err != nil {
		return "", err
	}
	defer cleanup()

	path, cleanup, err := core.MkdirTemp()
	if err != nil {
		return "", nil
	}
	defer cleanup()

	return imageName, sh.Run(
		"docker", "buildx", "build",
		"--platform", fmt.Sprintf("linux/%s", runtime.GOARCH),
		"-f", file,
		"-t", imageName,
		"--load",
		"--build-arg", fmt.Sprintf("%s=%s", "DYFF_VERSION", dyffVersion),
		"--build-arg", fmt.Sprintf("%s=%s", "TARGETARG", runtime.GOARCH),
		path,
	)
}
