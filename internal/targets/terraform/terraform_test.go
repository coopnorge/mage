package terraform

import (
	"os"
	"testing"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/targets/testhelpers"
	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargets(t *testing.T) {
	tests := []struct {
		name        string
		testProject string
		targets     []string
		wantErr     bool
	}{
		{
			name:        "Terraform Init target should succeed",
			testProject: "testdata/success",
			targets:     []string{"goapp:terraform:init"},
			wantErr:     false,
		},
		{
			name:        "Terraform Init target should fail",
			testProject: "testdata/fail-init",
			targets:     []string{"goapp:terraform:init"},
			wantErr:     true,
		},
		{
			name:        "Terraform Test target should succeed",
			testProject: "testdata/success",
			targets:     []string{"goapp:terraform:test"},
			wantErr:     false,
		},
		{
			name:        "Terraform Test target should fail",
			testProject: "testdata/fail-test",
			targets:     []string{"goapp:terraform:test"},
			wantErr:     true,
		},
		{
			name:        "Terraform Lint target should succeed",
			testProject: "testdata/success",
			targets:     []string{"goapp:terraform:lint"},
			wantErr:     false,
		},
		{
			name:        "Terraform LintFix target should succeed",
			testProject: "testdata/success",
			targets:     []string{"goapp:terraform:lintfix"},
			wantErr:     false,
		},
		{
			name:        "Terraform Lint target should fail on linting",
			testProject: "testdata/fail-lint",
			targets:     []string{"goapp:terraform:lint"},
			wantErr:     true,
		},
		{
			name:        "Terraform Lint target should fail on formatting",
			testProject: "testdata/fail-fmt",
			targets:     []string{"goapp:terraform:lint"},
			wantErr:     true,
		},
		{
			name:        "Terraform LintFix target should fix formatting",
			testProject: "testdata/fail-fmt",
			targets:     []string{"goapp:terraform:lintfix", "goapp:terraform:lint"},
			wantErr:     false,
		},
		{
			name:        "Terraform Security target should succeed",
			testProject: "testdata/success",
			targets:     []string{"goapp:terraform:security"},
			wantErr:     false,
		},
		{
			name:        "Terraform Security target should fail",
			testProject: "testdata/fail-security",
			targets:     []string{"goapp:terraform:security"},
			wantErr:     true,
		},
		{
			name:        "Terraform InitUpgrade target should succeed",
			testProject: "testdata/success",
			targets:     []string{"goapp:terraform:initupgrade"},
			wantErr:     false,
		},
		{
			name:        "Terraform LockProviders target should succeed",
			testProject: "testdata/success",
			targets:     []string{"goapp:terraform:lockproviders"},
			wantErr:     false,
		},
		{
			name:        "Terraform DocsValidate target should succeed",
			testProject: "testdata/module-success",
			targets:     []string{"terraformmodule:terraform:docsvalidate"},
			wantErr:     false,
		},
		{
			name:        "Terraform DocsValidate target should succeed for module with submodules",
			testProject: "testdata/module-with-submodule-success",
			targets:     []string{"terraformmodule:terraform:docsvalidate"},
			wantErr:     false,
		},
		{
			name:        "Terraform DocsValidate should fail for module without terraform-docs.yml",
			testProject: "testdata/fail-module-without-terraform-docs-yml",
			targets:     []string{"terraformmodule:terraform:docsvalidate"},
			wantErr:     true,
		},
		{
			name:        "Terraform DocsValidate target should fail",
			testProject: "testdata/fail-module-docs",
			targets:     []string{"terraformmodule:terraform:docsvalidate"},
			wantErr:     true,
		},
		{
			name:        "Terraform DocsValidateFix target should fix",
			testProject: "testdata/fail-module-docs",
			targets:     []string{"terraformmodule:terraform:docsvalidatefix", "terraformmodule:terraform:docsvalidate"},
			wantErr:     false,
		},
		{
			name:        "Terraform validate should succeed when lock file exists",
			testProject: "testdata/project-success",
			targets:     []string{"goapp:terraform:validate"},
			wantErr:     false,
		},
		{
			name:        "Terraform validate should fail when lock file doesn't exists",
			testProject: "testdata/fail-project-lockfile-missing",
			targets:     []string{"goapp:terraform:validate"},
			wantErr:     true,
		},
		{
			name:        "Terraform validate should fail for module with lockfile",
			testProject: "testdata/fail-module-with-lockfile",
			targets:     []string{"terraformmodule:terraform:validate"},
			wantErr:     true,
		},
		{
			name:        "Terraform validate should succeed for project with module without lockfile",
			testProject: "testdata/project-with-submodule-success",
			targets:     []string{"goapp:terraform:validate"},
			wantErr:     false,
		},
	}

	goModuleFactory := testhelpers.CreateGoModuleFactory(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// copy stuff to a temp dir to not leave any trace in the testdata
			dir, cleanup, err := core.MkdirTemp()
			require.NoError(t, err)
			t.Cleanup(cleanup)
			err = os.CopyFS(dir, os.DirFS(tt.testProject))
			require.NoError(t, err)
			err = os.CopyFS(dir, os.DirFS("testdata/layout"))
			require.NoError(t, err)

			t.Chdir(dir)

			err = sh.Run("git", "init")
			require.NoError(t, err)
			err = sh.Run("git", "config", "user.email", "test@example.com")
			require.NoError(t, err)
			err = sh.Run("git", "config", "user.name", "Test User")
			require.NoError(t, err)
			err = sh.Run("git", "add", ".")
			require.NoError(t, err)
			err = sh.Run("git", "commit", "-m", "initial commit")
			require.NoError(t, err)

			goModuleFactory(t)

			args := []string{"tool", "mage", "-v"}
			args = append(args, tt.targets...)
			gotErr := sh.RunV("go", args...)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
		})
	}
}
