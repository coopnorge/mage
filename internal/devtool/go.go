package devtool

import (
	"fmt"
	"go/version"
	"os"
	"runtime"

	"github.com/magefile/mage/sh"
)

type Go struct{}

const (
	minGoVersion = "go1.25"
)

func (g Go) Run(env map[string]string, args ...string) error {
	if isCommandAvailable("go") {
		if g.versionOK() {
			fmt.Println("Using native go")
			return g.runNative(env, args...)
		}
		fmt.Printf("Go found but is older than %s. Falliong back to running the docker version\n", minGoVersion)
		return g.runInDocker(env, "go", args...)
	}
	fmt.Println("Go binary not found. Use 'brew install go' to install. Falling back to running the docker version")
	return g.runInDocker(env, "go", args...)
}

func (g Go) versionOK() bool {
	return version.Compare(runtime.Version(), minGoVersion) != -1
}

func (g Go) runNative(env map[string]string, args ...string) error {
	return sh.RunWith(env, "go", args...)
}

// DevtoolGo runs the devtool for Go
func (g Go) runInDocker(env map[string]string, cmd string, args ...string) error {
	// This is a bit hacky to use the local go binary instead of the container
	// this is used for running the integration tests on targets.

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	goModCache, err := sh.Output("go", "env", "GOMODCACHE")
	if err != nil {
		goModCache = "$HOME/go/pkg/mod"
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/go/pkg/mod", goModCache), // Mount downloaded go modules
		"--volume", "/var/run/docker.sock:/var/run/docker.sock", // Mount Docker socket for docker-in-docker
		"--volume", "$HOME/.cache:/root/.cache", // Mount caches, such as linter cache, Go build cache, etc.
		"--volume", "$HOME/.gitconfig:/root/.gitconfig", // Mount Git config, for access to private repos
		"--volume", "$HOME/.ssh:/root/.ssh", // Mount SSH config, for access to private repos
		"--volume", fmt.Sprintf("%s:/app", path), // Mount the source code
		"--env", "TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal", // For testcontainers to work when running with docker-in-docker
		"--env", "GOMODCACHE=/go/pkg/mod", // Ensure that the GOMODCACHE env is set correctly
		"--add-host", "host.docker.internal:host-gateway", // For testcontainers to work when running with docker-in-docker
		"--workdir", "/app",
	}

	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	return Run("golang", dockerArgs, cmd, args...)
}
