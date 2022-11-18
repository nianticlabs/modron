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

	GCPOrgIdPrefix = "organizations/"
	GCPRolePrefix  = "roles/"
)

var (
	AdminRoles = map[string]struct{}{
		strings.ToLower(GCPOwnerRole):         {},
		strings.ToLower(GCPEditorRole):        {},
		strings.ToLower(GCPSecurityAdminRole): {},
	}
)
