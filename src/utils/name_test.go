package utils_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/nianticlabs/modron/src/utils"
)

func TestGetHumanReadableName(t *testing.T) {
	tc := [][]string{
		{
			"//container.googleapis.com/projects/xyz/locations/us-central1/clusters/cluster-name/k8s/namespaces/kube-system/pods/my-pod-1",
			"my-pod-1",
		},
		{
			"//container.googleapis.com/projects/xyz/locations/us-central1/clusters/cluster-name/k8s/namespaces/kube-system",
			"kube-system",
		},
		{
			"//container.googleapis.com/projects/xyz/locations/us-central1/clusters/cluster-name",
			"cluster-name",
		},
		{
			"//container.googleapis.com/projects/xyz/locations/us-central1",
			"us-central1",
		},
		{
			"//container.googleapis.com/projects/xyz",
			"xyz",
		},
		{
			"//iam.googleapis.com/projects/example-project/serviceAccounts/my-service-account@example-project.iam.gserviceaccount.com",
			"my-service-account@example-project.iam.gserviceaccount.com",
		},
		{
			"//iam.googleapis.com/projects/example-project/serviceAccounts/3984989392373/keys/b8ceb3f5d69d4e46acc9e74bf224d4e9",
			"b8ceb3f5d69d4e46acc9e74bf224d4e9",
		},
		{
			"//compute.googleapis.com/projects/example-project/zones/us-central1-f/instances/my-instance-4897322-03032024-cxx1-test-a0b0c0",
			"my-instance-4897322-03032024-cxx1-test-a0b0c0",
		},
		{
			"//container.googleapis.com/projects/example-1/zones/us-central1-b/clusters/security-runners/k8s/namespaces/twistlock",
			"twistlock",
		},
		{
			"//container.googleapis.com/projects/example-2/zones/us-central1-b/clusters/security-runners",
			"security-runners",
		},
	}

	for _, c := range tc {
		want := c[1]
		got := utils.GetHumanReadableName(c[0])
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("GetHumanReadableName(%q) mismatch (-want +got):\n%s", c[0], diff)
		}
	}
}
