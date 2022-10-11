package rules

import (
	"context"
	"testing"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

type TestRule struct {
	info model.RuleInfo
}

func NewTestRule(name string) *TestRule {
	return &TestRule{
		info: model.RuleInfo{
			Name:                  name,
			AcceptedResourceTypes: []string{},
		},
	}
}

func (r *TestRule) Check(ctx context.Context, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	return []*pb.Observation{{}}, nil
}

func (r *TestRule) Info() *model.RuleInfo {
	return &r.info
}

func TestAddAndGetRule(t *testing.T) {
	want := NewTestRule("TEST_RULE_3")
	AddRule(want)

	got, err := GetRule("TEST_RULE_3")
	if err != nil {
		t.Errorf(`GetRule unexpected error "%v"`, err)
	}
	if got != want {
		t.Errorf(`GetRule unexpected diff (-want, +got): "-%v", "+%v"`, want, got)
	}
}

// TODO: Make test state-independent (i.e., irrespective of the presence of rules in the registry).
func TestGetRules(t *testing.T) {
	rulesToAdd := []model.Rule{NewTestRule("TEST_RULE1"), NewTestRule("TEST_RULE2")}
	for _, rule := range rulesToAdd {
		AddRule(rule)
	}

	rules := GetRules()

	for _, addedRule := range rulesToAdd {
		found := false

		for _, rule := range rules {
			if rule == addedRule {
				found = true
				break
			}
		}
		if !found {
			t.Errorf(`GetRules() does not contain rule "%v"`, addedRule.Info().Name)
		}
	}
}
