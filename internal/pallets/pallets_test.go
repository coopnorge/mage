package pallets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePalletOwners(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		kind        string
		wantErr     bool
		errContains string
	}{
		{name: "owner from component", owner: "team-a"},
		{name: "owner from system", owner: "team-b"},
		{name: "unknown owner", owner: "team-c", wantErr: true, errContains: "does not match an owner"},
		{name: "non-string owner", owner: "42", wantErr: true, errContains: "spec.owner must be a non-empty string"},
		{name: "missing owner", owner: "", wantErr: false},
		{name: "null owner", owner: "null", wantErr: true, errContains: "spec.owner must be a non-empty string"},
		{name: "other pallet kind ignores owner", owner: "", kind: "BranchProtection", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			require.NoError(t, os.WriteFile("catalog-info.yaml", []byte(`apiVersion: backstage.io/v1alpha1
kind: System
metadata:
  name: system-name
spec:
  owner: team-b
---
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: component-name
spec:
  type: service
  lifecycle: production
  owner: team-a
`), 0o600))

			palletDir := filepath.Join(dir, ".pallet")
			require.NoError(t, os.Mkdir(palletDir, 0o700))
			kind := tt.kind
			if kind == "" {
				kind = "RepositoryConfig"
			}
			pallet := "apiVersion: github.coop.no/v1alpha1\nkind: " + kind + "\nspec:\n"
			if tt.owner != "" {
				pallet += "  owner: " + tt.owner + "\n"
			}
			palletPath := filepath.Join(palletDir, "config.yaml")
			require.NoError(t, os.WriteFile(palletPath, []byte(pallet), 0o600))

			err := validatePalletOwners([]string{palletPath})
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			assert.NoError(t, err)
		})
	}
}
