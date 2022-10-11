package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

const (
	collectId = "collectId-1"
)

func TestCheckDetectsHighPrivilege(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:          testProjectName,
			Parent:        "",
			CollectionUid: collectId,
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "iam.serviceAccountAdmin",
						Principals: []string{
							"account-0",
						},
					},
					{
						Role: "dataflow.admin",
						Principals: []string{
							"account-1",
						},
					},
					{
						Role: "viewer",
						Principals: []string{
							"account-1",
						},
					},
				},
			},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:          "project-1",
			Parent:        "",
			CollectionUid: collectId,
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "iam.serviceAccountUser",
						Principals: []string{
							"account-2",
							"account-3",
						},
					},
				},
			},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "account-0",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			CollectionUid:     collectId,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{},
				},
			},
		},
		{
			Name:              "account-1",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			CollectionUid:     collectId,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{},
				},
			},
		},
		{
			Name:              "account-2",
			Parent:            "project-1",
			ResourceGroupName: "project-1",
			CollectionUid:     collectId,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{},
				},
			},
		},
		{
			Name:              "account-3",
			Parent:            "project-1",
			ResourceGroupName: "project-1",
			CollectionUid:     collectId,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{},
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iam.serviceAccountAdmin"),
		},
		{
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("dataflow.admin"),
		},
		{
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iam.serviceAccountUser"),
		},
		{
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iam.serviceAccountUser"),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewTooHighPrivilegesRule()})

	// Sort observations lexicographically by resource name.
	slices.SortStableFunc(got, func(lhs, rhs *pb.Observation) bool {
		return lhs.Resource.Name < rhs.Resource.Name
	})

	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
