//go:build integration

package gormstorage

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/storage/test"
)

func getPostgresDB(ctx context.Context, t *testing.T) (*postgres.PostgresContainer, error) {
	t.Helper()
	return postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("modron"),
		testcontainers.WithLogger(testcontainers.TestLogger(t)),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(6*time.Second),
				wait.ForListeningPort("5432"),
			),
		),
		testcontainers.WithHostPortAccess(5432),
	)
}

func newPostgresTestDb(ctx context.Context, t *testing.T) model.Storage {
	t.Helper()
	pgDb, err := getPostgresDB(ctx, t)
	if err != nil {
		t.Fatalf("unable to create postgres container: %v", err)
	}
	connStr, err := pgDb.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("unable to get connection string: %v", err)
	}
	st, err := NewPostgres(Config{
		BatchSize: 10,
	}, connStr)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	return st
}

func TestPostgresStorageResource(t *testing.T) {
	test.StorageResource(t, newPostgresTestDb(context.Background(), t))
}

func TestPostgresStorageObservation(t *testing.T) {
	test.StorageObservation(t, newPostgresTestDb(context.Background(), t))
}

func TestPostgresStorageListObservationsActive(t *testing.T) {
	test.StorageListObservations2(t, newPostgresTestDb(context.Background(), t))
}
