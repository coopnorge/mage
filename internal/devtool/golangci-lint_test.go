package devtool

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func TestGolangCILintVersionIsSameMajorAndMinor(t *testing.T) {
	type args struct {
		devtoolVer string
		currentVer string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "exact match",
			args: args{
				devtoolVer: "2.10.1",
				currentVer: "2.10.1",
			},
			wantErr: false,
		},
		{
			name: "same major and minor, different patch lower",
			args: args{
				devtoolVer: "2.10.1",
				currentVer: "2.10.0",
			},
			wantErr: false,
		},
		{
			name: "same major and minor, different patch higher",
			args: args{
				devtoolVer: "2.10.0",
				currentVer: "2.10.1",
			},
			wantErr: false,
		},
		{
			name: "different minor higher",
			args: args{
				devtoolVer: "2.9.0",
				currentVer: "2.10.0",
			},
			wantErr: true,
		},
		{
			name: "different minor lower",
			args: args{
				devtoolVer: "2.10.0",
				currentVer: "2.9.0",
			},
			wantErr: true,
		},
		{
			name: "different major higher",
			args: args{
				devtoolVer: "1.0.0",
				currentVer: "2.0.0",
			},
			wantErr: true,
		},
		{
			name: "different major lower",
			args: args{
				devtoolVer: "2.0.0",
				currentVer: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "different major lower, but minor higher",
			args: args{
				devtoolVer: "2.0.0",
				currentVer: "1.1.0",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := GoLangCILint{}

			dVer, err := version.NewVersion(tt.args.devtoolVer)
			require.NoError(t, err)

			cVer, err := version.NewVersion(tt.args.currentVer)
			require.NoError(t, err)

			err = gl.versionIsSameMajorAndMinor(dVer, cVer)
			if tt.wantErr {
				require.Error(t, err, "expected error but got nil")
			} else {
				require.NoError(t, err, "go unexpected error")
			}
		})
	}
}
