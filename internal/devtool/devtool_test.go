package devtool

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionWithinMajor(t *testing.T) {
	tests := []struct {
		name        string
		devtoolVer  string
		currentVer  string
		minorOffset int
		wantErr     bool
	}{
		// accepted
		{name: "exact match", devtoolVer: "3.22.1", currentVer: "3.22.1", minorOffset: 5, wantErr: false},
		{name: "lowest allowed minor", devtoolVer: "3.22.1", currentVer: "3.17.0", minorOffset: 5, wantErr: false},
		{name: "newer minor in same major", devtoolVer: "3.22.1", currentVer: "3.30.0", minorOffset: 5, wantErr: false},
		{name: "zero offset same minor", devtoolVer: "1.25.5", currentVer: "1.25.0", minorOffset: 0, wantErr: false},
		{name: "major zero within range", devtoolVer: "0.7.0", currentVer: "0.6.0", minorOffset: 1, wantErr: false},

		// not accepted
		{name: "below lowest allowed minor", devtoolVer: "3.22.1", currentVer: "3.16.9", minorOffset: 5, wantErr: true},
		{name: "next major", devtoolVer: "3.22.1", currentVer: "4.0.0", minorOffset: 5, wantErr: true},
		{name: "next major with v prefix", devtoolVer: "3.22.1", currentVer: "v4.3.2", minorOffset: 5, wantErr: true},
		{name: "next major prerelease", devtoolVer: "3.22.1", currentVer: "4.0.0-rc.1", minorOffset: 5, wantErr: true},
		{name: "previous major", devtoolVer: "3.22.1", currentVer: "2.99.0", minorOffset: 5, wantErr: true},
		{name: "zero offset older minor", devtoolVer: "1.25.5", currentVer: "1.24.9", minorOffset: 0, wantErr: true},
		{name: "zero offset next major", devtoolVer: "1.25.5", currentVer: "2.0.0", minorOffset: 0, wantErr: true},
		{name: "major zero below range", devtoolVer: "0.7.0", currentVer: "0.5.0", minorOffset: 1, wantErr: true},
		{name: "major zero to one", devtoolVer: "0.7.0", currentVer: "1.0.0", minorOffset: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devtool, err := version.NewVersion(tt.devtoolVer)
			require.NoError(t, err)
			current, err := version.NewVersion(tt.currentVer)
			require.NoError(t, err)

			err = versionWithinMajor(devtool, current, tt.minorOffset)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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
			want:    "ocreg.invalid/github.com/coopnorge/mage/golang-devtool:latest",
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
