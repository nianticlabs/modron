package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const apiKeyOverbroadScopeRuleName = "API_KEY_WITH_OVERBROAD_SCOPE" //nolint:gosec

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

type APIKeyOverbroadScopeRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewAPIKeyOverbroadScopeRule())
}

func NewAPIKeyOverbroadScopeRule() model.Rule {
	return &APIKeyOverbroadScopeRule{
		info: model.RuleInfo{
			Name: apiKeyOverbroadScopeRuleName,
			AcceptedResourceTypes: []proto.Message{
				&pb.APIKey{},
			},
		},
	}
}

// TODO: move this to the collector for the selfLink field
func toURLKey(name string) string {
	name = getGcpReadableResourceName(name)
	return name[strings.LastIndex(name, "/")+1:]
}

func (r *APIKeyOverbroadScopeRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	key := rsrc.GetApiKey()

	// If the key has no scopes, it is unrestricted.
	if len(key.Scopes) == 0 {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue("restricted"),
			ObservedValue: structpb.NewStringValue("unrestricted"),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"API key [%q](https://console.cloud.google.com/apis/credentials/key/%s?project=%s) is unrestricted, which allows it to be used against any enabled GCP API",
					toURLKey(rsrc.Name),
					toURLKey(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
				Recommendation: fmt.Sprintf(
					"Restrict API key [%q](https://console.cloud.google.com/apis/credentials/key/%s?project=%s) strictly to the APIs it is supposed to call. [More details available in our documentation.](https://github.com/nianticlabs/modron/blob/main/docs/FINDINGS.md)",
					toURLKey(rsrc.Name),
					toURLKey(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		}
		obs = append(obs, ob)
		return
	}

	for _, scope := range key.Scopes {
		if slices.Contains(overbroadScopes, scope) {
			ob := &pb.Observation{
				Uid:           uuid.NewString(),
				Timestamp:     timestamppb.Now(),
				ResourceRef:   utils.GetResourceRef(rsrc),
				Name:          r.Info().Name,
				ExpectedValue: structpb.NewStringValue(""),
				ObservedValue: structpb.NewStringValue(scope),
				Remediation: &pb.Remediation{
					Description: fmt.Sprintf(
						"API key [%q](https://console.cloud.google.com/apis/credentials/key/%s?project=%s) may have over-broad scope %q",
						toURLKey(rsrc.Name),
						toURLKey(rsrc.Name),
						constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
						scope,
					),
					Recommendation: fmt.Sprintf(
						"Remove scope %q from API key [%q](https://console.cloud.google.com/apis/credentials/key/%s?project=%s) unless it is used. [More details available in our documentation.](https://github.com/nianticlabs/modron/blob/main/docs/FINDINGS.md)",
						scope,
						toURLKey(rsrc.Name),
						toURLKey(rsrc.Name),
						constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
					),
				},
				Severity: pb.Severity_SEVERITY_MEDIUM,
			}
			obs = append(obs, ob)
		}
	}

	return
}

func (r *APIKeyOverbroadScopeRule) Info() *model.RuleInfo {
	return &r.info
}
