package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/magefile/mage/sh"
)

type ghIssueComment struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

type ghIssueComments struct {
	Comments []ghIssueComment `json:"comments"`
}

const prNumberEnvVar = "PR_NUMBER"

// FindCommentInPR searches the current PR for a string in a comment.
// It will return true if found and the comment ID. If muiltiple comments are
// found it will return the most recent. If no comment found it will return
// false
func FindCommentInPR(searchString string) (bool, string, error) {
	prNumber, found := os.LookupEnv(prNumberEnvVar)
	if !found {
		return false, "", fmt.Errorf("the environment variable %s is required but not found", prNumberEnvVar)
	}
	// jq := fmt.Sprintf(".comments[] | select(.body | contains(\\\"%s\\\")) | .id\, searchString)
	out, err := sh.Output("gh", "pr", "view", prNumber, "--json", "comments")
	if err != nil {
		return false, "", err
	}
	var comments ghIssueComments
	err = json.Unmarshal([]byte(out), &comments)
	if err != nil {
		return false, "", err
	}
	for _, comment := range slices.Backward(comments.Comments) {
		if strings.Contains(comment.Body, searchString) {
			return true, comment.ID, nil
		}
	}
	// nothing found
	return false, "", nil
}

// HideComment hides a comment
func HideComment(id string) error {
	// gh api graphql -F id='COMMENT_NODE_ID' -f query='
	// mutation($id: ID!) { minimizeComment(input: {subjectId: $id, classifier: OUTDATED}) {minimizedComment {isMinimized}}}'

	query := "query=mutation($id:ID!){ minimizeComment(input:{subjectId:$id,classifier:OUTDATED}){minimizedComment{isMinimized}}}"
	idArg := fmt.Sprintf("id=%s", id)

	// not using mage sh library because it will remove $
	// https://github.com/magefile/mage/pull/505
	// return sh.Run("gh", "api", "graphql", "-F", idArg, "-f", query)
	cmd := exec.Command("gh", "api", "graphql", "-F", idArg, "-f", query)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("failed to run command. Error %s\n", stderr.String())
		return nil
	}
	return nil
}

// ReplaceCommentInPR replaces a comment with the id id and the body sources from
// the supplied filename. It
// will return an error if the body is to big or the command fails
func ReplaceCommentInPR(id string, filename string) error {
	// gh api -X PATCH repos/{owner}/{repo}/issues/comments/{comment_id} -f body=@path/to/your/comment.md

	err := validateCommentBody(filename)
	if err != nil {
		return err
	}
	pathArg := fmt.Sprintf("repos/{owner}/{repo}/issues/comments/%s", id)
	bodyArg := fmt.Sprintf("body=@%s", filename)
	return sh.Run("gh", "api", "-X", "PATCH", pathArg, "-f", bodyArg)
}

// CreateCommentInPR creates a comment with the id id and the body sources from
// the supplied filename. It
// will return an error if the body is to big or the command fails
func CreateCommentInPR(filename string) error {
	err := validateCommentBody(filename)
	if err != nil {
		return err
	}
	prNumber, found := os.LookupEnv(prNumberEnvVar)
	if !found {
		return fmt.Errorf("the environment variable %s is required but not found", prNumberEnvVar)
	}

	return sh.Run("gh", "pr", "comment", prNumber, "--body-file", filename)
}

// PrintActionMessage prints a action message in github action using the
// ::<level> format. It makes sure the encoding is correct. The first input the level, the
// second is the is the title and the third the message
// level can be debug, notice, warning, error. It will return a error if the
// level is not allowed.
func PrintActionMessage(level, title, message string) error {
	allowedLevels := []string{"debug", "notice", "warning", "error"}
	if !slices.Contains(allowedLevels, level) {
		return fmt.Errorf("supplied level %s is not in the list %s", level, strings.Join(allowedLevels, ","))
	}
	fmt.Printf("::%s title=%s::%s", level, gitHubActionsEscape(title), gitHubActionsEscape(message))
	return nil
}

func gitHubActionsEscape(s string) string {
	r := strings.NewReplacer(
		"%", "%25",
		"\n", "%0A",
		"\r", "%0D",
	)
	return r.Replace(s)
}

// StartLogGroup starts a log group if running in github actions
func StartLogGroup(name string) {
	if InCI() {
		fmt.Printf("::group::%s\n", gitHubActionsEscape(name))
	}
}

// EndLogGroup ends a log group if running in github actions
func EndLogGroup() {
	if InCI() {
		fmt.Println("::endgroup::")
	}
}

// InCI returns a true if you are running in Github Actions
func InCI() bool {
	_, found := os.LookupEnv("CI")
	return found
}

func validateCommentBody(filename string) error {
	body, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if utf8.RuneCountInString(string(body)) > 65536 {
		return fmt.Errorf("body is %d characters which is more than the max of 65536", utf8.RuneCountInString(string(body)))
	}
	return nil
}

type options struct {
	httpClient *http.Client
	token      string
	baseURL    string
	owner      string
	repo       string
}

// Option are options for the github http client
type Option func(*options)

func defaultOptions() (*options, error) {
	opts := &options{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// token fallback
	if t, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		opts.token = t
	} else {
		return nil, fmt.Errorf("missing GITHUB_TOKEN")
	}

	// repo fallback
	repo, err := getRepoInfo()
	if err != nil && InCI() {
		return nil, fmt.Errorf("failed to get repo info: %w", err)
	}
	opts.owner = repo.Owner
	opts.repo = repo.Repo
	opts.baseURL = repo.APIURL

	return opts, nil
}

// WithHTTPClient overrides the http client for github api requests. Mainly useful
// when with testing
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) {
		o.httpClient = c
	}
}

// GetLatestReleaseTagWithPrefix gets the latest release filtred by a prefix
// of the release name (not the tag name). It returns the tag and a error. If
// no release is found the tag will be an empty string.
func GetLatestReleaseTagWithPrefix(prefix string, opts ...Option) (string, error) {
	o, err := defaultOptions()
	if err != nil {
		return "", err
	}

	for _, opt := range opts {
		opt(o)
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases", o.baseURL, o.owner, o.repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+o.token)
	if err != nil {
		return "", err
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call GitHub API: %w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("Failed to close body, ignoring: %s\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got status %d, expected is %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var releases []ghRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", fmt.Errorf("failed to parse: %s\nerr: %w", string(body), err)
	}

	for _, r := range releases {
		if strings.HasPrefix(r.Name, prefix) {
			return r.TagName, nil
		}
	}
	return "", nil
}

type ghRelease struct {
	Name       string `json:"name"`
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// ghRepo stores information about the repo
// when running in github actions CI
type ghRepo struct {
	Owner  string
	Repo   string
	APIURL string
}

func getRepoInfo() (ghRepo, error) {
	info := ghRepo{}
	val, found := os.LookupEnv("GITHUB_REPOSITORY")
	if !found {
		return info, fmt.Errorf("environment variable GITHUB_REPOSITORY not found, unable to determine repository info")
	}
	url, found := os.LookupEnv("GITHUB_API_URL")
	if !found {
		return info, fmt.Errorf("environment variable GITHUB_API_URL not found, unable to determine api url")
	}
	info.APIURL = url
	info.Owner = strings.Split(val, "/")[0]
	info.Repo = strings.Split(val, "/")[1]
	return info, nil
}
