package terraform

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
)

var (
	devtoolTerraform devtool.Terraform
	devtoolTFLint    devtool.TFLint
	devtoolTrivy     devtool.Trivy
	devtoolTFDocs    devtool.TerraformDocs
)

// IsTerraformProject returns true if a directory contains a go module.
func IsTerraformProject(p string, d fs.DirEntry) bool {
	if !d.IsDir() {
		return false
	}
	// skip the examples dir form validation this could be more advanced
	if filepath.Base(p) == "examples" {
		return false
	}
	files, err := filepath.Glob(p + "/*.tf")
	if err != nil {
		panic("Unable to list .tf files")
	}
	if len(files) == 0 {
		return false
	}
	return true
}

// FindTerraformProjects will search through the base directory to find the
// all terraform projects
func FindTerraformProjects(base string) ([]string, error) {
	directories := []string{}

	err := filepath.WalkDir(base, func(workDir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if core.IsDotDirectory(workDir, d) {
			return filepath.SkipDir
		}
		if !IsTerraformProject(workDir, d) {
			return nil
		}

		directories = append(directories, workDir)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return directories, nil
}

// HasChanges checks if the current branch has any terraform changes compared
// to the main branch
func HasChanges(terraformProjects []string) (bool, error) {
	changedFiles, err := git.DiffToMain()
	if err != nil {
		return false, err
	}
	// always trigger on go.mod/sum and workflows because of changes in ci.
	additionalGlobs := append([]string{"go.mod", "go.sum", ".github/workflows/*"}, strings.Split(os.Getenv("ADDITIONAL_GLOBS_TERRAFORM"), ",")...)
	return core.CompareChangesToPaths(changedFiles, terraformProjects, additionalGlobs)
}

// Test automates testing the packages named by the import paths, see also: go
// test.
func Test(directory string) error {
	return devtoolTerraform.Run(nil, directory, "validate")
	// return DevtoolTerraform(nil, directory, "validate")
}

// Lint runs the linters
func Lint(directory, tfLintCfg string) error {
	lintCfg, cleanup, err := core.WriteTempFile(directory, "tflint.hcl", tfLintCfg)
	if err != nil {
		return err
	}
	defer cleanup()
	err = devtoolTerraform.Run(nil, directory, "fmt", "-diff", "-check")
	if err != nil {
		return err
	}
	err = devtoolTFLint.Run(nil, directory, "--init", "--color", fmt.Sprintf("--config=%s", filepath.Base(lintCfg)))
	if err != nil {
		return err
	}
	err = devtoolTFLint.Run(nil, directory, "--color", fmt.Sprintf("--config=%s", filepath.Base(lintCfg)))
	if err != nil {
		return err
	}

	return nil
}

// LintFix fixes found issues (if it's supported by the linters)
func LintFix(directory, tfLintCfg string) error {
	lintCfg, cleanup, err := core.WriteTempFile(directory, ".tflint.hcl", tfLintCfg)
	if err != nil {
		return err
	}
	defer cleanup()

	err = devtoolTerraform.Run(nil, directory, "fmt", "-diff")
	if err != nil {
		return err
	}

	err = devtoolTFLint.Run(nil, directory, "--init", "--color", fmt.Sprintf("--config=%s", filepath.Base(lintCfg)))
	if err != nil {
		return err
	}

	err = devtoolTFLint.Run(nil, directory, "--fix", "--color", fmt.Sprintf("--config=%s", filepath.Base(lintCfg)))
	if err != nil {
		return err
	}

	return nil
}

// Init downloads Terraform modules locally
func Init(directory string) error {
	log.Printf("Running terraform init for  %q", directory)

	return devtoolTerraform.Run(nil, directory, "init")
}

// InitUpgrade downloads and updates Terraform modules locally
func InitUpgrade(directory string) error {
	log.Printf("Running terraform init -upgrade for  %q", directory)
	return devtoolTerraform.Run(nil, directory, "init", "-upgrade")
}

// CheckLock checks that the lockfile exists
func CheckLock(directory string) error {
	log.Printf("Checking for terraform lockfile in %q", directory)

	lockfile := ".terraform.lock.hcl"
	lockfilePath := filepath.Join(directory, lockfile)
	hasLockFile := false
	if _, err := os.Stat(lockfilePath); err == nil {
		hasLockFile = true
	}

	if HasTerraformDocsConfig(directory) {
		if hasLockFile {
			return fmt.Errorf("lockfile %q found in directory %q, but it looks like a module (has terraform-docs.yml)", lockfile, directory)
		}
		return nil
	}

	if !hasLockFile {
		return fmt.Errorf("lockfile %q not found in directory %q as expected", lockfile, directory)
	}

	log.Printf("Lockfile %q found in %q", lockfile, directory)
	return nil
}

// ProviderLock updates the provider lock file locking poviders for a list of
// os architecures
func ProviderLock(directory string) error {
	log.Printf("Running terraform provider lock  %q", directory)
	return devtoolTerraform.Run(nil, directory, "providers", "lock",
		"-platform=linux_arm64",
		"-platform=linux_amd64",
		"-platform=darwin_amd64",
		"-platform=darwin_arm64",
		"-platform=windows_amd64",
	)
}

// Clean cache in a terraform directory
func Clean(directory string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cache := path.Join(cwd, directory)
	cache = path.Join(cache, ".terraform")
	log.Printf("(NOT YET IMPLEMENTED: Deleting content in %q", cache)

	// return os.RemoveAll()
	return nil
}

// Security validates security of the terraform project
// config --exit-code 1 --misconfig-scanners=terraform
func Security(directory string) error {
	return devtoolTrivy.Run(nil, directory, "config", "--exit-code", "1", "--misconfig-scanners=terraform", "./")
}

// HasTerraformDocsConfig checks whether the given directory
// contains a terraform-docs.yml configuration file.
func HasTerraformDocsConfig(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "terraform-docs.yml"))
	return err == nil
}

// Docs validate if the README of a module are up to date with the
// content of the module
func Docs(directory string) error {
	if !HasTerraformDocsConfig(directory) {
		return nil
	}

	return devtoolTFDocs.Run(nil, directory, ".", "-c", "terraform-docs.yml", "--output-check")
}

// DocsFix updates the README to the configuration of the module
func DocsFix(directory string) error {
	if !HasTerraformDocsConfig(directory) {
		return nil
	}

	return devtoolTFDocs.Run(nil, directory, ".", "-c", "terraform-docs.yml")
}
