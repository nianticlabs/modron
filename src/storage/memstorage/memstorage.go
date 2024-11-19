package memstorage

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/storage/gormstorage"
	storageutils "github.com/nianticlabs/modron/src/storage/utils"
)

const DefaultBatchSize = 100

var logger = logrus.StandardLogger()

func New() model.Storage {
	dbPath := storageutils.GetSqliteMemoryDbPath()
	logger.Debugf("Using SQLite storage with path: %s", dbPath)
	st, err := gormstorage.NewSQLite(gormstorage.Config{
		BatchSize:     DefaultBatchSize,
		LogAllQueries: os.Getenv("LOG_ALL_SQL_QUERIES") == "true",
	}, dbPath)
	if err != nil {
		// It's fine to panic here, memstorage should only be used in tests
		panic(err)
	}
	return st
}
