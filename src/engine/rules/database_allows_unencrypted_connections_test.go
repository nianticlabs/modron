package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckDetectsDatabaseAllowsUnencryptedConnections(t *testing.T) {
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
			Name:              "database-no-force-tls",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Database{
				Database: &pb.Database{
					Type:        "cloudsql",
					Version:     "123",
					TlsRequired: false,
				},
			},
		},
		{
			Name:              "database-force-tls",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Database{
				Database: &pb.Database{
					Type:        "cloudsql",
					Version:     "123",
					TlsRequired: true,
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name: DatabaseAllowsUnencryptedConnections,
			Resource: &pb.Resource{
				Name: "database-no-force-tls",
			},
			ObservedValue: structpb.NewBoolValue(false),
			ExpectedValue: structpb.NewBoolValue(true),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewDatabaseAllowsUnencryptedConnectionsRule()})

	// Check that the observations are correct.
	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
