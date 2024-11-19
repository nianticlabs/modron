package rules

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

func TestCheckDetectsOverbroadScope(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "folders/123",
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "projects/project-1",
			Parent:            "folders/234",
			ResourceGroupName: "projects/project-1",
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "api-key-unrestricted-0",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_ApiKey{
				ApiKey: &pb.APIKey{
					Scopes: nil,
				},
			},
		},
		{
			Name:              "api-key-with-overbroad-scope-1",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_ApiKey{
				ApiKey: &pb.APIKey{
					Scopes: []string{
						// overbroad
						"iamcredentials.googleapis.com",
						// not overbroad
						"storage_api",
						// overbroad
						"apikeys",
					},
				},
			},
		},
		{
			Name:              "api-key-without-overbroad-scope",
			Parent:            "projects/project-1",
			ResourceGroupName: "projects/project-1",
			Type: &pb.Resource_ApiKey{
				ApiKey: &pb.APIKey{
					Scopes: []string{"bigquerystorage.googleapis.com"},
				},
			},
		},
	}

	res1 := pb.Resource{Name: "api-key-with-overbroad-scope-1",
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{
					"iamcredentials.googleapis.com",
					"storage_api",
					"apikeys",
				},
			},
		},
		ResourceGroupName: "projects/project-0",
		Parent:            "projects/project-0",
	}

	want := []*pb.Observation{
		{
			Name:          apiKeyOverbroadScopeRuleName,
			ResourceRef:   utils.GetResourceRef(&res1),
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("iamcredentials.googleapis.com"),
			Remediation: &pb.Remediation{
				Description:    "API key [\"api-key-with-overbroad-scope-1\"](https://console.cloud.google.com/apis/credentials/key/api-key-with-overbroad-scope-1?project=project-0) may have over-broad scope \"iamcredentials.googleapis.com\"",
				Recommendation: "Remove scope \"iamcredentials.googleapis.com\" from API key [\"api-key-with-overbroad-scope-1\"](https://console.cloud.google.com/apis/credentials/key/api-key-with-overbroad-scope-1?project=project-0) unless it is used. [More details available in our documentation.](https://github.com/nianticlabs/modron/blob/main/docs/FINDINGS.md)",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name:          apiKeyOverbroadScopeRuleName,
			ResourceRef:   utils.GetResourceRef(&res1),
			ExpectedValue: structpb.NewStringValue(""),
			ObservedValue: structpb.NewStringValue("apikeys"),
			Remediation: &pb.Remediation{
				Description:    "API key [\"api-key-with-overbroad-scope-1\"](https://console.cloud.google.com/apis/credentials/key/api-key-with-overbroad-scope-1?project=project-0) may have over-broad scope \"apikeys\"",
				Recommendation: "Remove scope \"apikeys\" from API key [\"api-key-with-overbroad-scope-1\"](https://console.cloud.google.com/apis/credentials/key/api-key-with-overbroad-scope-1?project=project-0) unless it is used. [More details available in our documentation.](https://github.com/nianticlabs/modron/blob/main/docs/FINDINGS.md)",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name: apiKeyOverbroadScopeRuleName,
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-2"),
				GroupName:     "projects/project-0",
				ExternalId:    proto.String("api-key-unrestricted-0"),
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			ExpectedValue: structpb.NewStringValue("restricted"),
			ObservedValue: structpb.NewStringValue("unrestricted"),
			Remediation: &pb.Remediation{
				Description:    "API key [\"api-key-unrestricted-0\"](https://console.cloud.google.com/apis/credentials/key/api-key-unrestricted-0?project=project-0) is unrestricted, which allows it to be used against any enabled GCP API",
				Recommendation: "Restrict API key [\"api-key-unrestricted-0\"](https://console.cloud.google.com/apis/credentials/key/api-key-unrestricted-0?project=project-0) strictly to the APIs it is supposed to call. [More details available in our documentation.](https://github.com/nianticlabs/modron/blob/main/docs/FINDINGS.md)",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewAPIKeyOverbroadScopeRule()}, want)
}
