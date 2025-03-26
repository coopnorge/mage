package git_test

import (
	"testing"

	"github.com/coopnorge/mage/internal/git"
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
