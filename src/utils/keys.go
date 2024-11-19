package utils

import (
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/nianticlabs/modron/src/constants"
)

var log = logrus.StandardLogger().WithField(constants.LogKeyPkg, "utils")

const keyParts = 6

// GetKeyID converts a key reference (projects/my-project/serviceAccounts/sa-1/keys/abc) to a key ID (abc).
func GetKeyID(keyRef string) string {
	if !strings.HasPrefix(keyRef, constants.GCPProjectsNamePrefix) {
		log.Errorf("keyRef %s does not start with %s", keyRef, constants.GCPProjectsNamePrefix)
		return keyRef
	}

	split := strings.Split(keyRef, "/")
	if len(split) < keyParts {
		log.Errorf("keyRef %s has less than 6 parts", keyRef)
		return keyRef
	}

	return split[5]
}

func GetServiceAccountNameFromKeyRef(keyRef string) string {
	if !strings.HasPrefix(keyRef, constants.GCPProjectsNamePrefix) {
		log.Errorf("keyRef %s does not start with %s", keyRef, constants.GCPProjectsNamePrefix)
		return keyRef
	}

	split := strings.Split(keyRef, "/")
	if len(split) < keyParts {
		log.Errorf("keyRef %s has less than 6 parts", keyRef)
		return keyRef
	}

	return split[3]
}
