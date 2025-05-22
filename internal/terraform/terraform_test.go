package terraform

import (
	_ "embed"
	"os"
	"testing"

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
