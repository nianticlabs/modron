package rules

import (
	"context"
	"fmt"
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
			AcceptedResourceTypes: []proto.Message{
				&pb.VmInstance{},
			},
		},
	}
}

func (r *VMHasPublicIPRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	vm := rsrc.GetVmInstance()

	if vm.PublicIp != "" && !strings.HasPrefix(rsrc.GetName(), "gke-") && len([]rune(rsrc.GetName())) <= 30 {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
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
			Severity: pb.Severity_SEVERITY_HIGH,
		}
		obs = append(obs, ob)
	}

	return
}

func (r *VMHasPublicIPRule) Info() *model.RuleInfo {
	return &r.info
}
