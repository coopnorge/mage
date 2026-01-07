package devtool

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
)

// PolicyBot holds the devtool for policy-bot
type PolicyBot struct{}

// PolicyBotConfigCheckDocker the content of policy-bot.Dockerfile
//
//go:embed policy-bot/policy-bot.Dockerfile
var PolicyBotConfigCheckDocker string

// Run runs the policy-bot devtool
func (pb PolicyBot) Run(env map[string]string, args ...string) error {
	// for now only support running in Docker
	return pb.runInDocker(env, args...)
}

func (pb PolicyBot) runInDocker(env map[string]string, args ...string) error {
	image, err := pb.buildImage()
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

func (pb PolicyBot) buildImage() (string, error) {
	devtool, err := getTool(ToolsDockerfile, "policy-bot-version-tracker")
	if err != nil {
		return "", err
	}
	// use upstream if amd64
	if runtime.GOARCH == "amd64" {
		err := sh.Run("docker", "pull", devtool.image)
		if err != nil {
			return "", err
		}
		return devtool.image, nil
	}

	imageName := fmt.Sprintf("%s:%s", devtool.registry, devtool.version)

	// use cached if locally available
	out, err := sh.Output("docker", "inspect", imageName, "--format", `{{.Architecture}}`)
	if out == runtime.GOARCH && err == nil {
		return imageName, nil
	}

	file, cleanup, err := core.WriteTempFile(core.OutputDir, fmt.Sprintf("%s.Dockerfile", "policy-bot"), PolicyBotConfigCheckDocker)
	if err != nil {
		return "", err
	}
	defer cleanup()

	path, cleanup, err := core.MkdirTemp()
	if err != nil {
		return "", nil
	}
	defer cleanup()

	return imageName, sh.RunV(
		"docker", "buildx", "build",
		"--platform", fmt.Sprintf("linux/%s", runtime.GOARCH),
		"-f", file,
		"-t", imageName,
		"--load",
		"--build-arg", fmt.Sprintf("%s=%s", "TARGETARCH", runtime.GOARCH),
		"--build-arg", fmt.Sprintf("%s=%s", "POLICY_BOT_VERSION", devtool.version),
		path,
	)
}
