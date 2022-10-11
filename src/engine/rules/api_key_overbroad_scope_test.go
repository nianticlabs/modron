package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestCheckDetectsOverbroadScope(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:   testProjectName,
			Parent: "",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:   "project-1",
			Parent: "",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "api-key-unrestricted-0",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_ApiKey{
				ApiKey: &pb.APIKey{
					Scopes: nil,
				},
			},
		},
		{
			Name:              "api-key-unrestricted-1",
			Parent:            "project-1",
			ResourceGroupName: "project-1",
			Type: &pb.Resource_ApiKey{
				ApiKey: &pb.APIKey{
					Scopes: []string{},
				},
			},
		},
		{
			Name:              "api-key-with-overbroad-scope-1",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_ApiKey{
				ApiKey: &pb.APIKey{
					Scopes: []string{
						// overbroad
						"iamcredentials.googleapis.com",
						// not overbroad
						"storage_api",
						// overbroad
						"apikeys",
					},
				},
			},
		},
		{
			Name:              "api-key-without-overbroad-scope",
			Parent:            "project-1",
			ResourceGroupName: "project-1",
			Type: &pb.Resource_ApiKey{
				ApiKey: &pb.APIKey{
					Scopes: []string{"bigquerystorage.googleapis.com"},
				},
			},
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewApiKeyOverbroadScopeRule()})

	// Expected values are ordered lexicographically.
	want := []*pb.Observation{
		{
			Name:          "api-key-unrestricted-0",
			ExpectedValue: structpb.NewStringValue("restricted"),
			ObservedValue: structpb.NewStringValue("unrestricted"),
		},
		{
			Name:          "api-key-unrestricted-1",
			ExpectedValue: structpb.NewStringValue("restricted"),
			ObservedValue: structpb.NewStringValue("unrestricted"),
		},
		{
			Name:          "api-key-with-overbroad-scope-1",
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iamcredentials.googleapis.com"),
		},
		{
			Name:          "api-key-without-overbroad-scope",
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("apikeys"),
		},
	}
	// Sort observations lexicographically by resource name.
	slices.SortStableFunc(got, func(lhs, rhs *pb.Observation) bool {
		return lhs.Resource.Name < rhs.Resource.Name
	})

	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
