package constants

import (
	"strings"
)

const (
	OrgIdEnvVar     = "ORG_ID"
	OrgSuffixEnvVar = "ORG_SUFFIX"

	GCPOwnerRole         = "owner"
	GCPOrgIdPrefix       = "organizations/"
	GCPEditorRole        = "editor"
	GCPSecurityAdminRole = "iam.securityAdmin"
	GCPRolePrefix        = "roles/"
)

var (
	AdminRoles = map[string]struct{}{
		strings.ToLower(GCPOwnerRole):         {},
		strings.ToLower(GCPEditorRole):        {},
		strings.ToLower(GCPSecurityAdminRole): {},
	}
)
