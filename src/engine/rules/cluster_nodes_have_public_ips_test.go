package rules

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func TestPublicClusterNodesDetection(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "",
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "private-cluster",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					PrivateCluster: true,
				},
			},
		},
		{
			Name:              "public-cluster",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					PrivateCluster: false,
				},
			},
		},
	}

	// Expected values are ordered lexicographically.
	want := []*pb.Observation{
		{
			Name: ClusterNodesHavePublicIps,
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-0"),
				GroupName:     "projects/project-0",
				ExternalId:    proto.String("public-cluster"),
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			ExpectedValue: structpb.NewStringValue("private"),
			ObservedValue: structpb.NewStringValue("public"),
			Remediation: &pb.Remediation{
				Description:    "Cluster [\"public-cluster\"](https://console.cloud.google.com/kubernetes/list/overview?project=project-0) has a public IP, which could make it accessible by anyone on the internet",
				Recommendation: "Unless strictly needed, redeploy cluster [\"public-cluster\"](https://console.cloud.google.com/kubernetes/list/overview?project=project-0) as a [private cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/private-clusters)",
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewClusterNodesHavePublicIpsRule()}, want)
}
