package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

const HumanWithOverprivilegedBasicRole = "HUMAN_WITH_OVERPRIVILEGED_BASIC_ROLE"

// Define basic roles.
var basicRoles = map[constants.Role]struct{}{
	constants.GCPEditorRole:        {},
	constants.GCPOwnerRole:         {},
	constants.GCPSecurityAdminRole: {},
	constants.GCPViewerRole:        {},
}

type HumanWithOverprivilegedBasicRoleRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewHumanWithOverprivilegedBasicRoleRule())
}

func NewHumanWithOverprivilegedBasicRoleRule() model.Rule {
	return &HumanWithOverprivilegedBasicRoleRule{
		info: model.RuleInfo{
			Name: HumanWithOverprivilegedBasicRole,
			AcceptedResourceTypes: []proto.Message{
				&pb.ResourceGroup{},
			},
		},
	}
}

func (r *HumanWithOverprivilegedBasicRoleRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	policy := rsrc.GetIamPolicy()

	hasBasicRoles := map[string][]string{}
	if policy != nil {
		for _, perm := range policy.Permissions {
			if _, ok := basicRoles[constants.ToRole(perm.Role)]; !ok {
				continue
			}
			for _, principal := range perm.Principals {
				// We intentionally don't check service accounts.
				// Service accounts could be member of the group, this would make a false positive if the group
				// only has service accounts as members.
				if strings.HasPrefix(principal, constants.GCPUserAccountPrefix) || strings.HasPrefix(principal, constants.GCPAccountGroupPrefix) {
					hasBasicRoles[principal] = append(hasBasicRoles[principal], perm.Role)
				}
			}
		}
	}

	obs := []*pb.Observation{}
	for principal, roles := range hasBasicRoles {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue("No basic roles"),
			ObservedValue: structpb.NewStringValue(strings.Join(roles, ", ")),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Human account or group %s has overprivileged basic roles on project [%s](https://console.cloud.google.com/iam-admin/iam?project=%s)",
					principal,
					constants.ResourceWithoutProjectsPrefix(rsrc.GetResourceGroupName()),
					constants.ResourceWithoutProjectsPrefix(rsrc.GetResourceGroupName()),
				),
				Recommendation: "Consider assigning \"Developer\" to the editors and \"Owner\" to the owners instead of using basic roles.",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		}
		obs = append(obs, ob)
	}
	return obs, nil
}

func (r *HumanWithOverprivilegedBasicRoleRule) Info() *model.RuleInfo {
	return &r.info
}
