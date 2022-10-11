package gcpcollector

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/api/apikeys/v2"
	"github.com/nianticlabs/modron/src/pb"
)

const (
	globalProjectResourceID = "projects/%s/locations/global"
)

func (collector *GCPCollector) ListApiKeys(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	name := fmt.Sprintf(globalProjectResourceID, resourceGroup.Name)

	apiKeys := []*pb.Resource{}

	req, err := collector.api.ListApiKeys(name)
	if err != nil {
		return nil, err
	}
	for _, key := range req.Keys {
		// TODO : handle other types of GCP API keys restrictions
		// example : BrowserKeyRestrictions , AndroidKeyRestrictions , etc..
		scopes := getApiKeyScopes(key)
		apiKeys = append(apiKeys, &pb.Resource{
			Uid:               collector.getNewUid(),
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
