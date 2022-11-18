// Package BigQueryStorage provides a storage backend that runs locally in memory. It is supposed to be used primarily for API testing.
package bigquerystorage

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
	"github.com/nianticlabs/modron/src/storage/memstorage"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

type BigQueryStorage struct {
	datasetID string
	projectID string

	observationTableID string
	operationTableID   string
	resourceTableID    string

	client     *bigquery.Client
	cache      model.Storage
	opsEntries []*BigQueryOperationEntry
}

func New(
	ctx context.Context,
	projectID string,
	datasetID string,
	resourceTableID string,
	observationTableID string,
	operationTableID string,
) (model.Storage, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	storage := &BigQueryStorage{
		projectID:          projectID,
		datasetID:          datasetID,
		resourceTableID:    resourceTableID,
		observationTableID: observationTableID,
		operationTableID:   operationTableID,
		client:             client,
		cache:              memstorage.New(),
	}
	if err := storage.fetchObservations(ctx); err != nil {
		return nil, fmt.Errorf("fetching Observation: %v", err)
	}
	if err := storage.fetchResources(ctx); err != nil {
		return nil, fmt.Errorf("fetching Resources: %v", err)
	}
	if err := storage.fetchOperations(ctx); err != nil {
		return nil, fmt.Errorf("fetching Operations: %v", err)
	}
	return storage, nil
}

// Add a set of Resources into the database
func (bq *BigQueryStorage) BatchCreateResources(ctx context.Context, resources []*pb.Resource) ([]*pb.Resource, error) {
	if _, err := bq.cache.BatchCreateResources(ctx, resources); err != nil {
		return nil, fmt.Errorf("could not write to cache: %v", err)
	}
	ins := bq.client.Dataset(bq.datasetID).Table(bq.resourceTableID).Inserter()
	entries := []BigQueryResourceEntry{}
	for _, res := range resources {
		if entry, err := toBigQueryResourceEntry(res); err != nil {
			return nil, err
		} else {
			entries = append(entries, *entry)
		}
	}
	// TODO: handle (individual) row insertion errors
	if err := ins.Put(ctx, entries); err != nil {
		return nil, err
	}
	return resources, nil
}

// Add a set of Observations into the database
func (bq *BigQueryStorage) BatchCreateObservations(ctx context.Context, observations []*pb.Observation) ([]*pb.Observation, error) {
	if _, err := bq.cache.BatchCreateObservations(ctx, observations); err != nil {
		return nil, fmt.Errorf("could not write to cache: %v", err)
	}
	ins := bq.client.Dataset(bq.datasetID).Table(bq.observationTableID).Inserter()
	entries := []BigQueryObservationEntry{}
	for _, res := range observations {
		if entry, err := toBigQueryObservationEntry(res); err != nil {
			return nil, err
		} else {
			entries = append(entries, *entry)
		}
	}
	// TODO: handle (individual) row insertion errors
	if err := ins.Put(ctx, entries); err != nil {
		return nil, err
	}
	return observations, nil
}

func (bq *BigQueryStorage) ListObservations(ctx context.Context, filter model.StorageFilter) ([]*pb.Observation, error) {
	return bq.cache.ListObservations(ctx, filter)
}

func (bq *BigQueryStorage) ListResources(ctx context.Context, filter model.StorageFilter) ([]*pb.Resource, error) {
	return bq.cache.ListResources(ctx, filter)
}

func (bq *BigQueryStorage) AddOperationLog(ctx context.Context, ops []model.Operation) error {
	for _, o := range ops {
		if entry, err := toBigqueryOperationEntry(o); err != nil {
			return err
		} else {
			bq.opsEntries = append(bq.opsEntries, entry)
		}
	}
	if len(bq.opsEntries) >= 100 {
		if err := bq.FlushOpsLog(ctx); err != nil {
			return err
		}
	}
	// Add the operations to the cache as well.
	if err := bq.cache.AddOperationLog(ctx, ops); err != nil {
		return err
	}
	return nil
}

func (bq *BigQueryStorage) FlushOpsLog(ctx context.Context) error {
	ins := bq.client.Dataset(bq.datasetID).Table(bq.operationTableID).Inserter()
	if err := ins.Put(ctx, bq.opsEntries); err != nil {
		return fmt.Errorf("put: %v", err)
	}
	bq.opsEntries = []*BigQueryOperationEntry{}
	return nil
}

func (bq *BigQueryStorage) fetchObservations(ctx context.Context) error {
	observations := []*pb.Observation{}
	var query *bigquery.Query
	var queryParameters = []bigquery.QueryParameter{}
	queryString := fmt.Sprintf(`SELECT DISTINCT 
	    observations.observationProto
		FROM %s.%s AS observations RIGHT JOIN (
		SELECT DISTINCT uuid
		  FROM %s.%s operations1
		  WHERE statusTime = (SELECT MAX(statusTime) 
							FROM %s.%s as operations2 
							WHERE status = "COMPLETED" 
							  AND opsType = "scan" 
							  AND operations1.resourceGroupName = operations2.resourceGroupName)
	  ) AS operations ON observations.scanID = operations.uuid
	`, bq.datasetID, bq.observationTableID, bq.datasetID, bq.operationTableID, bq.datasetID, bq.operationTableID)
	queryString += `;`
	query = bq.client.Query(queryString)
	query.Parameters = queryParameters

	it, err := bq.runQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("ListResources (ctx) failed for \nQuery: %v \nError: %v", queryString, err)
	}
	for {
		umObservation := &pb.Observation{}
		var row BigQueryObservationEntry
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if err := proto.Unmarshal(row.ObservationProto, umObservation); err != nil {
			return err
		}
		observations = append(observations, umObservation)
	}
	_, err = bq.cache.BatchCreateObservations(ctx, observations)
	if err != nil {
		return fmt.Errorf("could not write to cache : %v", err)
	}
	return nil
}

