package bigquerystorage

import (
	"time"

	"google.golang.org/protobuf/proto"
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

// This struct equals to the BigQuery schema to make storing and reading from BigQuery easier
type BigQueryResourceEntry struct {
	ResourceID        string
	ResourceName      string
	ParentName        string
	ResourceGroupName string
	CollectionID      string
	RecordTime        time.Time
	ResourceType      int
	ResourceProto     []byte
}

func toBigQueryResourceEntry(resource *pb.Resource) (*BigQueryResourceEntry, error) {
	mPbResource, err := proto.Marshal(resource)
	if err != nil {
		return nil, err
	}
	resourceTypeID, err := common.TypeFromResource(resource)
	if err != nil {
		return nil, err
	}

	entry := &BigQueryResourceEntry{
		ResourceID:        resource.Uid,
		ResourceName:      resource.Name,
		ParentName:        resource.Parent,
		ResourceGroupName: resource.ResourceGroupName,
		CollectionID:      resource.CollectionUid,
		RecordTime:        time.Now(),
		ResourceType:      resourceTypeID,
		ResourceProto:     mPbResource,
	}
	return entry, nil
}

// This struct equals to the BigQuery schema to make storing and reading from BigQuery easier
type BigQueryObservationEntry struct {
	UUID              string
	Name              string
	RecordTime        time.Time
	ScanID            string
	ObservationProto  []byte
	ResourceID        string
	ResourceGroupName string
}

func toBigQueryObservationEntry(pbo *pb.Observation) (*BigQueryObservationEntry, error) {
	mPbObservation, err := proto.Marshal(pbo)
	if err != nil {
		return nil, err
	}

	entry := &BigQueryObservationEntry{
		UUID:              pbo.Uid,
		Name:              pbo.Name,
		RecordTime:        time.Now(),
		ScanID:            pbo.ScanUid,
		ObservationProto:  mPbObservation,
		ResourceID:        pbo.Resource.Uid,
		ResourceGroupName: pbo.Resource.ResourceGroupName,
	}
	return entry, nil
}

type BigQueryOperationEntry struct {
	UUID              string
	ResourceGroupName string
	OpsType           string
	StatusTime        time.Time
	Status            string
}

func toBigqueryOperationEntry(ops model.Operation) (*BigQueryOperationEntry, error) {
	return &BigQueryOperationEntry{
		UUID:              ops.ID,
		ResourceGroupName: ops.ResourceGroup,
		OpsType:           ops.OpsType,
		Status:            ops.Status.String(),
		StatusTime:        ops.StatusTime,
	}, nil
}
