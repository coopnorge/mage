package terraform

import (
	"context"
	_ "embed"

	"github.com/coopnorge/mage/internal/terraform"
	"github.com/coopnorge/mage/internal/targets/devtool"
	"github.com/magefile/mage/mg"
)

var (
	//go:embed tools.Dockerfile
	// TerraformToolsDockerfile the content of tools.Dockerfile
	TerraformToolsDockerfile string
)

// Test runs terraform validate
func Test(ctx context.Context) error {
	mg.CtxDeps(ctx, Init)
	directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}

	testDirs := []any{}
	for _, workDir := range directories {
		testDirs = append(testDirs, mg.F(test, workDir))
	}

	mg.CtxDeps(ctx, testDirs...)
	return nil
}

func test(ctx context.Context, workingDirectory string) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "terraform", TerraformToolsDockerfile))
	return terraform.Test(workingDirectory)
}

// Lint runs the linters
func Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, Init)
	directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}

	lintDirs := []any{}
	for _, workDir := range directories {
		lintDirs = append(lintDirs, mg.F(lint, workDir))
	}

	mg.CtxDeps(ctx, lintDirs...)
	return nil
}

func lint(ctx context.Context, workingDirectory string) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "tflint", TerraformToolsDockerfile), mg.F(devtool.Build, "terraform", TerraformToolsDockerfile))
	return terraform.Lint(workingDirectory)
}

// LintFix fixes found issues (if it's supported by the linters)
func LintFix(ctx context.Context) error {
	mg.CtxDeps(ctx, Init)
	directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}

	lintDirs := []any{}
	for _, workDir := range directories {
		lintDirs = append(lintDirs, mg.F(lintFix, workDir))
	}

	mg.SerialCtxDeps(ctx, lintDirs...)

	return nil
}

func lintFix(ctx context.Context, workingDirectory string) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "tflint", TerraformToolsDockerfile), mg.F(devtool.Build, "terraform", TerraformToolsDockerfile))
	return terraform.LintFix(workingDirectory)
}

// Init (re)initializes a terraform project
func Init(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "terraform", TerraformToolsDockerfile))
	directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	modules := []any{}
	for _, workDir := range directories {
		modules = append(modules, mg.F(initTerraform, workDir))
	}

	mg.SerialCtxDeps(ctx, modules...)
	return nil
}

func initTerraform(_ context.Context, directory string) error {
	return terraform.Init(directory)
}

// Lock providers locks the providers for a certain set of host systems
func LockProviders(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "terraform", TerraformToolsDockerfile))
    directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	modules := []any{}
	for _, workDir := range directories {
		modules = append(modules, mg.F(lockProviders, workDir))
	}

	mg.SerialCtxDeps(ctx, modules...)
	return nil
}

func lockProviders(_ context.Context, directory string) error {
	return terraform.ProviderLock(directory)
}


func Clean(ctx context.Context) error {
    directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	modules := []any{}
	for _, workDir := range directories {
		modules = append(modules, mg.F(clean, workDir))
	}

	mg.SerialCtxDeps(ctx, modules...)
	return nil

}

func clean(_ context.Context, directory string) error {
   return terraform.Clean(directory)
}
