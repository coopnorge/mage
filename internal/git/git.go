package git

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/sh"
)

// RepoURL returns the remote URL to the git repository
func RepoURL() (string, error) {
	remote, err := sh.Output("git", "remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return NormalizeGitURL(remote)
}

// SHA256 returns the hash of the current commit
func SHA256() (string, error) {
	return sh.Output("git", "rev-parse", "HEAD")
}

// LatestTag returns the
func LatestTag() (string, error) {
	return sh.Output("git", "describe", "--tags", "--abbrev=0")
}

// NormalizeGitURL parses git or https git URLs and returns an https URL.
func NormalizeGitURL(url string) (string, error) {
	if strings.HasPrefix(url, "https://") {
		return strings.TrimSuffix(url, ".git"), nil
	} else if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		parts := strings.Split(url, ":")
		url = fmt.Sprintf("https://%s/%s", parts[0], parts[1])
		return strings.TrimSuffix(url, ".git"), nil
	}
	return "", fmt.Errorf("unable to parse remote url: %s", url)
}
