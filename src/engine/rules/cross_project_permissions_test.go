package rules

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

func TestCheckDetectsCrossProjectAccount(t *testing.T) {
	collectID := uuid.NewString()
	testProjectName1 := "projects/project-1"
	testProjectIdentifier := "12345678"
	bucketName := "bucket1"
	accountZeroProjectZero := fmt.Sprintf("serviceAccount:account-0@%s.iam.gserviceaccount.com", constants.ResourceWithoutProjectsPrefix(testProjectName))
	devAccountProjectZero := fmt.Sprintf("serviceAccount:%s-compute@developer.gserviceaccount.com", testProjectIdentifier)
	accountOneProjectOne := fmt.Sprintf("serviceAccount:account-1@%s.iam.gserviceaccount.com", constants.ResourceWithoutProjectsPrefix(testProjectName1))
	accountTwoProjectOne := fmt.Sprintf("serviceAccount:account-2@%s.iam.gserviceaccount.com", constants.ResourceWithoutProjectsPrefix(testProjectName1))
	accountThreeProjectZero := fmt.Sprintf("serviceAccount:account-3@%s.iam.gserviceaccount.com", constants.ResourceWithoutProjectsPrefix(testProjectName))
	invalidAccount := "serviceAccount:thishasnoatsign"
	googleServiceAcc := "serviceAccount:service-123455@gcp-sa-firebase.iam.gserviceaccount.com"

	// TODO: Add a test case that involves a group
	testProjectNameResource := &pb.Resource{
		Name:              testProjectName,
		Parent:            "",
		ResourceGroupName: testProjectName,
		CollectionUid:     collectID,
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
	}

	testProjectName1Resource := &pb.Resource{
		Name:              testProjectName1,
		Parent:            "",
		ResourceGroupName: testProjectName1,
		CollectionUid:     collectID,
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
	}

	bucketResource := &pb.Resource{
		Name:              bucketName,
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		CollectionUid:     collectID,
		IamPolicy: &pb.IamPolicy{
			Resource: nil,
			Permissions: []*pb.Permission{
				{
					Role: "storage.admin",
					Principals: []string{
						devAccountProjectZero,
						accountOneProjectOne,
					},
				},
			},
		},
		Type: &pb.Resource_Bucket{
			Bucket: &pb.Bucket{},
		},
	}

	accountZeroResource := &pb.Resource{
		Name:              accountZeroProjectZero,
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		CollectionUid:     collectID,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_ServiceAccount{
			ServiceAccount: &pb.ServiceAccount{
				ExportedCredentials: []*pb.ExportedCredentials{},
			},
		},
	}

	accountOneResource := &pb.Resource{
		Name:              accountOneProjectOne,
		Parent:            testProjectName1,
		ResourceGroupName: testProjectName1,
		CollectionUid:     collectID,
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
	}

	accountTwoResource := &pb.Resource{
		Name:              accountTwoProjectOne,
		Parent:            testProjectName1,
		ResourceGroupName: testProjectName1,
		CollectionUid:     collectID,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_ServiceAccount{
			ServiceAccount: &pb.ServiceAccount{
				ExportedCredentials: []*pb.ExportedCredentials{},
			},
		},
	}

	accountThreeResource := &pb.Resource{
		Name:              accountThreeProjectZero,
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		CollectionUid:     collectID,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_ServiceAccount{
			ServiceAccount: &pb.ServiceAccount{
				ExportedCredentials: []*pb.ExportedCredentials{},
			},
		},
	}

	googleServiceAccountResource := &pb.Resource{
		Name:              googleServiceAcc,
		Parent:            testProjectName1,
		ResourceGroupName: testProjectName1,
		CollectionUid:     collectID,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_ServiceAccount{
			ServiceAccount: &pb.ServiceAccount{
				ExportedCredentials: []*pb.ExportedCredentials{},
			},
		},
	}

	devAccountProjectZeroResource := &pb.Resource{
		Name:              devAccountProjectZero,
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		CollectionUid:     collectID,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_ServiceAccount{
			ServiceAccount: &pb.ServiceAccount{
				ExportedCredentials: []*pb.ExportedCredentials{},
			},
		},
	}

	resources := []*pb.Resource{
		testProjectNameResource,
		testProjectName1Resource,
		bucketResource,
		accountZeroResource,
		accountOneResource,
		accountTwoResource,
		accountThreeResource,
		googleServiceAccountResource,
		devAccountProjectZeroResource,
	}

	project0 := "[project-0](https://console.cloud.google.com/welcome?project=project-0)"
	project1 := "[project-1](https://console.cloud.google.com/welcome?project=project-1)"
	bucket1 := "[bucket1](https://console.cloud.google.com/storage/browser/bucket1)"
	saAccount1 := "[account-1@project-1.iam.gserviceaccount.com](https://console.cloud.google.com/iam-admin/serviceaccounts/details/account-1@project-1.iam.gserviceaccount.com?project=project-1)"
	saAccount3 := "[account-3@project-0.iam.gserviceaccount.com](https://console.cloud.google.com/iam-admin/serviceaccounts/details/account-3@project-0.iam.gserviceaccount.com?project=project-0)"
	want := []*pb.Observation{
		{
			Name:          CrossProjectPermissionsRuleName,
			ResourceRef:   utils.GetResourceRef(testProjectNameResource),
			ExpectedValue: structpb.NewStringValue("project-0"),
			ObservedValue: structpb.NewStringValue("project-1"),
			Remediation: &pb.Remediation{
				Description:    `The project ` + project0 + ` gives the service account ` + saAccount1 + ` vast permissions through the role ` + "`dataflow.admin`.\n" + `This principal is defined in project "project-1", which means that anybody with rights in that project can use it to control the resources in this one`,
				Recommendation: `Replace the service account ` + saAccount1 + ` controlling project ` + project0 + ` with a principal created in the project "projects/project-0" that grants it the **smallest set of permissions** needed to operate.`,
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name:          CrossProjectPermissionsRuleName,
			ResourceRef:   utils.GetResourceRef(bucketResource),
			ExpectedValue: structpb.NewStringValue("project-0"),
			ObservedValue: structpb.NewStringValue("project-1"),
			Remediation: &pb.Remediation{
				Description:    `The bucket ` + bucket1 + ` is controlled by the service account ` + saAccount1 + ` with role ` + "`storage.admin`" + ` defined in project "project-1"`,
				Recommendation: `Replace the service account ` + saAccount1 + ` controlling bucket ` + bucket1 + ` with a principal created in the project "projects/project-0" that grants it the **smallest set of permissions** needed to operate.`,
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name:          CrossProjectPermissionsRuleName,
			ResourceRef:   utils.GetResourceRef(testProjectName1Resource),
			ExpectedValue: structpb.NewStringValue("project-1"),
			ObservedValue: structpb.NewStringValue("project-0"),
			Remediation: &pb.Remediation{
				Description:    `The project ` + project1 + ` gives the service account ` + saAccount3 + ` vast permissions through the role ` + "`compute.admin`.\n" + `This principal is defined in project "project-0", which means that anybody with rights in that project can use it to control the resources in this one`,
				Recommendation: `Replace the service account ` + saAccount3 + ` controlling project ` + project1 + ` with a principal created in the project "projects/project-1" that grants it the **smallest set of permissions** needed to operate.`,
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name:          CrossProjectPermissionsRuleName,
			ResourceRef:   utils.GetResourceRef(accountOneResource),
			ExpectedValue: structpb.NewStringValue("project-1"),
			ObservedValue: structpb.NewStringValue("project-0"),
			Remediation: &pb.Remediation{
				Description:    `The service account ` + saAccount1 + ` is controlled by the service account ` + saAccount3 + ` with role ` + "`iam.serviceAccountTokenCreator`" + ` defined in project "project-0"`,
				Recommendation: `Replace the service account ` + saAccount3 + ` controlling service account ` + saAccount1 + ` with a principal created in the project "projects/project-1" that grants it the **smallest set of permissions** needed to operate.`,
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewCrossProjectPermissionsRule()}, want)
}
