package utils

import (
	"fmt"
	"strings"

	"github.com/nianticlabs/modron/src/constants"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func IsGCPServiceAccountProject(project string) bool {
	_, ok := constants.GCPServiceAgentsProjects[project]
	return ok
}

func GetGCPProjectFromSAEmail(saEmail string) string {
	if strings.HasSuffix(saEmail, appspotServiceAccountSuffix) {
		return strings.TrimSuffix(saEmail, appspotServiceAccountSuffix)
	}

	if strings.HasSuffix(saEmail, iamGServiceAccountSuffix) {
		noSuffix := strings.TrimSuffix(saEmail, iamGServiceAccountSuffix)
		split := strings.Split(noSuffix, "@")
		if len(split) != 2 { //nolint:mnd
			log.Errorf("failed to split service account email: %s", saEmail)
			return ""
		}
		return split[1]
	}

	if strings.HasSuffix(saEmail, developerGserviceAccountSuffix) {
		// Can't handle this (we need a project ID, here we only have a project number)
		return ""
	}

	log.Warnf("unknown service account email format: %s", saEmail)
	return ""
}

type ResourceLink struct {
	Name string
	URL  string
	Type string
}

// TODO: Make sure all the observations use this function
func LinkGCPResource(resource *pb.Resource) ResourceLink {
	switch resource.Type.(type) {
	case *pb.Resource_Bucket:
		return ResourceLink{
			Name: resource.Name,
			URL:  "https://console.cloud.google.com/storage/browser/" + resource.Name,
			Type: "bucket",
		}
	case *pb.Resource_Database:
		return ResourceLink{
			Name: resource.Name,
			URL: fmt.Sprintf(
				"https://console.cloud.google.com/spanner/instances/%s/details/databases",
				resource.Name,
			),
			Type: "database",
		}

	case *pb.Resource_ServiceAccount:
		saEmail := strings.TrimPrefix(resource.Name, constants.GCPServiceAccountPrefix)
		return ResourceLink{
			Name: saEmail,
			URL: fmt.Sprintf(
				"https://console.cloud.google.com/iam-admin/serviceaccounts/details/%s?project=%s",
				saEmail,
				strings.TrimPrefix(resource.ResourceGroupName, constants.GCPProjectsNamePrefix),
			),
			Type: "service account",
		}
	case *pb.Resource_ResourceGroup:
		gcpType := ""
		switch {
		case strings.HasPrefix(resource.Name, constants.GCPProjectsNamePrefix):
			gcpType = "project"
		case strings.HasPrefix(resource.Name, constants.GCPFolderIDPrefix):
			gcpType = "folder"
		case strings.HasPrefix(resource.Name, constants.GCPOrgIDPrefix):
			gcpType = "organization"
		default:
			log.Warnf("LinkGCPResource: unknown resource type: %s", resource.Name)
		}
		return ResourceLink{
			Name: strings.TrimPrefix(resource.Name, constants.GCPProjectsNamePrefix),
			URL: fmt.Sprintf("https://console.cloud.google.com/welcome?project=%s",
				strings.TrimPrefix(resource.Name, constants.GCPProjectsNamePrefix),
			),
			Type: gcpType,
		}
	default:
		log.Warnf("LinkGCPResource: unknown resource type: %T", resource.Type)
	}
	return ResourceLink{
		Name: resource.Name,
	}
}
