package rules

import (
	"testing"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckDetectsPrivateGoogleAccessDisabled(t *testing.T) {
	networkNoPrivateAccess := &pb.Resource{
		Name:              "network-no-private-access",
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_Network{
			Network: &pb.Network{
				Ips:                      []string{"8.8.4.4"},
				GcpPrivateGoogleAccessV4: false,
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
		networkNoPrivateAccess,
		{
			Name:              "network-private-access",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					Ips:                      []string{"8.8.8.8"},
					GcpPrivateGoogleAccessV4: true,
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name:          PrivateGoogleAccessDisabled,
			ResourceRef:   utils.GetResourceRef(networkNoPrivateAccess),
			ObservedValue: structpb.NewStringValue("disabled"),
			ExpectedValue: structpb.NewStringValue("enabled"),
			Remediation: &pb.Remediation{
				Description:    "Network [\"network-no-private-access\"](https://console.cloud.google.com/networking/networks/details/network-no-private-access?project=project-0) has [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access) disabled. Private Google Access allows the workloads to access Google APIs via a private network which is safer than going over the public Internet",
				Recommendation: "Enable [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access) for Network [\"network-no-private-access\"](https://console.cloud.google.com/networking/networks/details/network-no-private-access?project=project-0)",
			},
			Severity: pb.Severity_SEVERITY_LOW,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewPrivateGoogleAccessDisabledRule()}, want)
}
