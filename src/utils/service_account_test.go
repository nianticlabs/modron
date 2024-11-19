package utils_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/nianticlabs/modron/src/utils"
)

func TestGetProjectFromSAEmail(t *testing.T) {
	tc := []struct {
		saEmail  string
		expected string
	}{
		{
			"gitlab-sa@example-project.iam.gserviceaccount.com",
			"example-project",
		},
		{
			"example-project@appspot.gserviceaccount.com",
			"example-project",
		},
		{
			"123456789012-compute@developer.gserviceaccount.com",
			"",
		},
	}

	for _, tt := range tc {
		if diff := cmp.Diff(tt.expected, utils.GetGCPProjectFromSAEmail(tt.saEmail)); diff != "" {
			t.Errorf("unexpected result (-want +got):\n%s", diff)
		}
	}
}
