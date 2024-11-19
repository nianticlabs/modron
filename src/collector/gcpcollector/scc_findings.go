package gcpcollector

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/google/uuid"
	"google.golang.org/api/securitycenter/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	modronmetric "github.com/nianticlabs/modron/src/metric"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

// https://cloud.google.com/security-command-center/docs/finding-classes
// https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings#findingclass
var reportingClasses = map[string]struct{}{
	"VULNERABILITY":     {},
	"MISCONFIGURATION":  {},
	"TOXIC_COMBINATION": {},
}

// This is a list of SCC categories that relate to containers
// we use this to filter out some packages that in the specified category do not make sense
// For example, the package "linux" is not relevant in the context of a container image
// because the kernel being used is not the one from the container image but the one from the host, thus
// we exclude it from `GKE_RUNTIME_OS_VULNERABILITY`
var skipPackagesByCategories = map[string]map[string]struct{}{
	"GKE_RUNTIME_OS_VULNERABILITY": {
		"linux": {},
	},
	"GKE_RUNTIME_LANG_VULNERABILITY": {},
}

func (collector *GCPCollector) ListSccFindings(ctx context.Context, rgName string) (observations []*pb.Observation, err error) {
	rgLogger := log.WithField(constants.LogKeyResourceGroup, rgName)
	rgLogger.Info("Listing SCC findings")
	sccFindings, err := collector.api.ListSccFindings(ctx, rgName)
	if err != nil {
		rgLogger.WithError(err).Errorf("Error listing SCC findings")
		return nil, err
	}
	rgLogger.Infof("Found %d SCC findings", len(sccFindings))

	for _, v := range sccFindings {
		if _, ok := reportingClasses[v.FindingClass]; !ok {
			continue
		}
		if _, ok := collector.allowedSccCategories[v.Category]; !ok {
			log.Infof("Skipping SCC finding with category %q", v.Category)
			continue
		}

		if v.Vulnerability != nil && v.Vulnerability.OffendingPackage != nil {
			offPkg := v.Vulnerability.OffendingPackage
			if pbc, ok := skipPackagesByCategories[v.Category]; ok {
				if _, ok := pbc[offPkg.PackageName]; ok {
					log.
						WithField(modronmetric.KeyCategory, v.Category).
						WithField(modronmetric.KeyOffendingPackage, offPkg).
						Infof("Skipping SCC finding because it has been excluded")
					continue
				}
			}
		}

		observations = append(observations, FindingToObservation(v, rgName))
		collector.metrics.SccCollectedObservations.
			Add(ctx, 1, metric.WithAttributes(
				attribute.String(modronmetric.KeyCategory, v.Category),
				attribute.String(modronmetric.KeySeverity, v.Severity),
			))
	}
	rgLogger.Infof("Collected %d observations", len(observations))
	return
}

func FindingToObservation(v *securitycenter.Finding, rgName string) *pb.Observation {
	obsTime, err := time.Parse(time.RFC3339, v.EventTime)
	if err != nil {
		log.Warnf("unable to parse time: %v", err)
		obsTime = time.Now()
	}
	return enrichObservation(&pb.Observation{
		Uid:           uuid.NewString(),
		Timestamp:     timestamppb.New(obsTime),
		Name:          v.Category,
		ExpectedValue: nil,
		ObservedValue: nil,
		Remediation:   getRemediation(v),
		ResourceRef: &pb.ResourceRef{
			GroupName:     rgName,
			ExternalId:    utils.RefOrNull(v.ResourceName),
			CloudPlatform: pb.CloudPlatform_GCP,
		},
		ExternalId: utils.RefOrNull(fmt.Sprintf("//securitycenter.googleapis.com/%s", v.CanonicalName)),
		Severity:   fromSccSeverity(v.Severity),
		Source:     pb.Observation_SOURCE_SCC,
		Category:   fromSccFindingClass(v.FindingClass),
	}, v)
}

func getPackageValue(pkg *securitycenter.Package) *structpb.Value {
	if pkg == nil {
		return nil
	}
	return structpb.NewStringValue(fmt.Sprintf("%s %s", pkg.PackageName, pkg.PackageVersion))
}

var kubernetesClusterRegex = regexp.MustCompile("^//container.googleapis.com/projects/[^/]+/locations/[^/]+/clusters/([^/]+)$")

