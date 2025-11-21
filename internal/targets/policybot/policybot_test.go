package policybot

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"
)

var goModTemplateString = `module dummy
go 1.25.0
require github.com/coopnorge/mage v0.7.0
require github.com/magefile/mage v1.15.0 // indirect
tool github.com/magefile/mage
replace github.com/coopnorge/mage => {{ . }}
`

func TestPolicyBotTargets(t *testing.T) {
	tests := []struct {
		name        string
		testProject string
		targets     []string
		wantErr     bool
	}{
		{
			name:        "PolicyBot Validate should succeed",
			testProject: "testdata/success",
			targets:     []string{"goapp:policyBotConfig:validate"},
			wantErr:     false,
		},
		{
			name:        "PolicyBot Validate should fail",
			testProject: "testdata/fail-validate",
			targets:     []string{"goapp:policyBotConfig:validate"},
			wantErr:     true,
		},
		{
			name:        "PolicyBot Changes should succeed when no changes",
			testProject: "testdata/success-no-policy-file",
			targets:     []string{"goapp:policyBotConfig:validate"},
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
			// create isolated temp project
			dir, cleanup, _ := core.MkdirTemp()
			err := os.CopyFS(dir, os.DirFS(tt.testProject))
			if err != nil {
				fmt.Printf("Error while copying %s", tt.testProject)
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
