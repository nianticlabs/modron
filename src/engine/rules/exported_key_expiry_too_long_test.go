package rules

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestExportedKeyTooOld(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(time.Hour * -24)
	tomorrow := now.Add(time.Hour * 24)
	oneYearAgo := now.Add(-time.Hour * 24 * 365)
	oneYearAhead := now.Add(time.Hour * 24 * 365)
	nonExpiring := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	resources := []*pb.Resource{
		{
			Name:      testProjectName,
			Parent:    "",
			IamPolicy: &pb.IamPolicy{},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "expiring-exported-key",
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
		{
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
		},
		{
			Name:              "expiring-in-one-year-exported-key",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ExportedCredentials{
				ExportedCredentials: &pb.ExportedCredentials{
					CreationDate:   timestamppb.New(oneYearAgo),
					ExpirationDate: &timestamppb.Timestamp{Seconds: oneYearAhead.Unix(), Nanos: 0},
				},
			},
		},
		{
			Name:              "non-expiring-exported-key",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ExportedCredentials{
				ExportedCredentials: &pb.ExportedCredentials{
					CreationDate:   timestamppb.New(oneYearAgo),
					ExpirationDate: timestamppb.New(nonExpiring),
				},
			},
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewExportedKeyIsTooOldRule()})

	// Expected values are ordered lexicographically.
	want := []*pb.Observation{
		{
			Name:          "expiring-in-one-year-exported-key",
			ExpectedValue: structpb.NewStringValue("sooner expiration date"),
			ObservedValue: structpb.NewStringValue(oneYearAhead.Format("2006-01-02 15:04:05 +0000 UTC")),
		},
		{
			Name:          "non-expiring-exported-key",
			ExpectedValue: structpb.NewStringValue("sooner expiration date"),
			ObservedValue: structpb.NewStringValue(nonExpiring.String()),
		},
	}
	// Sort observations lexicographically by resource name.
	slices.SortStableFunc(got, func(lhs, rhs *pb.Observation) bool {
		return lhs.Resource.Name < rhs.Resource.Name
	})

	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
