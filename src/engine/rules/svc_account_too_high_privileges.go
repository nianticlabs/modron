package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const TooHighPrivilegesRuleName = "SERVICE_ACCOUNT_TOO_HIGH_PRIVILEGES"

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
			AcceptedResourceTypes: []string{
				common.ResourceServiceAccount,
			},
		},
	}
}

func (r *TooHighPrivilegesRule) Check(ctx context.Context, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	rsrcGroup, err := engine.GetResource(ctx, rsrc.Parent)
	if err != nil {
		errs = append(errs, fmt.Errorf("error retrieving resource group of resource %q: %v", rsrc.Name, err))
		return
	}
	policy := rsrcGroup.IamPolicy
	roles := []string{}

	if policy != nil {
		for _, perm := range policy.Permissions {
			newRoles, err := engine.GetAccountRoles(perm, rsrc.Name)
			if err != nil {
				return nil, []error{err}
			}
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
				Resource:      rsrc,
				Name:          r.Info().Name,
				ExpectedValue: structpb.NewStringValue(""),
				ObservedValue: structpb.NewStringValue(role),
				Remediation: &pb.Remediation{
					Description: fmt.Sprintf(
						"Service account [%q](https://console.cloud.google.com/iam-admin/serviceaccounts?project=%s) has over-broad role %q",
						rsrc.Name,
						rsrcGroup.Name,
						role,
					),
					Recommendation: fmt.Sprintf(
						"Replace the role %q for service account [%q](https://console.cloud.google.com/iam-admin/serviceaccounts?project=%s) "+
							"with a predefined or custom role that grants it the **smallest set of permissions** needed to operate. "+
							"This role **can not** be any of the following: `%v` *Hint: The Security insights column can help you reduce the amount of permissions*",
						role,
						rsrc.Name,
						rsrcGroup.Name,
						dangerousRoles,
					),
				},
			}
			obs = append(obs, ob)
		}
	}

	return
}

func (r *TooHighPrivilegesRule) Info() *model.RuleInfo {
	return &r.info
}
