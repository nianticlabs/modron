package gcpcollector

import (
	"fmt"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	pb "github.com/nianticlabs/modron/src/proto/generated"

	"golang.org/x/net/context"
	"google.golang.org/api/apikeys/v2"
)

const (
	globalProjectResourceID = "%s/locations/global"
)

func (collector *GCPCollector) ListAPIKeys(ctx context.Context, rgName string) (apiKeys []*pb.Resource, err error) {
	name := fmt.Sprintf(globalProjectResourceID, constants.ResourceWithProjectsPrefix(rgName))

	keys, err := collector.api.ListAPIKeys(ctx, name)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		// TODO : handle other types of GCP API keys restrictions
		// example : BrowserKeyRestrictions , AndroidKeyRestrictions , etc..
		scopes := getAPIKeyScopes(key)
		apiKeys = append(apiKeys, &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              key.Name,
			Parent:            rgName,
			Type: &pb.Resource_ApiKey{
				ApiKey: &pb.APIKey{
					Scopes: scopes,
				},
			},
		})
	}
	return apiKeys, err
}

func getAPIKeyScopes(key *apikeys.V2Key) (scopes []string) {
	if key.Restrictions == nil || key.Restrictions.ApiTargets == nil {
		return nil
	}
	for _, apiTarget := range key.Restrictions.ApiTargets {
		scopes = append(scopes, apiTarget.Service)
	}
	return scopes
}
