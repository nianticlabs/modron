package rules

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestUnusedExportedKey(t *testing.T) {
	now := time.Now().UTC()
	yesterday := now.Add(time.Hour * -24)
	oneYearAgo := now.Add(-time.Hour * 24 * 365)
	threeMonthsAndOneDay := now.Add(-time.Hour * 24 * 91)
	oneYearAhead := now.Add(time.Hour * 24 * 365)
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
			Name:              "not-used-in-a-long-time",
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
			Name: UnusedExportedCredentials,
			Resource: &pb.Resource{
				Name: "not-used-in-a-long-time",
			},
			ExpectedValue: structpb.NewStringValue(fmt.Sprintf("%s <", oldestUsage.Format(time.RFC3339))),
			ObservedValue: structpb.NewStringValue(threeMonthsAndOneDay.Format(time.RFC3339)),
		},
	}

	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
