package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const TooHighPrivilegesRuleName = "SERVICE_ACCOUNT_TOO_HIGH_PRIVILEGES"

// Too Long

var dangerousRoles = []string{
	"editor",
	"owner",
	"composer.admin",
	"dataproc.admin",
	"dataproc.editor",
	"dataflow.admin",
	"dataflow.developer",
	"iam.serviceAccountAdmin",
	"iam.serviceAccountUser",
	"iam.serviceAccountTokenCreator",
}

type TooHighPrivilegesRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewTooHighPrivilegesRule())
}

func NewTooHighPrivilegesRule() model.Rule {
	return &TooHighPrivilegesRule{
		info: model.RuleInfo{
			Name: TooHighPrivilegesRuleName,
			AcceptedResourceTypes: []proto.Message{
				&pb.ServiceAccount{},
			},
		},
	}
}

func (r *TooHighPrivilegesRule) Check(ctx context.Context, e model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	rsrcGroup, err := e.GetResource(ctx, rsrc.Parent)
	if err != nil {
		errs = append(errs, fmt.Errorf("error retrieving resource group of resource %q: %w", rsrc.Name, err))
		return
	}
	policy := rsrcGroup.IamPolicy
	var roles []string

	if policy != nil {
		for _, perm := range policy.Permissions {
			newRoles := getAccountRoles(perm, rsrc.Name)
			roles = append(roles, newRoles...)
		}
	}

	if len(roles) > 0 {
		for _, role := range roles {
			if !slices.Contains(dangerousRoles, role) {
				continue
			}

			ob := &pb.Observation{
				Uid:           uuid.NewString(),
				Timestamp:     timestamppb.Now(),
				ResourceRef:   utils.GetResourceRef(rsrc),
				Name:          r.Info().Name,
				ExpectedValue: structpb.NewStringValue(""),
				ObservedValue: structpb.NewStringValue(role),
				Remediation: &pb.Remediation{
					Description: fmt.Sprintf(
						"Service account [%q](https://console.cloud.google.com/iam-admin/serviceaccounts?project=%s) has over-broad role %q",
						rsrc.Name,
						constants.ResourceWithoutProjectsPrefix(rsrcGroup.Name),
						role,
					),
					Recommendation: fmt.Sprintf(
						"Replace the role %q for service account [%q](https://console.cloud.google.com/iam-admin/serviceaccounts?project=%s) "+
							"with a predefined or custom role that grants it the **smallest set of permissions** needed to operate. "+
							"This role **cannot** be any of the following: `%v` *Hint: The Security insights column can help you reduce the amount of permissions*",
						role,
						rsrc.Name,
						constants.ResourceWithoutProjectsPrefix(rsrcGroup.Name),
						dangerousRoles,
					),
				},
				Severity: pb.Severity_SEVERITY_MEDIUM,
			}
			obs = append(obs, ob)
		}
	}

	return
}

func (r *TooHighPrivilegesRule) Info() *model.RuleInfo {
	return &r.info
}
