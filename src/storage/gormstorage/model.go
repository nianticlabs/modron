package gormstorage

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

// Resource represents a resource entry in the database.
type Resource struct {
	ID                string          `gorm:"column:resourceid;primaryKey"`
	Name              string          `gorm:"column:resourcename;index:idx_resources_resourcename"`
	DisplayName       string          `gorm:"column:display_name"`
	ResourceGroupName string          `gorm:"column:resourcegroupname;index:idx_resources_resourcegroupname;index:idx_resources_resourcetype_resourcegroupname;index:idx_collectionid_rgname"`
	CollectionID      string          `gorm:"column:collectionid;index:idx_resources_collectionid;index:idx_collectionid_rgname"`
	RecordTime        *time.Time      `gorm:"column:recordtime;index:idx_resource_recordtime"`
	ParentName        string          `gorm:"column:parentname"`
	Type              string          `gorm:"column:resourcetype;index:idx_resource_resourcetype;index:idx_resources_resourcetype_resourcegroupname"`
	Labels            json.RawMessage `gorm:"column:labels;type:jsonb"`
	Tags              json.RawMessage `gorm:"column:tags;type:jsonb"`
	Proto             []byte          `gorm:"column:resourceproto"`
}

// Observation represents an observation entry in the database.
type Observation struct {
	ID   string `gorm:"column:observationid;primaryKey"`
	Name string `gorm:"column:observationname;not null"`

	// Observations can be associated with either a scan (result of a Modron rule engine execution) or
	// a collection (result of fetching them from an external source)
	ScanID       *string `gorm:"column:scanid;index:idx_observation_scanid;index:idx_observation_resourcegroupname_scan"`
	CollectionID *string `gorm:"column:collectionid;index:idx_observation_collectionid"`

	RecordTime time.Time `gorm:"column:recordtime;not null;index:idx_observation_recordtime"`
	Proto      []byte    `gorm:"column:observationproto;not null"`

	// ResourceID is the UUID of the resource, as per the Resource table.
	Resource              *Resource           `gorm:"foreignKey:ResourceID;references:ID"`
	ResourceID            *string             `gorm:"column:resourceid"`
	ResourceGroupName     string              `gorm:"column:resourcegroupname;not null;index:idx_observation_resourcegroupname;index:idx_observation_resourcegroupname_scan"`
	ResourceExternalID    *string             `gorm:"default:null;index:idx_observation_resource_external_id"`
	ResourceCloudPlatform string              `gorm:"column:resourcecloudplatform;not null;default:'GCP'"`
	ExternalID            *string             `gorm:"index:idx_observation_external_id"`
	Source                ObservationSource   `gorm:"column:source;index:idx_observation_source"`
	Category              ObservationCategory `gorm:"column:category"`
	// SeverityScore represents the original severity (as set by the rule / external observation provider) without taking into account he impact
	SeverityScore *SeverityScore `gorm:"column:severity_score;default:null;index:idx_observation_severity_score"`
	// Impact is calculated by looking at the Resource Group details (e.g: environment, tags) so that it can be used to calculate the Risk Score
	Impact Impact `gorm:"column:impact;default:null"`
	// RiskScore represents the final risk score calculated by using the SeverityScore and Impact
	RiskScore *SeverityScore `gorm:"column:risk_score;default:null;index:idx_observation_risk_score"`
}

// ToObservationProto converts an Observation to a pb.Observation.
func (row Observation) ToObservationProto() (*pb.Observation, error) {
	obs := &pb.Observation{}
	err := proto.Unmarshal(row.Proto, obs)
	if err != nil {
		return nil, fmt.Errorf("unmarshal observation proto: %w", err)
	}
	obs.Uid = row.ID
	obs.Name = row.Name
	obs.ScanUid = row.ScanID
	obs.Timestamp = timestamppb.New(row.RecordTime)

	// DB fields
	obs.ResourceRef = &pb.ResourceRef{
		Uid:           row.ResourceID,
		GroupName:     row.ResourceGroupName,
		CloudPlatform: cloudPlatformFromString(row.ResourceCloudPlatform),
		ExternalId:    row.ResourceExternalID,
	}
	obs.Severity = ToSeverity(row.SeverityScore)
	obs.Source = pb.Observation_Source(row.Source)
	obs.Category = pb.Observation_Category(row.Category)
	return obs, nil
}

func ToSeverity(score *SeverityScore) pb.Severity {
	if score == nil {
		return pb.Severity_SEVERITY_UNKNOWN
	}
	return score.ToSeverity()
}

func cloudPlatformFromString(platform string) pb.CloudPlatform {
	switch platform {
	case "GCP":
		return pb.CloudPlatform_GCP
	case "AWS":
		return pb.CloudPlatform_AWS
	case "AZURE":
		return pb.CloudPlatform_AZURE
	}
	return pb.CloudPlatform_PLATFORM_UNKNOWN
}

// ToResourceProto converts a Resource to a pb.Resource.
func (row Resource) ToResourceProto() (*pb.Resource, error) {
	res := &pb.Resource{}
	err := proto.Unmarshal(row.Proto, res)
	if err != nil {
		return nil, fmt.Errorf("unmarshal resource proto: %w", err)
	}
	res.Uid = row.ID
	res.Name = row.Name
	res.ResourceGroupName = row.ResourceGroupName
	res.CollectionUid = row.CollectionID
	if row.RecordTime != nil {
		res.Timestamp = timestamppb.New(*row.RecordTime)
	}
	res.Parent = row.ParentName
	var labels map[string]string
	if len(row.Labels) != 0 {
		if err := json.Unmarshal(row.Labels, &labels); err != nil {
			return nil, fmt.Errorf("unmarshal labels: %w", err)
		}
		res.Labels = labels
	}
	var tags map[string]string
	if len(row.Tags) != 0 {
		if err := json.Unmarshal(row.Tags, &tags); err != nil {
			return nil, fmt.Errorf("unmarshal tags: %w", err)
		}
		res.Tags = tags
	}
	return res, nil
}
