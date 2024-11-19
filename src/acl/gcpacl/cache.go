package gcpacl

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/nianticlabs/modron/src/model"
)

type FSACLCache struct {
	LastUpdate time.Time      `json:"last_update"`
	Content    model.ACLCache `json:"content"`
}

var localACLCacheFile = os.TempDir() + "/modron-acl-cache.json"

const ownerRWPermissions = 0600

func (checker *GcpChecker) getLocalACLCache() (*FSACLCache, error) {
	log.Tracef("getting ACL cache from %s", localACLCacheFile)
	f, err := os.Open(localACLCacheFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil //nolint:nilnil
		}
		return nil, err
	}
	defer f.Close()

	var fsACLCache FSACLCache
	if err := json.NewDecoder(f).Decode(&fsACLCache); err != nil {
		return nil, err
	}
	return &fsACLCache, nil
}

func (checker *GcpChecker) saveLocalACLCache(res model.ACLCache) error {
	log.Tracef("saving ACL cache to %s", localACLCacheFile)
	fsACLCache := FSACLCache{
		LastUpdate: time.Now(),
		Content:    res,
	}
	f, err := os.OpenFile(localACLCacheFile, os.O_CREATE|os.O_WRONLY, ownerRWPermissions)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(fsACLCache)
}

func (checker *GcpChecker) deleteLocalACLCache() error {
	return os.Remove(localACLCacheFile)
}
