package cataloginfo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coopnorge/mage/targets/cataloginfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mg.CtxDeps runs each target once per process and caches the result, so only
// one scenario per target can be exercised in-process. Failure paths are
// covered by the subprocess tests in internal/targets/cataloginfo.
func TestValidate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp, err := json.Marshal([]map[string]string{{"slug": "engineering"}})
		require.NoError(t, err)
		_, err = w.Write(resp)
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)
	t.Setenv("GITHUB_TOKEN", "test-token")
	t.Setenv("GITHUB_REPOSITORY", "coopnorge/test")
	t.Setenv("GITHUB_API_URL", server.URL)

	t.Run("success", func(t *testing.T) {
		t.Chdir("../../internal/targets/cataloginfo/testdata/success")
		assert.NoError(t, cataloginfo.CatalogInfo{}.Validate(context.Background()))
	})
}

func TestChanges(t *testing.T) {
	t.Run("changes", func(t *testing.T) {
		t.Setenv("CHANGED_FILES", "catalog-info.yaml")
		assert.NoError(t, cataloginfo.CatalogInfo{}.Changes(context.Background()))
	})
}
