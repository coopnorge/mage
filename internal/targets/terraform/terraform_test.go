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
			name:    "Terraform Init should succeed",
			workdir:  "tests/succes",
			wantErr: false,
		},
        {
			name:    "Terraform Init should fail",
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
			gotErr := Init(context.Background())
			 if tt.wantErr {
			 	assert.Error(t, gotErr)
			 } else {
			 	assert.NoError(t, gotErr)
			 }

		})
	}
}
