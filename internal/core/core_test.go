package core_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/coopnorge/mage/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestGetRelativeRootPath(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		rootPath   string
		workDirRel string
		want       string
		wantErr    bool
	}{
		{
			name:       "child",
			rootPath:   "/temp",
			workDirRel: "src",
			want:       "..",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := core.GetRelativeRootPath(tt.rootPath, tt.workDirRel)
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWriteTempFile(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		directory string
		suffix    string
		content   string
		wantErr   bool
	}{
		{
			name:      "base case",
			directory: ".",
			suffix:    "example.txt",
			content:   "Hello, World",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, cleanup, gotErr := core.WriteTempFile(tt.directory, tt.suffix, tt.content)
			assert.NoError(t, gotErr)
			assert.Contains(t, got, tt.suffix)
			assert.FileExists(t, got)
			gotBytes, gotErr := os.ReadFile(got)
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.content, string(gotBytes))
			cleanup()
		})
	}
}

func TestMkdirTemp(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		wantErr bool
	}{
		{
			name:    "base case",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, cleanup, gotErr := core.MkdirTemp()
			assert.NoError(t, gotErr)
			assert.DirExists(t, got)
			//assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("%s/.+", os.TempDir())), got)
			assert.Regexp(t, regexp.MustCompile(filepath.Join(os.TempDir(), "/.+")), got)
			cleanup()
		})
	}
}

func TestCompareChangesToPaths(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		changes         []string
		paths           []string
		additionalGlobs []string
		want            bool
		wantErr         bool
	}{
		{
			name:            "simple match",
			changes:         []string{"a/b.txt"},
			paths:           []string{"a"},
			additionalGlobs: []string{},
			want:            true,
			wantErr:         false,
		},
		{
			name:            "no match",
			changes:         []string{"a/b.txt"},
			paths:           []string{"b/c/a.yaml"},
			additionalGlobs: []string{""},
			want:            false,
			wantErr:         false,
		},
		{
			name:            "match on additional glob",
			changes:         []string{"a/b.txt"},
			paths:           []string{"b/c/a.yaml"},
			additionalGlobs: []string{"**/*.txt"},
			want:            true,
			wantErr:         false,
		},
		{
			name:            "multiple additionalGlobs",
			changes:         []string{"a/b.txt"},
			paths:           []string{"b/c"},
			additionalGlobs: []string{"**/*.yaml", "**/*.txt"},
			want:            true,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := core.CompareChangesToPaths(tt.changes, tt.paths, tt.additionalGlobs)
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
