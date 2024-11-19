package rules

import (
	"regexp"
	"strings"

	"github.com/nianticlabs/modron/src/constants"
	pb "github.com/nianticlabs/modron/src/proto/generated"

	"github.com/sirupsen/logrus"
)

var log = logrus.StandardLogger().WithField(constants.LogKeyPkg, "rules")

func getAccountRoles(perm *pb.Permission, account string) (roles []string) {
	for _, principal := range perm.Principals {
		if strings.HasPrefix(principal, PrincipalDeleted) {
			continue
		}
		s := strings.Split(principal, ":")
		if len(s) != 2 { // nolint:mnd
			log.Warn("invalid principal in org policy: ", principal)
			continue
		}
		p := strings.Split(principal, ":")[1]
		if strings.EqualFold(p, account) {
			roles = append(roles, perm.Role)
		}
	}
	return roles
}

const (
	PrincipalDeleted = "deleted:"
)

// TODO: Add SelfLink and HumanReadableName field to Protobuf and move this logic to the collector.
func getGcpReadableResourceName(resourceName string) string {
	if !(strings.Contains(resourceName, "[") && strings.Contains(resourceName, "]")) {
		return resourceName
	}
	m := regexp.MustCompile(`\[.*\]$`)
	return m.ReplaceAllLiteralString(resourceName, "")
}
