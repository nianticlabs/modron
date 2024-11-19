package gormstorage

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

type ObservationCategory pb.Observation_Category

func (o ObservationCategory) Value() (driver.Value, error) {
	return pb.Observation_Category_name[int32(o)], nil
}

func (o *ObservationCategory) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}
	v, ok := pb.Observation_Category_value[str]
	if !ok {
		return fmt.Errorf("invalid ObservationCategory: %q", str)
	}
	*o = ObservationCategory(v)
	return nil
}

var _ sql.Scanner = (*ObservationCategory)(nil)
var _ driver.Valuer = (*ObservationCategory)(nil)
