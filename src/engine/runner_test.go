package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/exp/slices"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

type TestRule struct {
	info model.RuleInfo
}

func NewTestRule() *TestRule {
	return &TestRule{
		info: model.RuleInfo{
			Name:                  "TEST_RULE",
			AcceptedResourceTypes: []string{common.ResourceApiKey, common.ResourceServiceAccount},
		},
	}
}

func (r *TestRule) Check(ctx context.Context, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	if strings.Contains(rsrc.Name, "fail") {
		return nil, []error{fmt.Errorf(rsrc.Name)}
	} else {
		return []*pb.Observation{{Name: rsrc.Name, Resource: &pb.Resource{}}}, nil
	}
}

func (r *TestRule) Info() *model.RuleInfo {
	return &r.info
}

type TestRule1 struct {
	model.Rule
	info model.RuleInfo
}

type TestRule2 struct {
	model.Rule
	info model.RuleInfo
}

func NewTestRule1(name string) *TestRule1 {
	return &TestRule1{
		Rule: NewTestRule(),
		info: model.RuleInfo{
			Name:                  name,
			AcceptedResourceTypes: []string{common.ResourceVmInstance, common.ResourceLoadBalancer},
		},
	}
}

func NewTestRule2(name string) *TestRule2 {
	return &TestRule2{
		Rule: NewTestRule(),
		info: model.RuleInfo{
			Name:                  name,
			AcceptedResourceTypes: []string{common.ResourceVmInstance, common.ResourceNetwork},
		},
	}
}

func (r *TestRule1) Info() *model.RuleInfo {
	return &r.info
}

func (r *TestRule2) Info() *model.RuleInfo {
	return &r.info
}

func TestCheckRuleHandlesAllResourcesCorrectly(t *testing.T) {
	storage := memstorage.New()
	rule1 := NewTestRule1("TEST_RULE1")
	rule2 := NewTestRule2("TEST_RULE2")
	resources := []*pb.Resource{
		{
			Name:              "instance-0",
			Parent:            "project-0",
			ResourceGroupName: "project-0",
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp:  "1.1.1.1",
					PrivateIp: "2.2.2.2",
					Identity:  "svc-account-0",
				},
			},
		},
		{
			Name:              "loadbalancer-0",
			Parent:            "project-0",
			ResourceGroupName: "project-0",
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_INTERNAL,
				},
			},
		},
		{
			Name:              "loadbalancer-1",
			Parent:            "project-1",
			ResourceGroupName: "project-1",
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_INTERNAL,
				},
			},
		},
		{
			Name:              "network-0",
			Parent:            "project-0",
			ResourceGroupName: "project-0",
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					Ips:                      []string{"8.8.8.8"},
					GcpPrivateGoogleAccessV4: false,
				},
			},
		},
		{
			Name:              "some-instance-0-fail",
			Parent:            "project-1",
			ResourceGroupName: "project-1",
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp:  "3.3.3.3",
					PrivateIp: "4.4.4.4",
					Identity:  "svc-account-1",
				},
			},
		},
	}
	rules := []model.Rule{rule1, rule2}

	ctx := context.Background()
	if _, err := storage.BatchCreateResources(ctx, resources); err != nil {
		t.Fatalf(`unexpected error: "%v"`, err)
	}
	re := New(storage, rules)

	type Want struct {
		value         string
		isErrorString bool
	}

	want := []Want{}
	for _, rsrc := range resources {
		for _, rule := range rules {
			if ty, err := common.TypeFromResourceAsString(rsrc); err != nil {
				t.Errorf("common.TypeFromResourceAsString unexpected error: %v", err)
			} else if slices.Contains(rule.Info().AcceptedResourceTypes, ty) {
				want = append(want, Want{rsrc.Name, strings.Contains(rsrc.Name, "fail")})
			}
		}
	}
	slices.SortStableFunc(want, func(lhs, rhs Want) bool {
		return lhs.value < rhs.value
	})

	ctx = context.Background()
	obs, errs := re.CheckRules(ctx, "", []string{"project-0", "project-1"})
	if len(obs) != 5 {
		t.Errorf("len(obs) got %d, want %d", len(obs), 5)
	}

	type Got = Want

	got := []Got{}
	for _, ob := range obs {
		got = append(got, Got{
			value:         ob.Name,
			isErrorString: false,
		})
	}
	for _, err := range errs {
		errStr := strings.Split(err.Error(), ": ")
		if len(errStr) != 2 {
			t.Errorf("len(errStr) got %v, want %v", len(errStr), 2)
		}
		got = append(got, Got{
			value:         errStr[1],
			isErrorString: true,
		})
	}
	slices.SortStableFunc(got, func(lhs, rhs Got) bool {
		return lhs.value < rhs.value
	})

	for i := range want {
		if want[i] != got[i] {
			t.Errorf(`unexpected diff, got "%v", want "%v"`, got[i], want[i])
		}
	}
}

func TestCheckRulesHandlesCheckCancellation(t *testing.T) {
	rules := []model.Rule{
		NewTestRule1("TEST_RULE1"),
		NewTestRule2("TEST_RULE2"),
	}
	re := New(memstorage.New(), rules)

	ctx := context.Background()
	ctxWithCancel, cancelFn := context.WithCancel(ctx)

	// Cancel check execution in advance.
	cancelFn()

	// Check rules.
	_, err := re.CheckRules(ctxWithCancel, "", []string{})
	if err == nil {
		t.Errorf(`host.CheckRules got nil, expected error`)
	}

	if len(err) != 2 {
		t.Errorf("len(err) got %d, want 2", len(err))
	}
}
