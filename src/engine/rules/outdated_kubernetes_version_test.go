package rules

import (
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

func TestOutdatedKubernetesVersionDetection(t *testing.T) {
	clusterWithOutdatedNodesVersion := &pb.Resource{
		Name:              "cluster-with-outdated-nodes-version",
		Parent:            testProjectName,
		ResourceGroupName: testProjectName,
		IamPolicy:         &pb.IamPolicy{},
		Type: &pb.Resource_KubernetesCluster{
			KubernetesCluster: &pb.KubernetesCluster{
				PrivateCluster: true,
				MasterVersion:  "1.27.10-gke.600",
				NodesVersion:   "1.15.10-gke.600",
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
			Name:              "up-to-date-k8s-cluster",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					PrivateCluster: true,
					MasterVersion:  "1.27.10-gke.600",
					NodesVersion:   "1.27.10-gke.600",
				},
			},
		},
		clusterWithOutdatedNodesVersion,
	}

	want := []*pb.Observation{
		{
			Name:          OutDatedKubernetesVersion,
			ResourceRef:   utils.GetResourceRef(clusterWithOutdatedNodesVersion),
			ExpectedValue: structpb.NewStringValue(fmt.Sprintf("version > %.2f", currentK8sVersion)),
			ObservedValue: structpb.NewStringValue("1.15.10-gke.600"),
			Remediation: &pb.Remediation{
				Description:    "Cluster [\"cluster-with-outdated-nodes-version\"](https://console.cloud.google.com/kubernetes/list/overview?project=project-0) uses an outdated Kubernetes version",
				Recommendation: "Update the Kubernetes version on cluster [\"cluster-with-outdated-nodes-version\"](https://console.cloud.google.com/kubernetes/list/overview?project=project-0) to at least 1.27. For more details on this process, see [this article](https://cloud.google.com/kubernetes-engine/docs/how-to/upgrading-a-cluster)",
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewOutDatedKubernetesVersionRule()}, want)
}
