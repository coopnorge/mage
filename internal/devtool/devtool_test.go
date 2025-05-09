package devtool

import (
	"fmt"
	"runtime"
	"testing"

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
			got, gotErr := GetImageName(tt.target)
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDevtoolARCHToolSelector(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		target     string
		dockerfile string
		want       string
		wantErr    bool
	}{
		{
			name:       "universal",
			dockerfile: `FROM docker.io/library/golang:1.24.2@sha256:1ecc479bc712a6bdb56df3e346e33edcc141f469f82840bab9f4bc2bc41bf91d AS golang`,
			target:     "golang",
			want:       "golang",
			wantErr:    false,
		},
		{
			name: "platform-specifc",
			dockerfile: `FROM docker.io/library/golang:1.24.2@sha256:1ecc479bc712a6bdb56df3e346e33edcc141f469f82840bab9f4bc2bc41bf91d AS golang
			  FROM docker.io/library/golang:1.24.2@sha256:1ecc479bc712a6bdb56df3e346e33edcc141f469f82840bab9f4bc2bc41bf91d AS golang-arm64
			  FROM docker.io/library/golang:1.24.2@sha256:1ecc479bc712a6bdb56df3e346e33edcc141f469f82840bab9f4bc2bc41bf91d AS golang-amd64`,
			target:  "golang",
			want:    fmt.Sprintf("golang-%s", runtime.GOARCH),
			wantErr: false,
		},
		{
			name: "fallback to universal",
			dockerfile: `FROM docker.io/library/golang:1.24.2@sha256:1ecc479bc712a6bdb56df3e346e33edcc141f469f82840bab9f4bc2bc41bf91d AS golang
			  FROM docker.io/library/golang:1.24.2@sha256:1ecc479bc712a6bdb56df3e346e33edcc141f469f82840bab9f4bc2bc41bf91d AS golang-my-hobby-arch`,
			target:  "golang",
			want:    "golang",
			wantErr: false,
		},
		{
			name:       "error on not found",
			dockerfile: `FROM docker.io/library/golang:1.24.2@sha256:1ecc479bc712a6bdb56df3e346e33edcc141f469f82840bab9f4bc2bc41bf91d AS golang-my-hobby-arch`,
			target:     "golang",
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := archSelector(tt.target, tt.dockerfile)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
