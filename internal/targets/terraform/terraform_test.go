package terraform

import (
	"os"
	"testing"
    "context"
	"github.com/stretchr/testify/assert"
)

func TestTestTarget(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir  string
		wantErr bool
	}{
		{
			name:    "Terraform Test should succeed",
			workdir:  "tests/succes",
			wantErr: false,
		},
        {
			name:    "Terraform Test should fail",
			workdir:  "tests/test-fail",
			wantErr: true,
		},
	}
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Chdir(tt.workdir)
            t.Cleanup(func(){
		      os.RemoveAll(".terraform.lock.hcl")
		      os.RemoveAll(".terraform")
		    })
			ctx := context.Background()
			gotErr := Test(ctx)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.NoError(t, gotErr)
			}
		//assert.Equal(t, tt.want, got)

		})
	}
}
