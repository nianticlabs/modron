package risk

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/nianticlabs/modron/src/constants"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

const (
	tagKeyParts   = 2
	tagValueParts = 3
)

var log = logrus.StandardLogger().WithField(constants.LogKeyPkg, "risk")

type TagConfig struct {
	ImpactMap    map[string]pb.Impact
	Environment  string
	EmployeeData string
	CustomerData string
}

func GetRiskScore(impact pb.Impact, severity pb.Severity) pb.Severity {
	log.Debugf("risk score: impact=%s, severity=%s", impact, severity)
	switch impact {
	case pb.Impact_IMPACT_HIGH:
		switch severity {
		case pb.Severity_SEVERITY_CRITICAL:
			return pb.Severity_SEVERITY_CRITICAL
		case pb.Severity_SEVERITY_HIGH:
			return pb.Severity_SEVERITY_CRITICAL
		case pb.Severity_SEVERITY_MEDIUM:
			return pb.Severity_SEVERITY_HIGH
		case pb.Severity_SEVERITY_LOW:
			return pb.Severity_SEVERITY_MEDIUM
		case pb.Severity_SEVERITY_INFO:
			return pb.Severity_SEVERITY_LOW
		}
	case pb.Impact_IMPACT_MEDIUM:
		return severity
	case pb.Impact_IMPACT_LOW:
		switch severity {
		case pb.Severity_SEVERITY_CRITICAL:
			return pb.Severity_SEVERITY_HIGH
		case pb.Severity_SEVERITY_HIGH:
			return pb.Severity_SEVERITY_MEDIUM
		case pb.Severity_SEVERITY_MEDIUM:
			return pb.Severity_SEVERITY_LOW
		case pb.Severity_SEVERITY_LOW:
			return pb.Severity_SEVERITY_INFO
		case pb.Severity_SEVERITY_INFO:
			return pb.Severity_SEVERITY_INFO
		}
	}
	return pb.Severity_SEVERITY_UNKNOWN
}

type impactReason struct {
	impact pb.Impact
	reason string
}

func GetEnvironment(tagConfig TagConfig, hierarchy map[string]*pb.RecursiveResource, rgName string) string {
	parent := rgName
	mergedTags := map[string]string{}
	for {
		if parent == "" {
			break
		}
		v, ok := hierarchy[parent]
		if !ok {
			break
		}

		for k, v := range v.Tags {
			if _, ok := mergedTags[k]; !ok {
				// Label not found, adding
				mergedTags[k] = v
			}
		}
		parent = v.Parent
	}

	if v, ok := mergedTags[tagConfig.Environment]; ok {
		tagValue := humanReadableTagValue(v)
		return tagValue
	}
	return ""
}

// humanReadableTagKey converts a tag name in the format "111111111111/employee_data" to employee_data
func humanReadableTagKey(tagKey string) string {
	split := strings.SplitN(tagKey, "/", tagKeyParts)
	if len(split) != tagKeyParts {
		log.Warnf("unexpected tag key format: %q", tagKey)
		return tagKey
	}
	return split[1]
}

// humanReadableTagValue converts a tag value from the format "111111111111/environment/prod" to prod
func humanReadableTagValue(tagValue string) string {
	split := strings.SplitN(tagValue, "/", tagValueParts)
	if len(split) != tagValueParts {
		log.Warnf("unexpected tag value format: %q", tagValue)
		return tagValue
	}
	return split[2]
}

// GetImpact computes the impact of a resource group based on the tags set in its hierarchy.
// The label at the deepest level of the hierarchy is the one that will be considered first.
// If a tags is set at the organization label, it might be overwritten by the project label - because they
// take precedence in our impact analysis.
//
// If we find something that has a high impact, we immediately return since this is the highest possible impact.
// In case of multiple impacts, we return the highest one.
func GetImpact(tagConfig TagConfig, hierarchy map[string]*pb.RecursiveResource, rgName string) (impact pb.Impact, reason string) {
	parent := rgName
	mergedTags := map[string]string{}
	for {
		if parent == "" {
			break
		}
		v, ok := hierarchy[parent]
		if !ok {
			break
		}

		for k, v := range v.Tags {
			if _, ok := mergedTags[k]; !ok {
				// Label not found, adding
				mergedTags[k] = v
			}
		}
		parent = v.Parent
	}
	impactLogger := log.WithField("resource_group", rgName)

	var impacts []impactReason
	env := GetEnvironment(tagConfig, hierarchy, rgName)
	if env != "" {
		impacts = append(impacts, impactReason{
			impact: impactFromEnvironment(tagConfig.ImpactMap, env),
			reason: fmt.Sprintf("%s=%s", humanReadableTagKey(tagConfig.Environment), env),
		})
	}

	if v, ok := mergedTags[tagConfig.EmployeeData]; ok {
		tagValue := humanReadableTagValue(v)
		switch strings.ToLower(tagValue) {
		case "yes":
			impacts = append(impacts, impactReason{
				impact: constants.ImpactEmployeeData,
				reason: fmt.Sprintf("%s=%s", constants.ResourceLabelEmployeeData, tagValue),
			})
		case "no":
			// All good
		default:
			impactLogger.Warnf("unknown value for label %s: %q", constants.ResourceLabelEmployeeData, tagValue)
		}
	}

	if v, ok := mergedTags[tagConfig.CustomerData]; ok {
		tagValue := humanReadableTagValue(v)
		switch strings.ToLower(tagValue) {
		case "yes":
			impacts = append(impacts, impactReason{
				impact: constants.ImpactCustomerData,
				reason: fmt.Sprintf("%s=%s", constants.ResourceLabelCustomerData, tagValue),
			})
		case "no":
			// All good
		default:
			impactLogger.Warnf("unknown value for label %s: %q", constants.ResourceLabelCustomerData, tagValue)
		}
	}

	log.Debugf("mergedTags=%+v, impacts=%+v", mergedTags, impacts)
	if len(impacts) == 0 {
		log.Debugf("no facts that would change the impact")
		return pb.Impact_IMPACT_MEDIUM, reason
	}
	highestImpact := pb.Impact_IMPACT_UNKNOWN
	for _, i := range impacts {
		if i.impact > highestImpact {
			highestImpact = i.impact
			reason = i.reason
		}
	}
	return highestImpact, reason
}

func impactFromEnvironment(impactMap map[string]pb.Impact, env string) pb.Impact {
	v, ok := impactMap[env]
	if !ok {
		defaultImpact := pb.Impact_IMPACT_MEDIUM
		log.Warnf("no impact found for environment %q, using %s", env, defaultImpact.String())
		return defaultImpact
	}
	return v
}
