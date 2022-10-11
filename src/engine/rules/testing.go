package rules

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

const testProjectName = "project-0"

func TestRuleRun(t *testing.T, resources []*pb.Resource, rules []model.Rule) []*pb.Observation {
	t.Helper()
	ctx := context.Background()

	storage := memstorage.New()
	if _, err := storage.BatchCreateResources(ctx, resources); err != nil {
		t.Fatalf("AddResources unexpected error: %v", err)
	}

	allGroups := groupsFromResources(resources)

	obs, err := engine.New(storage, rules).CheckRules(ctx, "", allGroups)
	if err != nil {
		t.Fatalf("CheckRules unexpected error: %v", err)
	}
	return obs
}

func groupsFromResources(resources []*pb.Resource) (allGroups []string) {
	resourceGroups := map[string]struct{}{}
	for _, r := range resources {
		switch r.Type.(type) {
		case *pb.Resource_ResourceGroup:
			resourceGroups[r.Name] = struct{}{}
		}
	}
	for k := range resourceGroups {
		allGroups = append(allGroups, k)
	}
	return allGroups
}

func observationComparer(o1, o2 *pb.Observation) bool {
	if o1 == nil || o2 == nil {
		return false
	}
	if fmt.Sprintf("%T", o1.ExpectedValue) != fmt.Sprintf("%T", o2.ExpectedValue) {
		return false
	}
	if fmt.Sprintf("%T", o1.ObservedValue) != fmt.Sprintf("%T", o2.ObservedValue) {
		return false
	}
	switch o1.ExpectedValue.Kind.(type) {
	case *structpb.Value_StringValue:
		if o1.ExpectedValue.GetStringValue() == o2.ExpectedValue.GetStringValue() && o1.ObservedValue.GetStringValue() == o2.ObservedValue.GetStringValue() {
			return true
		}
		return false
	case *structpb.Value_NumberValue:
		if o1.ExpectedValue.GetNumberValue() == o2.ExpectedValue.GetNumberValue() && o1.ObservedValue.GetNumberValue() == o2.ObservedValue.GetNumberValue() {
			return true
		}
		return false
	case *structpb.Value_BoolValue:
		if o1.ExpectedValue.GetBoolValue() == o2.ExpectedValue.GetBoolValue() && o1.ObservedValue.GetBoolValue() == o2.ObservedValue.GetBoolValue() {
			return true
		}
		return false
	default:
		panic(fmt.Sprintf("comparison for type %T not implemented", o1.ExpectedValue.Kind))
	}
}
