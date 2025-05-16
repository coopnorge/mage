package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindTerraformFolders(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		want    []string
		wantErr bool
	}{
		{
			name:    "Should find all relevant folders",
			workdir: "testdata/folders",
			want:    []string{"a", "b"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(tt.workdir)
			got, gotErr := FindTerraformProjects(".")
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
