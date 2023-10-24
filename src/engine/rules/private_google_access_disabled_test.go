package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckDetectsPrivateGoogleAccessDisabled(t *testing.T) {
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
			Name:              "network-no-private-access",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					Ips:                      []string{"8.8.4.4"},
					GcpPrivateGoogleAccessV4: false,
				},
			},
		},
		{
			Name:              "network-private-access",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					Ips:                      []string{"8.8.8.8"},
					GcpPrivateGoogleAccessV4: true,
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name: PrivateGoogleAccessDisabled,
			Resource: &pb.Resource{
				Name: "network-no-private-access",
			},
			ObservedValue: structpb.NewStringValue("disabled"),
			ExpectedValue: structpb.NewStringValue("enabled"),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewPrivateGoogleAccessDisabledRule()})

	// Check that the observations are correct.
	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
