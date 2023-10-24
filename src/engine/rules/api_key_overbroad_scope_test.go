package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestCheckDetectsOverbroadScope(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "folders/123",
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "projects/project-1",
			Parent:            "folders/234",
			ResourceGroupName: "projects/project-1",
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
			Parent:            "projects/project-1",
			ResourceGroupName: "projects/project-1",
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
			Name: ApiKeyOverbroadScopeRuleName,
			Resource: &pb.Resource{
				Name: "api-key-unrestricted-0",
			},
			ExpectedValue: structpb.NewStringValue("restricted"),
			ObservedValue: structpb.NewStringValue("unrestricted"),
		},
		{
			Name: ApiKeyOverbroadScopeRuleName,
			Resource: &pb.Resource{
				Name: "api-key-with-overbroad-scope-1",
			},
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iamcredentials.googleapis.com"),
		},
		{
			Name: ApiKeyOverbroadScopeRuleName,
			Resource: &pb.Resource{
				Name: "api-key-with-overbroad-scope-1",
			},
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("apikeys"),
		},
	}

	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
