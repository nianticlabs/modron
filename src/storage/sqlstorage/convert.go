package sqlstorage

import (
	"time"

	"google.golang.org/protobuf/proto"
	"github.com/nianticlabs/modron/src/pb"
)

type SQLResourceRow struct {
	ResourceID        string
	ResourceName      string
	ResourceGroupName string
	CollectionID      string
	RecordTime        time.Time
	ParentName        string
	ResourceType      string
	ResourceProto     []byte
}

func (rr SQLResourceRow) toResourceProto() (*pb.Resource, error) {
	res := new(pb.Resource)
	err := proto.Unmarshal(rr.ResourceProto, res)
	return res, err
}

type SQLObservationRow struct {
	ObservationID     string
	ObservationName   string
	ResourceID        string
	ResourceGroupName string
	ScanID            string
	RecordTime        time.Time
	ObservationProto  []byte
}

func (or SQLObservationRow) toObservationProto() (*pb.Observation, error) {
	obs := new(pb.Observation)
	err := proto.Unmarshal(or.ObservationProto, obs)
	return obs, err
}
