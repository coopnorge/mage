package goapp

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/golang"
	devtoolTarget "github.com/coopnorge/mage/internal/targets/devtool"
	golangTargets "github.com/coopnorge/mage/internal/targets/golang"

	"github.com/magefile/mage/mg"
)

// Go is the magefile namespace to group Go commands
type Go mg.Namespace

const (
	cmdDir = "cmd"
	binDir = "bin"
)

var (
	// OsArchMatrix defines what CPU architectures to build binaries for
	OsArchMatrix = []map[string]string{
		{
			"GOOS": "darwin", "GOARCH": "arm64",
		},
		{
			"GOOS": "linux", "GOARCH": "amd64",
		},
		{
			"GOOS": "linux", "GOARCH": "arm64",
		},
	}
)

// Generate files
func (Go) Generate(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.Generate)
	return nil
}

// Build binaries from main package when found in cmd directories
func (Go) Build(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Validate)

	rootPath, err := os.Getwd()
	if err != nil {
		return err
	}

	goModules, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}

	cmds, err := findCommands(goModules)
	if err != nil {
		return err
	}

	bins := []any{}
	for _, command := range cmds {
		relativeRootPath, err := core.GetRelativeRootPath(rootPath, command.goModule)
		if err != nil {
			return err
		}
		for _, osArch := range OsArchMatrix {
			output := path.Join(relativeRootPath, binaryOutputPath(command.goModule, osArch["GOOS"], osArch["GOARCH"], command.binary))
			bins = append(bins, mg.F(Go.build, command.goModule, command.pkg, output, osArch["GOOS"], osArch["GOARCH"]))
		}
	}

	mg.CtxDeps(ctx, bins...)

	return nil
}

type cmd struct {
	goModule string
	pkg      string
	binary   string
}

func findCommands(goModules []string) ([]cmd, error) {
	result := []cmd{}
	for _, goModule := range goModules {
		if _, err := os.Stat(path.Join(goModule, cmdDir)); os.IsNotExist(err) {
			continue
		}
		entries, err := os.ReadDir(path.Join(goModule, cmdDir))
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			cmd := cmd{
				goModule: goModule,
				pkg:      fmt.Sprintf("./%s", path.Join(cmdDir, entry.Name())),
				binary:   entry.Name(),
			}
			result = append(result, cmd)
		}
	}
	return result, nil
}

func (Go) build(ctx context.Context, workingDirectory, input, output, goos, goarch string) error {
	mg.CtxDeps(ctx, Go.Validate, mg.F(devtoolTarget.Build, "golang", golangTargets.GolangToolsDockerfile))

	environmentalVariables := map[string]string{"GOOS": goos, "GOARCH": goarch}

	return devtool.RunWith(
		environmentalVariables,
		"golang",
		"go",
		"-C", workingDirectory,
		"build",
		"-v",
		"-tags='datadog.no_waf'",
		"-o", output,
		input)
}

// Run runs a cmd: mage run <cmd>
func (Go) Run(ctx context.Context, workingDirectory, bin string) error {
	mg.CtxDeps(ctx, Go.Validate, mg.F(devtoolTarget.Build, "golang", golangTargets.GolangToolsDockerfile))
	return devtool.Run("golang", "go", "-C", workingDirectory, "run", fmt.Sprintf("%s/%s/main.go", cmdDir, bin))
}

// Validate files
func (Go) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Test, Go.Lint)
	return nil
}

// Fix files
func (Go) Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.LintFix)
	return nil
}

// Test runs the test suite
func (Go) Test(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(golangTargets.Test))
	return nil
}

// Lint all source code
func (Go) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, golangTargets.Lint)
	return nil
}

// LintFix linting issues
func (Go) LintFix(ctx context.Context) error {
	mg.CtxDeps(ctx, golangTargets.LintFix)
	return nil
}

func binaryOutputPath(app, os, arch, binary string) string {
	return path.Join(binaryOutputBasePath(app), os, arch, binary)
}

func binaryOutputBasePath(app string) string {
	return path.Join(core.OutputDir, app, binDir)
}
