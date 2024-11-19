package gormstorage

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

type Impact pb.Impact

func (i *Impact) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}
	v, ok := pb.Impact_value[str]
	if !ok {
		return fmt.Errorf("invalid Impact: %q", str)
	}
	*i = Impact(v)
	return nil
}

func (i Impact) Value() (driver.Value, error) {
	return pb.Impact_name[int32(i)], nil
}

var _ sql.Scanner = (*Impact)(nil)
var _ driver.Valuer = (*Impact)(nil)
