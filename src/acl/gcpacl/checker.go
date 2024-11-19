package gcpacl

import (
	"fmt"
	"strings"
	"time"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/metadata"
)

const (
	defaultACLUpdateInterval = 5 * time.Minute
)

var (
	log = logrus.StandardLogger().WithField(constants.LogKeyPkg, "gcpacl")
)

type Config struct {
	AdminGroups  []string
	CacheTimeout time.Duration
	// PersistentCache is a flag to enable/disable the local ACL cache.
	PersistentCache bool
	// PersistentCacheTimeout is the amount of time we keep the ACLs on the filesystem before we fetch them again.
	PersistentCacheTimeout time.Duration
	// SkipIap enables or disables the IAP check - when this is enabled, all users are considered admins. Of course this should be always disabled in prod.
	SkipIap bool
}

type GcpChecker struct {
	cfg              Config
	collector        model.Collector
	cloudIdentitySvc *cloudidentity.Service

	aclCache       model.ACLCache
	adminsCache    map[string]ACLCacheEntry
	adminGroupsIDs []string
}

func (checker *GcpChecker) GetACL() model.ACLCache {
	return checker.aclCache
}

func New(ctx context.Context, collector model.Collector, cfg Config) (model.Checker, error) {
	cloudIdentitySvc, err := cloudidentity.NewService(ctx)
	if err != nil {
		return nil, err
	}
	gcpChecker := GcpChecker{
		collector:        collector,
		cfg:              cfg,
		cloudIdentitySvc: cloudIdentitySvc,
		aclCache:         make(model.ACLCache),
		adminsCache:      make(map[string]ACLCacheEntry),
		adminGroupsIDs:   []string{},
	}
	for _, ag := range cfg.AdminGroups {
		groupEmail := strings.TrimPrefix(ag, "group:")
		group, err := cloudIdentitySvc.Groups.Lookup().GroupKeyId(groupEmail).Do()
		if err != nil {
			log.Errorf("cannot fetch group %q: %v", ag, err)
			continue
		}
		gcpChecker.adminGroupsIDs = append(gcpChecker.adminGroupsIDs, group.Name)
	}
	if cfg.PersistentCache {
		log.Debugf("using on-disk ACL cache")
		if err := gcpChecker.loadACLCache(ctx); err != nil {
			return nil, err
		}
	} else {
		log.Debugf("using in-memory ACL cache")
		var res model.ACLCache
		res, err = gcpChecker.getACLAndStore(ctx)
		if err != nil {
			return nil, err
		}
		gcpChecker.aclCache = res
	}
	gcpChecker.updateACLs(ctx)
	return &gcpChecker, nil
}

func (checker *GcpChecker) updateACLs(ctx context.Context) {
	ticker := time.NewTicker(defaultACLUpdateInterval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := checker.loadACLCache(ctx); err != nil {
					log.Warnf("cannot update ACLs: %v", err)
				}
			}
		}
	}()
}

func (checker *GcpChecker) ListResourceGroupNamesOwned(ctx context.Context) (map[string]struct{}, error) {
	adminResources := checker.aclCache["*"]
	if checker.cfg.SkipIap {
		// When we decide to skip the IAP check (insecure, only for local development), we return all resources as admin.
		log.Warnf("IAP check is disabled, users are all admins. If you see this in production, reach out to the security team.")
		return adminResources, nil
	}
	user, err := checker.GetValidatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if checker.isAdmin(user) {
		return adminResources, nil
	}
	if _, ok := checker.aclCache[user]; !ok {
		return map[string]struct{}{}, nil
	}
	return checker.aclCache[user], nil
}

func (checker *GcpChecker) loadACLCache(ctx context.Context) error {
	aclFsCache, err := checker.getLocalACLCache()
	if err != nil {
		return err
	}
	if aclFsCache != nil {
		if time.Since(aclFsCache.LastUpdate) < checker.cfg.PersistentCacheTimeout {
			log.Tracef("ACL cache hit")
			checker.aclCache = aclFsCache.Content
			return nil
		}
		if err := checker.deleteLocalACLCache(); err != nil {
			return fmt.Errorf("ACL cache: %w", err)
		}
	}
	log.Tracef("ACL cache miss")
	res, err := checker.collector.ListResourceGroupAdmins(ctx)
	if err != nil {
		return err
	}
	checker.aclCache = res
	if err := checker.saveLocalACLCache(res); err != nil {
		return fmt.Errorf("save ACL cache: %w", err)
	}
	return nil
}

func (checker *GcpChecker) getACLAndStore(ctx context.Context) (model.ACLCache, error) {
	res, err := checker.collector.ListResourceGroupAdmins(ctx)
	if err != nil {
		return nil, err
	}
	checker.aclCache = res
	return res, nil
}

func (checker *GcpChecker) GetValidatedUser(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("could not parse context metadata")
	}
	values := md.Get("x-goog-iap-jwt-assertion")
	if len(values) == 0 {
		return "", fmt.Errorf("no IAP JWT in context metadata")
	}
	// We skip audience validation because we are allowing both internal and load balancer ingress.
	payload, err := idtoken.Validate(ctx, values[0], "")
	if err != nil {
		return "", err
	}
	user, ok := payload.Claims["email"].(string)
	if !ok {
		return "", fmt.Errorf("authorization payload does not contain email")
	}
	return user, nil
}

func (checker *GcpChecker) isAdmin(user string) bool {
	// We do some memoizing here as this might be called a lot over a short period of time.
	if v, ok := checker.adminsCache[user]; ok {
		if v.time.Add(checker.cfg.CacheTimeout).After(time.Now()) {
			return v.access
		}
	}
	for _, g := range checker.adminGroupsIDs {
		resp, err := checker.cloudIdentitySvc.Groups.Memberships.CheckTransitiveMembership(g).Query(fmt.Sprintf("member_key_id == '%s'", user)).Do()
		if err != nil {
			log.Warnf("cannot check membership: %v", err)
			continue
		}
		if resp.HasMembership {
			checker.adminsCache[user] = ACLCacheEntry{time: time.Now(), access: true}
			return true
		}
	}
	checker.adminsCache[user] = ACLCacheEntry{time: time.Now(), access: false}
	return false
}

type ACLCacheEntry struct {
	time   time.Time
	access bool
}
