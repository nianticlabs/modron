package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

const (
	ContainerRunningRuleName = "CONTAINER_NOT_RUNNING"
	separator                = "@@"
)

type ContainerRunningConfig struct {
	// RequiredContainers is a map of namespaces to a list of prefixes of the pods that should be running: for example {"default": ["my-pod-"]}
	RequiredContainers map[string][]string `json:"requiredContainers"`
}

var (
	found = map[string]bool{}
)

type ContainerRunningRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewContainerRunningRule())
}

func NewContainerRunningRule() model.Rule {
	return &ContainerRunningRule{
		info: model.RuleInfo{
			Name: ContainerRunningRuleName,
			AcceptedResourceTypes: []proto.Message{
				&pb.KubernetesCluster{},
			},
		},
	}
}

func (r *ContainerRunningRule) Check(ctx context.Context, e model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	var cfg ContainerRunningConfig
	if err := utils.GetRuleConfig(ctx, e, r.info.Name, &cfg); err != nil {
		return nil, []error{fmt.Errorf("unable to parse rule config: %w", err)}
	}
	clusterChildren, err := e.GetChildren(ctx, rsrc.Name)
	if err != nil {
		errs = append(errs, err)
		return obs, errs
	}

	for _, namespace := range clusterChildren {
		// This is safe, if the function returns an error, the value will be ""
		if n, _ := utils.TypeFromResource(namespace); n != common.ResourceNamespace {
			continue
		}
		if _, ok := cfg.RequiredContainers[namespace.Name]; !ok {
			// There are no pods to check in this namespace.
			continue
		}

		pods, err := e.GetChildren(ctx, namespace.Name)
		if err != nil {
			errs = append(errs, err)
		}
		for _, pod := range pods {
			if p, _ := utils.TypeFromResource(pod); p != common.ResourcePod {
				continue
			}
			for _, prefix := range cfg.RequiredContainers[namespace.Name] {
				if pod.GetPod().GetPhase() == pb.Pod_RUNNING && strings.HasPrefix(pod.Name, prefix) {
					found[namespace.Name+separator+prefix] = true
				}
			}
		}
	}

	for namespace, prefixes := range cfg.RequiredContainers {
		for _, prefix := range prefixes {
			if !found[namespace+separator+prefix] {
				// TODO: Improve the observation by reporting the state of the existing pod if any.
				obs = append(obs, &pb.Observation{
					Uid:           uuid.NewString(),
					Timestamp:     timestamppb.Now(),
					ResourceRef:   utils.GetResourceRef(rsrc),
					Name:          r.Info().Name,
					ExpectedValue: structpb.NewStringValue(prefix),
					ObservedValue: structpb.NewStringValue(""),
					Remediation: &pb.Remediation{
						Description:    fmt.Sprintf("Cluster %s doesn't run any running container starting with %s", rsrc.DisplayName, prefix),
						Recommendation: "This cluster is likely missing an important part of infrastructure. Check the cluster configuration or reach out to your tech support or SRE.",
					},
					Severity: pb.Severity_SEVERITY_LOW,
				})
			}
		}
	}
	return
}

func (r *ContainerRunningRule) Info() *model.RuleInfo {
	return &r.info
}
