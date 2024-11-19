package rules

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	OutDatedKubernetesVersion = "OUTDATED_KUBERNETES_VERSION"
	// https://cloud.google.com/kubernetes-engine/docs/release-schedule
	currentK8sVersion = 1.27
)

type OutDatedKubernetesVersionRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewOutDatedKubernetesVersionRule())
}

func NewOutDatedKubernetesVersionRule() model.Rule {
	return &OutDatedKubernetesVersionRule{
		info: model.RuleInfo{
			Name: OutDatedKubernetesVersion,
			AcceptedResourceTypes: []proto.Message{
				&pb.KubernetesCluster{},
			},
		},
	}
}

func (r *OutDatedKubernetesVersionRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	k8s := rsrc.GetKubernetesCluster()
	var obs []*pb.Observation
	var errs []error
	if k8s == nil {
		errs = append(errs, fmt.Errorf("no kubernetes cluster resource provided"))
		return obs, errs
	}
	if k8s.MasterVersion == "" {
		errs = append(errs, fmt.Errorf("kubernetes master verion of cluster %s/%s is empty", rsrc.ResourceGroupName, rsrc.Name))
	}
	if k8s.NodesVersion == "" {
		errs = append(errs, fmt.Errorf("kubernetes nodes version of cluster %s/%s is empty", rsrc.ResourceGroupName, rsrc.Name))
	}
	if len(errs) > 0 {
		return obs, errs
	}
	masterVersion, err := versionNumberFromVersionString(k8s.MasterVersion)
	if err != nil {
		errs = append(errs, err)
	}
	nodesVersion, err := versionNumberFromVersionString(k8s.NodesVersion)
	if err != nil {
		errs = append(errs, err)
	}
	if masterVersion < currentK8sVersion {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue(fmt.Sprintf("version > %.2f", currentK8sVersion)),
			ObservedValue: structpb.NewStringValue(k8s.MasterVersion),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Cluster [%q](https://console.cloud.google.com/kubernetes/list/overview?project=%q) uses an outdated Kubernetes master version",
					getGcpReadableResourceName(rsrc.Name),
					rsrc.ResourceGroupName,
				),
				Recommendation: fmt.Sprintf(
					"Update the Kubernetes master version on cluster [%q](https://console.cloud.google.com/kubernetes/list/overview?project=%s) to at least %.2f. For more details on this process, see [this article](https://cloud.google.com/kubernetes-engine/docs/how-to/upgrading-a-cluster)",
					getGcpReadableResourceName(rsrc.Name),
					rsrc.ResourceGroupName,
					currentK8sVersion,
				),
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		}
		obs = append(obs, ob)
	}

	if nodesVersion < currentK8sVersion {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue(fmt.Sprintf("version > %.2f", currentK8sVersion)),
			ObservedValue: structpb.NewStringValue(k8s.NodesVersion),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Cluster [%q](https://console.cloud.google.com/kubernetes/list/overview?project=%s) uses an outdated Kubernetes version",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
				Recommendation: fmt.Sprintf(
					"Update the Kubernetes version on cluster [%q](https://console.cloud.google.com/kubernetes/list/overview?project=%s) to at least %.2f. For more details on this process, see [this article](https://cloud.google.com/kubernetes-engine/docs/how-to/upgrading-a-cluster)",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
					currentK8sVersion,
				),
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		}
		obs = append(obs, ob)
	}

	return obs, errs
}

func (r *OutDatedKubernetesVersionRule) Info() *model.RuleInfo {
	return &r.info
}

func versionNumberFromVersionString(s string) (float64, error) {
	// Version from GKE 1.22.10-gke.600, we want 1.22
	v := strings.Split(s, "-")[0]
	tokenised := strings.Split(v, ".")
	vStr := strings.Join(tokenised[:2], ".")
	return strconv.ParseFloat(vStr, 64)
}
