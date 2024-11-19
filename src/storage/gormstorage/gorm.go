package gormstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	gormtracing "gorm.io/plugin/opentelemetry/tracing"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

const (
	createBatchSize              = 500
	listResourcesLimit           = 1_000_000
	listObservationsLimit        = 1_000_000
	maxUUIDRetries               = 3
	sqlDBConnectionTimeout       = 3 * time.Second
	sqlDBConnectionTimeoutFactor = 3
	dbMaxBootUpWaitTime          = 30 * time.Second

	collectionOp opType = "collection"
	scanOp       opType = "scan"
)

type opType string

var (
	log    = logrus.StandardLogger().WithField(constants.LogKeyPkg, "gormstorage")
	tracer = otel.Tracer("github.com/nianticlabs/modron/src/storage/gormstorage")
)

// BatchCreateResources creates a batch of resources.
func (svc *Service) BatchCreateResources(ctx context.Context, resources []*pb.Resource) ([]*pb.Resource, error) {
	ctx, span := tracer.Start(ctx, "BatchCreateResources")
	defer span.End()
	span.SetAttributes(
		attribute.Int("num_resources", len(resources)),
	)
	var toCreateResources []Resource
	for _, res := range resources {
		if res.Uid == "" {
			res.Uid = common.GetUUID(maxUUIDRetries)
		}
		resPb, err := proto.Marshal(res)
		if err != nil {
			log.Warnf("proto marshal %q: %v", res.Uid, err)
			continue
		}
		t, err := utils.TypeFromResource(res)
		if err != nil {
			log.Warnf("type of %q: %v", res.Uid, err)
		}
		var recordTime *time.Time
		ts := res.Timestamp
		if ts != nil {
			pbTime := ts.AsTime()
			recordTime = &pbTime
		}
		labelsBytes, err := json.Marshal(res.Labels)
		if err != nil {
			log.Errorf("failed to marshal labels: %v", err)
			return nil, err
		}
		tagsBytes, err := json.Marshal(res.Tags)
		if err != nil {
			log.Errorf("failed to marshal tags: %v", err)
			return nil, err
		}
		toCreateResources = append(toCreateResources, Resource{
			ID:                res.Uid,
			Name:              res.Name,
			DisplayName:       res.DisplayName,
			ResourceGroupName: res.ResourceGroupName,
			CollectionID:      res.CollectionUid,
			RecordTime:        recordTime,
			ParentName:        res.Parent,
			Type:              t,
			Proto:             resPb,
			Labels:            labelsBytes,
			Tags:              tagsBytes,
		})
	}
	if len(toCreateResources) > 0 {
		if err := svc.db.WithContext(ctx).CreateInBatches(&toCreateResources, createBatchSize).Error; err != nil {
			return nil, fmt.Errorf("insert: %w", err)
		}
	}
	return resources, nil
}

