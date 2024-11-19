package rules

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

func TestExportedKeyWithAdminPrivileges(t *testing.T) {
	account1 := &pb.Resource{
		Name:              "account-1",
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		CollectionUid:     collectID,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_ServiceAccount{
			ServiceAccount: &pb.ServiceAccount{
				ExportedCredentials: []*pb.ExportedCredentials{
					{CreationDate: timestamppb.Now()},
				},
			},
		},
	}
	account2 := &pb.Resource{
		Name:              "account-2",
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		CollectionUid:     collectID,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_ServiceAccount{
			ServiceAccount: &pb.ServiceAccount{
				ExportedCredentials: []*pb.ExportedCredentials{
					{CreationDate: timestamppb.Now()},
					{CreationDate: timestamppb.Now()},
				},
			},
		},
	}

	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			ResourceGroupName: testProjectName,
			CollectionUid:     collectID,
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "owner",
						Principals: []string{
							"serviceAccount:account-1",
						},
					},
					{
						Role: "iam.securityAdmin",
						Principals: []string{
							"serviceAccount:account-2",
						},
					},
					{
						Role: "viewer",
						Principals: []string{
							"serviceAccount:account-3-no-admin-privileges",
						},
					},
					{
						Role: "editor",
						Principals: []string{
							"serviceAccount:account-no-exported-credentials",
						},
					},
				},
			},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		account1,
		account2,
		{
			Name:              "account-3-no-admin-privileges",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			CollectionUid:     collectID,
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
			CollectionUid:     collectID,
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
			Name:          ExportedKeyWithAdminPrivileges,
			ResourceRef:   utils.GetResourceRef(account1),
			ExpectedValue: structpb.NewStringValue("0 keys"),
			ObservedValue: structpb.NewStringValue("1 keys"),
			Remediation: &pb.Remediation{
				Description:    "Service account [\"account-1\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=project-0) has 1 exported keys with admin privileges",
				Recommendation: "Avoid exporting keys of service accounts with admin privileges, they can be copied and used outside of Niantic. Revoke the exported key by clicking on service account [\"account-1\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=project-0), switch to the KEYS tab and delete the exported key. Instead of exporting keys, make use of [workload identity](https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity) or similar concepts",
			},
			Severity: pb.Severity_SEVERITY_CRITICAL,
		},
		{
			Name:          ExportedKeyWithAdminPrivileges,
			ResourceRef:   utils.GetResourceRef(account2),
			ExpectedValue: structpb.NewStringValue("0 keys"),
			ObservedValue: structpb.NewStringValue("2 keys"),
			Remediation: &pb.Remediation{
				Description:    "Service account [\"account-2\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=project-0) has 2 exported keys with admin privileges",
				Recommendation: "Avoid exporting keys of service accounts with admin privileges, they can be copied and used outside of Niantic. Revoke the exported key by clicking on service account [\"account-2\"](https://console.cloud.google.com/iam-admin/serviceaccounts?project=project-0), switch to the KEYS tab and delete the exported key. Instead of exporting keys, make use of [workload identity](https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity) or similar concepts",
			},
			Severity: pb.Severity_SEVERITY_CRITICAL,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewExportedKeyWithAdminPrivilegesRule()}, want)
}
