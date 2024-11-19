package gcpcollector

import (
	"encoding/json"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/container/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/api/core/v1"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

func (collector *GCPCollector) ListKubernetesClusters(ctx context.Context, rgName string) (kubernetesClusters []*pb.Resource, err error) {
	clusters, err := collector.api.ListClustersByZone(ctx, rgName, "-")
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		nodeVersion := ""
		for _, nodePool := range cluster.NodePools {
			nodeVersion = nodePool.Version
		}
		var masterAuthorizedNetworks []string
		if cluster.MasterAuthorizedNetworksConfig != nil {
			for _, cidrBlock := range cluster.MasterAuthorizedNetworksConfig.CidrBlocks {
				masterAuthorizedNetworks = append(masterAuthorizedNetworks, cidrBlock.CidrBlock)
			}
		}
		privateCluster := false
		if cluster.PrivateClusterConfig != nil {
			privateCluster = cluster.PrivateClusterConfig.EnablePrivateNodes
		}

		kubernetesClusters = append(kubernetesClusters, &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              cluster.Name,
			Parent:            rgName,
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					Location:                 cluster.Location,
					PrivateCluster:           privateCluster,
					MasterAuthorizedNetworks: masterAuthorizedNetworks,
					MasterVersion:            cluster.CurrentMasterVersion,
					NodesVersion:             nodeVersion,
					Security: &pb.KubernetesCluster_Security{
						VulnerabilityScanning: toPbSecurityVulnScanningType(cluster.SecurityPostureConfig),
					},
				},
			},
		})
	}

	return kubernetesClusters, nil
}

func toPbSecurityVulnScanningType(config *container.SecurityPostureConfig) pb.KubernetesCluster_Security_VulnScanning {
	if config == nil {
		return pb.KubernetesCluster_Security_VULN_SCAN_DISABLED
	}

	// https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1beta1/projects.locations.clusters#Cluster.VulnerabilityMode
	switch strings.ToUpper(config.VulnerabilityMode) {
	case "VULNERABILITY_DISABLED":
		return pb.KubernetesCluster_Security_VULN_SCAN_DISABLED
	case "VULNERABILITY_BASIC":
		return pb.KubernetesCluster_Security_VULN_SCAN_BASIC
	case "VULNERABILITY_ENTERPRISE":
		return pb.KubernetesCluster_Security_VULN_SCAN_ADVANCED
	}
	return pb.KubernetesCluster_Security_VULN_SCAN_UNKNOWN
}

func (collector *GCPCollector) ListKubernetesNamespaces(ctx context.Context, rgName string) (namespaces []*pb.Resource, err error) {
	ns, err := collector.api.ListNamespaces(ctx, rgName)
	if err != nil {
		return nil, err
	}
	for _, n := range ns {
		createTime := parseTimeOrZero(n.CreateTime)
		namespaces = append(namespaces, &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Parent:            n.ParentFullResourceName,
			Name:              n.Name,
			Type: &pb.Resource_Namespace{
				Namespace: &pb.Namespace{
					Cluster:      n.ParentFullResourceName,
					CreationTime: timestamppb.New(createTime),
				},
			},
		})
	}
	return namespaces, nil
}

func (collector *GCPCollector) ListKubernetesPods(ctx context.Context, rgName string) (pods []*pb.Resource, err error) {
	ps, err := collector.api.ListPods(ctx, rgName)
	if err != nil {
		return nil, err
	}
	for _, p := range ps {
		var pod v1.Pod
		if len(p.VersionedResources) == 0 {
			log.WithField(constants.LogKeyResourceGroup, rgName).
				Warnf("no versioned resources found for pod %q", p.Name)
			continue
		}
		if err := json.Unmarshal(p.VersionedResources[0].Resource, &pod); err != nil {
			return nil, err
		}
		removeSensitiveFields(&pod)

		createTime := parseTimeOrZero(p.CreateTime)

		// Cleanup some fields we don't need in ObjectMeta
		pod.ObjectMeta.ManagedFields = []apimachineryv1.ManagedFieldsEntry{}

		pods = append(pods, &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Parent:            p.ParentFullResourceName,
			Name:              utils.GetHumanReadableName(p.Name),
			Type: &pb.Resource_Pod{
				Pod: &pb.Pod{
					// Extract the cluster name from the namespace name.
					// Namespace names follow this format:
					// "//container.googleapis.com/projects/project-id/locations/location/clusters/cluster-name/k8s/namespaces/namespace-name",
					Cluster:      removeNamespaceFromResourceName(p.ParentFullResourceName),
					CreationTime: timestamppb.New(createTime),
					Namespace:    p.ParentFullResourceName,
					Phase:        phaseFromString(p.State),
					Spec:         &pod.Spec,
					ObjectMeta:   &pod.ObjectMeta,
				},
			},
		})
	}
	return pods, nil
}

func removeNamespaceFromResourceName(resourceName string) string {
	index := strings.LastIndex(resourceName, "/k8s/namespaces/")
	if index == -1 {
		return resourceName
	}
	return resourceName[:index]
}

// removeSensitiveFields removes certain fields of the pod spec that might contain sensitive information.
// generally this shouldn't be a problem as secrets shouldn't be stored in the pod spec directly (but rather in a Secret),
// but since there is a likelihood that this might happen, we remove these fields just in case.
// TODO: remove this function and replace it with a more specific check based on entropy - and create observations for these cases
func removeSensitiveFields(pod *v1.Pod) {
	for i := range pod.Spec.Containers {
		for k, e := range pod.Spec.Containers[i].Env {
			if e.Value != "" {
				pod.Spec.Containers[i].Env[k].Value = "REDACTED"
			}
		}
	}
	for i := range pod.Spec.InitContainers {
		for k, e := range pod.Spec.InitContainers[i].Env {
			if e.Value != "" {
				pod.Spec.InitContainers[i].Env[k].Value = "REDACTED"
			}
		}
	}
}

func parseTimeOrZero(timeString string) time.Time {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		log.Errorf("cannot parse namespace creation time %q: %v", timeString, err)
		return time.Time{}
	}
	return t
}

func phaseFromString(s string) pb.Pod_Phase {
	switch strings.ToUpper(s) {
	case "RUNNING":
		return pb.Pod_RUNNING
	case "PENDING":
		return pb.Pod_PENDING
	case "SUCCEEDED":
		return pb.Pod_SUCCEEDED
	case "FAILED":
		return pb.Pod_FAILED
	case "UNKNOWN":
		return pb.Pod_UNKNOWN
	default:
		return pb.Pod_UNKNOWN_PHASE
	}
}
