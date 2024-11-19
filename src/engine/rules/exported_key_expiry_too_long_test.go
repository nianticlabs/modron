package rules

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

func TestExportedKeyTooOld(t *testing.T) {
	now := time.Now().UTC()
	yesterday := now.Add(time.Hour * -24)
	tomorrow := now.Add(time.Hour * 24)
	oneYearAgo := now.Add(-time.Hour * 24 * 365)

	outdatedExportedKey := &pb.Resource{
		Name:              "outdated-exported-key",
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_ExportedCredentials{
			ExportedCredentials: &pb.ExportedCredentials{
				CreationDate:   timestamppb.New(oneYearAgo),
				ExpirationDate: timestamppb.New(tomorrow),
			},
		},
	}
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "",
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "rotated-exported-key",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ExportedCredentials{
				ExportedCredentials: &pb.ExportedCredentials{
					CreationDate:   timestamppb.New(yesterday),
					ExpirationDate: timestamppb.New(tomorrow),
				},
			},
		},
		outdatedExportedKey,
	}

	// Expected values are ordered lexicographically.
	want := []*pb.Observation{
		{
			Name:          ExportedKeyIsTooOld,
			ResourceRef:   utils.GetResourceRef(outdatedExportedKey),
			ExpectedValue: structpb.NewStringValue("later creation date"),
			ObservedValue: structpb.NewStringValue(oneYearAgo.Format("2006-01-02 15:04:05 +0000 UTC")),
			Remediation: &pb.Remediation{
				Description:    "Exported key [\"outdated-exported-key\"](https://console.cloud.google.com/apis/credentials?project=project-0) is too long lived",
				Recommendation: "Rotate the exported key [\"outdated-exported-key\"](https://console.cloud.google.com/apis/credentials?project=project-0) every 6 months",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewExportedKeyIsTooOldRule()}, want)
}
