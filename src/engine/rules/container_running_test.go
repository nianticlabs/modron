package rules

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func TestCheckContainerNotRunning(t *testing.T) {
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
			Name:              "Cluster",
			DisplayName:       "cluster-display-name",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{},
			},
		},
		{
			Name:              "Namespace",
			Parent:            "Cluster",
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Namespace{
				Namespace: &pb.Namespace{},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name: ContainerRunningRuleName,
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-0"),
				GroupName:     "projects/project-0",
				ExternalId:    proto.String("Cluster"),
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			ObservedValue: structpb.NewStringValue(""),
			ExpectedValue: structpb.NewStringValue("pod-prefix-1-"),
			Remediation: &pb.Remediation{
				Description:    "Cluster cluster-display-name doesn't run any running container starting with pod-prefix-1-",
				Recommendation: "This cluster is likely missing an important part of infrastructure. Check the cluster configuration or reach out to your tech support or SRE.",
			},
			Severity: pb.Severity_SEVERITY_LOW,
		},
		{
			Name: ContainerRunningRuleName,
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-0"),
				GroupName:     "projects/project-0",
				ExternalId:    proto.String("Cluster"),
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			ObservedValue: structpb.NewStringValue(""),
			ExpectedValue: structpb.NewStringValue("pod-prefix-2-"),
			Remediation: &pb.Remediation{
				Description:    "Cluster cluster-display-name doesn't run any running container starting with pod-prefix-2-",
				Recommendation: "This cluster is likely missing an important part of infrastructure. Check the cluster configuration or reach out to your tech support or SRE.",
			},
			Severity: pb.Severity_SEVERITY_LOW,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewContainerRunningRule()}, want)
}
