package core_test

import (
	"os"
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
