package git_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/git"
	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"
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
	dir, cleanup, _ := core.MkdirTemp()
	t.Chdir(dir)
	t.Cleanup(cleanup)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, command := range tt.commands {
				cmd := strings.Fields(command)
				assert.NoError(t, sh.RunWith(env, cmd[0], cmd[1:]...))
			}
			if tt.changedFilesEnv != "" {
				os.Setenv("CHANGED_FILES", tt.changedFilesEnv)
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
