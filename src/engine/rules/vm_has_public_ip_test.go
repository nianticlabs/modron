package rules

import (
	"testing"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckVMHasPublicIP(t *testing.T) {
	publicIPResource := &pb.Resource{
		Name:              "public-ip",
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_VmInstance{
			VmInstance: &pb.VmInstance{
				PublicIp: "8.8.8.8",
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
		publicIPResource,
		{
			Name:              "gke-public-ip",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp: "8.8.8.8",
				},
			},
		},
		{
			Name:              "public-ip-automatically-created-34",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp: "8.8.8.8",
				},
			},
		},
		{
			Name:              "no-public-ip",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name:          VMHasPublicIPRuleName,
			ResourceRef:   utils.GetResourceRef(publicIPResource),
			ObservedValue: structpb.NewStringValue("8.8.8.8"),
			ExpectedValue: structpb.NewStringValue("empty"),
			Remediation: &pb.Remediation{
				Description:    "VM \"public-ip\" has a public IP assigned",
				Recommendation: "Compute instances should not be configured to have external IP addresses. Update network-settings of [public-ip](https://console.cloud.google.com/compute/instances?project=project-0). You can connect to Linux VMs that do not have public IP addresses by using Identity-Aware Proxy for TCP forwarding. [Learn more](https://cloud.google.com/compute/docs/instances/connecting-advanced#sshbetweeninstances)",
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewVMHasPublicIPRule()}, want)
}
