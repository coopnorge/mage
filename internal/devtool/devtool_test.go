package devtool_test

import (
	"testing"

	"github.com/coopnorge/mage/internal/devtool"
	"github.com/stretchr/testify/assert"
)

func TestGetImageName(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		target  string
		want    string
		wantErr bool
	}{
		{
			name:    "base case",
			target:  "golang",
			want:    "ocreg.invalid/coopnorge/devtool/golang-devtool:latest",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := devtool.GetImageName(tt.target)
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
