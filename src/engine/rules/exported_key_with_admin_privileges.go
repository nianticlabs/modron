package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

const ExportedKeyWithAdminPrivileges = "EXPORTED_KEY_WITH_ADMIN_PRIVILEGES"

// Too Long

var adminRoles = map[string]struct{}{
	"composer.admin":                 {},
	"compute.admin":                  {},
	"editor":                         {},
	"iam.securityAdmin":              {},
	"iam.serviceAccounts.actAs":      {},
	"iam.serviceAccountTokenCreator": {},
	"owner":                          {},
	"dataflow.admin":                 {},
}

type ExportedKeyWithAdminPrivilegesRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewExportedKeyWithAdminPrivilegesRule())
}

func NewExportedKeyWithAdminPrivilegesRule() model.Rule {
	return &ExportedKeyWithAdminPrivilegesRule{
		info: model.RuleInfo{
			Name: ExportedKeyWithAdminPrivileges,
			AcceptedResourceTypes: []proto.Message{
				&pb.ServiceAccount{},
			},
		},
	}
}

func (r *ExportedKeyWithAdminPrivilegesRule) Check(ctx context.Context, e model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	sa := rsrc.GetServiceAccount()
	rsrcGroup, err := e.GetResource(ctx, rsrc.Parent)
	if err != nil {
		return nil, []error{fmt.Errorf("error retrieving resource group of resource %q: %w", rsrc.Name, err)}
	}
	policy := rsrcGroup.IamPolicy
	hasAdminRoles := false

	if policy != nil {
		for _, perm := range policy.Permissions {
			newRoles := getAccountRoles(perm, rsrc.Name)
			for _, r := range newRoles {
				if _, ok := adminRoles[r]; ok {
					hasAdminRoles = true
				}
			}
		}
	}

	obs := []*pb.Observation{}
	if hasAdminRoles {
		nbEx := len(sa.ExportedCredentials)
		if nbEx > 0 {
			ob := &pb.Observation{
				Uid:           uuid.NewString(),
				Timestamp:     timestamppb.Now(),
				ResourceRef:   utils.GetResourceRef(rsrc),
				Name:          r.Info().Name,
				ExpectedValue: structpb.NewStringValue("0 keys"),
				ObservedValue: structpb.NewStringValue(fmt.Sprintf("%v keys", nbEx)),
				Remediation: &pb.Remediation{
					Description: fmt.Sprintf(
						"Service account [%q](https://console.cloud.google.com/iam-admin/serviceaccounts?project=%s) has %d exported keys with admin privileges",
						rsrc.Name,
						constants.ResourceWithoutProjectsPrefix(rsrcGroup.Name),
						nbEx,
					),
					Recommendation: fmt.Sprintf(
						"Avoid exporting keys of service accounts with admin privileges, they can be copied and used outside of Niantic. Revoke the exported key by clicking on service account [%q](https://console.cloud.google.com/iam-admin/serviceaccounts?project=%s), switch to the KEYS tab and delete the exported key. Instead of exporting keys, make use of [workload identity](https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity) or similar concepts",
						rsrc.Name,
						constants.ResourceWithoutProjectsPrefix(rsrcGroup.Name),
					),
				},
				Severity: pb.Severity_SEVERITY_CRITICAL,
			}
			obs = append(obs, ob)
		}

	}
	return obs, nil
}

func (r *ExportedKeyWithAdminPrivilegesRule) Info() *model.RuleInfo {
	return &r.info
}
