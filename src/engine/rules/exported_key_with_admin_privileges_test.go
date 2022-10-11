package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestExportedKeyWithAdminPrivileges(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:          testProjectName,
			CollectionUid: collectId,
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "owner",
						Principals: []string{
							"account-1",
						},
					},
					{
						Role: "admin",
						Principals: []string{
							"account-2",
						},
					},
					{
						Role: "viewer",
						Principals: []string{
							"account-3-no-admin-privileges",
						},
					},
					{
						Role: "editor",
						Principals: []string{
							"account-no-exported-credentials",
						},
					},
				},
			},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
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
					ExportedCredentials: []*pb.ExportedCredentials{
						{CreationDate: timestamppb.Now()},
					},
				},
			},
		},
		{
			Name:              "account-2",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			CollectionUid:     collectId,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{
						{CreationDate: timestamppb.Now()},
						{CreationDate: timestamppb.Now()},
					},
				},
			},
		},
		{
			Name:              "account-3-no-admin-privileges",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			CollectionUid:     collectId,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{
						{CreationDate: timestamppb.Now()},
						{CreationDate: timestamppb.Now()},
						{CreationDate: timestamppb.Now()},
					},
				},
			},
		},
		{
			Name:              "account-no-exported-credentials",
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
	}

	want := []*pb.Observation{
		{
			ExpectedValue: structpb.NewStringValue("0 keys"),
			ObservedValue: structpb.NewStringValue("1 keys"),
		},
		{
			ExpectedValue: structpb.NewStringValue("0 keys"),
			ObservedValue: structpb.NewStringValue("2 keys"),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewExportedKeyWithAdminPrivilegesRule()})

	// Sort observations lexicographically by resource name.
	slices.SortStableFunc(got, func(lhs, rhs *pb.Observation) bool {
		return lhs.Resource.Name < rhs.Resource.Name
	})

	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
