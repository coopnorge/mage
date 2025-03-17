package docker_test

import (
	"testing"

	"github.com/coopnorge/mage/internal/docker"
	"github.com/stretchr/testify/assert"
)

func TestParseMetadata(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		imageName string
		want      docker.Metadata
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "good metadata",
			file:      "./testdata/parse-metadata/good_metadata.json",
			imageName: "ocreg.invalid/coopnorge/helloworld/helloworld:v2025.03.11135857",
		},
		{
			name:    "only latest tag",
			file:    "./testdata/parse-metadata/only-latest-tag_metadata.json",
			wantErr: true,
			errMsg:  "image name not found in:",
		},
		{
			name:    "empty json",
			file:    "./testdata/parse-metadata/empty_metadata.json",
			wantErr: true,
			errMsg:  "no metadata found in:",
		},
		{
			name:    "file not found",
			file:    "./testdata/parse-metadata/does-not-exist_metadata.json",
			wantErr: true,
			errMsg:  "no such file or director",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := docker.ParseMetadata(tt.file)
			if tt.wantErr {
				assert.ErrorContains(t, gotErr, tt.errMsg)
			} else {
				assert.NoError(t, gotErr)
				assert.NotZero(t, got)
				assert.Equal(t, tt.imageName, got.ImageName)
			}
		})
	}
}

func TestFindMetadataFiles(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		want    []string
		wantErr bool
	}{
		{
			name: "base case",
			base: "./testdata/send-metadata-to-github",
			want: []string{
				"testdata/send-metadata-to-github/app1/oci/binary1/metadata.json",
				"testdata/send-metadata-to-github/app1/oci/binary2/metadata.json",
				"testdata/send-metadata-to-github/app2/oci/binary1/metadata.json",
				"testdata/send-metadata-to-github/app2/oci/binary2/metadata.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := docker.FindMetadataFiles(tt.base)
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestImages(t *testing.T) {
	tests := []struct {
		name     string
		imageDir string
		want     map[string]map[string]map[string]string
		wantErr  bool
	}{
		{
			name:     "base case",
			imageDir: "./testdata/send-metadata-to-github",
			want: map[string]map[string]map[string]string{
				"app1": {
					"binary1": {
						"image": "ocreg.invalid/coopnorge/app1/binary1:v2025.03.11135857",
						"tag":   "v2025.03.11135857",
					},
					"binary2": {
						"image": "ocreg.invalid/coopnorge/app1/binary2:v2025.03.11135857",
						"tag":   "v2025.03.11135857",
					},
				},
				"app2": {
					"binary1": {
						"image": "ocreg.invalid/coopnorge/app2/binary1:v2025.03.11135857",
						"tag":   "v2025.03.11135857",
					},
					"binary2": {
						"image": "ocreg.invalid/coopnorge/app2/binary2:v2025.03.11135857",
						"tag":   "v2025.03.11135857",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := docker.Images(tt.imageDir)
			assert.NoError(t, gotErr)
			assert.NotZero(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}
