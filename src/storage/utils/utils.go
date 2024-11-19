package storageutils

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

func GetSqliteMemoryDbPath() string {
	debugDbPath := os.Getenv("DEBUG_DB_PATH")
	if debugDbPath != "" {
		return debugDbPath
	}
	// We use an uniqueID, so that two tests running in parallel do not conflict with each other
	uniqueID := uuid.NewString()
	// Do not use `:memory:` here! https://github.com/mattn/go-sqlite3/issues/204
	return fmt.Sprintf("file:%s?mode=memory&cache=shared", uniqueID)
}
