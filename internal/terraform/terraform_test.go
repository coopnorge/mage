package terraform

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
)

var (
	//go:embed testdata/tools.Dockerfile
	// TerraformToolsDockerfile the content of tools.Dockerfile
	TerraformToolsDockerfile string
)

func TestFindTerraformFolders(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		want    []string
		wantErr bool
	}{
		{
			name:    "Should find all relevant folders",
			workdir: "testdata/folders",
			want:    []string{"a", "b"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(tt.workdir)
			got, gotErr := FindTerraformProjects(".")
			assert.NoError(t, gotErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestInitUpgradet(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
	}{
		{
			name:    "Terraform InitUpgrade upgrades versions within constraints",
			workdir: "testdata/init-upgrade",
		},
	}

	err := devtool.Build("terraform", TerraformToolsDockerfile)
	if err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup, _ := core.MkdirTemp()
			_ = os.CopyFS(dir, os.DirFS(tt.workdir))
			t.Chdir(dir)
			t.Cleanup(func() {
				cleanup()
			})

			gotErr := InitUpgrade(".")
			assert.NoError(t, gotErr)
			// check for dirs and files
			assert.FileExists(t, ".terraform.lock.hcl")
			assert.DirExists(t, ".terraform")

			lockfile, gotErr := os.ReadFile(".terraform.lock.hcl")
			assert.NoError(t, gotErr)
			assert.Contains(t, string(lockfile), "version     = \"3.7.2\"")
		})
	}
}

func TestLockProviders(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		wantErr bool
	}{
		{
			name:    "Terraform LockProviders target should succeed",
			workdir: "testdata/providers-lock",
		},
	}
	err := devtool.Build("terraform", TerraformToolsDockerfile)
	if err != nil {
		panic(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup, _ := core.MkdirTemp()
			_ = os.CopyFS(dir, os.DirFS(tt.workdir))
			t.Chdir(dir)
			t.Cleanup(func() {
				cleanup()
			})

			gotErr := ProviderLock(".")
			assert.NoError(t, gotErr)

			lockfile, gotErr := os.ReadFile(".terraform.lock.hcl")
			assert.NoError(t, gotErr)
			resultlockfile, gotErr := os.ReadFile("result.terraform.lock.hcl")
			assert.NoError(t, gotErr)

			assert.Equal(t, string(lockfile), string(resultlockfile))
		})
	}
}

func TestCheckLock(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		notTracked []string
		checkDir   string
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Project with lockfile should succeed",
			files: map[string]string{
				".terraform.lock.hcl": "lockfile content",
				"main.tf":             "resource \"null_resource\" \"this\" {}",
			},
			wantErr: false,
		},
		{
			name: "Project without lockfile should fail",
			files: map[string]string{
				"main.tf": "resource \"null_resource\" \"this\" {}",
			},
			wantErr: true,
			errMsg:  "is not tracked by git in directory",
		},
		{
			name: "Project with untracked lockfile should fail",
			files: map[string]string{
				".terraform.lock.hcl": "lockfile content",
				"main.tf":             "resource \"null_resource\" \"this\" {}",
			},
			notTracked: []string{".terraform.lock.hcl"},
			wantErr:    true,
			errMsg:     "is not tracked by git in directory",
		},
		{
			name: "Module with lockfile should fail if tracked",
			files: map[string]string{
				"terraform-docs.yml":  "config",
				".terraform.lock.hcl": "lockfile content",
				"main.tf":             "resource \"null_resource\" \"this\" {}",
			},
			wantErr: true,
			errMsg:  "but it looks like a module",
		},
		{
			name: "Module without lockfile should succeed",
			files: map[string]string{
				"terraform-docs.yml": "config",
				"main.tf":            "resource \"null_resource\" \"this\" {}",
			},
			wantErr: false,
		},
		{
			name: "Submodule without lockfile should succeed",
			files: map[string]string{
				"main.tf":        "resource \"null_resource\" \"this\" {}",
				"subdir/main.tf": "resource \"null_resource\" \"this\" {}",
			},
			checkDir: "subdir",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			tempDir, err := filepath.EvalSymlinks(tempDir)
			if err != nil {
				t.Fatalf("Failed to eval symlinks for temp dir: %v", err)
			}

			projectBase := filepath.Join(tempDir, "workspace", "project")
			if err := os.MkdirAll(projectBase, 0755); err != nil {
				t.Fatalf("Failed to create project base: %v", err)
			}

			for p, content := range tt.files {
				fullPath := filepath.Join(projectBase, p)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatalf("Failed to create directory for file %s: %v", p, err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", p, err)
				}
			}

			t.Chdir(projectBase)

			// Initialize git and track all files by default
			sh.Run("git", "init")
			sh.Run("git", "add", ".")
			for _, f := range tt.notTracked {
				sh.Run("git", "rm", "--cached", f)
			}
			sh.Run("git", "commit", "-m", "initial commit")

			testDir := projectBase
			if tt.checkDir != "" {
				testDir = filepath.Join(projectBase, tt.checkDir)
			}

			err = CheckLock(testDir)
			if tt.wantErr {
				if assert.Error(t, err, tt.name) && tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err, tt.name)
			}
		})
	}
}
