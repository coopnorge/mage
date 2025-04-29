package terraform

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/coopnorge/mage/internal/devtool"
)


// IsTerraformProject returns true if a directory contains a go module.
func IsTerraformProject(p string, d fs.DirEntry) bool {
	if !d.IsDir() {
		return false
	}
	if _, err := os.Stat(path.Join(p, ".terraform.hcl.lock")); os.IsNotExist(err) {
		return false
	}
	return true
}

// FindTerraformProjects will search through the base directory to find the all the
// Go modules.
func FindTerraformProjects(base string) ([]string, error) {
	directories := []string{}

	err := filepath.WalkDir(base, func(workDir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if isDotDirectory(workDir, d) {
			return nil
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

func isDotDirectory(path string, d fs.DirEntry) bool {
	if !d.IsDir() {
		return false
	}
	if path == "." {
		return false
	}
	return strings.HasPrefix(path, ".")
}



// Test automates testing the packages named by the import paths, see also: go
// test.
func Test(directory string) error {
	return DevtoolTerraform(nil, directory, "terraform", "validate")
}

// Lint runs the linters
func Lint(directory string) error {

	err := DevtoolTerraform(nil, directory, "terraform", "fmt","-diff","-check")
	if err != nil {
		return err
	}

    err = DevtoolTFLint(nil, directory, "tflint")
	if err != nil {
		return err
	}

	return nil
}

// LintFix fixes found issues (if it's supported by the linters)
func LintFix(directory string) error {

    err := DevtoolTerraform(nil, directory, "terraform", "fmt","-diff")
	if err != nil {
		return err
	}

    err = DevtoolTFLint(nil, directory, "tflint","-fix")
	if err != nil {
		return err
	}

	return nil
}

// DownloadModules downloads Go modules locally
func Init(directory string) error {
	log.Printf("Running terraform init for  %q", directory)
	return DevtoolTerraform(nil, directory, "terraform", "init")
}


func Clean(directory string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cache := path.Join(cwd,directory)
	cache = path.Join(cache,".terraform")
	log.Printf("(NOT YET IMPLEMENTED: Deleting content in %q", cache)

	// return os.RemoveAll()
	return nil
}

// DevtoolGo runs the devtool for Go
func DevtoolTerraform(env map[string]string, directory string ,cmd string, args ...string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

    // TODO: add provider cache
	dockerArgs := []string{
		"--volume", "$HOME/.cache:/root/.cache",
		"--volume", "$HOME/.terraform.d:/root/.terraform.d",
		"--volume", "$HOME/.gitconfig:/root/.gitconfig",
		"--volume", "$HOME/.ssh:/root/.ssh",
		"--volume", fmt.Sprintf("%s:/src", cwd),
		"--workdir", path.Join("/src",directory),
	}

	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	return devtool.Run("terraform", dockerArgs, cmd, args...)
}

// DevtoolGolangCILint runs the devtool for Golangci-lint
func DevtoolTFLint(env map[string]string, directory string, cmd string, args ...string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	dockerArgs := []string{
        "--volume", "$HOME/.cache:/root/.cache",
		"--volume", "$HOME/.terraform.d:/root/.terraform.d",
		"--volume", fmt.Sprintf("%s:/src", cwd),
		"--workdir", path.Join("/src",directory),
	}

	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	return devtool.Run("golangci-lint", dockerArgs, cmd, args...)
}
