package terraform

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"
)

var goModTemplateString = `module dummy
go 1.24.0
require github.com/coopnorge/mage v0.4.3
require github.com/magefile/mage v1.15.0 // indirect
tool github.com/magefile/mage
replace github.com/coopnorge/mage => {{ . }}
`

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
	}

	goModMage, err := sh.Output("go", "env", "GOMOD")
	if err != nil {
		panic(err)
	}
	mageRoot := filepath.Dir(goModMage)
	goModTemplate, err := template.New("gomod").Parse(goModTemplateString)
	if err != nil {
		panic(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// copy stuff to a temp dir to not leave any trace in the testdata
			dir, cleanup, _ := core.MkdirTemp()
			err := os.CopyFS(dir, os.DirFS(tt.testProject))
			if err != nil {
				panic(err)
			}
			err = os.CopyFS(dir, os.DirFS("testdata/layout"))
			if err != nil {
				panic(err)
			}

			t.Chdir(dir)

			goMod, err := os.Create("go.mod")
			if err != nil {
				panic(err)
			}
			goModTemplate.Execute(goMod, mageRoot)

			t.Cleanup(func() {
				goMod.Close()
				cleanup()
			})
			args := []string{"tool", "mage"}
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
