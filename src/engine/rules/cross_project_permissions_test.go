package rules

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestCheckDetectsCrossProjectAccount(t *testing.T) {
	collectId := uuid.NewString()
	testProjectName1 := "projects/project-1"
	testProjectIdentifier := "12345678"
	accountZeroProjectZero := fmt.Sprintf("serviceAccount:account-0@%s.iam.gserviceaccount.com", constants.ResourceWithoutProjectsPrefix(testProjectName))
	devAccountProjectZero := fmt.Sprintf("serviceAccount:%s-compute@developer.gserviceaccount.com", testProjectIdentifier)
	accountOneProjectOne := fmt.Sprintf("serviceAccount:account-1@%s.iam.gserviceaccount.com", constants.ResourceWithoutProjectsPrefix(testProjectName1))
	accountTwoProjectOne := fmt.Sprintf("serviceAccount:account-2@%s.iam.gserviceaccount.com", constants.ResourceWithoutProjectsPrefix(testProjectName1))
	accountThreeProjectZero := fmt.Sprintf("serviceAccount:account-3@%s.iam.gserviceaccount.com", constants.ResourceWithoutProjectsPrefix(testProjectName))
	invalidAccount := "serviceAccount:thishasnoatsign"
	googleServiceAcc := "serviceAccount:service-123455@gcp-sa-firebase.iam.gserviceaccount.com"

	// TODO: Add a test case that involves a group
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "",
			ResourceGroupName: testProjectName,
			CollectionUid:     collectId,
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "iam.serviceAccountAdmin",
						Principals: []string{
							accountZeroProjectZero,
						},
					},
					{
						Role: "dataflow.admin",
						Principals: []string{
							accountOneProjectOne,
							devAccountProjectZero,
						},
					},
					{
						Role: "viewer",
						Principals: []string{
							accountOneProjectOne,
						},
					},
				},
			},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: testProjectIdentifier,
				},
			},
		},
		{
			Name:              testProjectName1,
			Parent:            "",
			ResourceGroupName: testProjectName1,
			CollectionUid:     collectId,
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "compute.admin",
						Principals: []string{
							accountTwoProjectOne,
							accountThreeProjectZero,
							googleServiceAcc,
						},
					},
				},
			},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: "9876543221",
				},
			},
		},
		{
			Name:              accountZeroProjectZero,
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
			Name:              accountOneProjectOne,
			Parent:            testProjectName1,
			ResourceGroupName: testProjectName1,
			CollectionUid:     collectId,
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "iam.serviceAccountTokenCreator",
						Principals: []string{
							accountTwoProjectOne,
							accountThreeProjectZero,
							invalidAccount,
						},
					},
				},
			},
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{},
				},
			},
		},
		{
			Name:              accountTwoProjectOne,
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
			Name:              accountThreeProjectZero,
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
			Name:              googleServiceAcc,
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
			Name: CrossProjectPermissionsRuleName,
			Resource: &pb.Resource{
				Name: testProjectName,
			},
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue(fmt.Sprintf("%q with role %s", accountOneProjectOne, "dataflow.admin")),
		},
		{
			Name: CrossProjectPermissionsRuleName,
			Resource: &pb.Resource{
				Name: testProjectName1,
			},
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue(fmt.Sprintf("%q with role %s", accountThreeProjectZero, "compute.admin")),
		},
		{
			Name: CrossProjectPermissionsRuleName,
			Resource: &pb.Resource{
				Name: accountOneProjectOne,
			},
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue(fmt.Sprintf("%q with role %s", accountThreeProjectZero, "iam.serviceAccountTokenCreator")),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewCrossProjectPermissionsRule()})

	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
