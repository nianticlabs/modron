package gcpcollector

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/compute/v1"
	"google.golang.org/protobuf/testing/protocmp"

	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

type testSlowGCPAPI struct {
	GCPApi
}

func (t *testSlowGCPAPI) ListRegions(_ context.Context, _ string) ([]*compute.Region, error) {
	return []*compute.Region{
		{
			Name: "us-central1",
		},
		{
			Name: "us-west1",
		},
		{
			Name: "us-east1",
		},
	}, nil
}

func (t *testSlowGCPAPI) ListSubNetworksByRegion(_ context.Context, _ string, region string) ([]*compute.Subnetwork, error) {
	time.Sleep(1 * time.Second)
	return []*compute.Subnetwork{
		{
			Name:    "subnet-" + region,
			Purpose: "PRIVATE",
		},
	}, nil
}

func TestListNetworks(t *testing.T) {
	ctx := context.Background()
	st := memstorage.New()
	tagConfig := risk.TagConfig{
		Environment:  "111111111111/environment",
		EmployeeData: "111111111111/employee_data",
		CustomerData: "111111111111/customer_data",
	}
	coll, err := New(ctx, st, "111111", "example.com", []string{}, tagConfig, []string{})
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}
	gcpColl := coll.(*GCPCollector)
	gcpColl.api = &testSlowGCPAPI{}
	got, err := coll.(*GCPCollector).ListNetworks(ctx, "projects/test-rg")
	if err != nil {
		t.Fatalf("failed to list networks: %v", err)
	}
	sort.Sort(resourceByName(got))
	want := []*pb.Resource{
		{
			Name:              "subnet-us-central1",
			Parent:            "projects/test-rg",
			ResourceGroupName: "projects/test-rg",
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					Ips:                      []string{""},
					GcpPrivateGoogleAccessV4: false,
				},
			},
		},
		{
			Name:              "subnet-us-east1",
			Parent:            "projects/test-rg",
			ResourceGroupName: "projects/test-rg",
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					Ips:                      []string{""},
					GcpPrivateGoogleAccessV4: false,
				},
			},
		},
		{
			Name:              "subnet-us-west1",
			Parent:            "projects/test-rg",
			ResourceGroupName: "projects/test-rg",
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					Ips:                      []string{""},
					GcpPrivateGoogleAccessV4: false,
				},
			},
		},
	}
	if diff := cmp.Diff(want, got, protocmp.Transform(), protocmp.IgnoreFields(&pb.Resource{}, "uid")); diff != "" {
		t.Errorf("ListNetworks() mismatch (-want +got):\n%s", diff)
	}
}

type resourceByName []*pb.Resource

func (r resourceByName) Len() int {
	return len(r)
}

func (r resourceByName) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

func (r resourceByName) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

var _ sort.Interface = resourceByName(nil)
