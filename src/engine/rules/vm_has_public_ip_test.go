package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckVMHasPublicIP(t *testing.T) {
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
			Name:              "public-ip",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp: "8.8.8.8",
				},
			},
		},
		{
			Name:              "gke-public-ip",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp: "8.8.8.8",
				},
			},
		},
		{
			Name:              "public-ip-automatically-created-34",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp: "8.8.8.8",
				},
			},
		},
		{
			Name:              "no-public-ip",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name: VMHasPublicIPRuleName,
			Resource: &pb.Resource{
				Name: "public-ip",
			},
			ObservedValue: structpb.NewStringValue("8.8.8.8"),
			ExpectedValue: structpb.NewStringValue("empty"),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewVMHasPublicIPRule()})

	// Check that the observations are correct.
	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
