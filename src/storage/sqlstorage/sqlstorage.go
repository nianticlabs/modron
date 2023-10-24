package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"google.golang.org/protobuf/proto"
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

const (
	maxAttempts = 5
)

var (
	dbMaxBootUpWaitTime = 30 * time.Second // TODO: Convert to ENV variable with fallback
	insertMut           = sync.Mutex{}
	updateMut           = sync.Mutex{}
)

type Service struct {
	db  *sql.DB
	cfg Config

	insertOpsCache []model.Operation
	updateOpsCache []model.Operation
}

type Config struct {
	ResourceTableID    string
	ObservationTableID string
	OperationTableID   string
	BatchSize          int32
}

func waitForSqlDb(db *sql.DB) (err error) {
	var lastError error
	timeOut := 3 * time.Second
	for timeOut < dbMaxBootUpWaitTime {
		if err = db.Ping(); err == nil {
			return nil
		}
		time.Sleep(timeOut)
		timeOut = 3 * timeOut
	}
	return lastError
}

func New(db *sql.DB, cfg Config) (model.Storage, error) {
	if err := waitForSqlDb(db); err != nil {
		return nil, err
	}
	return Service{
		db:             db,
		cfg:            cfg,
		insertOpsCache: make([]model.Operation, 0),
		updateOpsCache: make([]model.Operation, 0),
	}, nil
}

func (svc Service) BatchCreateResources(ctx context.Context, resources []*pb.Resource) ([]*pb.Resource, error) {
	stmt, err := svc.db.PrepareContext(ctx, fmt.Sprintf("INSERT INTO %s (resourceID, resourceName, resourceGroupName, collectionID, recordTime, parentName, resourceType, resourceProto) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)", svc.cfg.ResourceTableID))
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}
	for _, res := range resources {
		if res.Uid == "" {
			res.Uid = common.GetUUID(3)
		}
		resPb, err := proto.Marshal(res)
		if err != nil {
			glog.Warningf("proto marshal %q: %v", res.Uid, err)
			continue
		}
		t, err := common.TypeFromResourceAsString(res)
		if err != nil {
			glog.Warningf("type of %q: %v", res.Uid, err)
		}
		if _, err := stmt.ExecContext(ctx, res.Uid, res.Name, res.ResourceGroupName, res.CollectionUid, res.Timestamp.AsTime(), res.Parent, t, resPb); err != nil {
			glog.Warningf("insert %q: %v", res.Uid, err)
			continue
		}
	}
	return resources, nil
}

