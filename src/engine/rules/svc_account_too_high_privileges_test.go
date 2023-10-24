package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

const (
	collectId = "collectId-1"
)

func TestCheckDetectsHighPrivilege(t *testing.T) {
	// Because of the new memoization we need a specific project name for this test.
	testProjectName := "projects/test-project" + uuid.NewString()
	testProjectName1 := "projects/test-project1" + uuid.NewString()
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "folders/123",
			ResourceGroupName: testProjectName,
			CollectionUid:     collectId,
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
			Name:              testProjectName1,
			Parent:            "folders/234",
			ResourceGroupName: testProjectName1,
			CollectionUid:     collectId,
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
			Uid:               uuid.NewString(),
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
			Uid:               uuid.NewString(),
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
			Uid:               uuid.NewString(),
			Name:              "account-2",
			Parent:            testProjectName1,
			ResourceGroupName: testProjectName1,
			CollectionUid:     collectId,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{},
				},
			},
		},
		{
			Uid:               uuid.NewString(),
			Name:              "account-3",
			Parent:            testProjectName1,
			ResourceGroupName: testProjectName1,
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
			Name: TooHighPrivilegesRuleName,
			Resource: &pb.Resource{
				Name: "account-0",
			},
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iam.serviceAccountAdmin"),
		},
		{
			Name: TooHighPrivilegesRuleName,
			Resource: &pb.Resource{
				Name: "account-1",
			},
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("dataflow.admin"),
		},
		{
			Name: TooHighPrivilegesRuleName,
			Resource: &pb.Resource{
				Name: "account-2",
			},
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iam.serviceAccountUser"),
		},
		{
			Name: TooHighPrivilegesRuleName,
			Resource: &pb.Resource{
				Name: "account-3",
			},
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iam.serviceAccountUser"),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewTooHighPrivilegesRule()})

	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
