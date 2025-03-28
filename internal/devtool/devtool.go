package devtool

import (
	"fmt"
	"os"
	"path"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
)

// GetImageName returns the name of a devtools OCI image
func GetImageName(target string) (string, error) {
	rootPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ocreg.invalid/coopnorge/%s/%s-devtool:latest", path.Base(rootPath), target), nil
}

// Run will run the specified command with arguments in the
// specified Docker image
func Run(tool, cmd string, args ...string) error {
	return RunWith(nil, tool, cmd, args...)
}

// RunWith will run the specified command with arguments in the
// specified Docker image with environment variables defined.
func RunWith(env map[string]string, tool, cmd string, args ...string) error {
	image, err := GetImageName(tool)
	if err != nil {
		return err
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	goModCache, err := sh.Output("go", "env", "GOMODCACHE")
	if err != nil {
		return err
	}

	call := []string{
		"run",
		"--rm",
		"-v", fmt.Sprintf("%s:/go/pkg/mod", goModCache),
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", "$HOME/.cache:/root/.cache",
		"-v", "$HOME/.gitconfig:/root/.gitconfig",
		"-v", "$HOME/.ssh:/root/.ssh",
		"-e", "TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal",
		"--add-host", "host.docker.internal:host-gateway",
		"-v", fmt.Sprintf("%s:/app", path),
		"-w", "/app",
	}

	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		call = append(call, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	call = append(call, image, cmd)
	call = append(call, args...)
	return sh.RunV("docker", call...)
}

// Build allow a mage target to depend on a Docker image. This will
// pull the image from a Docker registry.
func Build(tool, dockerfile string) error {
	file, cleanup, err := core.WriteTempFile(core.OutputDir, fmt.Sprintf("%s.Dockerfile", tool), dockerfile)
	if err != nil {
		return err
	}
	defer cleanup()

	imageName, err := GetImageName(tool)
	if err != nil {
		return err
	}

	path, cleanup, err := core.MkdirTemp()
	if err != nil {
		return nil
	}
	defer cleanup()

	return sh.RunV(
		"docker", "buildx", "build",
		"-f", file,
		"--target", tool,
		"-t", imageName,
		"--load",
		path,
	)
}
