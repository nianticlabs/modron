package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/structpb"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestMasterAuthorizedNetworksNotSet(t *testing.T) {
	notSetResourceName := "master-authorized-networks-not-set"
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "",
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "master-authorized-networks-set",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					MasterAuthorizedNetworks: []string{"10.0.0.0/8"},
				},
			},
		},
		{
			Name:              notSetResourceName,
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					MasterAuthorizedNetworks: []string{},
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name: MasterAuthorizedNetworksNotSet,
			Resource: &pb.Resource{
				Name: notSetResourceName,
			},
			ExpectedValue: structpb.NewStringValue("not empty"),
			ObservedValue: structpb.NewStringValue("empty"),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewMasterAuthorizedNetworksNotSetRule()})

	// Check that the observations are correct.
	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
