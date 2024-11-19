package rules

import (
	"testing"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

const (
	collectID = "collectID-1"
)

func TestCheckDetectsHighPrivilege(t *testing.T) {
	// Because of the new memoization we need a specific project name for this test.
	testProjectName := "projects/check-detects-high-privilege-0"
	testProjectName1 := "projects/check-detects-high-privilege-1"

	account0 := &pb.Resource{
		Uid:               uuid.NewString(),
		Name:              "account-0",
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

	account1 := &pb.Resource{
		Uid:               uuid.NewString(),
		Name:              "account-1",
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

	account2 := &pb.Resource{
		Uid:               uuid.NewString(),
		Name:              "account-2",
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

	account3 := &pb.Resource{
		Uid:               uuid.NewString(),
		Name:              "account-3",
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

	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "folders/123",
			ResourceGroupName: testProjectName,
			CollectionUid:     collectID,
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "iam.serviceAccountAdmin",
						Principals: []string{
							"serviceAccount:account-0",
						},
					},
					{
						Role: "dataflow.admin",
						Principals: []string{
							"serviceAccount:account-1",
						},
					},
					{
						Role: "viewer",
						Principals: []string{
							"serviceAccount:account-1",
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
			CollectionUid:     collectID,
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "iam.serviceAccountUser",
						Principals: []string{
							"serviceAccount:account-2",
							"serviceAccount:account-3",
						},
					},
				},
			},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		account0,
		account1,
		account2,
		account3,
	}

	want := []*pb.Observation{
		{
			Name:          TooHighPrivilegesRuleName,
			ResourceRef:   utils.GetResourceRef(account0),
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iam.serviceAccountAdmin"),
			Remediation: &pb.Remediation{
				Description:    "Service account [\"account-0\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=check-detects-high-privilege-0) has over-broad role \"iam.serviceAccountAdmin\"",
				Recommendation: "Replace the role \"iam.serviceAccountAdmin\" for service account [\"account-0\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=check-detects-high-privilege-0) with a predefined or custom role that grants it the **smallest set of permissions** needed to operate. This role **cannot** be any of the following: `[editor owner composer.admin dataproc.admin dataproc.editor dataflow.admin dataflow.developer iam.serviceAccountAdmin iam.serviceAccountUser iam.serviceAccountTokenCreator]` *Hint: The Security insights column can help you reduce the amount of permissions*",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name:          TooHighPrivilegesRuleName,
			ResourceRef:   utils.GetResourceRef(account1),
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("dataflow.admin"),
			Remediation: &pb.Remediation{
				Description:    "Service account [\"account-1\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=check-detects-high-privilege-0) has over-broad role \"dataflow.admin\"",
				Recommendation: "Replace the role \"dataflow.admin\" for service account [\"account-1\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=check-detects-high-privilege-0) with a predefined or custom role that grants it the **smallest set of permissions** needed to operate. This role **cannot** be any of the following: `[editor owner composer.admin dataproc.admin dataproc.editor dataflow.admin dataflow.developer iam.serviceAccountAdmin iam.serviceAccountUser iam.serviceAccountTokenCreator]` *Hint: The Security insights column can help you reduce the amount of permissions*",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name:          TooHighPrivilegesRuleName,
			ResourceRef:   utils.GetResourceRef(account2),
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iam.serviceAccountUser"),
			Remediation: &pb.Remediation{
				Description:    "Service account [\"account-2\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=check-detects-high-privilege-1) has over-broad role \"iam.serviceAccountUser\"",
				Recommendation: "Replace the role \"iam.serviceAccountUser\" for service account [\"account-2\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=check-detects-high-privilege-1) with a predefined or custom role that grants it the **smallest set of permissions** needed to operate. This role **cannot** be any of the following: `[editor owner composer.admin dataproc.admin dataproc.editor dataflow.admin dataflow.developer iam.serviceAccountAdmin iam.serviceAccountUser iam.serviceAccountTokenCreator]` *Hint: The Security insights column can help you reduce the amount of permissions*",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name:          TooHighPrivilegesRuleName,
			ResourceRef:   utils.GetResourceRef(account3),
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iam.serviceAccountUser"),
			Remediation: &pb.Remediation{
				Description:    "Service account [\"account-3\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=check-detects-high-privilege-1) has over-broad role \"iam.serviceAccountUser\"",
				Recommendation: "Replace the role \"iam.serviceAccountUser\" for service account [\"account-3\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=check-detects-high-privilege-1) with a predefined or custom role that grants it the **smallest set of permissions** needed to operate. This role **cannot** be any of the following: `[editor owner composer.admin dataproc.admin dataproc.editor dataflow.admin dataflow.developer iam.serviceAccountAdmin iam.serviceAccountUser iam.serviceAccountTokenCreator]` *Hint: The Security insights column can help you reduce the amount of permissions*",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewTooHighPrivilegesRule()}, want)
}
