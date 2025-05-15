package terraform

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/coopnorge/mage/internal/core"

	"github.com/stretchr/testify/assert"
)

func TestInitTarget(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		wantErr bool
	}{
		{
			name:    "Terraform Init target should succeed",
			workdir: "testdata/success",
			wantErr: false,
		},
		{
			name:    "Terraform Init target should fail",
			workdir: "testdata/invalid-code",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("recovered in f, this is expected", r)
				}
			}()

			dir, cleanup, _ := core.MkdirTemp()
			_ = os.CopyFS(dir, os.DirFS(tt.workdir))
			t.Chdir(dir)
			t.Cleanup(func() {
				cleanup()
			})
			gotErr := Init(context.Background())
			if tt.wantErr {
				//nolint:errcheck
				assert.Panics(t, func() { Init(context.Background()) }, "this function should panic")
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
		})
	}
}

func TestTestTarget(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		wantErr bool
	}{
		{
			name:    "Terraform Test target should succeed",
			workdir: "testdata/success",
			wantErr: false,
		},
		{
			name:    "Terraform Test target should fail",
			workdir: "testdata/invalid-code",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("recovered in f, this is expected", r)
				}
			}()
			dir, cleanup, _ := core.MkdirTemp()
			_ = os.CopyFS(dir, os.DirFS(tt.workdir))
			t.Chdir(dir)
			t.Cleanup(func() {
				cleanup()
			})

			gotErr := Test(context.Background())
			if tt.wantErr {
				//nolint:errcheck
				assert.Panics(t, func() { Test(context.Background()) }, "this function should panic")
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
		})
	}
}

func TestLintTarget(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		wantErr bool
	}{
		{
			name:    "Terraform Lint target should succeed",
			workdir: "testdata/success",
			wantErr: false,
		},
		{
			name:    "Terraform Lint target should on terraform fmt fail",
			workdir: "testdata/formatting-fail",
			wantErr: true,
		},
		{
			name:    "Terraform Lint target should on tflint fail",
			workdir: "testdata/linting-fail",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("recovered in f, this is expected", r)
				}
			}()
			dir, cleanup, _ := core.MkdirTemp()
			_ = os.CopyFS(dir, os.DirFS(tt.workdir))
			t.Chdir(dir)
			t.Cleanup(func() {
				cleanup()
			})

			gotErr := Lint(context.Background())
			if tt.wantErr {
				//nolint:errcheck
				assert.Panics(t, func() { Lint(context.Background()) }, "this function should panic")
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
		})
	}
}

func TestLintFixTarget(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		wantErr bool
	}{
		{
			name:    "Terraform Lint target should succeed",
			workdir: "testdata/success",
			wantErr: false,
		},
		{
			name:    "Terraform Lint target should fail",
			workdir: "testdata/formatting-fail",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("recovered in f, this is expected", r)
				}
			}()
			dir, cleanup, _ := core.MkdirTemp()
			_ = os.CopyFS(dir, os.DirFS(tt.workdir))
			t.Chdir(dir)
			t.Cleanup(func() {
				cleanup()
			})

			gotErr := LintFix(context.Background())
			if tt.wantErr {
				//nolint:errcheck
				assert.Panics(t, func() { LintFix(context.Background()) }, "this function should panic")
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}

			// Lint should pass after fix
			gotErr = Lint(context.Background())
			assert.NoError(t, gotErr)
		})
	}
}

func TestInitUpgradeTarget(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		wantErr bool
	}{
		{
			name:    "Terraform Init Upgrade target should succeed",
			workdir: "testdata/locking",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("recovered in f, this is expected", r)
				}
			}()
			dir, cleanup, _ := core.MkdirTemp()
			_ = os.CopyFS(dir, os.DirFS(tt.workdir))
			t.Chdir(dir)
			t.Cleanup(func() {
				cleanup()
			})

			gotErr := InitUpgrade(context.Background())
			if tt.wantErr {
				//nolint:errcheck
				assert.Panics(t, func() { InitUpgrade(context.Background()) }, "this function should panic")
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}

			// check for dirs and files
			assert.FileExists(t, "code-5/.terraform.lock.hcl")
			assert.DirExists(t, "code-5/.terraform")

			lockfile, gotErr := os.ReadFile("code-5/.terraform.lock.hcl")
			assert.NoError(t, gotErr)
			// version     = "3.7.1"
			assert.Contains(t, string(lockfile), "version     = \"3.7.2\"")
		})
	}
}

func TestLockProvidersTarget(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		wantErr bool
	}{
		{
			name:    "Terraform LockProviders target should succeed",
			workdir: "testdata/locking",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("recovered in f, this is expected", r)
				}
			}()
			dir, cleanup, _ := core.MkdirTemp()
			_ = os.CopyFS(dir, os.DirFS(tt.workdir))
			t.Chdir(dir)
			t.Cleanup(func() {
				cleanup()
			})

			gotErr := LockProviders(context.Background())
			if tt.wantErr {
				//nolint:errcheck
				assert.Panics(t, func() { LockProviders(context.Background()) }, "this function should panic")
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}

			// check for dirs and files
			assert.FileExists(t, "code-5/.terraform.lock.hcl")
			assert.DirExists(t, "code-5/.terraform")

			lockfile, gotErr := os.ReadFile("code-5/.terraform.lock.hcl")
			assert.NoError(t, gotErr)
			//assert.Contains(t,string(lockfile),"version     = \"3.7.2\"")
			resultlockfile, gotErr := os.ReadFile("code-5/.terraform.lock.hcl")
			assert.NoError(t, gotErr)

			assert.Equal(t, lockfile, resultlockfile)
		})
	}
}