// EnrichObservation enriches the observation by modifying some fields based on the finding
func enrichObservation(p *pb.Observation, v *securitycenter.Finding) *pb.Observation {
	if v == nil {
		return p
	}
	obsLogger := log.
		WithField("observation", p.Name).
		WithField("scc_finding_name", v.Name).
		WithField("scc_finding_class", v.FindingClass).
		WithField("scc_finding_category", v.Category)
	if v.Vulnerability != nil {
		if v.Vulnerability.FixedPackage != nil {
			p.ExpectedValue = getPackageValue(v.Vulnerability.FixedPackage)
		}
		if v.Vulnerability.OffendingPackage != nil {
			p.ObservedValue = getPackageValue(v.Vulnerability.OffendingPackage)
		}
	}
	if v.Kubernetes != nil { //nolint:nestif
		// We have a Kubernetes reference, let's use it
		if kubernetesClusterRegex.MatchString(v.ResourceName) {
			if len(v.Kubernetes.Objects) > 0 {
				obj := v.Kubernetes.Objects[0]
				extID := ""
				if p.ResourceRef.ExternalId != nil {
					extID = *p.ResourceRef.ExternalId
				}
				switch obj.Kind {
				case "Deployment":
					extID = fmt.Sprintf("%s/k8s/namespaces/%s/apps/deployments/%s", extID, obj.Ns, obj.Name)
				case "DaemonSet":
					extID = fmt.Sprintf("%s/k8s/namespaces/%s/apps/daemonsets/%s", extID, obj.Ns, obj.Name)
				case "NodePool":
					extID = fmt.Sprintf("%s/nodePools/%s", extID, obj.Name)
				default:
					obsLogger.Warnf("unhandled Kubernetes object kind: %s", obj.Kind)
				}
				p.ResourceRef.ExternalId = utils.RefOrNull(extID)
			} else {
				obsLogger.Warnf("no Kubernetes objects found for %s", v.ResourceName)
			}
		}
	}
	return p
}

func getRemediation(v *securitycenter.Finding) *pb.Remediation {
	if v == nil {
		return nil
	}
	vuln := v.Vulnerability
	if vuln != nil {
		if vuln.OffendingPackage != nil && vuln.FixedPackage != nil {
			return &pb.Remediation{
				Description: fmt.Sprintf("%s in %s %s%s: update to %s %s",
					vuln.Cve.Id,
					vuln.OffendingPackage.PackageName, vuln.OffendingPackage.PackageVersion,
					getContainerImage(v.Kubernetes),
					vuln.FixedPackage.PackageName, vuln.FixedPackage.PackageVersion,
				) + "\n\n" + v.Description,
				Recommendation: v.NextSteps,
			}
		}
	}
	return &pb.Remediation{
		Description:    v.Description,
		Recommendation: v.NextSteps,
	}
}

func getContainerImage(kubernetes *securitycenter.Kubernetes) string {
	if kubernetes == nil {
		return ""
	}

	if len(kubernetes.Objects) == 0 {
		return ""
	}

	firstObj := kubernetes.Objects[0]
	if len(firstObj.Containers) == 0 {
		return ""
	}

	// TODO: We assume the first container of the first object is vulnerable - this might be wrong
	return fmt.Sprintf(" (%s)", firstObj.Containers[0].ImageId)
}

// https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings#findingclass
func fromSccFindingClass(findingClass string) pb.Observation_Category {
	switch findingClass {
	case "VULNERABILITY":
		return pb.Observation_CATEGORY_VULNERABILITY
	case "MISCONFIGURATION":
		return pb.Observation_CATEGORY_MISCONFIGURATION
	case "TOXIC_COMBINATION":
		return pb.Observation_CATEGORY_TOXIC_COMBINATION
	}
	return pb.Observation_CATEGORY_UNKNOWN
}

func fromSccSeverity(severity string) pb.Severity {
	switch severity {
	case "CRITICAL":
		return pb.Severity_SEVERITY_CRITICAL
	case "HIGH":
		return pb.Severity_SEVERITY_HIGH
	case "MEDIUM":
		return pb.Severity_SEVERITY_MEDIUM
	case "LOW":
		return pb.Severity_SEVERITY_LOW
	}
	return pb.Severity_SEVERITY_UNKNOWN
}
