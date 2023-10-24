package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const VMHasPublicIPRuleName = "VM_HAS_PUBLIC_IP"

type VMHasPublicIPRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewVMHasPublicIPRule())
}

func NewVMHasPublicIPRule() model.Rule {
	return &VMHasPublicIPRule{
		info: model.RuleInfo{
			Name: VMHasPublicIPRuleName,
			AcceptedResourceTypes: []string{
				common.ResourceVmInstance,
			},
		},
	}
}

func (r *VMHasPublicIPRule) Check(ctx context.Context, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	vm := rsrc.GetVmInstance()

	if vm.PublicIp != "" && !strings.HasPrefix(rsrc.GetName(), "gke-") && len([]rune(rsrc.GetName())) <= 30 {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue("empty"),
			ObservedValue: structpb.NewStringValue(vm.PublicIp),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"VM %q has a public IP assigned",
					rsrc.Name,
				),
				Recommendation: fmt.Sprintf(
					"Compute instances should not be configured to have external IP addresses. Update network-settings of [%s](https://console.cloud.google.com/compute/instances?project=%s). You can connect to Linux VMs that do not have public IP addresses by using Identity-Aware Proxy for TCP forwarding. [Learn more](https://cloud.google.com/compute/docs/instances/connecting-advanced#sshbetweeninstances)",
					rsrc.Name,
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
			},
		}
		obs = append(obs, ob)
	}

	return
}

func (r *VMHasPublicIPRule) Info() *model.RuleInfo {
	return &r.info
}
