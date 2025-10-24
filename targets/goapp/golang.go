package goapp

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/coopnorge/mage/internal/core"
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
	// OsArchMatrix defines the CPU architectures to build binaries for
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
	mg.CtxDeps(ctx, golangTargets.DownloadModules)
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

// DownloadModules download the go modules
func (Go) DownloadModules(ctx context.Context) error {
	mg.CtxDeps(ctx, golangTargets.DownloadModules)
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
	mg.CtxDeps(ctx, mg.F(devtoolTarget.Build, "golang", golangTargets.GolangToolsDockerfile))

	environmentalVariables := map[string]string{"GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": "0"}

	return golang.DevtoolGo(
		environmentalVariables,
		"go",
		"-C",
		workingDirectory,
		"build",
		"-v",
		"-tags='datadog.no_waf'",
		"-o", output,
		input)
}

// Validate runs validation check on the Go source code in the repository.
//
// For details see [Go.Test] and [Go.Lint].
func (Go) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Test, Go.Lint)
	return nil
}

// Fix runs auto fixes on the Go source code in the repository.
//
// For details see [Go.LintFix].
func (Go) Fix(ctx context.Context) error {
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

func binaryOutputPath(app, os, arch, binary string) string {
	return path.Join(binaryOutputBasePath(app), os, arch, binary)
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