func (bq *BigQueryStorage) fetchOperations(ctx context.Context) error {
	var query *bigquery.Query
	var queryParameters = []bigquery.QueryParameter{}
	queryString := fmt.Sprintf(`SELECT DISTINCT uuid, resourceGroupName, opsType, statusTime, status
		  FROM %s.%s operations1
		  WHERE statusTime = (SELECT MAX(statusTime) 
							FROM %s.%s as operations2 
							WHERE status = "COMPLETED" 
							  AND opsType = "scan" 
							  AND operations1.resourceGroupName = operations2.resourceGroupName) ORDER BY statusTime ASC`,
		bq.datasetID, bq.operationTableID, bq.datasetID, bq.operationTableID)
	queryString += `;`
	query = bq.client.Query(queryString)
	query.Parameters = queryParameters

	it, err := bq.runQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("fetchOperations (ctx) failed for \nQuery: %v \nError: %v", queryString, err)
	}
	operations := []model.Operation{}
	for {
		var row BigQueryOperationEntry
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		op, err := row.toModelOperation()
		if err != nil {
			glog.Warningf("invalid operation %v: %v", row, err)
			continue
		}
		operations = append(operations, op)
	}
	if err := bq.cache.AddOperationLog(ctx, operations); err != nil {
		return err
	}
	return nil
}

func (bq *BigQueryStorage) fetchResources(ctx context.Context) error {
	resources := []*pb.Resource{}
	var query *bigquery.Query
	var queryParameters = []bigquery.QueryParameter{}
	// TODO: add a global time window limit to ensure we don't sift through all data with every query
	queryString := fmt.Sprintf(`SELECT resourceProto FROM %s.%s AS res INNER JOIN (
								SELECT collectionID, resourceGroupName, ROW_NUMBER() OVER (
										PARTITION BY resourceGroupName
										ORDER BY recordTime
										DESC
									) AS rowNumber
								FROM %s.%s
							) AS collections
							ON (res.collectionID = collections.collectionID AND res.resourceGroupName = collections.resourceGroupName AND collections.rowNumber = 1)
							`, bq.datasetID, bq.resourceTableID, bq.datasetID, bq.resourceTableID)
	queryString += `ORDER BY recordTime DESC `
	/*
		if filter.Limit != nil {
			queryString += `LIMIT @limit `
			queryParameters = append(queryParameters, bigquery.QueryParameter{Name: "limit", Value: *filter.Limit})
		}
	*/
	queryString += `;`

	query = bq.client.Query(queryString)
	query.Parameters = queryParameters

	it, err := bq.runQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("ListResources (ctx) failed for \nQuery: %v \nError: %v", queryString, err)
	}
	for {
		umResource := &pb.Resource{}
		var row BigQueryResourceEntry
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if err := proto.Unmarshal(row.ResourceProto, umResource); err != nil {
			return err
		}
		resources = append(resources, umResource)
	}
	_, err = bq.cache.BatchCreateResources(ctx, resources)
	if err != nil {
		return fmt.Errorf("could not write to cache : %v", err)
	}
	return nil
}

func (bq *BigQueryStorage) runQuery(ctx context.Context, q *bigquery.Query) (*bigquery.RowIterator, error) {
	j, err := q.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("error: %v; for Query: %q, Parameters: %v", err, q.Q, q.Parameters)
	}
	s, err := j.Wait(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	it, err := j.Read(ctx)
	if err != nil {
		return nil, err
	}
	return it, nil
}

func BuildQueryFromFilter[T int | string](columnName string, baseParameterName string, expConnector string, filterValues []T) (string, []bigquery.QueryParameter, error) {
	q := ""
	qParams := []bigquery.QueryParameter{}

	// more than one element to filter present
	if len(filterValues) > 1 {
		q = "("
		expression := ``

		for i, f := range filterValues {
			qParam := fmt.Sprintf(`%s%d`, baseParameterName, i)
			q += fmt.Sprintf(`%s %s = @%s `, expression, columnName, qParam)
			qParams = append(qParams, bigquery.QueryParameter{Name: qParam, Value: f})
			expression = `OR`
		}
		q += ")"
	} else {
		q = fmt.Sprintf(`%s = @%s`, columnName, baseParameterName)
		qParams = append(qParams, bigquery.QueryParameter{Name: baseParameterName, Value: filterValues[0]})
	}
	q = fmt.Sprintf(`%s %s `, expConnector, q) // add expression connector
	return q, qParams, nil
}
