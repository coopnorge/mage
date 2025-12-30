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
	return DevtoolTerraform(nil, directory, "validate")
}

// Lint runs the linters
func Lint(directory, tfLintCfg string) error {
	lintCfg, cleanup, err := core.WriteTempFile(directory, "tflint.hcl", tfLintCfg)
	if err != nil {
		return err
	}
	defer cleanup()

	err = DevtoolTerraform(nil, directory, "fmt", "-diff", "-check")
	if err != nil {
		return err
	}
	err = DevtoolTFLint(nil, directory, "--init", "--color", fmt.Sprintf("--config=%s", filepath.Base(lintCfg)))
	if err != nil {
		return err
	}
	err = DevtoolTFLint(nil, directory, "--color", fmt.Sprintf("--config=%s", filepath.Base(lintCfg)))
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

	err = DevtoolTerraform(nil, directory, "fmt", "-diff")
	if err != nil {
		return err
	}
	err = DevtoolTFLint(nil, directory, "--init", "--color", fmt.Sprintf("--config=%s", filepath.Base(lintCfg)))
	if err != nil {
		return err
	}
	err = DevtoolTFLint(nil, directory, "--fix", "--color", fmt.Sprintf("--config=%s", filepath.Base(lintCfg)))
	if err != nil {
		return err
	}

	return nil
}

// Init downloads Terraform modules locally
func Init(directory string) error {
	log.Printf("Running terraform init for  %q", directory)
	err := DevtoolTerraform(nil, directory, "init")
	if err != nil {
		return err
	}
	return nil
}

// InitUpgrade downloads and updates Terraform modules locally
func InitUpgrade(directory string) error {
	log.Printf("Running terraform init -upgrade for  %q", directory)
	err := DevtoolTerraform(nil, directory, "init", "-upgrade")
	if err != nil {
		return err
	}
	return nil
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

	isModule := HasTerraformDocsConfig(directory) || IsTerraformSubmodule(directory)

	if isModule {
		if hasLockFile {
			return fmt.Errorf("lockfile %q found in directory %q, but it looks like a module (has terraform-docs.yml or is a submodule)", lockfile, directory)
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
	err := DevtoolTerraform(nil, directory, "providers", "lock",
		"-platform=linux_arm64",
		"-platform=linux_amd64",
		"-platform=darwin_amd64",
		"-platform=darwin_arm64",
		"-platform=windows_amd64",
	)
	if err != nil {
		return err
	}
	return nil
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
	return DevtoolTrivy(nil, directory, "config", "--exit-code", "1", "--misconfig-scanners=terraform", "./")
}

// HasTerraformDocsConfig checks whether the given directory
// contains a terraform-docs.yml configuration file.
func HasTerraformDocsConfig(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "terraform-docs.yml"))
	return err == nil
}

// IsTerraformSubmodule returns true if the directory is a subdirectory
// of another terraform project.
func IsTerraformSubmodule(dir string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return false
	}
	absCwd, err = filepath.EvalSymlinks(absCwd)
	if err != nil {
		return false
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}

	absDir, err = filepath.EvalSymlinks(absDir)
	if err != nil {
		return false
	}

	parent := filepath.Dir(absDir)
	for parent != absDir && strings.HasPrefix(parent, absCwd) {
		files, err := filepath.Glob(filepath.Join(parent, "*.tf"))
		if err != nil {
			return false
		}
		if len(files) > 0 {
			return true
		}

		absDir = parent
		parent = filepath.Dir(absDir)
	}

	return false
}

// Docs validate if the README of a module are up to date with the
// content of the module
func Docs(directory string) error {
	if !HasTerraformDocsConfig(directory) {
		return nil
	}

	return DevtoolTerraformDocs(nil, directory, ".", "-c", "terraform-docs.yml", "--output-check")
}

// DocsFix updates the README to the configuration of the module
func DocsFix(directory string) error {
	if !HasTerraformDocsConfig(directory) {
		return nil
	}

	return DevtoolTerraformDocs(nil, directory, ".", "-c", "terraform-docs.yml")
}

// DevtoolTerraform runs the devtool for terraform
func DevtoolTerraform(env map[string]string, directory string, cmd string, args ...string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// TODO: add provider cache
	dockerArgs := []string{
		"--volume", "$HOME/.cache:/root/.cache",
		"--volume", "$HOME/.terraform.d:/root/.terraform.d",
		"--volume", "$HOME/.ssh:/root/.ssh",
		"--volume", fmt.Sprintf("%s:/src", cwd),
		"--workdir", path.Join("/src", directory),
	}

	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	return devtool.Run("terraform", dockerArgs, cmd, args...)
}

// DevtoolTFLint runs the devtool for tflint
func DevtoolTFLint(env map[string]string, directory string, cmd string, args ...string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if env == nil {
		env = make(map[string]string)
	}

	if _, exists := env["GITHUB_TOKEN"]; !exists {
		if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
			env["GITHUB_TOKEN"] = token
		}
	}

	dockerArgs := []string{
		"--volume", "$HOME/.tflint.d:/root/.tflint.d",
		"--volume", fmt.Sprintf("%s:/src", cwd),
		"--workdir", path.Join("/src", directory),
	}

	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	return devtool.Run("tflint", dockerArgs, cmd, args...)
}

// DevtoolTrivy runs the devtool for trivy
func DevtoolTrivy(env map[string]string, directory string, cmd string, args ...string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/src", cwd),
		"--volume", "$HOME/.cache/trivy:/root/.cache/trivy",
		"--workdir", path.Join("/src", directory),
	}

	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	return devtool.Run("trivy", dockerArgs, cmd, args...)
}

// DevtoolTerraformDocs the devtool for trivy
func DevtoolTerraformDocs(env map[string]string, directory string, cmd string, args ...string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/src", cwd),
		"--workdir", path.Join("/src", directory),
	}

	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	return devtool.Run("terraform-docs", dockerArgs, cmd, args...)
}
