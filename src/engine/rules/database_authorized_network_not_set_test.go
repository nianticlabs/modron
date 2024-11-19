package rules

import (
	"testing"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckDetectsDatabaseAuthorizedNetworksNotSet(t *testing.T) {
	databasePublicAndNoAuthorizedNetworks := &pb.Resource{
		Name:              "database-public-and-no-authorized-networks",
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_Database{
			Database: &pb.Database{
				Type:                               "cloudsql",
				Version:                            "123",
				AuthorizedNetworksSettingAvailable: pb.Database_AUTHORIZED_NETWORKS_NOT_SET,
				IsPublic:                           true,
			},
		},
	}
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
		databasePublicAndNoAuthorizedNetworks,
		{
			Name:              "database-private-no-authorized-networks",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Database{
				Database: &pb.Database{
					Type:                               "cloudsql",
					Version:                            "123",
					AuthorizedNetworksSettingAvailable: pb.Database_AUTHORIZED_NETWORKS_NOT_SET,
					IsPublic:                           false,
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
			Name:          DatabaseAuthorizedNetworksNotSet,
			ResourceRef:   utils.GetResourceRef(databasePublicAndNoAuthorizedNetworks),
			ObservedValue: structpb.NewStringValue("AUTHORIZED_NETWORKS_NOT_SET"),
			ExpectedValue: structpb.NewStringValue("AUTHORIZED_NETWORKS_SET"),
			Remediation: &pb.Remediation{
				Description:    "Database database-public-and-no-authorized-networks is reachable from any IP on the Internet.",
				Recommendation: "Enable the authorized network setting in the database settings to restrict what networks can access database-public-and-no-authorized-networks.",
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewDatabaseAuthorizedNetworksNotSetRule()}, want)
}
