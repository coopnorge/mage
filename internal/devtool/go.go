package devtool

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/hashicorp/go-version"
	"github.com/magefile/mage/sh"
)

type Go struct{}

func (g Go) Run(env map[string]string, args ...string) error {
	if !isCommandAvailable("go") {
		fmt.Println("Go binary not found. Use 'brew install go' to install. Falling back to running the docker version")
		return g.runInDocker(env, args...)
	}

	err := g.versionOK()
	if err != nil {
		fmt.Printf("Go does not meet version constraints. Falling back to docker verion\n error: %s\n", err)
		return g.runInDocker(env, args...)
	}

	fmt.Println("Using native go")
	return g.runNative(env, args...)
}

func (g Go) versionOK() error {
	devtoolData, err := getTool(ToolsDockerfile, "golang")
	if err != nil {
		return err
	}
	current, err := version.NewVersion(strings.TrimPrefix(runtime.Version(), "go"))
	if err != nil {
		return err
	}
	devtool, err := version.NewVersion(devtoolData.version)
	if err != nil {
		return err
	}
	// set constraint that minor version should be minimum
	constraintString := fmt.Sprintf(">= %s.%s", strconv.Itoa(devtool.Segments()[0]), strconv.Itoa(devtool.Segments()[1]+1))
	constraint, err := version.NewConstraint(constraintString)
	if err != nil {
		return err
	}
	if !constraint.Check(current) {
		return fmt.Errorf("version does not match constrant %s", constraintString)
	}
	return nil
}

func (g Go) runNative(env map[string]string, args ...string) error {
	if core.Verbose() {
		return sh.RunWith(env, "go", args...)
	}
	out, err := sh.OutputWith(env, "go", args...)
	if err != nil {
		fmt.Println(out)
		return err
	}
	return err
}

// DevtoolGo runs the devtool for Go
func (g Go) runInDocker(env map[string]string, args ...string) error {
	devtool, err := getTool(ToolsDockerfile, "golang")
	if err != nil {
		return err
	}

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

	runArgs := []string{
		"run",
		"--rm",
	}
	runArgs = append(runArgs, dockerArgs...)
	runArgs = append(runArgs, devtool.image)
	runArgs = append(runArgs, "go")
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
