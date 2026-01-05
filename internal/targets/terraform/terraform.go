package terraform

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/coopnorge/mage/internal/terraform"
	"github.com/magefile/mage/mg"
)

// TFlintCfg is the config for tflint
//
//go:embed .tflint.hcl
var TFlintCfg string

// Test runs terraform validate
func Test(ctx context.Context) error {
	directories, err := terraform.FindTerraformProjects(".")
	fmt.Println("found test dirs", directories)
	if err != nil {
		return err
	}

	var testDirs []any
	var checkLocks []any
	for _, workDir := range directories {
		testDirs = append(testDirs, mg.F(test, workDir))
		checkLocks = append(checkLocks, mg.F(checkLock, workDir))
	}

	mg.CtxDeps(ctx, checkLocks...)
	mg.CtxDeps(ctx, Init)
	mg.CtxDeps(ctx, testDirs...)
	return nil
}

func test(ctx context.Context, workingDirectory string) error {
	return terraform.Test(workingDirectory)
}

func checkLock(_ context.Context, workingDirectory string) error {
	return terraform.CheckLock(workingDirectory)
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
	return terraform.Lint(workingDirectory, TFlintCfg)
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
	return terraform.LintFix(workingDirectory, TFlintCfg)
}

// Init initializes a terraform project
func Init(ctx context.Context) error {
	directories, err := terraform.FindTerraformProjects(".")
	fmt.Println("found dirs", directories)
	if err != nil {
		return err
	}
	modules := []any{}
	for _, workDir := range directories {
		fmt.Println("adding dep initTerraform", workDir)
		modules = append(modules, mg.F(initTerraform, workDir))
	}

	mg.SerialCtxDeps(ctx, modules...)
	return nil
}

func initTerraform(_ context.Context, directory string) error {
	fmt.Println("Runninng initTerraform", directory)
	return terraform.Init(directory)
}

// InitUpgrade initializes and upgrades the provides and modules within
// the version constraints
func InitUpgrade(ctx context.Context) error {
	directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	modules := []any{}
	for _, workDir := range directories {
		modules = append(modules, mg.F(initUpgrade, workDir))
	}

	mg.SerialCtxDeps(ctx, modules...)
	return nil
}

func initUpgrade(_ context.Context, directory string) error {
	return terraform.InitUpgrade(directory)
}

// LockProviders locks the providers for a certain set of host systems
func LockProviders(ctx context.Context) error {
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

// Clean implements cleaning the module and provider cache of the terraform
// projects (Unimplemented)
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

// Security implements security related targets
func Security(ctx context.Context) error {
	directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	modules := []any{}
	for _, workDir := range directories {
		modules = append(modules, mg.F(security, workDir))
	}

	mg.SerialCtxDeps(ctx, modules...)
	return nil
}

func security(_ context.Context, directory string) error {
	return terraform.Security(directory)
}

// DocsValidate implements validation of terraform module documentation
func DocsValidate(ctx context.Context) error {
	if err := checkTerraformDocsConfig("."); err != nil {
		return err
	}

	directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	modules := []any{}
	for _, workDir := range directories {
		modules = append(modules, mg.F(terraformDocs, workDir))
	}

	mg.SerialCtxDeps(ctx, modules...)
	return nil
}

func terraformDocs(_ context.Context, directory string) error {
	return terraform.Docs(directory)
}

// DocsValidateFix implements fixing of terraform module documentation
func DocsValidateFix(ctx context.Context) error {
	if err := checkTerraformDocsConfig("."); err != nil {
		return err
	}

	directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	modules := []any{}
	for _, workDir := range directories {
		modules = append(modules, mg.F(terraformDocsFix, workDir))
	}

	mg.SerialCtxDeps(ctx, modules...)
	return nil
}

func terraformDocsFix(_ context.Context, directory string) error {
	return terraform.DocsFix(directory)
}

func checkTerraformDocsConfig(directory string) error {
	if !terraform.HasTerraformDocsConfig(directory) {
		return fmt.Errorf("terraform-docs.yml config not found in module root")
	}
	return nil
}

// Changes implements a target that check if the current branch has changes
// related to main branch
func Changes(_ context.Context) error {
	directories, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	changes, err := terraform.HasChanges(directories)
	if err != nil {
		return err
	}

	if changes {
		fmt.Println("true")
		return nil
	}
	fmt.Println("false")
	return nil
}
