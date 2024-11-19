package rules

import (
	"testing"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckDetectsDatabaseAllowsUnencryptedConnections(t *testing.T) {
	databaseNoForceTLS := &pb.Resource{
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
		databaseNoForceTLS,
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
			Name:          DatabaseAllowsUnencryptedConnections,
			ResourceRef:   utils.GetResourceRef(databaseNoForceTLS),
			ObservedValue: structpb.NewBoolValue(false),
			ExpectedValue: structpb.NewBoolValue(true),
			Remediation: &pb.Remediation{
				Description:    "Database database-no-force-tls allows for unencrypted connections.",
				Recommendation: "Enable the require SSL setting in the database settings to allow only encrypted connections to database-no-force-tls.",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewDatabaseAllowsUnencryptedConnectionsRule()}, want)
}
