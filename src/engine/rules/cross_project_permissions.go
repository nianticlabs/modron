package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

const CrossProjectPermissionsRuleName = "CROSS_PROJECT_PERMISSIONS"

var (
	rolesToWatch = []string{
		"artifactregistry.writer",
		"cloudbuild.builds.builder",
		"composer.admin",
		"compute.admin",
		"compute.instanceAdmin",
		"compute.loadBalancerAdmin",
		"container.admin",
		"container.serviceAgent",
		"containerregistry.serviceAgent",
		"dataproc.admin",
		"dataproc.editor",
		"dataflow.admin",
		"dataflow.developer",
		"editor",
		"iam.securityAdmin",
		"iam.serviceAccountAdmin",
		"iam.serviceAccountTokenCreator",
		"iam.serviceAccountUser",
		"iam.workloadIdentityUser",
		"owner",
		"spanner.admin",
		"storage.admin",
		"storage.legacyBucketOwner",
	}
)

type CrossProjectPermissionsRule struct {
	info                 model.RuleInfo
	cloudIdentityService *cloudidentity.Service
}

type FindingPrincipal struct {
	account   string
	projectID string
	role      string
}

func init() {
	AddRule(NewCrossProjectPermissionsRule())
}

func NewCrossProjectPermissionsRule() model.Rule {
	return &CrossProjectPermissionsRule{
		info: model.RuleInfo{
			Name: CrossProjectPermissionsRuleName,
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

func getResourceSpecificString(rsrc *pb.Resource) string {
	resourceLink := utils.LinkGCPResource(rsrc)
	return fmt.Sprintf("%s [%s](%s)", resourceLink.Type, resourceLink.Name, resourceLink.URL)
}

func getRemediationByResourceType(rsrc *pb.Resource, fp *FindingPrincipal) *pb.Remediation {
	resourceString := getResourceSpecificString(rsrc)
	principalString := getResourceSpecificString(&pb.Resource{
		Name:              fp.account,
		ResourceGroupName: constants.ResourceWithProjectsPrefix(fp.projectID),
		Type:              &pb.Resource_ServiceAccount{},
	})
	recommendation := fmt.Sprintf(
		"Replace the %s controlling %s with a principal created in the project %q that grants it the **smallest set of permissions** needed to operate.",
		principalString,
		resourceString,
		rsrc.ResourceGroupName,
	)
	switch rsrc.Type.(type) {
	case *pb.Resource_Bucket, *pb.Resource_Database, *pb.Resource_ServiceAccount:
		return &pb.Remediation{
			Description: fmt.Sprintf(
				"The %s is controlled by the %s with role `%s` defined in project %q",
				resourceString,
				principalString,
				fp.role,
				fp.projectID,
			),
			Recommendation: recommendation,
		}
	case *pb.Resource_ResourceGroup:
		return &pb.Remediation{
			Description: fmt.Sprintf(
				"The %s gives the %s vast permissions through the role `%s`.\n"+
					"This principal is defined in project %q, which means that anybody with rights in that project can use it to control the resources in this one",
				resourceString,
				principalString,
				fp.role,
				fp.projectID,
			),
			Recommendation: recommendation,
		}
	default:
		return &pb.Remediation{}
	}
}

func isCrossProjectServiceAccount(user string, sourceProjectID string) (bool, string) {
	if user == "allUsers" || user == "allAuthenticatedUsers" {
		return false, ""
	}
	svcAccnt := strings.Split(user, "@")
	if len(svcAccnt) < 2 { //nolint:mnd
		log.Warnf("unknown service account %q", user)
		return false, ""
	}
	if !strings.HasSuffix(user, "iam.gserviceaccount.com") {
		return false, ""
	}
	saProject := utils.GetGCPProjectFromSAEmail(user)
	if saProject == "" {
		log.Warnf("could not get project from service account %q", user)
		return false, ""
	}

	if utils.IsGCPServiceAccountProject(saProject) {
		return false, ""
	}
	if saProject == sourceProjectID {
		return false, ""
	}
	return true, saProject
}

func (r *CrossProjectPermissionsRule) createObservation(rsrc *pb.Resource, fp *FindingPrincipal) *pb.Observation {
	return &pb.Observation{
		Uid:           uuid.NewString(),
		Timestamp:     timestamppb.Now(),
		ResourceRef:   utils.GetResourceRef(rsrc),
		Name:          r.Info().Name,
		ExpectedValue: structpb.NewStringValue(strings.TrimPrefix(rsrc.ResourceGroupName, constants.GCPProjectsNamePrefix)),
		ObservedValue: structpb.NewStringValue(fp.projectID),
		Remediation:   getRemediationByResourceType(rsrc, fp),
		Severity:      pb.Severity_SEVERITY_MEDIUM,
	}
}

func (r *CrossProjectPermissionsRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	policy := rsrc.IamPolicy
	if policy == nil {
		return nil, []error{fmt.Errorf("resource %s has no IAM policy", rsrc.Name)}
	}

	for _, perm := range policy.Permissions {
		if slices.Contains(rolesToWatch, perm.Role) {
			for _, principal := range perm.GetPrincipals() {
				// TODO: Check for cross project service accounts in groups
				xProjectSA, projectID := isCrossProjectServiceAccount(
					principal,
					strings.TrimPrefix(
						rsrc.ResourceGroupName,
						constants.GCPProjectsNamePrefix,
					),
				)
				if xProjectSA {
					obs = append(obs, r.createObservation(rsrc, &FindingPrincipal{principal, projectID, perm.Role}))
				}
			}
		}
	}
	return obs, errs
}

func (r *CrossProjectPermissionsRule) Info() *model.RuleInfo {
	return &r.info
}
