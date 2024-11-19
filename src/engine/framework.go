package engine

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

type TransportProvider func(ctx context.Context, cluster *container.Cluster) (http.RoundTripper, error)

type Storage struct {
	model.Storage
}

var (
	cleanupInterval = 30 * time.Minute
	cleanupTicker   = time.NewTicker(cleanupInterval)
)

type cachedResource struct {
	resource  *pb.Resource
	timestamp time.Time
}

func (e *RuleEngine) GetChildren(ctx context.Context, parent string) ([]*pb.Resource, error) {
	filter := model.StorageFilter{
		ParentNames: []string{parent},
	}
	return e.storage.ListResources(ctx, filter)
}

func (e *RuleEngine) GetResource(ctx context.Context, resourceName string) (*pb.Resource, error) {
	if cache, exists := e.memoizationMap.Load(resourceName); exists {
		res := cache.(*cachedResource)
		return res.resource, nil
	}

	filter := model.StorageFilter{
		Limit:         1,
		ResourceNames: []string{resourceName},
	}
	collectionID, ok := ctx.Value(constants.CollectIDKey).(string)
	if ok {
		filter.OperationID = collectionID
	} else {
		log.Warnf("collection ID not found in context")
	}
	res, err := e.storage.ListResources(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("resource %q could not be fetched: %w", resourceName, err)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("resource %q does not exist", resourceName)
	}

	cachedRes := &cachedResource{
		resource:  res[0],
		timestamp: time.Now(),
	}
	e.memoizationMap.Store(resourceName, cachedRes)
	return res[0], nil
}

func (e *RuleEngine) startCacheCleanup() {
	for range cleanupTicker.C {
		e.clearExpiredResources()
	}
}

func (e *RuleEngine) clearExpiredResources() {
	e.memoizationMap.Range(func(key, value interface{}) bool {
		cachedRes := value.(*cachedResource)
		if time.Since(cachedRes.timestamp) >= cleanupInterval {
			e.memoizationMap.Delete(key)
		}
		return true
	})
}

func GetKubernetesClient(ctx context.Context, clusterName string, httpClient *http.Client, getTransport TransportProvider) (*kubernetes.Clientset, error) {
	tokenSource, err := google.DefaultTokenSource(ctx, compute.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get a token source: %w", err)
	}
	if httpClient == http.DefaultClient {
		httpClient = oauth2.NewClient(ctx, tokenSource)
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	}
	containerService, err := container.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("could not create client for Google Container Engine: %w", err)
	}

	cluster, err := containerService.Projects.Locations.Clusters.Get(clusterName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("cluster %q: %w", clusterName, err)
	}
	// This is a very ugly dependency injection, but we have to do it otherwise the unit test would require a complete oauth2 backend.
	tr, err := getTransport(ctx, cluster)
	if err != nil {
		return nil, err
	}
	kubeHTTPClient := httpClient
	kubeHTTPClient.Transport = tr
	kubeClient, err := kubernetes.NewForConfigAndClient(
		&rest.Config{
			Host: "https://" + cluster.Endpoint,
		},
		kubeHTTPClient,
	)
	if err != nil {
		return nil, fmt.Errorf("kubernetes HTTP client could not be created: %w", err)
	}
	return kubeClient, nil
}

func GetOauthTransport(ctx context.Context, cluster *container.Cluster) (http.RoundTripper, error) {
	tokenSource, err := google.DefaultTokenSource(ctx, compute.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get a token source: %w", err)
	}
	// Connect to Kubernetes using OAuth authentication, trusting its CA.
	caPool := x509.NewCertPool()
	caCertPEM, err := base64.StdEncoding.DecodeString(cluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 in ClusterCaCertificate: %w", err)
	}
	caPool.AppendCertsFromPEM(caCertPEM)
	return &oauth2.Transport{
		Source: tokenSource,
		Base:   http.DefaultTransport,
	}, nil
}
