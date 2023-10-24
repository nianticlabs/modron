package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/structpb"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestPublicClusterNodesDetection(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "",
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "private-cluster",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					PrivateCluster: true,
				},
			},
		},
		{
			Name:              "public-cluster",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					PrivateCluster: false,
				},
			},
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewClusterNodesHavePublicIpsRule()})

	// Expected values are ordered lexicographically.
	want := []*pb.Observation{
		{
			Name: ClusterNodesHavePublicIps,
			Resource: &pb.Resource{
				Name: "public-cluster",
			},
			ExpectedValue: structpb.NewStringValue("private"),
			ObservedValue: structpb.NewStringValue("public"),
		},
	}

	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
