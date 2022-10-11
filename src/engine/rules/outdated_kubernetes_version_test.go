package rules

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/structpb"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestOutdatedKubernetesVersionDetection(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:      testProjectName,
			Parent:    "",
			IamPolicy: &pb.IamPolicy{},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "up-to-date-k8s-cluster",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					PrivateCluster: true,
					MasterVersion:  "1.22.10-gke.600",
					NodesVersion:   "1.22.10-gke.600",
				},
			},
		},
		{
			Name:              "cluster-with-outdated-nodes-version",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					PrivateCluster: true,
					MasterVersion:  "1.22.10-gke.600",
					NodesVersion:   "1.15.10-gke.600",
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			ExpectedValue: structpb.NewStringValue(fmt.Sprintf("version > %.2f", currentK8sVersion)),
			ObservedValue: structpb.NewStringValue("1.15.10-gke.600"),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewOutDatedKubernetesVersionRule()})

	// Check that the observations are correct.
	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
