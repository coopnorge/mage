package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
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

// FindCommentInPR searches the current PR for a string in a comment.
// It will return true if found and the comment ID. If muiltiple comments are
// found it will return the most recent. If no comment found it will return
// false
func FindCommentInPR(searchString string) (bool, string, error) {
	args := []string{"pr", "view"}
	if prNumber, found := os.LookupEnv("PR_NUMBER"); found {
		args = append(args, prNumber)
	}
	args = append(args, "--json", "comments")
	// jq := fmt.Sprintf(".comments[] | select(.body | contains(\\\"%s\\\")) | .id\, searchString)
	out, err := sh.Output("gh", args...)
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

	args := []string{"pr", "comment"}
	if prNumber, found := os.LookupEnv("PR_NUMBER"); found {
		args = append(args, prNumber)
	}
	args = append(args, "--body-file", filename)

	return sh.Run("gh", args...)
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
