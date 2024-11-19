package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/utils"
)

const CrossEnvironmentPermissionsRuleName = "CROSS_ENVIRONMENT_PERMISSIONS"

type CrossEnvironmentPermissionsRule struct {
	info                 model.RuleInfo
	cloudIdentityService *cloudidentity.Service
}

func init() {
	AddRule(NewCrossEnvironmentPermissionsRule())
}

func NewCrossEnvironmentPermissionsRule() model.Rule {
	return &CrossEnvironmentPermissionsRule{
		info: model.RuleInfo{
			Name: CrossEnvironmentPermissionsRuleName,
			AcceptedResourceTypes: []proto.Message{
				&pb.KubernetesCluster{},
				&pb.ServiceAccount{},
				&pb.Bucket{},
				// &pb.Database{}, // TODO: Uncomment when we start collecting DBs IAM Policies
				&pb.ResourceGroup{},
			},
		},
		cloudIdentityService: nil,
	}
}

func (r *CrossEnvironmentPermissionsRule) Check(ctx context.Context, e model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	log = log.WithField("rule", r.info.Name)
	policy := rsrc.IamPolicy
	if policy == nil {
		return nil, []error{fmt.Errorf("resource %s has no IAM policy", rsrc.Name)}
	}

	hierarchy, err := e.GetHierarchy(ctx, rsrc.CollectionUid)
	if err != nil {
		return nil, []error{err}
	}

	myEnv := risk.GetEnvironment(e.GetTagConfig(), hierarchy, rsrc.ResourceGroupName)
	principals := make(map[string]string)
	for _, perm := range policy.Permissions {
		for _, p := range perm.Principals {
			if !strings.HasPrefix(p, constants.GCPServiceAccountPrefix) {
				log.Debugf("skipping principal %s", p)
				continue
			}
			saEmail := strings.TrimPrefix(p, constants.GCPServiceAccountPrefix)
			saProject := utils.GetGCPProjectFromSAEmail(saEmail)
			if saProject == "" {
				log.Warnf("got an invalid project for service account %s", saEmail)
				continue
			}
			if utils.IsGCPServiceAccountProject(saProject) {
				log.Debugf("service account %s is a GCP service account", saEmail)
				continue
			}
			if saProject == strings.TrimPrefix(rsrc.ResourceGroupName, constants.GCPProjectsNamePrefix) {
				log.Debugf("service account %s is in the same project as the resource", saEmail)
				continue
			}
			// Cross project
			env := risk.GetEnvironment(e.GetTagConfig(), hierarchy, constants.GCPProjectsNamePrefix+saProject)
			if env != myEnv {
				principals[saEmail] = env
			}
		}
	}

	for principal, otherEnv := range principals {
		obs = append(obs, &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Name:          r.info.Name,
			ExpectedValue: structpb.NewStringValue(myEnv),
			ObservedValue: structpb.NewStringValue(otherEnv),
			Remediation: &pb.Remediation{
				Description:    fmt.Sprintf("%s is in a different environment than the resource %q", principal, rsrc.Name),
				Recommendation: fmt.Sprintf("Revoke the access of %q to the resource %q", principal, rsrc.Name),
			},
			ResourceRef: utils.GetResourceRef(rsrc),
			Category:    pb.Observation_CATEGORY_MISCONFIGURATION,
			Severity:    pb.Severity_SEVERITY_HIGH,
		})
	}

	return obs, errs
}

func (r *CrossEnvironmentPermissionsRule) Info() *model.RuleInfo {
	return &r.info
}