// ListResources retrieves resources based on the provided filter.
func (svc *Service) ListResources(ctx context.Context, filter model.StorageFilter) (resources []*pb.Resource, err error) {
	var sqlResourceRows []Resource
	if filter.Limit == 0 {
		filter.Limit = listResourcesLimit
	}
	whereFilter, params := getFilter(filter, collectionOp)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err =
			svc.db.WithContext(ctx).
				Model(&Resource{}).
				Where(whereFilter, params...).
				Limit(filter.Limit).
				Order("resourcename ASC").
				Find(&sqlResourceRows).Error
		if err == nil {
			break
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			log.Warnf("retrying with exponential backoff (%d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Duration(math.Pow(2.0, float64(i))) * time.Second) //nolint:mnd
			continue
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	for _, row := range sqlResourceRows {
		res, err := row.ToResourceProto()
		if err != nil {
			log.Warnf("unmarshal: %v", err)
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func getFilter(filter model.StorageFilter, operationName opType) (string, []any) {
	var allFilters []string
	var params []any

	if filter.OperationID == "" {
		// When we don't have the operation ID, we have to get the latest one for the resource group
		// This is an expensive operation!
		addFilter := ""
		if len(filter.ResourceGroupNames) > 0 {
			addFilter = "AND resourcegroupname IN ?\n"
		}
		operationFilter := fmt.Sprintf(`(%s,resourcegroupname) IN (
			SELECT DISTINCT operationid,resourcegroupname FROM operations WHERE (resourcegroupname, endtime) IN (
					SELECT resourcegroupname, timestamp FROM (
						SELECT resourcegroupname, max(endtime) as timestamp 
						FROM operations AS ts 
						WHERE 
						opstype = '%s' AND 
						status = 'COMPLETED' %s
						GROUP BY resourcegroupname
					) 
					AS ts1
				)
			)`,
			operationName+"id",
			operationName,
			addFilter,
		)
		allFilters = append(allFilters, operationFilter)
		if addFilter != "" {
			params = append(params, filter.ResourceGroupNames)
		}
	}

	paramFilters, sqlFilterParams := sqlFilterFromModel(filter, operationName)
	params = append(params, sqlFilterParams...)
	allFilters = append(allFilters, paramFilters...)
	finalFilter := strings.Join(allFilters, " AND ")
	return finalFilter, params
}

// BatchCreateObservations creates a batch of observations.
func (svc *Service) BatchCreateObservations(ctx context.Context, observations []*pb.Observation) ([]*pb.Observation, error) {
	ctx, span := tracer.Start(ctx, "BatchCreateObservations")
	defer span.End()
	span.SetAttributes(
		attribute.Int("num_observations", len(observations)),
	)
	var dbObs []Observation
	for _, obs := range observations {
		myObs := obs
		obsPb, err := proto.Marshal(myObs)
		if err != nil {
			log.Warnf("proto marshal %q: %v", myObs.Uid, err)
			continue
		}
		rsrcRef := myObs.ResourceRef
		var groupName, cloudPlatform string
		var externalID, resourceID *string
		if rsrcRef != nil {
			groupName = rsrcRef.GroupName
			resourceID = rsrcRef.Uid
			cloudPlatform = rsrcRef.CloudPlatform.String()
			externalID = rsrcRef.ExternalId
		}

		dbObs = append(dbObs, Observation{
			ID:                    myObs.Uid,
			Name:                  myObs.Name,
			Resource:              nil,
			ResourceID:            resourceID,
			ResourceGroupName:     groupName,
			ResourceCloudPlatform: cloudPlatform,
			ResourceExternalID:    externalID,
			ScanID:                myObs.ScanUid,
			CollectionID:          myObs.CollectionId,
			RecordTime:            myObs.Timestamp.AsTime(),
			Proto:                 obsPb,
			ExternalID:            myObs.ExternalId,
			Source:                ObservationSource(myObs.Source),
			Category:              ObservationCategory(myObs.Category),
			SeverityScore:         FromSeverityPb(myObs.Severity),
			Impact:                Impact(myObs.Impact),
			RiskScore:             FromSeverityPb(myObs.Severity),
		})

	}
	if len(dbObs) > 0 {
		if err := svc.db.WithContext(ctx).CreateInBatches(&dbObs, createBatchSize).Error; err != nil {
			return nil, fmt.Errorf("insert: %w", err)
		}
	}
	return observations, nil
}

func validObservation(db *gorm.DB) *gorm.DB {
	return db.Where("severity_score >= 0")
}

// ListObservations retrieves observations based on the provided filter.
func (svc *Service) ListObservations(ctx context.Context, filter model.StorageFilter) (observations []*pb.Observation, err error) {
	ctx, span := tracer.Start(ctx, "ListObservations")
	defer span.End()
	var sqlObservationRows []Observation
	if filter.Limit > listObservationsLimit {
		return nil, fmt.Errorf("limit too high: %d", filter.Limit)
	}
	if filter.Limit == 0 {
		filter.Limit = listObservationsLimit
	}
	// Observations generated from a scan (Modron)
	scanFilter, scanParams := getFilter(filter, scanOp)
	// Observations generated from a collection (SCC, external integrations)
	collectFilter, collectParams := getFilter(filter, collectionOp)
	err =
		svc.db.WithContext(ctx).
			Model(&Observation{}).
			Scopes(validObservation).
			Where(
				svc.db.Where(scanFilter, scanParams...).Or(collectFilter, collectParams...),
			).
			Order("risk_score DESC, source ASC").
			Limit(filter.Limit).
			Find(&sqlObservationRows).
			Error
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	span.SetAttributes(
		attribute.Int(constants.TraceKeyNumObservations, len(sqlObservationRows)),
	)
	for _, row := range sqlObservationRows {
		obs, err := row.ToObservationProto()
		if err != nil {
			log.Warnf("unmarshal: %v", err)
		}
		observations = append(observations, obs)
	}
	return observations, nil
}

func fromPbOperation(o *pb.Operation) Operation {
	var startTime time.Time
	var endTime *time.Time

	switch o.Status {
	case pb.Operation_STARTED:
		startTime = o.StatusTime.AsTime()
	case pb.Operation_COMPLETED, pb.Operation_FAILED:
		t1 := o.StatusTime.AsTime()
		endTime = &t1
	default:
		log.Warnf("unknown status: %v", o.Status)
	}

	return Operation{
		ID:            o.Id,
		ResourceGroup: o.ResourceGroup,
		OpsType:       o.Type,
		Status:        OperationStatus(o.Status),
		StartTime:     startTime,
		EndTime:       endTime,
		Reason:        o.Reason,
	}
}

// AddOperationLog adds a new operation log entry.
func (svc *Service) AddOperationLog(ctx context.Context, operations []*pb.Operation) error {
	var errArr []error
	var addOps []Operation
	for _, o := range operations {
		if o.ResourceGroup == "" {
			log.WithField("operation_id", o.Id).Warn("missing resource group for operation")
		}
		if o.Status != pb.Operation_STARTED {
			// We first add new ops
			continue
		}
		addOps = append(addOps, fromPbOperation(o))
	}
	if len(addOps) > 0 {
		if err := svc.db.WithContext(ctx).Create(&addOps).Error; err != nil {
			log.WithError(err).Error("failed to insert operations")
			errArr = append(errArr, err)
		}
	}

	for _, o := range operations {
		if o.Status == pb.Operation_STARTED {
			// Insert ops were already added
			continue
		}
		opLogger := log.WithFields(logrus.Fields{
			"operation_id": o.Id,
			"status":       o.Status,
		})
		var foundOp Operation
		err := svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			tx.Find(&foundOp,
				"operationid = ? AND opstype = ? AND resourcegroupname = ?",
				o.Id, o.Type, o.ResourceGroup,
			)
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return fmt.Errorf("operation %q not found", o.Id)
			} else if tx.Error != nil {
				return fmt.Errorf("unable to find operation: %w", tx.Error)
			}

			foundOp.Status = OperationStatus(o.Status)
			if o.StatusTime != nil {
				t := o.StatusTime.AsTime()
				foundOp.EndTime = &t
			}
			foundOp.Reason = o.Reason

			if err := tx.Save(&foundOp).Error; err != nil {
				return fmt.Errorf("failed to update operation: %w", err)
			}
			return nil
		})
		if err != nil {
			opLogger.WithError(err).Errorf("failed to perform operation")
			errArr = append(errArr, err)
			continue
		}
	}
	return errors.Join(errArr...)
}

// PurgeIncompleteOperations purges incomplete operations from the database.
func (svc *Service) PurgeIncompleteOperations(ctx context.Context) error {
	return svc.db.WithContext(ctx).
		Model(&Operation{}).
		Where("endtime IS NULL").
		Update("endtime", time.Now()).
		Update("status", pb.Operation_FAILED).Error
}

// Deprecated FlushOpsLog flushes the operation log cache to the database.
func (svc *Service) FlushOpsLog(_ context.Context) error {
	// NO-OP
	return nil
}

func sqlFilterFromModel(filter model.StorageFilter, operationName opType) (filters []string, params []any) {
	if len(filter.ResourceNames) > 0 {
		filters = append(filters, "resourceName IN ?")
		params = append(params, filter.ResourceNames)
	}
	if len(filter.ResourceTypes) > 0 {
		filters = append(filters, "resourceType IN ?")
		params = append(params, filter.ResourceTypes)
	}
	if len(filter.ResourceGroupNames) > 0 {
		filters = append(filters, "resourceGroupName IN ?")
		params = append(params, filter.ResourceGroupNames)
	}
	if len(filter.ResourceIDs) > 0 {
		filters = append(filters, "resourceID IN ?")
		params = append(params, filter.ResourceIDs)
	}
	if len(filter.ParentNames) > 0 {
		filters = append(filters, "parentName IN ?")
		params = append(params, filter.ParentNames)
	}
	if filter.OperationID != "" {
		switch operationName {
		case collectionOp:
			filters = append(filters, "collectionID = ?")
			params = append(params, filter.OperationID)
		case scanOp:
			filters = append(filters, "scanID = ?")
			params = append(params, filter.OperationID)
		default:
			log.Warnf("unknown operation type: %v", operationName)
		}
	}
	if !filter.StartTime.IsZero() {
		filters = append(filters, "recordTime >= ?")
		params = append(params, filter.StartTime)
	}
	return
}

type Service struct {
	db  *gorm.DB
	cfg Config
}

// GetChildrenOfResource gets the _children_ of a given resource ID
// when parentResourceName is empty, the whole tree (from the root) is returned
func (svc *Service) GetChildrenOfResource(
	ctx context.Context,
	collectID string,
	parentResourceName string,
	resourceType *string,
) (map[string]*pb.RecursiveResource, error) {
	var resources []Resource
	tx := svc.db.WithContext(ctx)
	if resourceType != nil {
		tx = tx.Where("resourcetype = ?", *resourceType)
	}
	if parentResourceName == "" {
		tx = tx.Find(&resources, "collectionid = ?", collectID)
	} else {
		// We use a recursive CTE only in case we have a parentResourceName, otherwise the call is too expensive
		tx = tx.
			Raw(`WITH RECURSIVE resource_hierarchy(id) AS (
			VALUES(?)
			UNION ALL
			SELECT resourcename FROM
			resources, resource_hierarchy
			WHERE
			resources.parentname = resource_hierarchy.id AND
			resources.collectionid = ?
		)
		SELECT resourcename,display_name,parentname,resourcetype,labels 
		FROM resources 
		WHERE resourcename IN (SELECT * FROM resource_hierarchy) AND 
		collectionid = ?`,
				parentResourceName, collectID, collectID)
		tx = tx.Scan(&resources)
	}

	if tx.Error != nil {
		return nil, tx.Error
	}

	var pbResources []*pb.Resource
	for _, r := range resources {
		pbRes, err := r.ToResourceProto()
		if err != nil {
			return nil, err
		}
		pbResources = append(pbResources, pbRes)
	}
	return utils.ComputeRgHierarchy(pbResources)
}

type byTypeAndName []*pb.RecursiveResource

func (b byTypeAndName) Len() int {
	return len(b)
}

func (b byTypeAndName) Less(i, j int) bool {
	// We're lucky that folders/ comes before projects/ (in the alphabet), so it's
	// enough to just sort them by their name. Folders will always come before projects.
	return b[i].Name < b[j].Name
}

func (b byTypeAndName) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

var _ sort.Interface = byTypeAndName{}

type Config struct {
	BatchSize     int32
	LogAllQueries bool
}

func waitForSQLDatabase(db *gorm.DB) (err error) {
	var lastError error
	timeOut := sqlDBConnectionTimeout
	sqlDb, err := db.DB()
	if err != nil {
		return err
	}
	for timeOut < dbMaxBootUpWaitTime {
		if err = sqlDb.Ping(); err == nil {
			return nil
		}
		time.Sleep(timeOut)
		timeOut = sqlDBConnectionTimeoutFactor * timeOut
	}
	return lastError
}

func New(db *gorm.DB, cfg Config) (model.Storage, error) {
	if err := db.Use(gormtracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("setup tracing for gorm: %w", err)
	}
	if cfg.BatchSize <= 0 {
		return nil, fmt.Errorf("batch size must be greater than 0")
	}
	if err := waitForSQLDatabase(db); err != nil {
		return nil, err
	}
	sqlDb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}
	err = sqlDb.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	// Automigrate
	for _, v := range []any{
		&Resource{},
		&Observation{},
		&Operation{},
	} {
		if err := db.AutoMigrate(v); err != nil {
			return nil, fmt.Errorf("auto migrate: %w", err)
		}
	}

	if cfg.LogAllQueries {
		db.Config.Logger = gormlogger.Default.LogMode(gormlogger.Info)
	}

	return &Service{
		db:  db,
		cfg: cfg,
	}, nil
}

func NewSQLite(cfg Config, dbPath string) (model.Storage, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}
	phyDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB: %w", err)
	}
	phyDB.SetMaxOpenConns(1) // https://github.com/mattn/go-sqlite3/issues/274#issuecomment-191597862
	return New(db, cfg)
}

// NewPostgres creates a new SQL storage service using PostgreSQL.
func NewPostgres(cfg Config, dsn string) (model.Storage, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	return New(db, cfg)
}

func NewDB(driver string, connectionString string, config Config) (model.Storage, error) {
	switch driver {
	case "sqlite3":
		return NewSQLite(config, connectionString)
	case "postgres":
		return NewPostgres(config, connectionString)
	}
	return nil, fmt.Errorf("unsupported driver: %s", driver)
}
