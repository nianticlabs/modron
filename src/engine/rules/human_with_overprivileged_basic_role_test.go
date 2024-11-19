package rules

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func TestHumanWithOverprivilegedBasicRole(t *testing.T) {
	iamPolicy := &pb.IamPolicy{
		Resource: &pb.Resource{
			Name: testProjectName,
		},
		Permissions: []*pb.Permission{
			{
				Role: "owner",
				Principals: []string{
					"user:human-account-owner",
					"group:group-owner",
				},
			},
			{
				Role: "iam.securityAdmin",
				Principals: []string{
					"user:human-account-securityadmin",
				},
			},
			{
				Role: "viewer",
				Principals: []string{
					"user:human-account-viewer",
				},
			},
			{
				Role: "editor",
				Principals: []string{
					"user:human-account-editor",
				},
			},
			{
				Role: "non-basic",
				Principals: []string{
					"human-account-non-basic",
				},
			},
		},
	}
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "",
			ResourceGroupName: testProjectName,
			CollectionUid:     collectID,
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
			IamPolicy: iamPolicy,
		},
	}

	want := []*pb.Observation{
		{
			ScanUid: proto.String("unit-test-scan"),
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-0"),
				ExternalId:    proto.String("projects/project-0"),
				GroupName:     "projects/project-0",
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			Name:          "HUMAN_WITH_OVERPRIVILEGED_BASIC_ROLE",
			ExpectedValue: structpb.NewStringValue("No basic roles"),
			ObservedValue: structpb.NewStringValue("owner"),
			Remediation: &pb.Remediation{
				Description:    "Human account or group user:human-account-owner has overprivileged basic roles on project [project-0](https://console.cloud.google.com/iam-admin/iam?project=project-0)",
				Recommendation: "Consider assigning \"Developer\" to the editors and \"Owner\" to the owners instead of using basic roles.",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			ScanUid: proto.String("unit-test-scan"),
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-0"),
				ExternalId:    proto.String("projects/project-0"),
				GroupName:     "projects/project-0",
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			Name:          "HUMAN_WITH_OVERPRIVILEGED_BASIC_ROLE",
			ExpectedValue: structpb.NewStringValue("No basic roles"),
			ObservedValue: structpb.NewStringValue("iam.securityAdmin"),
			Remediation: &pb.Remediation{
				Description:    "Human account or group user:human-account-securityadmin has overprivileged basic roles on project [project-0](https://console.cloud.google.com/iam-admin/iam?project=project-0)",
				Recommendation: "Consider assigning \"Developer\" to the editors and \"Owner\" to the owners instead of using basic roles.",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			ScanUid: proto.String("unit-test-scan"),
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-0"),
				ExternalId:    proto.String("projects/project-0"),
				GroupName:     "projects/project-0",
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			Name:          "HUMAN_WITH_OVERPRIVILEGED_BASIC_ROLE",
			ExpectedValue: structpb.NewStringValue("No basic roles"),
			ObservedValue: structpb.NewStringValue("owner"),
			Remediation: &pb.Remediation{
				Description:    "Human account or group group:group-owner has overprivileged basic roles on project [project-0](https://console.cloud.google.com/iam-admin/iam?project=project-0)",
				Recommendation: "Consider assigning \"Developer\" to the editors and \"Owner\" to the owners instead of using basic roles.",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			ScanUid: proto.String("unit-test-scan"),
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-0"),
				ExternalId:    proto.String("projects/project-0"),
				GroupName:     "projects/project-0",
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			Name:          "HUMAN_WITH_OVERPRIVILEGED_BASIC_ROLE",
			ExpectedValue: structpb.NewStringValue("No basic roles"),
			ObservedValue: structpb.NewStringValue("viewer"),
			Remediation: &pb.Remediation{
				Description:    "Human account or group user:human-account-viewer has overprivileged basic roles on project [project-0](https://console.cloud.google.com/iam-admin/iam?project=project-0)",
				Recommendation: "Consider assigning \"Developer\" to the editors and \"Owner\" to the owners instead of using basic roles.",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			ScanUid: proto.String("unit-test-scan"),
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-0"),
				ExternalId:    proto.String("projects/project-0"),
				GroupName:     "projects/project-0",
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			Name:          "HUMAN_WITH_OVERPRIVILEGED_BASIC_ROLE",
			ExpectedValue: structpb.NewStringValue("No basic roles"),
			ObservedValue: structpb.NewStringValue("editor"),
			Remediation: &pb.Remediation{
				Description:    "Human account or group user:human-account-editor has overprivileged basic roles on project [project-0](https://console.cloud.google.com/iam-admin/iam?project=project-0)",
				Recommendation: "Consider assigning \"Developer\" to the editors and \"Owner\" to the owners instead of using basic roles.",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewHumanWithOverprivilegedBasicRoleRule()}, want)
}
