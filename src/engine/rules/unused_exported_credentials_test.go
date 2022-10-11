package rules

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestUnusedExportedKey(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(time.Hour * -24)
	oneYearAgo := now.Add(-time.Hour * 24 * 365)
	threeMonthsAndOneDay := now.Add(-time.Hour * 24 * 91)
	oneYearAhead := now.Add(time.Hour * 24 * 365)
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
			Name:              "not-used-in-a-year",
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
		},
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
	}

	got := TestRuleRun(t, resources, []model.Rule{NewUnusedExportedCredentialsRule()})

	// Expected values are ordered lexicographically.
	want := []*pb.Observation{
		{
			Name:          "not-used-in-a-year",
			ExpectedValue: structpb.NewStringValue(fmt.Sprintf("%s <", oldestUsage.Format(time.RFC3339))),
			ObservedValue: structpb.NewStringValue(threeMonthsAndOneDay.Format(time.RFC3339)),
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
