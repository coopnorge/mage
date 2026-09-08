package cataloginfo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func setupMockGitHubServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})

	t.Setenv("GITHUB_TOKEN", "test-token")
	t.Setenv("GITHUB_REPOSITORY", "coopnorge/test")
	t.Setenv("GITHUB_API_URL", server.URL)
}

func TestValidate(t *testing.T) {
	t.Run("success valid owner and matching github team", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/orgs/coopnorge/teams?per_page=100&page=1", r.URL.RequestURI())
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			resp, _ := json.Marshal([]map[string]string{{"slug": "team-a"}})
			_, _ = w.Write(resp)
		})

		absPath, err := filepath.Abs("testdata/valid-system-component-resource")
		require.NoError(t, err)

		t.Chdir(absPath)

		err = Validate()
		assert.NoError(t, err)
	})

	t.Run("fail owner mismatch across catalog objects", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp, _ := json.Marshal([]map[string]string{{"slug": "team-a"}, {"slug": "team-b"}})
			_, _ = w.Write(resp)
		})

		absPath, err := filepath.Abs("testdata/fail-owner-mismatch")
		require.NoError(t, err)

		t.Chdir(absPath)

		err = Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner mismatch across catalog objects")
	})

	t.Run("fail owner not found in github teams", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp, _ := json.Marshal([]map[string]string{{"slug": "other-team"}})
			_, _ = w.Write(resp)
		})

		absPath, err := filepath.Abs("testdata/valid-system-component-resource")
		require.NoError(t, err)

		t.Chdir(absPath)

		err = Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner \"team-a\" is not a valid GitHub team in coopnorge")
	})

	t.Run("fail github api error", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		absPath, err := filepath.Abs("testdata/valid-system-component-resource")
		require.NoError(t, err)

		t.Chdir(absPath)

		err = Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch GitHub teams")
	})

	t.Run("fail empty owner or no entities", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		dir := t.TempDir()
		t.Chdir(dir)

		err := Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "catalog info is missing owner")
	})
}
