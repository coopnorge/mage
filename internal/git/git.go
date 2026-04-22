package git

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/github"
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
func NormalizeGitURL(rawURL string) (string, error) {
	// "https://<redacted>:x-oauth-basic@github.com/coopnorge/helloworld"

	if strings.HasPrefix(rawURL, "https://") {
		rawURL = strings.TrimSuffix(rawURL, ".git")
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return "", fmt.Errorf("unable to parse remote https url: %w", err)
		}
		parsedURL.User = nil // This removes any user info (username:password) from URL, if present
		return parsedURL.Redacted(), nil
	} else if strings.HasPrefix(rawURL, "git@") {
		rawURL = strings.TrimPrefix(rawURL, "git@")
		rawURL = strings.TrimSuffix(rawURL, ".git")
		parts := strings.Split(rawURL, ":")
		rawURL = fmt.Sprintf("https://%s/%s", parts[0], parts[1])
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return "", fmt.Errorf("unable to parse remote https url: %w", err)
		}
		parsedURL.User = nil // This removes any user info (username:password) from URL, if present
		return parsedURL.Redacted(), nil
	}
	return "", fmt.Errorf("unable to parse remote url: %s", rawURL)
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

// DiffToTagPattern returns a list of files that have been changed
// compared to the most recent tags of a certain pattern.
func DiffToTagPattern(releasePrefix string) ([]string, error) {
	// git diff
	// --name-only # only list file names
	// --no-renames # rename of file is shown as delete and add

	changedFiles := []string{}

	changedFilesFromEnv, ok := os.LookupEnv("CHANGED_FILES")
	if ok {
		changedFiles = strings.Split(changedFilesFromEnv, ",")
		return changedFiles, nil
	}

	ref := "origin/main"
	onMain, err := onMainBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to check if commit is on main branch: %w", err)
	}

	if onMain {
		releaseRef, createdAt, err := github.GetLatestReleaseTagWithPrefix(releasePrefix)
		if err != nil {
			return nil, fmt.Errorf("getting releases from github failed: %w", err)
		}
		// if no relelease is found, use CHANGES from dorny path filter. next release should
		// create release.
		// TODO: implement native model to do changes against github api
		if releaseRef == "" {
			changedFilesFromEnv, ok := os.LookupEnv("CHANGES")
			if !ok {
				return nil, fmt.Errorf("the environment varariable $CHANGES is required but not found. This is required to detect changes on main")
			}
			changedFiles = strings.Split(changedFilesFromEnv, ",")
			return changedFiles, nil
		}
		ref = releaseRef
		currentCommit, err := getTimeStampOfCurrentCommit()
		if err != nil {
			return nil, fmt.Errorf("failed to get timetamp of current commit: %w ", err)
		}
		if currentCommit.Before(createdAt) || currentCommit.Equal(createdAt) {
			return nil, fmt.Errorf("current commit creation date (%s) is created before or is equal the most recent release %s (%s)", currentCommit.String(), ref, createdAt.String())
		}
	}

	gitDiff, err := sh.Output("git", "diff", "--name-only", "--no-renames", ref)
	if err != nil {
		return changedFiles, err
	}
	changedFiles = append(changedFiles, strings.Split(gitDiff, "\n")...)
	return changedFiles, nil
}

// OnMainBranch returns true the current branch is the main branch
func onMainBranch() (bool, error) {
	branch, err := sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return false, err
	}
	return branch == "main", nil
}

func checkBranch(branch string) error {
	return sh.Run("git", "rev-parse", "--verify", branch)
}

// IsTracked returns true if the file is tracked by git
func IsTracked(path string) bool {
	return sh.Run("git", "ls-files", "--error-unmatch", path) == nil
}

// CurrentBranch returns the current branch
func CurrentBranch() (string, error) {
	return sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")
}

func getTimeStampOfCurrentCommit() (time.Time, error) {
	out, err := sh.Output("git", "show", "--no-patch", `--format=%cI`)
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("getting timestamp of commit using git failed: %w", err)
	}
	timestamp, err := time.Parse(time.RFC3339, out)
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("failed to parse timestamp %s: %w", out, err)
	}
	return timestamp, nil
}

// Worktree creates a new worktree for the given branch.
// It returns the absolute path to the worktree and an error if the operation fails.
func Worktree(branch string) (string, func(), error) {
	// Define target location (e.g., in a 'worktrees' directory outside the current repo).
	// Placing worktrees outside prevents recursive issues with tools scanning the main repo.
	targetDir, cleanupDir, err := core.MkdirTemp()
	if err != nil {
		return targetDir, cleanupDir, err
	}
	// Execute 'git worktree add <path> <branch>'
	err = sh.Run("git", "worktree", "add", targetDir, branch)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create worktree for branch %s: %w", branch, err)
	}

	// We use git worktree remove which cleans up the admin files and the directory.
	cleanup := func() {
		err = sh.Run("git", "worktree", "remove", targetDir)
		if err != nil {
			fmt.Printf("Failed to delete %s, error %s", targetDir, err)
		}
		cleanupDir()
	}

	return targetDir, cleanup, nil
}
