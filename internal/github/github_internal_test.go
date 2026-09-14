package github

import (
	"testing"

	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRepoInfoFromEnvironment(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "coopnorge/mage")
	t.Setenv("GITHUB_API_URL", "https://github.example/api/v3/")

	got, err := getRepoInfo()

	require.NoError(t, err)
	assert.Equal(t, ghRepo{Owner: "coopnorge", Repo: "mage", APIURL: "https://github.example/api/v3"}, got)
}

func TestGetRepoInfoFromGitRemote(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "")
	t.Setenv("GITHUB_API_URL", "")
	t.Chdir(t.TempDir())

	runGit(t, "init")
	runGit(t, "remote", "add", "origin", "git@github.com:coopnorge/mage.git")

	got, err := getRepoInfo()

	require.NoError(t, err)
	assert.Equal(t, ghRepo{Owner: "coopnorge", Repo: "mage", APIURL: "https://api.github.com"}, got)
}

func TestGetRepoInfoWithoutRepository(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "")
	t.Chdir(t.TempDir())

	_, err := getRepoInfo()

	assert.Error(t, err)
}

func TestRepositoryFromRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		remote  string
		owner   string
		repo    string
		wantErr bool
	}{
		{
			name:   "ssh",
			remote: "git@github.com:coopnorge/mage.git",
			owner:  "coopnorge",
			repo:   "mage",
		},
		{
			name:   "https",
			remote: "https://github.com/coopnorge/mage.git",
			owner:  "coopnorge",
			repo:   "mage",
		},
		{
			name:   "authenticated https",
			remote: "https://user:token@github.com/coopnorge/mage.git",
			owner:  "coopnorge",
			repo:   "mage",
		},
		{
			name:    "invalid",
			remote:  "not-a-repository",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := repositoryFromRemoteURL(tt.remote)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.owner, owner)
			assert.Equal(t, tt.repo, repo)
		})
	}
}

func runGit(t *testing.T, args ...string) {
	t.Helper()
	require.NoError(t, sh.Run("git", args...))
}
