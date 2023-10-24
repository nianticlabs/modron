package gcpcollector

import (
	"fmt"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/pb"

	"golang.org/x/net/context"
	"google.golang.org/api/apikeys/v2"
)

const (
	globalProjectResourceID = "%s/locations/global"
)

func (collector *GCPCollector) ListApiKeys(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	name := fmt.Sprintf(globalProjectResourceID, constants.ResourceWithProjectsPrefix(resourceGroup.Name))

	apiKeys := []*pb.Resource{}

	keys, err := collector.api.ListApiKeys(ctx, name)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		// TODO : handle other types of GCP API keys restrictions
		// example : BrowserKeyRestrictions , AndroidKeyRestrictions , etc..
		scopes := getApiKeyScopes(key)
		apiKeys = append(apiKeys, &pb.Resource{
			Uid:               common.GetUUID(3),
			ResourceGroupName: resourceGroup.Name,
			Name:              formatResourceName(key.Name, key.Uid),
			Parent:            resourceGroup.Name,
			Type: &pb.Resource_ApiKey{
				ApiKey: &pb.APIKey{
					Scopes: scopes,
				},
			},
		})
	}
	return apiKeys, err
}

func getApiKeyScopes(key *apikeys.V2Key) []string {
	if key.Restrictions == nil || key.Restrictions.ApiTargets == nil {
		return nil
	}
	scopes := []string{}
	for _, apiTarget := range key.Restrictions.ApiTargets {
		scopes = append(scopes, apiTarget.Service)
	}
	return scopes
}
