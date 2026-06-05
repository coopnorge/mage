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
			name: "it should work",
			payload: `[
		       {"name": "v1.2.0", "tag_name": "v1.2.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"},
		       {"name": "test-1.0.0", "tag_name": "test-1.0.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"}
	        ]`,
			prefix:     "v",
			httpStatus: 200,
			want:       "v1.2.0",
			wantErr:    false,
		},
		{
			name: "ordering should work",
			payload: `[
		       {"name": "v1.2.0", "tag_name": "v1.2.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"},
		       {"name": "v1.3.0", "tag_name": "v1.3.0", "draft": false, "prerelease": false, "created_at": "2014-02-27T19:35:32Z"},
		       {"name": "test-1.0.0", "tag_name": "test-1.0.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"}
	        ]`,
			prefix:     "v",
			httpStatus: 200,
			want:       "v1.3.0",
			wantErr:    false,
		},

		{
			name: "Skip draft",
			payload: `[
		       {"name": "v1.3.0", "tag_name": "v1.2.0", "draft": true, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"},
		       {"name": "v1.2.0", "tag_name": "v1.2.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"},
		       {"name": "test-1.0.0", "tag_name": "test-1.0.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"}
	        ]`,
			prefix:     "v",
			httpStatus: 200,
			want:       "v1.2.0",
			wantErr:    false,
		},
		{
			name: "Skip prerelease",
			payload: `[
		       {"name": "v1.3.0", "tag_name": "v1.2.0", "draft": false, "prerelease": true, "created_at": "2013-02-27T19:35:32Z"},
		       {"name": "v1.2.0", "tag_name": "v1.2.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"},
		       {"name": "test-1.0.0", "tag_name": "test-1.0.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"}
	        ]`,
			prefix:     "v",
			httpStatus: 200,
			want:       "v1.2.0",
			wantErr:    false,
		},
		{
			name: "No version",
			payload: `[
		       {"name": "v1.2.0", "tag_name": "v1.2.0", "draft": true, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"},
		       {"name": "1.2.0", "tag_name": "v1.2.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"},
		       {"name": "test-1.0.0", "tag_name": "test-1.0.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"}
	        ]`,
			prefix:     "v",
			httpStatus: 200,
			want:       "",
			wantErr:    false,
		},
		{
			name: "Failed payload",
			payload: `[
		       {"name": "1.2.0", "tag_name": "v1.2.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"},
		       test-1.0.0", "tag_name": "test-1.0.0", "draft": false, "prerelease": false, "created_at": "2013-02-27T19:35:32Z"}
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
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.httpStatus)
				_, err := w.Write([]byte(tt.payload))
				assert.NoError(t, err)
			}))
			t.Cleanup(func() {
				server.Close()
			})

			t.Setenv("GITHUB_TOKEN", "test")
			t.Setenv("GITHUB_REPOSITORY", "test/test")
			t.Setenv("GITHUB_API_URL", server.URL)

			tag, _, err := github.GetLatestReleaseTagWithPrefix(
				tt.prefix,
				github.WithHTTPClient(server.Client()),
			)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, tag)
		})
	}
}
