package github_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coopnorge/mage/internal/github"
	"github.com/stretchr/testify/assert"
)

func TestGetLatestReleaseTagWithPrefix(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		prefix     string
		httpStatus int
		want       string
		wantErr    bool
	}{
		{
			name: "Working example",
			payload: `[
		       {"name": "v1.2.0", "tag_name": "v1.2.0", "draft": false, "prerelease": false},
		       {"name": "test-1.0.0", "tag_name": "test-1.0.0", "draft": false, "prerelease": false}
	        ]`,
			prefix:     "v",
			httpStatus: 200,
			want:       "v1.2.0",
			wantErr:    false,
		},
		{
			name: "No version",
			payload: `[
		       {"name": "1.2.0", "tag_name": "v1.2.0", "draft": false, "prerelease": false},
		       {"name": "test-1.0.0", "tag_name": "test-1.0.0", "draft": false, "prerelease": false}
	        ]`,
			prefix:     "v",
			httpStatus: 200,
			want:       "",
			wantErr:    false,
		},
		{
			name: "Failed payload",
			payload: `[
		       {"name": "1.2.0", "tag_name": "v1.2.0", "draft": false, "prerelease": false},
		       test-1.0.0", "tag_name": "test-1.0.0", "draft": false, "prerelease": false}
	        ]`,
			prefix:     "v",
			httpStatus: 200,
			want:       "",
			wantErr:    true,
		},
		{
			name:       "Failed request",
			payload:    "503 Fail",
			prefix:     "v",
			httpStatus: 500,
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tt.httpStatus)
			_, err := w.Write([]byte(tt.payload))
			_ = err
		}))
		defer server.Close()

		t.Setenv("GITHUB_TOKEN", "test")
		t.Setenv("GITHUB_REPOSITORY", "test/test")
		t.Setenv("GITHUB_API_URL", server.URL)

		tag, err := github.GetLatestReleaseTagWithPrefix(
			tt.prefix,
			github.WithHTTPClient(server.Client()),
		)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, tt.want, tag)
	}
}
