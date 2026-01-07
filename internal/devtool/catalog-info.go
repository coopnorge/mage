package devtool

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
)

// CatalogInfo holds the devtool for policy-bot
type CatalogInfo struct{}

// CatalogInfoDockerfile the content of policy-bot.Dockerfile
//
//go:embed catalog-info/validate-entity.Dockerfile
var CatalogInfoDockerfile string

const validateEntityVersion = "0.5.1"

// Run runs the policy-bot devtool
func (cataloginfo CatalogInfo) Run(env map[string]string, args ...string) error {
	// for now only support running in Docker
	return cataloginfo.runInDocker(env, args...)
}

func (cataloginfo CatalogInfo) runInDocker(env map[string]string, args ...string) error {
	image, err := cataloginfo.buildImage()
	if err != nil {
		return err
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

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

func (cataloginfo CatalogInfo) buildImage() (string, error) {
	// Entity valiator does not really seem to be maintained. We should look
	// into alternatives in the future.
	//
	imageName := fmt.Sprintf("%s:%s", "back-stage-entity-validator", validateEntityVersion)

	file, cleanup, err := core.WriteTempFile(core.OutputDir, fmt.Sprintf("%s.Dockerfile", "validate-entity"), CatalogInfoDockerfile)
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
		"--build-arg", fmt.Sprintf("%s=%s", "BACKSTAGE_ENTITY_VALIDATOR_VERSION", validateEntityVersion),
		path,
	)
}
