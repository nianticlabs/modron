package constants

import (
	"strings"
)

const (
	OrgIdEnvVar     = "ORG_ID"
	OrgSuffixEnvVar = "ORG_SUFFIX"

	GCPEditorRole        = "editor"
	GCPOwnerRole         = "owner"
	GCPSecurityAdminRole = "iam.securityAdmin"

	GCPOrgIdPrefix          = "organizations/"
	GCPFolderIdPrefix       = "folders/"
	GCPProjectsNamePrefix   = "projects/"
	GCPRolePrefix           = "roles/"
	GCPAccountGroupPrefix   = "group:"
	GCPServiceAccountPrefix = "serviceAccount:"
	GCPUserAccountPrefix    = "user:"
)

var AdminRoles = map[string]struct{}{
	strings.ToLower(GCPOwnerRole):         {},
	strings.ToLower(GCPEditorRole):        {},
	strings.ToLower(GCPSecurityAdminRole): {},
}

func ResourceWithProjectsPrefix(resourceName string) string {
	if strings.HasPrefix(resourceName, GCPProjectsNamePrefix) {
		return resourceName
	}
	return GCPProjectsNamePrefix + resourceName
}

func ResourceWithoutProjectsPrefix(resourceName string) string {
	return strings.TrimPrefix(resourceName, GCPProjectsNamePrefix)
}
