package cataloginfo

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCatalogInfoFiles(t *testing.T) {
	tests := []struct {
		name          string
		testDir       string
		wantErr       bool
		errContains   string
		wantSystem    bool
		numComponents int
		numResources  int
	}{
		{
			name:          "valid system component and resource",
			testDir:       "testdata/valid-system-component-resource",
			wantErr:       false,
			wantSystem:    true,
			numComponents: 1,
			numResources:  1,
		},
		{
			name:          "valid zero system and multiple components",
			testDir:       "testdata/valid-multiple-components",
			wantErr:       false,
			wantSystem:    false,
			numComponents: 2,
			numResources:  0,
		},
		{
			name:        "fail more than one system",
			testDir:     "testdata/fail-multiple-systems",
			wantErr:     true,
			errContains: "more than one System defined",
		},
		{
			name:        "fail unknown kind",
			testDir:     "testdata/fail-unknown-kind",
			wantErr:     true,
			errContains: "unknown or unsupported entity kind \"Domain\"",
		},
		{
			name:        "fail missing apiVersion",
			testDir:     "testdata/fail-missing-apiversion",
			wantErr:     true,
			errContains: "missing required field apiVersion",
		},
		{
			name:        "fail missing metadata name",
			testDir:     "testdata/fail-missing-metadata-name",
			wantErr:     true,
			errContains: "missing required field metadata.name",
		},
		{
			name:        "fail missing component spec owner",
			testDir:     "testdata/fail-missing-spec-owner",
			wantErr:     true,
			errContains: "spec is missing owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath, err := filepath.Abs(tt.testDir)
			require.NoError(t, err)

			t.Chdir(absPath)

			data, err := parseCatalogInfoFiles()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, data)
				if tt.wantSystem {
					assert.NotNil(t, data.System)
				} else {
					assert.Nil(t, data.System)
				}
				assert.Len(t, data.Components, tt.numComponents)
				assert.Len(t, data.Resources, tt.numResources)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	absPath, err := filepath.Abs("testdata/valid-system-component-resource")
	require.NoError(t, err)

	t.Chdir(absPath)

	err = Validate()
	assert.NoError(t, err)
}
