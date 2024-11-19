package rules

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
	"github.com/nianticlabs/modron/src/utils"
)

const testProjectName = "projects/project-0"

var observationsSorter = func(lhs, rhs *pb.Observation) bool {
	if lhs.ResourceRef == nil || rhs.ResourceRef == nil {
		return lhs.Remediation.Description < rhs.Remediation.Description
	}

	if lhs.ResourceRef.ExternalId == nil || rhs.ResourceRef.ExternalId == nil {
		return lhs.Remediation.Description < rhs.Remediation.Description
	}

	if *lhs.ResourceRef.ExternalId < *rhs.ResourceRef.ExternalId {
		return true
	} else if *lhs.ResourceRef.ExternalId > *rhs.ResourceRef.ExternalId {
		return false
	}
	return lhs.Remediation.Description < rhs.Remediation.Description
}

func mustMarshal[T any](v T) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func testRuleRunHelper(t *testing.T, resources []*pb.Resource, rules []model.Rule) ([]*pb.Observation, []error) {
	t.Helper()
	ctx := context.Background()
	storage := memstorage.New()

	allGroups := utils.GroupsFromResources(resources)

	// Fake a collection event
	collectID := uuid.NewString()
	for i := range resources {
		resources[i].CollectionUid = collectID
	}
	// Fake a collection completed event, otherwise the resources cannot be found
	now := time.Now()
	for _, group := range allGroups {
		err := storage.AddOperationLog(ctx, []*pb.Operation{{
			Id:            collectID,
			ResourceGroup: group,
			Type:          "collection",
			StatusTime:    timestamppb.New(now),
			Status:        pb.Operation_STARTED,
			Reason:        "",
		}})
		if err != nil {
			t.Fatalf("AddOperationLog unexpected error: %v", err)
		}
	}
	// Flush Ops Log
	if err := storage.FlushOpsLog(ctx); err != nil {
		t.Fatalf("FlushOpsLog unexpected error: %v", err)
	}
	if _, err := storage.BatchCreateResources(ctx, resources); err != nil {
		t.Fatalf("AddResources unexpected error: %v", err)
	}
	end := time.Now()
	for _, group := range allGroups {
		err := storage.AddOperationLog(ctx, []*pb.Operation{{
			Id:            collectID,
			ResourceGroup: group,
			Type:          "collection",
			StatusTime:    timestamppb.New(end),
			Status:        pb.Operation_COMPLETED,
			Reason:        "",
		}})
		if err != nil {
			t.Fatalf("AddOperationLog unexpected error: %v", err)
		}
	}
	// Flush Ops Log
	if err := storage.FlushOpsLog(ctx); err != nil {
		t.Fatalf("FlushOpsLog unexpected error: %v", err)
	}

	scanID := uuid.NewString()
	e, err := engine.New(storage, rules, map[string]json.RawMessage{
		"CONTAINER_NOT_RUNNING": mustMarshal(ContainerRunningConfig{
			RequiredContainers: map[string][]string{
				"namespace-1": {"pod-prefix-1-", "pod-prefix-2-"},
			},
		}),
	}, []string{}, risk.TagConfig{
		Environment:  "111111111111/environment",
		EmployeeData: "111111111111/employee_data",
		CustomerData: "111111111111/customer_data",
	})
	if err != nil {
		t.Fatalf("New unexpected error: %v", err)
	}
	return e.CheckRules(ctx, scanID, "", allGroups, nil)
}

func errorStrings(errs []error) []string {
	var errStrs []string
	for _, err := range errs {
		errStrs = append(errStrs, err.Error())
	}
	return errStrs
}

func TestRuleShouldFail(t *testing.T, resources []*pb.Resource, rules []model.Rule, expectedErr []error) {
	_, err := testRuleRunHelper(t, resources, rules)
	if diff := cmp.Diff(errorStrings(expectedErr), errorStrings(err)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}

func TestRuleRun(t *testing.T, resources []*pb.Resource, rules []model.Rule, want []*pb.Observation) {
	got, errArr := testRuleRunHelper(t, resources, rules)
	if len(errArr) > 0 {
		t.Fatalf("CheckRules unexpected error: %v", errors.Join(errArr...))
	}
	for _, obs := range got {
		if obs.Uid == "" {
			t.Errorf("CheckRules unexpected empty UID")
		}
		if obs.Timestamp == nil {
			t.Errorf("CheckRules unexpected nil timestamp")
		}
		if obs.Severity == pb.Severity_SEVERITY_UNKNOWN {
			t.Errorf("CheckRules unexpected unknown severity")
		}
	}

	// We add some fields to `want` so that we don't have to change the test data every time
	// we modify one "meta" field:
	for i, ob := range want {
		ob.Source = pb.Observation_SOURCE_MODRON
		ob.RiskScore = ob.Severity
		ob.Impact = pb.Impact_IMPACT_MEDIUM
		want[i] = ob
	}

	if diff := cmp.Diff(want, got, protocmp.Transform(), protocmp.IgnoreFields(
		&pb.Observation{},
		"timestamp",
		"uid",
		"scan_uid",
	),
		protocmp.IgnoreFields(&pb.ResourceRef{}, "uid"),
		protocmp.IgnoreUnknown(),
		cmpopts.SortSlices(observationsSorter),
	); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
