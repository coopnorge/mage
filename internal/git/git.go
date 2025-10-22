package git

import (
	"fmt"
	"os"
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

// DiffToMain returns a list of files that have been changed
// compared to the main branch. Files have to be staged or committed.
func DiffToMain() ([]string, error) {
	// git diff
	// --name-only # only list file names
	// --no-renames # rename of file is shown as delete and add

	changedFiles := []string{}
	changedFilesFromEnv, ok := os.LookupEnv("CHANGED_FILES")
	if ok {
		changedFiles = strings.Split(changedFilesFromEnv, ",")
		return changedFiles, nil
	}
	// We assume the default branch is main, preferred origin/main. We should
	// add the ability for adding a specific branch as well.
	diffBranch := "origin/main"
	if checkBranch(diffBranch) != nil {
		diffBranch = "main"
	}
	if checkBranch(diffBranch) != nil {
		return changedFiles, fmt.Errorf("unable to find branch %s", diffBranch)
	}
	gitDiff, err := sh.Output("git", "diff", "--name-only", "--no-renames", diffBranch)
	if err != nil {
		return []string{}, err
	}
	changedFiles = append(changedFiles, strings.Split(gitDiff, "\n")...)
	return changedFiles, nil
}

func checkBranch(branch string) error {
	return sh.Run("git", "rev-parse", "--verify", branch)
}
