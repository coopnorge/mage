package git

import "github.com/magefile/mage/sh"

// RepoURL returns the remote URL to the git repository
func RepoURL() (string, error) {
	return sh.Output("git", "remote", "get-url", "origin")
}

// SHA256 returns the hash of the current commit
func SHA256() (string, error) {
	return sh.Output("git", "rev-parse", "HEAD")
}
