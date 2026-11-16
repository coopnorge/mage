package cataloginfo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
		numAPIs       int
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
			numAPIs:       1,
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
			name:    "valid API with OpenAPI reference",
			testDir: "testdata/valid-api-openapi-reference",
			wantErr: false,
			numAPIs: 1,
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
		{
			name:        "fail missing API definition",
			testDir:     "testdata/fail-missing-api-definition",
			wantErr:     true,
			errContains: "API \"my-api\" spec is missing definition",
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
				assert.Len(t, data.APIs, tt.numAPIs)
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
	t.Setenv("CI", "true")
}

func TestValidate(t *testing.T) {
	t.Run("success valid owner and matching github team", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/orgs/coopnorge/teams?per_page=100&page=1", r.URL.RequestURI())
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			resp, err := json.Marshal([]map[string]string{{"slug": "team-a"}})
			require.NoError(t, err)
			_, err = w.Write(resp)
			require.NoError(t, err)
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
			resp, err := json.Marshal([]map[string]string{{"slug": "team-a"}, {"slug": "team-b"}})
			require.NoError(t, err)
			_, err = w.Write(resp)
			require.NoError(t, err)
		})

		absPath, err := filepath.Abs("testdata/fail-owner-mismatch")
		require.NoError(t, err)

		t.Chdir(absPath)

		err = Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner mismatch across catalog objects")
		assert.Contains(t, err.Error(), `System "my-system"`)
		assert.Contains(t, err.Error(), `Component "my-component"`)
	})

	t.Run("fail owner not found in github teams", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp, err := json.Marshal([]map[string]string{{"slug": "other-team"}})
			require.NoError(t, err)
			_, err = w.Write(resp)
			require.NoError(t, err)
		})

		absPath, err := filepath.Abs("testdata/valid-system-component-resource")
		require.NoError(t, err)

		t.Chdir(absPath)

		err = Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner \"team-a\" is not a valid GitHub team in coopnorge")
	})

	t.Run("fail github api error in CI", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		t.Setenv("CI", "true")

		absPath, err := filepath.Abs("testdata/valid-system-component-resource")
		require.NoError(t, err)

		t.Chdir(absPath)

		err = Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch GitHub teams")
	})

	t.Run("warn github api error when not in CI", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		err := os.Unsetenv("CI")
		require.NoError(t, err)

		absPath, err := filepath.Abs("testdata/valid-system-component-resource")
		require.NoError(t, err)

		t.Chdir(absPath)

		err = Validate()
		require.NoError(t, err)
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

	t.Run("APIs may have different owners", func(t *testing.T) {
		calls := 0
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			calls++
			writeTeams(t, w, "team-a", "team-b", "team-c")
		})
		t.Chdir(writeCatalog(t, catalogDoc("System", "sys", "team-a")+
			catalogDoc("Component", "comp", "team-a")+
			catalogDoc("API", "api-1", "team-b")+
			catalogDoc("API", "api-2", "team-c")+
			catalogDoc("API", "api-3", "team-b")))

		require.NoError(t, Validate())
		assert.Equal(t, 1, calls)
	})

	t.Run("fail non-API entity with different owner next to API", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			writeTeams(t, w, "team-a", "team-b")
		})
		t.Chdir(writeCatalog(t, catalogDoc("System", "sys", "team-a")+
			catalogDoc("API", "api-1", "team-b")+
			catalogDoc("Component", "comp", "team-b")))

		err := Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner mismatch across catalog objects")
		assert.Contains(t, err.Error(), `Component "comp"`)
	})

	t.Run("fail API missing owner", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			writeTeams(t, w, "team-a")
		})
		t.Chdir(writeCatalog(t, catalogDoc("System", "sys", "team-a")+catalogDoc("API", "api-1", "")))

		err := Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing owner")
	})

	t.Run("fail unknown API owner", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			writeTeams(t, w, "team-a")
		})
		t.Chdir(writeCatalog(t, catalogDoc("System", "sys", "team-a")+catalogDoc("API", "api-1", "team-b")))

		err := Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), `owner "team-b" is not a valid GitHub team in coopnorge`)
		assert.NotContains(t, err.Error(), `"team-a"`)
	})

	t.Run("lists all unknown owners", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			writeTeams(t, w, "other-team")
		})
		t.Chdir(writeCatalog(t, catalogDoc("System", "sys", "team-a")+
			catalogDoc("API", "api-1", "team-b")+
			catalogDoc("API", "api-2", "team-c")))

		err := Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), `owners "team-a", "team-b", "team-c" are not valid GitHub teams in coopnorge`)
	})

	t.Run("only APIs with different owners", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			writeTeams(t, w, "team-a", "team-b")
		})
		t.Chdir(writeCatalog(t, catalogDoc("API", "api-1", "team-a")+catalogDoc("API", "api-2", "team-b")))

		require.NoError(t, Validate())
	})

	t.Run("fail API owners on github api error in CI", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		t.Chdir(writeCatalog(t, catalogDoc("System", "sys", "team-a")+catalogDoc("API", "api-1", "team-b")))

		err := Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch GitHub teams")
	})

	t.Run("warn API owners on github api error when not in CI", func(t *testing.T) {
		setupMockGitHubServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		require.NoError(t, os.Unsetenv("CI"))
		t.Chdir(writeCatalog(t, catalogDoc("System", "sys", "team-a")+catalogDoc("API", "api-1", "team-b")))

		require.NoError(t, Validate())
	})
}

