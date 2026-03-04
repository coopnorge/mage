package cataloginfo

import (
	"os"
	"testing"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/targets/testhelpers"
	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogInfoTargets(t *testing.T) {
	tests := []struct {
		name        string
		testProject string
		targets     []string
		wantErr     bool
	}{
		{
			name:        "CatalogInfo Validate should succeed",
			testProject: "testdata/success",
			targets:     []string{"goapp:catalogInfo:validate"},
			wantErr:     false,
		},
		{
			name:        "CatalogInfo Validate should fail",
			testProject: "testdata/fail-validate",
			targets:     []string{"goapp:catalogInfo:validate"},
			wantErr:     true,
		},
	}

	goModuleFactory := testhelpers.CreateGoModuleFactory(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create isolated temp project
			dir, cleanup, err := core.MkdirTemp()
			require.NoError(t, err)
			t.Cleanup(cleanup)
			err = os.CopyFS(dir, os.DirFS(tt.testProject))
			require.NoError(t, err)
			err = os.CopyFS(dir, os.DirFS("testdata/layout"))
			require.NoError(t, err)

			t.Chdir(dir)

			goModuleFactory(t)

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
