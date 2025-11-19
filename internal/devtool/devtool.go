package devtool

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
)

// GetImageName returns the name of a devtools OCI image
func GetImageName(target string) (string, error) {
	repository := "unknown-coopnorge"
	// Try to fetch repo-info from debug-info
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Path != "" {
			repository = info.Main.Path
		}
	}
	return fmt.Sprintf("ocreg.invalid/%s/%s-devtool:latest", repository, target), nil
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
	if cmd == "" {
		call = append(call, image)
	} else {
		call = append(call, image, cmd)
	}
	call = append(call, args...)
	return sh.RunWithV(env, "docker", call...)
}

// Build allow a mage target to depend on a Docker image. This will
// pull the image from a Docker registry.
func Build(tool string, dockerfile any) error {
	// This is a bit hacky to use the local go binary instead of the container.
	// We don't need to build a dependency here
	// this is used for running the integration tests on targets.
	if os.Getenv("GO_RUNTIME") == "local" && tool == "golang" {
		return nil
	}

	var (
		cleanup            func()
		err                error
		dockerfileContents string
		dockerFilePath     string
		dockerFileDir      string
	)

	switch v := dockerfile.(type) {
	case string:
		dockerFilePath, cleanup, err = core.WriteTempFile(core.OutputDir, fmt.Sprintf("%s.Dockerfile", tool), v)
		if err != nil {
			return err
		}
		defer cleanup()

		path, cleanup, err := core.MkdirTemp()
		if err != nil {
			return nil
		}
		defer cleanup()
		dockerFileDir = path

		dockerfileContents = v
	case embed.FS:
		dockerFileDir, cleanup, err = core.WriteTempFiles(core.OutputDir, fmt.Sprintf("%s-docker", tool), v)
		if err != nil {
			return err
		}
		defer cleanup()

		dockerFilePath = filepath.Join(dockerFileDir, "tools.Dockerfile")
		// Read Dockerfile contents from embedded FS
		data, err := v.ReadFile("tools.Dockerfile")
		if err != nil {
			return fmt.Errorf("embedded FS: could not read tools.Dockerfile: %w", err)
		}
		dockerfileContents = string(data)
	default:
		return fmt.Errorf("unsupported dockerfile type: %T", v)
	}

	imageName, err := GetImageName(tool)
	if err != nil {
		return err
	}

	selectedTool, err := archSelector(tool, dockerfileContents)
	if err != nil {
		return err
	}

	return sh.RunV(
		"docker", "buildx", "build",
		"-f", dockerFilePath,
		"--target", selectedTool,
		"-t", imageName,
		"--load",
		dockerFileDir,
	)
}

// archSelector tries to select a devtool for a certain architecture.
// It will select the architecture based first, then fallbackback on universal
// and errors if not fitting tool is found.
func archSelector(tool, dockerfile string) (string, error) {
	targets := []string{}
	for _, line := range strings.Split(dockerfile, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 3 {
			targets = append(targets, fields[3])
		}
	}

	archTool := fmt.Sprintf("%s-%s", tool, runtime.GOARCH)
	switch {
	case slices.Contains(targets, archTool):
		return archTool, nil
	case slices.Contains(targets, tool):
		return tool, nil
	default:
		return "", fmt.Errorf("unable to find devtool for tool \"%s\" for the host architecture %s or universal", tool, runtime.GOARCH)
	}
}
