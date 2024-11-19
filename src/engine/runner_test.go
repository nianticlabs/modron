package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
	"github.com/nianticlabs/modron/src/utils"
)

type TestRule struct {
	info model.RuleInfo
}

var impactMap = map[string]pb.Impact{
	"prod":       pb.Impact_IMPACT_HIGH,
	"pre-prod":   pb.Impact_IMPACT_MEDIUM,
	"dev":        pb.Impact_IMPACT_LOW,
	"playground": pb.Impact_IMPACT_LOW,
}

func NewTestRule() *TestRule {
	return &TestRule{
		info: model.RuleInfo{
			Name:                  "TEST_RULE",
			AcceptedResourceTypes: []proto.Message{&pb.APIKey{}, &pb.ServiceAccount{}},
		},
	}
}

func (r *TestRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	if strings.Contains(rsrc.Name, "fail") {
		return nil, []error{fmt.Errorf("%s", rsrc.Name)}
	}
	return []*pb.Observation{{
		Uid:  uuid.NewString(),
		Name: rsrc.Name,
		ResourceRef: &pb.ResourceRef{
			Uid:           proto.String(rsrc.Uid),
			GroupName:     rsrc.ResourceGroupName,
			CloudPlatform: pb.CloudPlatform_GCP,
			ExternalId:    nil,
		},
		Severity: pb.Severity_SEVERITY_LOW,
	}}, nil
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

type TestRuleBucketPublic struct{}

func (t TestRuleBucketPublic) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	if rsrc.GetBucket().GetAccessType() == pb.Bucket_PUBLIC {
		obs = append(obs, &pb.Observation{
			Uid:      uuid.NewString(),
			Name:     rsrc.Name,
			Severity: pb.Severity_SEVERITY_HIGH,
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String(rsrc.Uid),
				GroupName:     rsrc.ResourceGroupName,
				CloudPlatform: pb.CloudPlatform_GCP,
			},
		})
	}
	return
}

func (t TestRuleBucketPublic) Info() *model.RuleInfo {
	return &model.RuleInfo{
		Name:                  "TEST_RULE_BUCKET_PUBLIC",
		AcceptedResourceTypes: []proto.Message{&pb.Bucket{}},
	}
}

var _ model.Rule = (*TestRuleBucketPublic)(nil)

func NewTestRule1(name string) *TestRule1 {
	return &TestRule1{
		Rule: NewTestRule(),
		info: model.RuleInfo{
			Name:                  name,
			AcceptedResourceTypes: []proto.Message{&pb.VmInstance{}, &pb.LoadBalancer{}},
		},
	}
}

