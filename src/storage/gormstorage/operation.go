package gormstorage

import (
	"database/sql/driver"
	"fmt"
	"time"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

var _ driver.Valuer = (*OperationStatus)(nil)

type OperationStatus pb.Operation_Status

func (o OperationStatus) Value() (driver.Value, error) {
	return pb.Operation_Status(o).String(), nil
}

func (o *OperationStatus) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}
	v, ok := pb.Operation_Status_value[str]
	if !ok {
		return fmt.Errorf("invalid OperationStatus: %q", str)
	}
	*o = OperationStatus(v)
	return nil
}

type Operation struct {
	ID            string          `gorm:"column:operationid;primaryKey;index:operations_idx"`
	ResourceGroup string          `gorm:"column:resourcegroupname;index;primaryKey;index:operations_idx;index:operations_rgname_endtime"`
	OpsType       string          `gorm:"column:opstype;index;primaryKey;index:operations_idx"`
	StartTime     time.Time       `gorm:"column:starttime;index"`
	EndTime       *time.Time      `gorm:"column:endtime;index;index:operations_idx;index:operations_rgname_endtime"`
	Status        OperationStatus `gorm:"index;index:operations_idx"`
	Reason        string
}
