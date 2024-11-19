package rules

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

func TestMasterAuthorizedNetworksNotSet(t *testing.T) {
	notSetResourceName := "master-authorized-networks-not-set"
	notSetResource := &pb.Resource{
		Name:              notSetResourceName,
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_KubernetesCluster{
			KubernetesCluster: &pb.KubernetesCluster{
				MasterAuthorizedNetworks: []string{},
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
		{
			Name:              "master-authorized-networks-set",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					MasterAuthorizedNetworks: []string{"10.0.0.0/8"},
				},
			},
		},
		notSetResource,
	}

	want := []*pb.Observation{
		{
			Name:          MasterAuthorizedNetworksNotSet,
			ResourceRef:   utils.GetResourceRef(notSetResource),
			ExpectedValue: structpb.NewStringValue("not empty"),
			ObservedValue: structpb.NewStringValue("empty"),
			Remediation: &pb.Remediation{
				Description:    "Cluster [\"master-authorized-networks-not-set\"](https://console.cloud.google.com/kubernetes/list/overview?project=project-0) does not have a [Master Authorized Network](https://cloud.google.com/kubernetes-engine/docs/how-to/authorized-networks#create_cluster) set. Without this setting, the cluster control plane is accessible to anyone",
				Recommendation: "Set a [Master Authorized Network](https://cloud.google.com/kubernetes-engine/docs/how-to/authorized-networks#create_cluster) network range for cluster [\"master-authorized-networks-not-set\"](https://console.cloud.google.com/kubernetes/list/overview?project=project-0)",
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewMasterAuthorizedNetworksNotSetRule()}, want)
}
