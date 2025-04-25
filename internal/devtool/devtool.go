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
func Run(tool string, dockerRunArgs []string, cmd string, args ...string) error {
	return RunWith(nil, tool, dockerRunArgs, cmd, args...)
}

// RunWith will run the specified command with arguments in the
// specified Docker image with environment variables defined.
func RunWith(env map[string]string, tool string, dockerRunArgs []string, cmd string, args ...string) error {
	image, err := GetImageName(tool)
	if err != nil {
		return err
	}

	call := []string{
		"run",
		"--rm",
	}

	call = append(call, dockerRunArgs...)
	call = append(call, image, cmd)
	call = append(call, args...)
	return sh.RunWithV(env, "docker", call...)
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
