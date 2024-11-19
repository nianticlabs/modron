//go:build integration

package gcpcollector_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"k8s.io/api/core/v1"

	"github.com/nianticlabs/modron/src/collector/gcpcollector"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
	"github.com/nianticlabs/modron/src/utils"
)

func getKubernetesClusterCollector(t *testing.T) (*pb.Resource, *gcpcollector.GCPCollector) {
	t.Helper()
	ctx := context.Background()
	storage := memstorage.New()
	orgID := os.Getenv("ORG_ID")
	projectID := os.Getenv("PROJECT_ID")
	if orgID == "" {
		t.Skip("ORG_ID not set, skipping")
	}
	if projectID == "" {
		t.Skip("PROJECT_ID not set, skipping")
	}
	orgSuffix := ""
	tagConfig := risk.TagConfig{}
	coll, err := gcpcollector.New(ctx, storage, orgID, orgSuffix, []string{}, tagConfig, []string{})
	if err != nil {
		t.Fatal(err)
	}
	rsrc := &pb.Resource{
		Name:              "projects/" + projectID,
		ResourceGroupName: "projects/xyz",
		Type: &pb.Resource_ResourceGroup{
			ResourceGroup: &pb.ResourceGroup{
				Identifier: projectID,
			},
		},
	}
	return rsrc, coll.(*gcpcollector.GCPCollector)
}

func TestKubernetesCluster_ListKubernetesPods(t *testing.T) {
	rsrc, coll := getKubernetesClusterCollector(t)
	res, err := coll.ListKubernetesPods(context.Background(), rsrc.Name)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) == 0 {
		t.Fatal("No kubernetes pods found")
	}

	for _, p := range res {
		pod := p.GetPod()
		if pod == nil {
			t.Fatal("expected pod")
		}
		fmt.Printf("Pod %s\n", utils.GetHumanReadableName(p.Name))
		if pod.Spec == nil {
			t.Fatal("pod spec cannot be nil")
		}

		fmt.Printf("\tNamespace\t%s\n", utils.GetHumanReadableName(pod.Namespace))

		if len(pod.Spec.Containers) == 0 {
			t.Fatal("expected at least one container")
		}

		for _, c := range pod.Spec.InitContainers {
			fmt.Printf("\tInit Container\t%s\n", c.Name)
			verifyEnvVarsAreRedacted(t, c.Env)
		}
		for _, c := range pod.Spec.Containers {
			fmt.Printf("\tContainer\t%s\n", c.Name)
			verifyEnvVarsAreRedacted(t, c.Env)
		}
	}
}

func TestKubernetesCluster_ListClusters(t *testing.T) {
	rsrc, coll := getKubernetesClusterCollector(t)
	clusters, err := coll.ListKubernetesClusters(context.Background(), rsrc.Name)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range clusters {
		k8sCluster := c.GetKubernetesCluster()
		t.Logf("Cluster %s:\n%+v\n", utils.GetHumanReadableName(c.Name), k8sCluster)
	}
}

func verifyEnvVarsAreRedacted(t *testing.T, vars []v1.EnvVar) {
	for _, e := range vars {
		if e.Value != "" && e.Value != "REDACTED" {
			t.Errorf("expected env value for %s to be empty or REDACTED, got %s", e.Name, e.Value)
		}
	}
}
