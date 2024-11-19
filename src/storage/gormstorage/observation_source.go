package gormstorage

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

type ObservationSource pb.Observation_Source

func (o ObservationSource) Value() (driver.Value, error) {
	return pb.Observation_Source_name[int32(o)], nil
}

func (o *ObservationSource) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}
	v, ok := pb.Observation_Source_value[str]
	if !ok {
		return fmt.Errorf("invalid ObservationSource: %q", str)
	}
	*o = ObservationSource(v)
	return nil
}

var _ sql.Scanner = (*ObservationSource)(nil)
var _ driver.Valuer = (*ObservationSource)(nil)
