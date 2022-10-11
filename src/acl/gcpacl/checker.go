package gcpacl

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nianticlabs/modron/src/model"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/metadata"
)

const (
	AclUpdateIntervalSecEnvVar = "ACLINTERVAL"
)

var (
	aclUpdateIntervalSec = func() int {
		intVar, err := strconv.Atoi(os.Getenv(AclUpdateIntervalSecEnvVar))
		if err != nil {
			return 5 * 60
		}
		return intVar
	}()
)

type Config struct {
	CacheTimeout time.Duration
	AdminGroups  []string
}

type GcpChecker struct {
	cfg              Config
	collector        model.Collector
	cloudIdentitySvc *cloudidentity.Service

	aclCache       map[string]map[string]struct{}
	adminsCache    map[string]ACLCacheEntry
	adminGroupsIds []string
}

func (checker *GcpChecker) GetAcl() map[string]map[string]struct{} {
	return checker.aclCache
}

func New(ctx context.Context, collector model.Collector, cfg Config) (model.Checker, error) {
	cisvc, err := cloudidentity.NewService(ctx)
	if err != nil {
		return nil, err
	}
	gcpChecker := GcpChecker{
		collector:        collector,
		cfg:              cfg,
		cloudIdentitySvc: cisvc,
		aclCache:         make(map[string]map[string]struct{}),
		adminsCache:      make(map[string]ACLCacheEntry),
		adminGroupsIds:   []string{},
	}
	for _, ag := range cfg.AdminGroups {
		group, err := cisvc.Groups.Lookup().GroupKeyId(ag).Do()
		if err != nil {
			glog.Errorf("cannot fetch group %q: %v", ag, err)
			continue
		}
		gcpChecker.adminGroupsIds = append(gcpChecker.adminGroupsIds, group.Name)
	}
	if err := gcpChecker.loadAclCache(ctx); err != nil {
		return nil, err
	}
	gcpChecker.updateACLs(ctx)
	return &gcpChecker, nil
}

func (checker *GcpChecker) updateACLs(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(aclUpdateIntervalSec) * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := checker.loadAclCache(ctx); err != nil {
					glog.Warningf("cannot update ACLs: %v", err)
				}
			}
		}
	}()
}

func (checker *GcpChecker) ListResourceGroupNamesOwned(ctx context.Context) (map[string]struct{}, error) {
	user, err := checker.GetValidatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if checker.isAdmin(user) {
		return checker.aclCache["*"], nil
	}
	if _, ok := checker.aclCache[user]; !ok {
		return map[string]struct{}{}, nil
	}
	return checker.aclCache[user], nil
}

func (checker *GcpChecker) loadAclCache(ctx context.Context) error {
	res, err := checker.collector.ListResourceGroupAdmins(ctx)
	if err != nil {
		return err
	}
	checker.aclCache = res
	return nil
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

func (c *GcpChecker) isAdmin(user string) bool {
	// We do some memoizing here as this might be called a lot over a short period of time.
	if v, ok := c.adminsCache[user]; ok {
		if v.time.Add(c.cfg.CacheTimeout).After(time.Now()) {
			return v.access
		}
	}
	for _, g := range c.adminGroupsIds {
		resp, err := c.cloudIdentitySvc.Groups.Memberships.CheckTransitiveMembership(g).Query(fmt.Sprintf("member_key_id == '%s'", user)).Do()
		if err != nil {
			glog.Warningf("cannot check membership: %v", err)
		} else {
			if resp.HasMembership {
				c.adminsCache[user] = ACLCacheEntry{time: time.Now(), access: true}
				return true
			}
		}
	}
	c.adminsCache[user] = ACLCacheEntry{time: time.Now(), access: false}
	return false
}

type ACLCacheEntry struct {
	time   time.Time
	access bool
}