func NewTestRule2(name string) *TestRule2 {
	return &TestRule2{
		Rule: NewTestRule(),
		info: model.RuleInfo{
			Name:                  name,
			AcceptedResourceTypes: []proto.Message{&pb.VmInstance{}, &pb.Network{}},
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
	collectionUID := uuid.NewString()
	resources := []*pb.Resource{
		{
			Name:              "projects/project-0",
			Parent:            "folders/folder-0",
			ResourceGroupName: "projects/project-0",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: "projects/project-0",
					Name:       "project-0",
				},
			},
			Tags: map[string]string{
				"111111111111/customer_data": "111111111111/customer_data/yes",
			},
		},
		{
			Name:              "folders/folder-0",
			Parent:            "organizations/1",
			ResourceGroupName: "folders/folder-0",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: "folders/folder-0",
					Name:       "Dev",
				},
			},
			Tags: map[string]string{
				"111111111111/environment":   "111111111111/environment/development",
				"111111111111/employee_data": "111111111111/employee_data/no",
				"111111111111/customer_data": "111111111111/customer_data/no",
			},
		},
		{
			Name:              "organizations/1",
			Parent:            "",
			ResourceGroupName: "organizations/1",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: "organizations/1",
					Name:       "ACME Inc.",
				},
			},
			Tags: map[string]string{
				"111111111111/environment":   "111111111111/environment/production",
				"111111111111/customer_data": "111111111111/customer_data/yes",
				"111111111111/employee_data": "111111111111/employee_data/yes",
			},
		},
		{
			Name:              "instance-0",
			Parent:            "projects/project-0",
			ResourceGroupName: "projects/project-0",
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
			Parent:            "projects/project-0",
			ResourceGroupName: "projects/project-0",
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_INTERNAL,
				},
			},
		},
		{
			Name:              "loadbalancer-1",
			Parent:            "projects/project-1",
			ResourceGroupName: "projects/project-1",
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_INTERNAL,
				},
			},
		},
		{
			Name:              "network-0",
			Parent:            "projects/project-0",
			ResourceGroupName: "projects/project-0",
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					Ips:                      []string{"8.8.8.8"},
					GcpPrivateGoogleAccessV4: false,
				},
			},
		},
		{
			Name:              "some-instance-0-fail",
			Parent:            "projects/project-1",
			ResourceGroupName: "projects/project-1",
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp:  "3.3.3.3",
					PrivateIp: "4.4.4.4",
					Identity:  "svc-account-1",
				},
			},
		},
	}
	rgNames := map[string]struct{}{}
	for i, res := range resources {
		resources[i].CollectionUid = collectionUID
		rgNames[res.ResourceGroupName] = struct{}{}
	}
	rules := []model.Rule{rule1, rule2}

	ctx := context.Background()
	registerCollectOperation(ctx, t, resources, storage, collectionUID)
	if _, err := storage.BatchCreateResources(ctx, resources); err != nil {
		t.Fatalf(`unexpected error: "%v"`, err)
	}

	re, _ := New(storage, rules, map[string]json.RawMessage{}, []string{}, risk.TagConfig{
		Environment:  "111111111111/environment",
		EmployeeData: "111111111111/employee_data",
		CustomerData: "111111111111/customer_data",
	})
	type Want struct {
		value         string
		isErrorString bool
	}

	var want []Want
	for _, rsrc := range resources {
		for _, rule := range rules {
			if ty, err := utils.TypeFromResource(rsrc); err != nil {
				t.Errorf("common.TypeFromResource unexpected error: %v", err)
			} else if slices.Contains(utils.ProtoAcceptsTypes(rule.Info().AcceptedResourceTypes), ty) {
				want = append(want, Want{rsrc.Name, strings.Contains(rsrc.Name, "fail")})
			}
		}
	}
	slices.SortStableFunc(want, func(lhs, rhs Want) int {
		if lhs.value < rhs.value {
			return -1
		}
		if lhs.value > rhs.value {
			return 1
		}
		return 0
	})

	obs, errs := re.CheckRules(ctx, uuid.NewString(), collectionUID, []string{"projects/project-0", "projects/project-1"}, nil)
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
	slices.SortStableFunc(got, func(lhs, rhs Got) int {
		if lhs.value < rhs.value {
			return -1
		}
		if lhs.value > rhs.value {
			return 1
		}
		return 0
	})

	for i := range want {
		if want[i] != got[i] {
			t.Errorf(`unexpected diff, got "%v", want "%v"`, got[i], want[i])
		}
	}
}

