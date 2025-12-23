package goapp

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/golang"
	golangTargets "github.com/coopnorge/mage/internal/targets/golang"

	"github.com/magefile/mage/mg"
)

// Go is the magefile namespace to group Go commands
type Go mg.Namespace

const (
	cmdDir = "cmd"
	binDir = "bin"
)

// OsArchMatrix defines the CPU architectures to build binaries for
var OsArchMatrix = []map[string]string{
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

var toolGo devtool.Go

// Generate runs commands described by directives within existing files with
// the intent to generate Go code. Those commands can run any process but the
// intent is to create or update Go source files
//
// For details see [golang.Generate].
func (Go) Generate(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.Generate)
	return nil
}

// Build compiles all commands
//
// Build will recursively search for all Go modules in the repository
// containing a cmd package. The cmd package is expected to contain main
// packages. The binaries are written to [core.OutputDir].
//
// Given the input:
//
//	.
//	├── app1
//	│   ├── cmd
//	│   │   ├── dataloader
//	│   │   │   └── main.go
//	│   │   └── server
//	│   │       └── main.go
//	│   ├── go.mod
//	│   └── go.sum
//	└── app2
//	    ├── cmd
//	    │   ├── dataloader
//	    │   │   └── main.go
//	    │   └── server
//	    │       └── main.go
//	    ├── go.mod
//	    └── go.sum
//
// [Go.Build] will create:
//
//	./var
//	├── app1
//	│   └── bin
//	│       ├── darwin
//	│       │   └── arm64
//	│       │       ├── dataloader
//	│       │       └── server
//	│       └── linux
//	│           ├── amd64
//	│           │   ├── dataloader
//	│           │   └── server
//	│           └── arm64
//	│               ├── dataloader
//	│               └── server
//	└── app2
//	    └── bin
//	        ├── darwin
//	        │   └── arm64
//	        │       ├── dataloader
//	        │       └── server
//	        └── linux
//	            ├── amd64
//	            │   ├── dataloader
//	            │   └── server
//	            └── arm64
//	                ├── dataloader
//	                └── server
func (Go) Build(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.DownloadDevTools)
	mg.CtxDeps(ctx, Go.DownloadModules)
	mg.SerialCtxDeps(ctx, Go.Validate, Go.BuildBinaries)

	return nil
}

// BuildBinaries just finds and builds the binaries
// just like `go build`.
func (Go) BuildBinaries(ctx context.Context) error {
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
	for _, cmd := range cmds {
		relativeRootPath, err := core.GetRelativeRootPath(rootPath, cmd.goModule)
		if err != nil {
			return err
		}
		input := strings.Join(cmd.pkgs, " ")

		for _, osArch := range OsArchMatrix {
			output := path.Join(relativeRootPath, binaryOutputPathMulti(cmd.goModule, osArch["GOOS"], osArch["GOARCH"]))
			err := os.MkdirAll(path.Join(cmd.goModule, output), os.ModePerm)
			if err != nil {
				return err
			}
			bins = append(bins, mg.F(Go.build, cmd.goModule, input, output, osArch["GOOS"], osArch["GOARCH"]))
		}
	}

	mg.CtxDeps(ctx, bins...)

	return nil
}

// DownloadModules download the go modules
func (Go) DownloadModules(ctx context.Context) error {
	mg.CtxDeps(ctx, golangTargets.DownloadModules)
	return nil
}

type cmd struct {
	goModule string
	pkgs     []string
	binaries []string
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
		pkgs := []string{}
		bins := []string{}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			pkg := fmt.Sprintf("./%s", path.Join(cmdDir, entry.Name()))
			pkgs = append(pkgs, pkg)
			bins = append(bins, entry.Name())
		}
		if len(pkgs) == 0 {
			continue
		}
		cmd := cmd{
			goModule: goModule,
			pkgs:     pkgs,
			binaries: bins,
		}
		result = append(result, cmd)
	}
	return result, nil
}

func (Go) build(_ context.Context, workingDirectory, input, output, goos, goarch string) error {
	environmentalVariables := map[string]string{"GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": "0"}

	inputs := strings.Split(input, " ")

	args := []string{
		"-C",
		workingDirectory,
		"build",
		"-tags='datadog.no_waf'",
		"-o", output,
	}
	arguments := append(args, inputs...)

	return toolGo.Run(
		environmentalVariables,
		arguments...,
	)
}

// Validate runs validation check on the Go source code in the repository.
//
// For details see [Go.Test] and [Go.Lint].
func (Go) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.DownloadDevTools)
	mg.CtxDeps(ctx, Go.DownloadModules)
	mg.CtxDeps(ctx, Go.Test, Go.Lint)
	return nil
}

// Fix runs auto fixes on the Go source code in the repository.
//
// For details see [Go.LintFix].
func (Go) Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.DownloadDevTools)
	mg.CtxDeps(ctx, Go.DownloadModules)
	mg.CtxDeps(ctx, Go.LintFix)
	return nil
}

// Test automates testing the packages named by the import paths, see also: go
// test.
//
// For details see [golangTargets.Test].
func (Go) Test(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(golangTargets.Test))
	return nil
}

// Lint checks all Go source code for issues.
//
// For details see [golangTargets.Lint].
func (Go) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, golangTargets.Lint)
	return nil
}

// LintFix fixes found issues (if it's supported by the linters)
//
// For details see [golangTargets.LintFix].
func (Go) LintFix(ctx context.Context) error {
	mg.CtxDeps(ctx, golangTargets.LintFix)
	return nil
}

func binaryOutputPathMulti(app, os, arch string) string {
	return path.Join(binaryOutputBasePath(app), os, arch)
}

func binaryOutputBasePath(app string) string {
	return path.Join(core.OutputDir, app, binDir)
}

// Changes returns the string true or false depending on the fact that
// the current branch contains changes compared to the main branch.
func (Go) Changes(ctx context.Context) error {
	mg.CtxDeps(ctx, golangTargets.Changes)
	return nil
}

// DownloadDevTools download all devtools required for running the golang
// targets
func (Go) DownloadDevTools(ctx context.Context) error {
	mg.CtxDeps(
		ctx,
		mg.F(golangTargets.DownloadDevTool, "golang"),
		mg.F(golangTargets.DownloadDevTool, "golangci-lint"),
	)
	return nil
}
