package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
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
	account string
	group   string
	role    string
}

func init() {
	AddRule(NewCrossProjectPermissionsRule())
}

func NewCrossProjectPermissionsRule() model.Rule {
	return &CrossProjectPermissionsRule{
		info: model.RuleInfo{
			Name: CrossProjectPermissionsRuleName,
			AcceptedResourceTypes: []string{
				common.ResourceServiceAccount,
				common.ResourceBucket,
				common.ResourceDatabase,
				common.ResourceResourceGroup,
			},
		},
		cloudIdentityService: nil,
	}
}

func getResourceSpecificString(rsrc *pb.Resource) string {
	switch rsrc.Type.(type) {
	case *pb.Resource_Bucket:
		return "storage bucket [%q](https://console.cloud.google.com/storage/browser/%s)"
	case *pb.Resource_Database:
		return "database [%q](https://console.cloud.google.com/spanner/instances/%s/details/databases)"
	case *pb.Resource_ServiceAccount:
		return "service account [%q](https://console.cloud.google.com/iam-admin/serviceaccounts?project=%s)"
	case *pb.Resource_ResourceGroup:
		return "project [%q](https://console.cloud.google.com/welcome?project=%s)"
	default:
		return ""
	}
}

func getRemediationByResourceType(rsrc *pb.Resource, fp *FindingPrincipal) pb.Remediation {
	resourceString := getResourceSpecificString(rsrc)
	var throughGroup string
	var recommendation string
	var recommendationTemplate string
	var linkContent string
	switch rsrc.Type.(type) {
	case *pb.Resource_Bucket, *pb.Resource_Database:
		linkContent = rsrc.Name
	case *pb.Resource_ServiceAccount, *pb.Resource_ResourceGroup:
		linkContent = constants.ResourceWithoutProjectsPrefix(rsrc.Name)
	}
	if fp.group != "" {
		throughGroup = fmt.Sprintf(". The user is part of the group %s", fp.group)
		recommendationTemplate =
			"Remove the account %q from the group %q, remove the group from the " +
				resourceString +
				" or replace the principal %q with a principal created in the project %q that grants it the **smallest set of permissions** needed to operate"
		recommendation = fmt.Sprintf(recommendationTemplate,
			fp.account,
			fp.group,
			rsrc.Name,
			linkContent,
			fp.account,
			constants.ResourceWithoutProjectsPrefix(rsrc.Parent),
		)
	} else {
		throughGroup = ""
		recommendationTemplate =
			"Replace the principal %q controlling the " +
				resourceString +
				" with a principal created in the project %q that grants it the **smallest set of permissions** needed to operate"
		recommendation = fmt.Sprintf(
			recommendationTemplate,
			fp.account,
			rsrc.Name,
			linkContent,
			constants.ResourceWithoutProjectsPrefix(rsrc.Parent),
		)
	}
	switch rsrc.Type.(type) {
	case *pb.Resource_Bucket, *pb.Resource_Database, *pb.Resource_ServiceAccount:
		descriptionTemplate := "The " + resourceString + " is controlled by the principal %q with role %s defined in another project%s"
		return pb.Remediation{
			Description: fmt.Sprintf(
				descriptionTemplate,
				rsrc.Name,
				linkContent,
				fp.account,
				fp.role,
				throughGroup,
			),
			Recommendation: recommendation,
		}
	case *pb.Resource_ResourceGroup:
		descriptionTemplate :=
			"The " +
				resourceString +
				" gives the principal %q vast permissions through the role %s%s This principal is defined in another project which means that anybody with rights in that project can use it to control the resources in this one"
		return pb.Remediation{
			Description: fmt.Sprintf(
				descriptionTemplate,
				rsrc.Name,
				linkContent,
				fp.account,
				fp.role,
				"."+throughGroup,
			),
			Recommendation: recommendation,
		}
	default:
		return pb.Remediation{}
	}
}

func isCrossProjectServiceAccount(ctx context.Context, user string, rsrc *pb.Resource) (bool, error) {
	svcAccnt := strings.Split(user, "@")
	if user == "allUsers" || user == "allAuthenticatedUsers" {
		return false, nil
	}
	if len(svcAccnt) < 2 {
		glog.Warningf("unknown service account %q", user)
		return false, nil
	}
	_, contained := googleManagedAccountSuffixes[svcAccnt[1]]
	var resourceId string
	switch rsrc.Type.(type) { // ResourceGroup is only available on a Resource_ResourceGroup and not on others
	case *pb.Resource_ResourceGroup:
		resourceId = rsrc.GetResourceGroup().Identifier
	default:
		parent, err := engine.GetResource(ctx, rsrc.Parent)
		if err != nil {
			err = fmt.Errorf("could not get parent %v of resource %v", rsrc.Parent, rsrc.Name)
			glog.Error(err)
			return false, err
		}
		resourceId = parent.GetResourceGroup().Identifier
	}

	return strings.HasSuffix(user, ".gserviceaccount.com") && // It should be a service account
		!strings.HasPrefix(strings.TrimPrefix(user, constants.GCPServiceAccountPrefix), resourceId) && // Should not be an account with the project prefix
		!strings.Contains(user, constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName)) && // The service account was created in another project
		!strings.HasSuffix(user, "@cloudservices.gserviceaccount.com") &&
		!contained, nil
}

func (r *CrossProjectPermissionsRule) createObservation(rsrc *pb.Resource, fp *FindingPrincipal) *pb.Observation {
	remediation := getRemediationByResourceType(rsrc, fp)
	observed_value := structpb.NewStringValue(fmt.Sprintf("%q with role %s", fp.account, fp.role))
	if fp.group != "" {
		observed_value = structpb.NewStringValue(fmt.Sprintf("%q > %q", fp.group, fp.account))
	}
	ob := &pb.Observation{
		Uid:           uuid.NewString(),
		Timestamp:     timestamppb.Now(),
		Resource:      rsrc,
		Name:          r.Info().Name,
		ExpectedValue: structpb.NewStringValue(""),
		ObservedValue: observed_value,
		Remediation:   &remediation,
	}

	return ob
}

func (r *CrossProjectPermissionsRule) Check(ctx context.Context, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	policy := rsrc.IamPolicy

	if policy == nil {
		return nil, []error{fmt.Errorf("resource %s has no IAM policy", rsrc.Name)}
	}

	for _, perm := range policy.Permissions {
		if slices.Contains(rolesToWatch, perm.Role) {
			for _, principal := range perm.GetPrincipals() {
				// TODO: Check for cross project users in groups
				shouldFlag, err := isCrossProjectServiceAccount(ctx, principal, rsrc)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				if shouldFlag {
					obs = append(obs, r.createObservation(rsrc, &FindingPrincipal{principal, "", perm.Role}))
				}
			}
		}
	}

	return obs, errs
}

func (r *CrossProjectPermissionsRule) Info() *model.RuleInfo {
	return &r.info
}