func (svc Service) ListResources(ctx context.Context, filter model.StorageFilter) ([]*pb.Resource, error) {
	operationFilter := fmt.Sprintf(`(collectionID,resourceGroupName) IN (
                            SELECT DISTINCT operationID,resourceGroupName FROM %s WHERE (resourceGroupName, endTime) IN (
								SELECT resourceGroupName, timestamp FROM (
									SELECT resourceGroupName, max(endtime) as timestamp 
									FROM %s AS ts 
									WHERE opstype = 'collection' AND status = 'COMPLETED'
									GROUP BY resourceGroupName) AS ts1))`,
		svc.cfg.OperationTableID,
		svc.cfg.OperationTableID)
	paramFilters := sqlFilterFromModel(filter)
	allFilters := operationFilter
	if paramFilters != "" {
		allFilters = strings.Join([]string{paramFilters, allFilters}, " AND ")
	}
	q := fmt.Sprintf(`SELECT resourceID, resourceName, resourceGroupName, collectionID, recordTime, parentName, resourceType, resourceProto 
		             FROM %s WHERE %s %s`,
		svc.cfg.ResourceTableID,
		allFilters,
		limitClause(filter),
	)
	stmt, err := svc.db.PrepareContext(ctx, q)

	if err != nil {
		return nil, fmt.Errorf("prepare resource read %q: %w", q, err)
	}
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	defer rows.Close()
	resources := []*pb.Resource{}
	for rows.Next() {
		r := SQLResourceRow{}
		if err := rows.Scan(&r.ResourceID, &r.ResourceName, &r.ResourceGroupName, &r.CollectionID, &r.RecordTime, &r.ParentName, &r.ResourceType, &r.ResourceProto); err != nil {
			glog.Warningf("read row: %v", err)
			continue
		}
		res, err := r.toResourceProto()
		if err != nil {
			glog.Warningf("unmarshal: %v", err)
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func (svc Service) BatchCreateObservations(ctx context.Context, observations []*pb.Observation) ([]*pb.Observation, error) {
	stmt, err := svc.db.PrepareContext(ctx, fmt.Sprintf("INSERT INTO %s (observationID, observationName, resourceGroupName, resourceID, scanID, recordTime, observationProto) VALUES ($1,$2,$3,$4,$5,$6,$7);", svc.cfg.ObservationTableID))
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}
	for _, obs := range observations {
		obsPb, err := proto.Marshal(obs)
		if err != nil {
			glog.Warningf("proto marshal %q: %v", obs.Uid, err)
			continue
		}
		if _, err := stmt.ExecContext(ctx, obs.Uid, obs.Name, obs.Resource.ResourceGroupName, obs.Resource.Uid, obs.ScanUid, obs.Timestamp.AsTime(), obsPb); err != nil {
			glog.Warningf("insert %q: %v", obs.Uid, err)
			continue
		}
	}
	return observations, nil
}

func (svc Service) ListObservations(ctx context.Context, filter model.StorageFilter) ([]*pb.Observation, error) {
	operationFilter := fmt.Sprintf(`(scanID,resourceGroupName) IN (
                            SELECT DISTINCT operationID,resourceGroupName FROM %s WHERE (resourceGroupName, endTime) IN (
								SELECT resourceGroupName, timestamp FROM (
									SELECT resourceGroupName, max(endtime) as timestamp 
									FROM %s AS ts 
									WHERE opstype = 'scan' AND status = 'COMPLETED'
									GROUP BY resourceGroupName) AS ts1))`,
		svc.cfg.OperationTableID,
		svc.cfg.OperationTableID)
	paramFilters := sqlFilterFromModel(filter)
	allFilters := operationFilter
	if paramFilters != "" {
		allFilters = strings.Join([]string{paramFilters, allFilters}, " AND ")
	}
	q := fmt.Sprintf(`SELECT observationID, observationName, resourceGroupName, resourceID, scanID, recordTime, observationProto
		             FROM %s WHERE %s %s`,
		svc.cfg.ObservationTableID,
		allFilters,
		limitClause(filter),
	)
	stmt, err := svc.db.PrepareContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("prepare observations read %q: %w", q, err)
	}
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	defer rows.Close()
	observations := []*pb.Observation{}
	for rows.Next() {
		o := SQLObservationRow{}
		if err := rows.Scan(&o.ObservationID, &o.ObservationName, &o.ResourceGroupName, &o.ResourceID, &o.ScanID, &o.RecordTime, &o.ObservationProto); err != nil {
			glog.Warningf("read row: %v", err)
			continue
		}
		obs, err := o.toObservationProto()
		if err != nil {
			glog.Warningf("unmarshal: %v", err)
		}
		observations = append(observations, obs)
	}
	// TODO: Filter with SQL and not with Go.
	return observations, nil
}

func (svc Service) AddOperationLog(ctx context.Context, operations []model.Operation) error {
	// We need to batch these queries as this method is called very often and takes up a lot of connections to the DB that are needed to run the actual scan.
	for _, o := range operations {
		if o.Status == model.OperationStarted {
			insertMut.Lock()
			svc.insertOpsCache = append(svc.insertOpsCache, o)
			insertMut.Unlock()
			continue
		}
		updateMut.Lock()
		svc.updateOpsCache = append(svc.updateOpsCache, o)
		updateMut.Unlock()
	}
	if int32(len(svc.insertOpsCache))%svc.cfg.BatchSize == 0 || int32(len(svc.updateOpsCache))%svc.cfg.BatchSize == 0 {
		if err := svc.FlushOpsLog(ctx); err != nil {
			glog.Warningf("flush: %v", err)
		}
	}
	return nil
}

func (svc Service) PurgeIncompleteOperations(ctx context.Context) error {
	stmt, err := svc.db.PrepareContext(ctx, fmt.Sprintf("UPDATE %s SET endtime=CURRENT_TIMESTAMP, status='FAILED' WHERE endtime is NULL", svc.cfg.OperationTableID))
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (svc Service) FlushOpsLog(ctx context.Context) error {
	insertMut.Lock()
	insertOpsCache := svc.insertOpsCache
	//nolint:staticcheck // SA4005 ignore this.
	svc.insertOpsCache = svc.insertOpsCache[:0]
	insertMut.Unlock()
	updateMut.Lock()
	updateOpsCache := svc.updateOpsCache
	//nolint:staticcheck // SA4005 ignore this.
	svc.updateOpsCache = svc.updateOpsCache[:0]
	updateMut.Unlock()
	if len(insertOpsCache) > 0 {
		insertSqlStatement := fmt.Sprintf("INSERT INTO %s (operationID, resourceGroupName, opsType, startTime, status) VALUES ", svc.cfg.OperationTableID)
		values := []interface{}{}
		for i, ops := range insertOpsCache {
			insertSqlStatement += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)
			if i < len(insertOpsCache)-1 {
				insertSqlStatement += ","
			}
			values = append(values, ops.ID, ops.ResourceGroup, ops.OpsType, ops.StatusTime, ops.Status.String())
		}
		startStmt, err := svc.db.PrepareContext(ctx, insertSqlStatement)
		if err != nil {
			return fmt.Errorf("prepare insert: %w", err)
		}
		if _, err := startStmt.ExecContext(ctx, values...); err != nil {
			glog.Warningf("insert ops: %v", err)
		}
	}
	if len(updateOpsCache) > 0 {
		updateStmt, err := svc.db.PrepareContext(ctx, fmt.Sprintf("UPDATE %s SET endTime = $1, status = $2, reason = $3 WHERE operationID = $4", svc.cfg.OperationTableID))
		if err != nil {
			return fmt.Errorf("prepare update: %w", err)
		}
		for _, ops := range updateOpsCache {
			var err error
			attempts := 0
			wait := time.Second
			for err == nil {
				if _, err := updateStmt.ExecContext(ctx, ops.StatusTime, ops.Status.String(), ops.Reason, ops.ID); err != nil {

					if attempts < maxAttempts {
						time.Sleep(wait)
						wait = wait * 2
						continue
					} else {
						glog.Warningf("update %q, resource group %q: %v", ops.ID, ops.ResourceGroup, err)
					}
				}
				break
			}
		}
	}
	return nil
}
