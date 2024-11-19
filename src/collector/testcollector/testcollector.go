package testcollector

import (
	"fmt"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"

	"golang.org/x/net/context"
)

var _ model.Collector = (*TestCollector)(nil)

var errNotImplemented = fmt.Errorf("not implemented")

type TestCollector struct{}

func (t TestCollector) CollectAndStoreAll(_ context.Context, _ string, _ []string, _ []*pb.Resource) error {
	return errNotImplemented
}

func (t TestCollector) ListResourceGroupObservations(_ context.Context, _ string, _ string) ([]*pb.Observation, []error) {
	return nil, []error{errNotImplemented}
}

func (t TestCollector) GetResourceGroupWithIamPolicy(_ context.Context, _ string, _ string) (*pb.Resource, error) {
	return nil, errNotImplemented
}

func (t TestCollector) ListResourceGroups(_ context.Context, _ []string) ([]*pb.Resource, error) {
	return nil, errNotImplemented
}

func (t TestCollector) ListResourceGroupsWithIamPolicies(_ context.Context, _ []string) ([]*pb.Resource, error) {
	return nil, errNotImplemented
}

func (t TestCollector) ListResourceGroupNames(_ context.Context) ([]string, error) {
	return nil, errNotImplemented
}

func (t TestCollector) ListResourceGroupAdmins(_ context.Context) (model.ACLCache, error) {
	return model.ACLCache{
		"*": {
			"projects/modron-test":  {},
			"projects/super-secret": {},
		},
		"user@example.com": {
			"projects/modron-test": {},
		},
	}, nil
}

func (t TestCollector) ListResourceGroupResources(_ context.Context, _ string, _ string) ([]*pb.Resource, []error) {
	return nil, []error{errNotImplemented}
}
