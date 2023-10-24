package rules

import (
	"regexp"
	"strings"

	"github.com/nianticlabs/modron/src/pb"
)

func getAccountRoles(perm *pb.Permission, account string) (roles []string) {
	for _, principal := range perm.Principals {
		if strings.EqualFold(principal, account) {
			roles = append(roles, perm.Role)
		}
	}
	return roles
}

const (
	PrincipalServiceAccount        = "serviceAccount"
	PrincipalUser                  = "user"
	PrincipalGroup                 = "group"
	PrincipalAllUsers              = "allUsers"
	PrincipalAllAuthenticatedUsers = "allAuthenticatedUsers"
	PrincipalDomain                = "domain"
)

// TODO: Add SelfLink and HumanReadableName field to Protobuf and move this logic to the collector.
func getGcpReadableResourceName(resourceName string) string {
	if !(strings.Contains(resourceName, "[") && strings.Contains(resourceName, "]")) {
		return resourceName
	}
	m := regexp.MustCompile(`\[.*\]$`)
	return m.ReplaceAllLiteralString(resourceName, "")
}
