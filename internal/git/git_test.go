package git_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/git"
	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "git url",
			url:  "git@github.com:coopnorge/mage.git",
			want: "https://github.com/coopnorge/mage",
		},
		{
			name: "https url",
			url:  "https://github.com/coopnorge/mage.git",
			want: "https://github.com/coopnorge/mage",
		},
		{
			name: "https url with basic auth",
			url:  "https://gpg_dfdfl:x-basic-auth@github.com/coopnorge/mage.git",
			want: "https://github.com/coopnorge/mage",
		},
		{
			name: "https url with basic auth, without git suffix",
			url:  "https://gpg_dfdfl:x-basic-auth@github.com/coopnorge/mage",
			want: "https://github.com/coopnorge/mage",
		},
		{
			name: "https url with user",
			url:  "https://myuser@github.com/coopnorge/mage.git",
			want: "https://github.com/coopnorge/mage",
		},
		{
			name: "https url with user, without git suffix",
			url:  "https://myuser@github.com/coopnorge/mage",
			want: "https://github.com/coopnorge/mage",
		},
		{
			name:    "error url",
			url:     "github.com/coopnorge",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := git.NormalizeGitURL(tt.url)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGitDiff(t *testing.T) {
	tests := []struct {
		name            string
		commands        []string
		want            []string
		wantErr         bool
		changedFilesEnv string
	}{
		{
			name:     "no git directory",
			commands: []string{},
			want:     []string{},
			wantErr:  true,
		},
		{
			name: "no diff",
			commands: []string{
				"git init",
				`git config user.email "mage@coop.no"`,
				`git config user.name "Mage CI"`,
				"git checkout -b main",
				"cp $TD/1.txt ./1.txt",
				"git add 1.txt",
				`git commit -am "init"`,
				"git checkout -b diff-one",
			},
			want:    []string{""},
			wantErr: false,
		},
		{
			name: "add change with commit",
			commands: []string{
				"cp $TD/2.txt ./1.txt",
			},
			want:    []string{"1.txt"},
			wantErr: false,
		},
		{
			name: "commit the change",
			commands: []string{
				`git commit -am "change" `,
			},
			want:    []string{"1.txt"},
			wantErr: false,
		},
		{
			name: "add file but dont stage",
			commands: []string{
				"cp $TD/3.txt ./3.txt",
			},
			want:    []string{"1.txt"},
			wantErr: false,
		},
		{
			name: "stage file",
			commands: []string{
				"git add 3.txt",
			},
			want:    []string{"1.txt", "3.txt"},
			wantErr: false,
		},
		{
			name: "rename file",
			commands: []string{
				`git commit -am "add3"`,
				"git checkout main",
				"git merge diff-one",
				"git checkout -b diff-two",
				"mv 1.txt 4.txt",
			},
			want:    []string{"1.txt"},
			wantErr: false,
		},
		{
			name: "stage new after rename",
			commands: []string{
				"git add 4.txt",
			},
			want:    []string{"1.txt", "4.txt"},
			wantErr: false,
		},
		{
			name: "delete file",
			commands: []string{
				"rm 3.txt",
			},
			want:    []string{"1.txt", "3.txt", "4.txt"},
			wantErr: false,
		},
		{
			name:            "env override",
			commands:        []string{},
			want:            []string{"i", "am", `"over riden"`},
			changedFilesEnv: `i,am,"over riden"`,
			wantErr:         false,
		},
	}

	wd, err := os.Getwd()
	env := map[string]string{"TD": filepath.Join(wd, "testdata/files")}
	assert.NoError(t, err)
	dir, cleanup, err := core.MkdirTemp()
	require.NoError(t, err)
	t.Chdir(dir)
	t.Cleanup(cleanup)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, command := range tt.commands {
				cmd := strings.Fields(command)
				assert.NoError(t, sh.RunWith(env, cmd[0], cmd[1:]...))
			}
			if tt.changedFilesEnv != "" {
				err = os.Setenv("CHANGED_FILES", tt.changedFilesEnv)
				require.NoError(t, err)
			}
			got, gotErr := git.DiffToMain()
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiffToTagPatternComparesCommitTimestamps(t *testing.T) {
	tests := []struct {
		name      string
		currentAt string
		want      []string
		wantErr   string
	}{
		{
			name:      "current commit is newer than release",
			currentAt: "2026-10-19T10:00:00Z",
			want:      []string{"after-release.txt"},
		},
		{
			name:      "current commit has same timestamp as release",
			currentAt: "2026-10-19T09:00:00Z",
			wantErr:   "is created before or is equal the most recent release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsetEnv(t, "CHANGED_FILES")
			unsetEnv(t, "CHANGES")
			t.Setenv("GITHUB_TOKEN", "test")
			t.Setenv("GITHUB_REPOSITORY", "coopnorge/mage")
			t.Setenv("CI", "true")

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, err := w.Write([]byte(`[{"name":"v1.0.0","tag_name":"v1.0.0","draft":false,"prerelease":false}]`))
				require.NoError(t, err)
			}))
			t.Cleanup(server.Close)
			t.Setenv("GITHUB_API_URL", server.URL)

			t.Chdir(t.TempDir())
			runGitTest(t, "init", "-b", "main")
			runGitTest(t, "config", "user.email", "mage@coop.no")
			runGitTest(t, "config", "user.name", "Mage CI")
			runGitTest(t, "config", "commit.gpgsign", "false")
			require.NoError(t, os.WriteFile("before-release.txt", []byte("before"), 0o600))
			runGitTest(t, "add", "before-release.txt")
			require.NoError(t, sh.RunWith(map[string]string{
				"GIT_AUTHOR_DATE":    "2026-10-19T09:00:00Z",
				"GIT_COMMITTER_DATE": "2026-10-19T09:00:00Z",
			}, "git", "commit", "-m", "release"))
			runGitTest(t, "tag", "v1.0.0")
			require.NoError(t, os.WriteFile("after-release.txt", []byte("after"), 0o600))
			runGitTest(t, "add", "after-release.txt")
			require.NoError(t, sh.RunWith(map[string]string{
				"GIT_AUTHOR_DATE":    tt.currentAt,
				"GIT_COMMITTER_DATE": tt.currentAt,
			}, "git", "commit", "-m", "after release"))

			got, err := git.DiffToTagPattern("v")
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	value, wasSet := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if wasSet {
			assert.NoError(t, os.Setenv(key, value))
		} else {
			assert.NoError(t, os.Unsetenv(key))
		}
	})
}

func runGitTest(t *testing.T, args ...string) {
	t.Helper()
	require.NoError(t, sh.Run("git", args...))
}
