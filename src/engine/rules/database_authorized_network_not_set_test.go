package rules

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckDetectsDatabaseAuthorizedNetworksNotSet(t *testing.T) {
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
			Name:              "database-no-authorized-networks",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Database{
				Database: &pb.Database{
					Type:                               "cloudsql",
					Version:                            "123",
					AuthorizedNetworksSettingAvailable: pb.Database_AUTHORIZED_NETWORKS_NOT_SET,
				},
			},
		},
		{
			Name:              "database-authorized-networks",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Database{
				Database: &pb.Database{
					Type:                               "cloudsql",
					Version:                            "123",
					AuthorizedNetworksSettingAvailable: pb.Database_AUTHORIZED_NETWORKS_SET,
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name:          "database-no-authorized-networks",
			ObservedValue: structpb.NewStringValue("AUTHORIZED_NETWORKS_NOT_SET"),
			ExpectedValue: structpb.NewStringValue("AUTHORIZED_NETWORKS_SET"),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewDatabaseAuthorizedNetworksNotSetRule()})

	// Check that the observations are correct.
	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer)); diff != "" {
		fmt.Println(want)
		fmt.Println(got)
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