func catalogDoc(kind, name, owner string) string {
	spec := "  owner: " + owner + "\n"
	if owner == "" {
		spec = ""
	}
	switch kind {
	case "API":
		spec += "  type: openapi\n  lifecycle: production\n  definition: |\n    openapi: 3.0.0\n"
	case "Component":
		spec += "  type: service\n  lifecycle: production\n"
	}
	return "---\napiVersion: backstage.io/v1alpha1\nkind: " + kind + "\nmetadata:\n  name: " + name + "\nspec:\n" + spec
}

func writeCatalog(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "catalog-info.yaml"), []byte(content), 0o600))
	return dir
}

func writeTeams(t *testing.T, w http.ResponseWriter, slugs ...string) {
	t.Helper()
	teams := make([]map[string]string, len(slugs))
	for i, slug := range slugs {
		teams[i] = map[string]string{"slug": slug}
	}
	resp, err := json.Marshal(teams)
	require.NoError(t, err)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	require.NoError(t, err)
}

func TestMetadataNameLength(t *testing.T) {
	component := func(name string) string {
		return "apiVersion: backstage.io/v1alpha1\nkind: Component\nmetadata:\n  name: " + name +
			"\nspec:\n  type: service\n  lifecycle: production\n  owner: team-a\n"
	}
	api := func(name string) string {
		return "apiVersion: backstage.io/v1alpha1\nkind: API\nmetadata:\n  name: " + name +
			"\nspec:\n  type: grpc\n  lifecycle: production\n  owner: team-a\n  definition: some-definition\n"
	}
	name63 := strings.Repeat("a", 63)
	name64 := strings.Repeat("a", 64)

	tests := []struct {
		name        string
		content     string
		errContains []string
	}{
		{name: "63 characters passes", content: component(name63)},
		{name: "64 characters fails", content: component(name64), errContains: []string{
			"Component", `"` + name64 + `"`, "64 characters", "maximum of 63",
		}},
		{name: "long API name fails with kind", content: api(strings.Repeat("b", 118)), errContains: []string{
			"API", "118 characters",
		}},
		{name: "multi-document file passes when all within limit", content: component("short") + "---\n" + api(name63)},
		{name: "multi-document file fails on later document", content: component("short") + "---\n" + api(name64), errContains: []string{
			"doc 2", "API", "64 characters",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "catalog-info.yaml"), []byte(tt.content), 0o600))
			t.Chdir(dir)

			_, err := parseCatalogInfoFiles()
			if len(tt.errContains) == 0 {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			for _, s := range tt.errContains {
				assert.Contains(t, err.Error(), s)
			}
		})
	}
}
