package rules

import (
	"fmt"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

func TestUnusedExportedKey(t *testing.T) {
	now := time.Now().UTC()
	yesterday := now.Add(time.Hour * -24)
	oneYearAgo := now.Add(-time.Hour * 24 * 365)
	threeMonthsAndOneDay := now.Add(-time.Hour * 24 * 91)
	oneYearAhead := now.Add(time.Hour * 24 * 365)

	resourceNotUsedInALongTime := &pb.Resource{
		Name:              testProjectName + "/serviceAccounts/unused-svc-account@project-id.iam.gserviceaccount.com/keys/d88bb32b79ee4193a05ee178447e09a4",
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_ExportedCredentials{
			ExportedCredentials: &pb.ExportedCredentials{
				CreationDate:   timestamppb.New(oneYearAgo),
				ExpirationDate: &timestamppb.Timestamp{Seconds: oneYearAhead.Unix(), Nanos: 0},
				LastUsage:      timestamppb.New(threeMonthsAndOneDay),
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
		resourceNotUsedInALongTime,
		{
			Name:              "used-yesterday",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ExportedCredentials{
				ExportedCredentials: &pb.ExportedCredentials{
					CreationDate:   timestamppb.New(oneYearAgo),
					ExpirationDate: &timestamppb.Timestamp{Seconds: oneYearAhead.Unix(), Nanos: 0},
					LastUsage:      timestamppb.New(yesterday),
				},
			},
		},
		{
			Name:              "created-yesterday-unused-do-not-report",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ExportedCredentials{
				ExportedCredentials: &pb.ExportedCredentials{
					CreationDate:   timestamppb.New(yesterday),
					ExpirationDate: &timestamppb.Timestamp{Seconds: oneYearAhead.Unix(), Nanos: 0},
					LastUsage:      nil,
				},
			},
		},
	}

	// Expected values are ordered lexicographically.
	want := []*pb.Observation{
		{
			Name:          unusedExportedCredentials,
			ResourceRef:   utils.GetResourceRef(resourceNotUsedInALongTime),
			ExpectedValue: structpb.NewStringValue(fmt.Sprintf("%s <", oldestUsage.Format(time.RFC3339))),
			ObservedValue: structpb.NewStringValue(threeMonthsAndOneDay.Format(time.RFC3339)),
			Remediation: &pb.Remediation{
				Description:    "Exported key `d88bb32b79ee4193a05ee178447e09a4` of [\"unused-svc-account@project-id.iam.gserviceaccount.com\"](https://console.cloud.google.com/iam-admin/serviceaccounts/details/unused-svc-account@project-id.iam.gserviceaccount.com/keys?project=project-0) has not been used in the last 3 months",
				Recommendation: "Consider deleting the exported key `d88bb32b79ee4193a05ee178447e09a4` of [\"unused-svc-account@project-id.iam.gserviceaccount.com\"](https://console.cloud.google.com/iam-admin/serviceaccounts/details/unused-svc-account@project-id.iam.gserviceaccount.com/keys?project=project-0) which is no longer in use",
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewUnusedExportedCredentialsRule()}, want)
}
