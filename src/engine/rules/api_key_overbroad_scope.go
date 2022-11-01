package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const ApiKeyOverbroadScopeRuleName = "API_KEY_WITH_OVERBROAD_SCOPE"

// TODO: Complete and/or remove excess scopes.
// Uncommon scopes for API keys that should be marked as overbroad.
var overbroadScopes = []string{
	"apikeys",
	"cloudasset.googleapis.com",
	"autoscaling.googleapis.com",
	"cloudbuild.googleapis.com",
	"deploymentmanager",
	"containerfilesystem.googleapis.com",
	"containerregistry.googleapis.com",
	"iam",
	"iamcredentials.googleapis.com",
	"source.googleapis.com",
	"vmmigration.googleapis.com",
}

type ApiKeyOverbroadScopeRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewApiKeyOverbroadScopeRule())
}

func NewApiKeyOverbroadScopeRule() model.Rule {
	return &ApiKeyOverbroadScopeRule{
		info: model.RuleInfo{
			Name: ApiKeyOverbroadScopeRuleName,
			AcceptedResourceTypes: []string{
				common.ResourceApiKey,
			},
		},
	}
}

// TODO: move this to the collector for the selfLink field
func toURLKey(name string) string {
	name = engine.GetGcpReadableResourceName(name)
	return name[strings.LastIndex(name, "/")+1:]
}

func (r *ApiKeyOverbroadScopeRule) Check(ctx context.Context, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	key := rsrc.GetApiKey()

	// If the key has no scopes, it is unrestricted.
	if len(key.Scopes) == 0 {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue("restricted"),
			ObservedValue: structpb.NewStringValue("unrestricted"),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"API key [%q](https://console.cloud.google.com/apis/credentials/key/%s?project=%s) is unrestricted, which allows it to be used against any enabled GCP API.",
					toURLKey(rsrc.Name),
					toURLKey(rsrc.Name),
					rsrc.ResourceGroupName,
				),
				Recommendation: fmt.Sprintf(
					"Restrict API key [%q](https://console.cloud.google.com/apis/credentials/key/%s?project=%s) strictly to the APIs it is supposed to call.",
					toURLKey(rsrc.Name),
					toURLKey(rsrc.Name),
					rsrc.ResourceGroupName,
				),
			},
		}
		obs = append(obs, ob)
		return
	}

	for _, scope := range key.Scopes {
		if slices.Contains(overbroadScopes, scope) {
			ob := &pb.Observation{
				Uid:           uuid.NewString(),
				Timestamp:     timestamppb.Now(),
				Resource:      rsrc,
				Name:          r.Info().Name,
				ExpectedValue: structpb.NewStringValue(""),
				ObservedValue: structpb.NewStringValue(scope),
				Remediation: &pb.Remediation{
					Description: fmt.Sprintf(
						"API key [%q](https://console.cloud.google.com/apis/credentials/key/%s?project=%s) may have over-broad scope %q.",
						toURLKey(rsrc.Name),
						toURLKey(rsrc.Name),
						rsrc.ResourceGroupName,
						scope,
					),
					Recommendation: fmt.Sprintf(
						"Remove scope %q from API key [%q](https://console.cloud.google.com/apis/credentials/key/%s?project=%s) unless it is used.",
						scope,
						toURLKey(rsrc.Name),
						toURLKey(rsrc.Name),
						rsrc.ResourceGroupName,
					),
				},
			}
			obs = append(obs, ob)
		}
	}

	return
}

func (r *ApiKeyOverbroadScopeRule) Info() *model.RuleInfo {
	return &r.info
}
