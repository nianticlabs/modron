package fakeacl

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
)

type GcpCheckerFake struct{}

var log = logrus.StandardLogger().WithField(constants.LogKeyPkg, "fakeacl")
var _ model.Checker = (*GcpCheckerFake)(nil)

func New() model.Checker {
	log.Warnf("If you see this on production, contact security")
	return &GcpCheckerFake{}
}

func (checker *GcpCheckerFake) GetACL() model.ACLCache {
	return nil
}

func (checker *GcpCheckerFake) GetValidatedUser(_ context.Context) (string, error) {
	return "", nil
}

func (checker *GcpCheckerFake) ListResourceGroupNamesOwned(_ context.Context) (map[string]struct{}, error) {
	return map[string]struct{}{"projects/modron-test": {}}, nil
}