func TestCheckRuleSeverity(t *testing.T) {
	st := memstorage.New()
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	ctx := context.Background()
	collID := uuid.NewString()

	resources := []*pb.Resource{
		{
			Name:              "organizations/1",
			Parent:            "",
			ResourceGroupName: "organizations/1",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: "organizations/1",
					Name:       "ACME Inc.",
				},
			},
			Tags: map[string]string{
				"111111111111/environment":   "111111111111/environment/prod",
				"111111111111/customer_data": "111111111111/customer_data/yes",
				"111111111111/employee_data": "111111111111/employee_data/yes",
			},
			CollectionUid: collID,
		},
		{
			Name:              "folders/dev-folder",
			Parent:            "organizations/1",
			ResourceGroupName: "folders/dev-folder",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: "folders/dev-folder",
					Name:       "Dev",
				},
			},
			Tags: map[string]string{
				"111111111111/environment":   "111111111111/environment/dev",
				"111111111111/customer_data": "111111111111/customer_data/no",
				"111111111111/employee_data": "111111111111/employee_data/no",
			},
		},
		{
			Name:              "folders/pre-prod-folder",
			Parent:            "organizations/1",
			ResourceGroupName: "folders/pre-prod-folder",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: "folders/pre-prod-folder",
					Name:       "Pre-Prod",
				},
			},
			Tags: map[string]string{
				"111111111111/environment":   "111111111111/environment/pre-prod",
				"111111111111/customer_data": "111111111111/customer_data/yes",
				"111111111111/employee_data": "111111111111/employee_data/no",
			},
		},
		{
			Name:              "projects/project-0",
			Parent:            "folders/dev-folder",
			ResourceGroupName: "projects/project-0",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: "projects/project-0",
					Name:       "project-0",
				},
			},
			Tags: map[string]string{
				"111111111111/customer_data": "111111111111/customer_data/no",
			},
		},
		{
			Name:              "bucket-0",
			Parent:            "projects/project-0",
			ResourceGroupName: "projects/project-0",
			Type: &pb.Resource_Bucket{
				Bucket: &pb.Bucket{
					AccessType: pb.Bucket_PUBLIC,
				},
			},
		},
		{
			Name:              "projects/project-1",
			Parent:            "folders/pre-prod-folder",
			ResourceGroupName: "projects/project-1",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: "projects/project-1",
					Name:       "project-1",
				},
			},
			Tags: map[string]string{
				"111111111111/customer_data": "111111111111/customer_data/no",
				"111111111111/employee_data": "111111111111/employee_data/no",
			},
		},
		{
			Name:              "bucket-1",
			Parent:            "projects/project-1",
			ResourceGroupName: "projects/project-1",
			Type: &pb.Resource_Bucket{
				Bucket: &pb.Bucket{
					AccessType: pb.Bucket_PUBLIC,
				},
			},
		},
	}
	for k := range resources {
		resources[k].CollectionUid = collID
	}
	registerCollectOperation(ctx, t, resources, st, collID)
	if _, err := st.BatchCreateResources(ctx, resources); err != nil {
		t.Fatalf(`unexpected error: "%v"`, err)
	}
	_ = st.FlushOpsLog(ctx)
	engine, _ := New(st, []model.Rule{TestRuleBucketPublic{}}, map[string]json.RawMessage{}, []string{}, risk.TagConfig{
		ImpactMap:    impactMap,
		Environment:  "111111111111/environment",
		EmployeeData: "111111111111/employee_data",
		CustomerData: "111111111111/customer_data",
	})
	scanID := uuid.NewString()
	got, err := engine.CheckRules(ctx, scanID, collID, []string{"projects/project-0", "projects/project-1"}, resources)
	if err != nil {
		t.Fatalf(`unexpected error: "%v"`, err)
	}
	want := []*pb.Observation{
		{
			Name:    "bucket-0",
			ScanUid: proto.String(scanID),
			Source:  pb.Observation_SOURCE_MODRON,
			ResourceRef: &pb.ResourceRef{
				GroupName:     "projects/project-0",
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			Impact:       pb.Impact_IMPACT_LOW,
			ImpactReason: "environment=dev",
			Severity:     pb.Severity_SEVERITY_HIGH,
			RiskScore:    pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name:    "bucket-1",
			ScanUid: proto.String(scanID),
			Source:  pb.Observation_SOURCE_MODRON,
			ResourceRef: &pb.ResourceRef{
				GroupName:     "projects/project-1",
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			Impact:       pb.Impact_IMPACT_MEDIUM,
			ImpactReason: "environment=pre-prod",
			Severity:     pb.Severity_SEVERITY_HIGH,
			RiskScore:    pb.Severity_SEVERITY_HIGH,
		},
	}
	if diff := cmp.Diff(want, got, protocmp.Transform(),
		protocmp.IgnoreFields(&pb.ResourceRef{}, "uid"),
		protocmp.IgnoreFields(&pb.Observation{}, "uid"),
	); diff != "" {
		t.Errorf(`unexpected diff (-want +got): %s`, diff)
	}
}

func registerCollectOperation(ctx context.Context, t *testing.T, resources []*pb.Resource, storage model.Storage, collectionUID string) {
	now := time.Now()
	for _, group := range utils.GroupsFromResources(resources) {
		err := storage.AddOperationLog(ctx, []*pb.Operation{{
			Id:            collectionUID,
			ResourceGroup: group,
			Type:          "collection",
			Status:        pb.Operation_STARTED,
			StatusTime:    timestamppb.New(now),
		}})
		if err != nil {
			t.Fatalf(`unexpected error: "%v"`, err)
		}
	}
	// Flush
	if err := storage.FlushOpsLog(ctx); err != nil {
		t.Fatalf(`unexpected error: "%v"`, err)
	}
	now = time.Now()
	for _, group := range utils.GroupsFromResources(resources) {
		err := storage.AddOperationLog(ctx, []*pb.Operation{{
			Id:            collectionUID,
			ResourceGroup: group,
			Type:          "collection",
			Status:        pb.Operation_COMPLETED,
			StatusTime:    timestamppb.New(now),
		}})
		if err != nil {
			t.Fatalf(`unexpected error: "%v"`, err)
		}
	}
	// Flush
	if err := storage.FlushOpsLog(ctx); err != nil {
		t.Fatalf(`unexpected error: "%v"`, err)
	}
}

func getRecursiveResource(recRes *pb.RecursiveResource, children []*pb.RecursiveResource) *pb.RecursiveResource {
	newRecRes := proto.Clone(recRes).(*pb.RecursiveResource)
	newRecRes.Children = children
	return newRecRes
}

func TestGetImpact(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	org1 := &pb.RecursiveResource{Name: "organizations/1",
		DisplayName: "ACME Inc.",
		Type:        "ResourceGroup",
		Parent:      "",
		Tags: map[string]string{
			"111111111111/environment":   "111111111111/environment/prod",
			"111111111111/customer_data": "111111111111/customer_data/yes",
			"111111111111/employee_data": "111111111111/employee_data/yes",
		},
		Children: []*pb.RecursiveResource{
			{
				Name:        "folders/dev-folder",
				DisplayName: "Dev",
				Type:        "ResourceGroup",
				Parent:      "organizations/1",
				Labels: map[string]string{
					"111111111111/environment":   "111111111111/environment/dev",
					"111111111111/customer_data": "111111111111/customer_data/no",
					"111111111111/employee_data": "111111111111/employee_data/no",
				},
			},
		},
	}
	folder1 := &pb.RecursiveResource{
		Name:        "folders/dev-folder",
		DisplayName: "Dev",
		Type:        "ResourceGroup",
		Parent:      "organizations/1",
		Tags: map[string]string{
			"111111111111/environment":   "111111111111/environment/dev",
			"111111111111/customer_data": "111111111111/customer_data/no",
			"111111111111/employee_data": "111111111111/employee_data/no",
		},
	}
	folder2 := &pb.RecursiveResource{
		Name:        "folders/prod-folder",
		DisplayName: "111111111111/environment/prod",
		Type:        "ResourceGroup",
		Parent:      "organizations/1",
		Tags: map[string]string{
			"111111111111/environment":   "111111111111/environment/prod",
			"111111111111/customer_data": "111111111111/customer_data/yes",
			"111111111111/employee_data": "111111111111/employee_data/yes",
		},
	}
	prj0 := &pb.RecursiveResource{
		Name:        "projects/project-0",
		DisplayName: "Project 0",
		Type:        "ResourceGroup",
		Parent:      "folders/dev-folder",
		Tags: map[string]string{
			"111111111111/customer_data": "111111111111/customer_data/no",
		},
	}
	prj1 := &pb.RecursiveResource{
		Name:        "projects/project-1",
		DisplayName: "Project 1",
		Type:        "ResourceGroup",
		Parent:      "folders/prod-folder",
		Labels: map[string]string{
			"contact1": "alice@example.com",
			"contact2": "bob@example.com",
		},
	}
	prj2 := &pb.RecursiveResource{
		Name:        "projects/project-2",
		DisplayName: "Project 2",
		Type:        "ResourceGroup",
		Parent:      "folders/dev-folder",
		Tags: map[string]string{
			"111111111111/customer_data": "111111111111/customer_data/yes",
		},
	}
	prj3 := &pb.RecursiveResource{
		Name:        "projects/project-3",
		DisplayName: "Project 3",
		Type:        "ResourceGroup",
		Parent:      "folders/dev-folder",
		Tags: map[string]string{
			"111111111111/employee_data": "111111111111/employee_data/yes",
		},
	}
	// A project that is a direct child of the organization, with no default labels
	prj4 := &pb.RecursiveResource{
		Name:        "projects/project-4",
		DisplayName: "Project 4",
		Type:        "ResourceGroup",
		Parent:      "organizations/1",
	}

	rgHierarchy := map[string]*pb.RecursiveResource{
		"":                    org1,
		"organizations/1":     getRecursiveResource(org1, []*pb.RecursiveResource{prj4}),
		"folders/dev-folder":  getRecursiveResource(folder1, []*pb.RecursiveResource{prj0, prj2, prj3}),
		"folders/prod-folder": getRecursiveResource(folder2, []*pb.RecursiveResource{prj1}),
		"projects/project-0":  prj0,
		"projects/project-1":  prj1,
		"projects/project-2":  prj2,
		"projects/project-3":  prj3,
		"projects/project-4":  prj4,
	}

	testForImpact(t, rgHierarchy, "projects/project-0", pb.Impact_IMPACT_LOW)
	testForImpact(t, rgHierarchy, "projects/project-1", constants.ImpactCustomerData)
	testForImpact(t, rgHierarchy, "projects/project-2", pb.Impact_IMPACT_HIGH)
	testForImpact(t, rgHierarchy, "projects/project-3", constants.ImpactEmployeeData)
	testForImpact(t, rgHierarchy, "projects/project-4", pb.Impact_IMPACT_HIGH)
}

func testForImpact(t *testing.T, rgHierarchy map[string]*pb.RecursiveResource, rgName string, want pb.Impact) {
	t.Helper()
	got, _ := risk.GetImpact(risk.TagConfig{
		ImpactMap:    impactMap,
		Environment:  "111111111111/environment",
		EmployeeData: "111111111111/employee_data",
		CustomerData: "111111111111/customer_data",
	}, rgHierarchy, rgName)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf(`unexpected diff for %s (-want +got): %s`, rgName, diff)
	}
}

// TODO fix flaky test.
// func TestCheckRulesHandlesCheckCancellation(t *testing.T) {
// 	rules := []model.Rule{
// 		NewTestRule1("TEST_RULE1"),
// 		NewTestRule2("TEST_RULE2"),
// 	}
// 	re := New(memstorage.New(), rules)

// 	ctx, cancelFn := context.WithCancel(context.Background())
// 	// Cancel check execution in advance.
// 	cancelFn()

// 	// Check rules.
// 	_, err := re.CheckRules(ctx, "", []string{})
// 	if err == nil {
// 		t.Errorf(`host.CheckRules got nil, expected error`)
// 	}

// 	if len(err) != 2 {
// 		t.Errorf("len(err) got %d, want 2", len(err))
// 	}
// }
