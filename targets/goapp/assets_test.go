package goapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssets(t *testing.T) {
	tests := []struct {
		name string
		// files are created under the assets root, a trailing slash
		// creating an empty directory.
		files  []string
		root   string
		apps   []string
		unset  bool
		want   map[string]string
		errMsg string
	}{
		{
			name:  "unset",
			unset: true,
			apps:  []string{"app1"},
			want:  map[string]string{},
		},
		{
			name:   "root missing",
			root:   "does-not-exist",
			apps:   []string{"app1"},
			errMsg: "GOAPP_ASSETS: stat",
		},
		{
			name:   "root is a file",
			files:  []string{"file"},
			root:   "file",
			apps:   []string{"app1"},
			errMsg: "GOAPP_ASSETS: <root>/file is not a directory",
		},
		{
			name:   "app directory missing",
			files:  []string{"app1/index.html"},
			apps:   []string{"app1", "app2"},
			errMsg: "GOAPP_ASSETS: <root> has no directory for app2; every app needs one, empty for an app without assets",
		},
		{
			name:   "several app directories missing",
			files:  []string{"app2/"},
			apps:   []string{"app1", "app2", "app3"},
			errMsg: "GOAPP_ASSETS: <root> has no directory for app1, app3; every app needs one, empty for an app without assets",
		},
		{
			name:   "app directory is a file",
			files:  []string{"app1"},
			apps:   []string{"app1"},
			errMsg: "GOAPP_ASSETS: <root>/app1 is not a directory",
		},
		{
			name:  "empty app directory",
			files: []string{"app1/"},
			apps:  []string{"app1"},
			want:  map[string]string{},
		},
		{
			name:  "app directories with and without content",
			files: []string{"app1/index.html", "app2/"},
			apps:  []string{"app1", "app2"},
			want:  map[string]string{"app1": "<root>/app1"},
		},
		{
			name:  "nested app",
			files: []string{"apps/app1/web/index.html"},
			apps:  []string{"apps/app1"},
			want:  map[string]string{"apps/app1": "<root>/apps/app1"},
		},
		{
			name:  "directory without an app is ignored",
			files: []string{"app1/index.html", "unknown/index.html"},
			apps:  []string{"app1"},
			want:  map[string]string{"app1": "<root>/app1"},
		},
		{
			name:   "app in the repository root",
			files:  []string{"index.html"},
			apps:   []string{"."},
			errMsg: "GOAPP_ASSETS: assets are not supported for an app in the repository root",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			for _, file := range test.files {
				path := filepath.Join(root, file)
				if file[len(file)-1] == '/' {
					require.NoError(t, os.MkdirAll(path, 0o755))
					continue
				}
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte("content"), 0o644))
			}

			if test.unset {
				t.Setenv(AssetsEnv, "")
			} else {
				t.Setenv(AssetsEnv, filepath.Join(root, test.root))
			}

			got, err := findAssets(test.apps)

			if test.errMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), replaceRoot(test.errMsg, root))
				return
			}
			require.NoError(t, err)
			want := map[string]string{}
			for app, directory := range test.want {
				want[app] = replaceRoot(directory, root)
			}
			assert.Equal(t, want, got)
		})
	}
}

func replaceRoot(text, root string) string {
	return strings.ReplaceAll(text, "<root>", root)
}
