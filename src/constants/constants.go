package constants

import (
	"strings"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

type Role string

const (
	GCPEditorRole        Role = "editor"
	GCPOwnerRole         Role = "owner"
	GCPSecurityAdminRole Role = "iam.securityAdmin"
	GCPViewerRole        Role = "viewer"
)

func (r Role) String() string {
	return string(r)
}

const (
	GCPOrgIDPrefix          = "organizations/"
	GCPFolderIDPrefix       = "folders/"
	GCPProjectsNamePrefix   = "projects/"
	GCPRolePrefix           = "roles/"
	GCPSysProjectPrefix     = "sys-"
	GCPAccountGroupPrefix   = "group:"
	GCPServiceAccountPrefix = "serviceAccount:"
	GCPUserAccountPrefix    = "user:"

	MetricsPrefix = "modron_"

	ResourceLabelCustomerData = "customer_data"
	ResourceLabelEmployeeData = "employee_data"

	LabelContact1 = "contact1"
	LabelContact2 = "contact2"

	ImpactEmployeeData = pb.Impact_IMPACT_MEDIUM
	ImpactCustomerData = pb.Impact_IMPACT_HIGH
)

const (
	LogKeyCollectID          = "collect_id"
	LogKeyCollector          = "collector"
	LogKeyObservationName    = "observation_name"
	LogKeyObservationUID     = "observation_uid"
	LogKeyPkg                = "package"
	LogKeyResourceGroup      = "resource_group"
	LogKeyResourceGroupNames = "resource_group_names"
	LogKeyRule               = "rule"
	LogKeyScanID             = "scan_id"
)

const (
	TraceKeyCollectID          = "collect_id"
	TraceKeyCollector          = "collector"
	TraceKeyMethod             = "method"
	TraceKeyName               = "name"
	TraceKeyNumNotifications   = "num_notifications"
	TraceKeyNumObservations    = "num_observations"
	TraceKeyNumResources       = "num_resources"
	TraceKeyObservationUID     = "observation_uid"
	TraceKeyPath               = "path"
	TraceKeyResourceGroup      = "resource_group"
	TraceKeyResourceGroupNames = "resource_group_names"
	TraceKeyRule               = "rule"
	TraceKeyScanID             = "scan_id"
	TraceKeyScanType           = "scan_type"
)

const (
	MetricKeyStatus = "status"
)

var AdminRoles = map[Role]struct{}{
	GCPOwnerRole:         {},
	GCPEditorRole:        {},
	GCPSecurityAdminRole: {},
}

func ToRole(role string) Role {
	return Role(strings.TrimPrefix(role, GCPRolePrefix))
}

func ResourceWithProjectsPrefix(resourceName string) string {
	if strings.HasPrefix(resourceName, GCPProjectsNamePrefix) {
		return resourceName
	}
	return GCPProjectsNamePrefix + resourceName
}

func ResourceWithoutProjectsPrefix(resourceName string) string {
	return strings.TrimPrefix(resourceName, GCPProjectsNamePrefix)
}
