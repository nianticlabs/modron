package gcpcollector

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"regexp"
	"time"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/pb"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	externalMatch    = regexp.MustCompile("EXTERNAL")
	internalMatch    = regexp.MustCompile("INTERNAL")
	defaultSslPolicy = &pb.SslPolicy{ // Default is COMPATIBLE with min TLS1.0 https://cloud.google.com/load-balancing/docs/ssl-policies-concepts
		// TODO: Get this via the API
		MinTlsVersion: pb.SslPolicy_TLS_1_0,
		Profile:       pb.SslPolicy_COMPATIBLE,
		CreationDate:  timestamppb.New(time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)),
		Name:          "GcpDefaultSslPolicy",
	}
	policyProfileMap = map[string]pb.SslPolicy_Profile{
		"COMPATIBLE": pb.SslPolicy_COMPATIBLE,
		"MODERN":     pb.SslPolicy_MODERN,
		"RESTRICTED": pb.SslPolicy_RESTRICTED,
		"CUSTOM":     pb.SslPolicy_CUSTOM,
	}
	policyMinTlsVersionMap = map[string]pb.SslPolicy_MinTlsVersion{
		"TLS_1_0": pb.SslPolicy_TLS_1_0,
		"TLS_1_1": pb.SslPolicy_TLS_1_1,
		"TLS_1_2": pb.SslPolicy_TLS_1_2,
		"TLS_1_3": pb.SslPolicy_TLS_1_3,
	}
)

func certsFromPemChain(pem_chain string) (certs []*x509.Certificate, err error) {
	if len(pem_chain) == 0 {
		return nil, fmt.Errorf("certificate chain is empty")
	}

	next := []byte(pem_chain)
	for len(next) > 0 {
		block, rest := pem.Decode(next)
		if block == nil {
			err = fmt.Errorf("unable to decode PEM-encoded certificate chain")
			rest = nil
		} else {
			if cert, parseErr := x509.ParseCertificate(block.Bytes); parseErr != nil {
				err = fmt.Errorf("X509 parsing: %v", parseErr)
			} else {
				certs = append(certs, cert)
			}
		}
		next = rest
	}
	return
}

func getSslPolicyForService(
	service *compute.BackendService,
	proxies []*compute.TargetHttpsProxy,
	urlMaps []*compute.UrlMap,
	sslPolicies []*compute.SslPolicy) (*pb.SslPolicy, error) {

	getSslPolicyWithServiceMatched := func(proxy *compute.TargetHttpsProxy) (*compute.SslPolicy, error) {
		for _, sslPolicy := range sslPolicies {
			if sslPolicy.SelfLink == proxy.SslPolicy {
				return sslPolicy, nil
			}
		}
		return nil, fmt.Errorf("sslPolicy for proxy %s not found", proxy.Name)
	}

	handlePathMatchers := func(pathMatchers []*compute.PathMatcher, proxy *compute.TargetHttpsProxy) (*compute.SslPolicy, error) {
		var sslPolicy *compute.SslPolicy
		var err error
		for _, pathMatch := range pathMatchers {
			sslPolicy = nil
			if service.SelfLink == pathMatch.DefaultService {
				sslPolicy, err = getSslPolicyWithServiceMatched(proxy)
			} else {
				for _, pathRule := range pathMatch.PathRules {
					if pathRule.Service == service.SelfLink {
						sslPolicy, err = getSslPolicyWithServiceMatched(proxy)
						break
					}
				}
			}
			if sslPolicy != nil {
				return sslPolicy, err
			}
		}
		return nil, fmt.Errorf("sslPolicy for proxy %s not found", proxy.Name)
	}

	getPolicyForProxy := func(proxy *compute.TargetHttpsProxy) (*compute.SslPolicy, error) {
		for _, urlMap := range urlMaps {
			if proxy.UrlMap == urlMap.SelfLink && service.SelfLink == urlMap.DefaultService {
				return getSslPolicyWithServiceMatched(proxy)
			} else {
				if proxy.UrlMap == urlMap.SelfLink {
					sslPolicy, err := handlePathMatchers(urlMap.PathMatchers, proxy)
					if sslPolicy != nil {
						return sslPolicy, err
					}
				}
				for _, pathMatch := range urlMap.PathMatchers {
					if proxy.UrlMap == urlMap.SelfLink && service.SelfLink == pathMatch.DefaultService {
						sslPolicy, err := getSslPolicyWithServiceMatched(proxy)
						if sslPolicy != nil {
							return sslPolicy, err
						}
					}
				}
			}
		}
		return nil, fmt.Errorf("sslPolicy for proxy %s not found", proxy.Name)
	}

	usedPolicy := defaultSslPolicy
	for _, proxy := range proxies {
		if proxy.SslPolicy != "" {
			policy, err := getPolicyForProxy(proxy)
			if err != nil {
				// proxy uses the GCP Default Policy
				continue
			}
			timeStamp, err := time.Parse(time.RFC3339, policy.CreationTimestamp)
			if err != nil {
				glog.Errorf("SslPolicy %s: %s. %v", policy.Name, policy.CreationTimestamp, err)
				continue
			}
			usedPolicy = &pb.SslPolicy{
				CreationDate:    timestamppb.New(timeStamp),
				Name:            policy.Name,
				Profile:         policyProfileMap[policy.Profile],
				CustomFeatures:  policy.CustomFeatures,
				EnabledFeatures: policy.EnabledFeatures,
				MinTlsVersion:   policyMinTlsVersionMap[policy.MinTlsVersion],
			}
			break
		}
	}
	return usedPolicy, nil
}

func getBackendServiceCerts(
	service *compute.BackendService,
	proxies []*compute.TargetHttpsProxy,
	certs map[string]*compute.SslCertificate,
	urlMaps []*compute.UrlMap,
) ([]*compute.SslCertificate, error) {
	getCertsForUrlMap := func(proxy *compute.TargetHttpsProxy, urlMap *compute.UrlMap) []*compute.SslCertificate {
		newCerts := []*compute.SslCertificate{}
		// TODO: `DefaultService` is not enough. Check `HostRules` too.
		if proxy.UrlMap == urlMap.SelfLink && urlMap.DefaultService == service.SelfLink {
			for _, url := range proxy.SslCertificates {
				newCerts = append(newCerts, certs[url])
			}
		}
		return newCerts
	}
	getCertsForProxy := func(proxy *compute.TargetHttpsProxy) []*compute.SslCertificate {
		newCerts := []*compute.SslCertificate{}
		for _, urlMap := range urlMaps {
			newCerts = append(newCerts, getCertsForUrlMap(proxy, urlMap)...)
		}
		return newCerts
	}
	serviceCerts := []*compute.SslCertificate{}
	for _, proxy := range proxies {
		serviceCerts = append(serviceCerts, getCertsForProxy(proxy)...)
	}
	return serviceCerts, nil
}

func loadBalancerFromBackendService(
	service *compute.BackendService,
	proxiesByScope []*compute.TargetHttpsProxy,
	certs map[string]*compute.SslCertificate,
	urlMapsByScope []*compute.UrlMap,
	sslPoliciesByScope []*compute.SslPolicy,
) (*pb.LoadBalancer, error) {
	loadBalancerType := pb.LoadBalancer_UNKNOWN_TYPE
	if externalMatch.MatchString(service.LoadBalancingScheme) {
		loadBalancerType = pb.LoadBalancer_EXTERNAL
	}
	if internalMatch.MatchString(service.LoadBalancingScheme) {
		loadBalancerType = pb.LoadBalancer_INTERNAL
	}

	serviceCerts := []*compute.SslCertificate{}

	if newServiceCerts, err := getBackendServiceCerts(
		service,
		proxiesByScope,
		certs,
		urlMapsByScope,
	); err != nil {
		return nil, fmt.Errorf("unable to retrieve certificates for backend service %q: %v", service.Name, err)
	} else {
		serviceCerts = append(serviceCerts, newServiceCerts...)
	}

	pbCerts := []*pb.Certificate{}
	for _, cert := range serviceCerts {
		certType, err := common.TypeFromSslCertificate(cert)
		if err != nil {
			return nil, fmt.Errorf("retrieve %q: %v", cert.Name, err)
		}
		creationDate, err := time.Parse(time.RFC3339, cert.CreationTimestamp)
		if err != nil {
			return nil, fmt.Errorf("creation timestamp of %q: %v", cert.Name, err)
		}
		expirationDate, err := time.Parse(time.RFC3339, cert.ExpireTime)
		if err != nil {
			return nil, fmt.Errorf("expiration timestamp of certificate %q: %v", cert.Name, err)
		}
		// Parse the certificate chain. The certificate at index 0 is the leaf certificate.
		certs, err := certsFromPemChain(cert.Certificate)
		if err != nil {
			return nil, fmt.Errorf("parse certificate chain of %q: %v", cert.Name, err)
		}
		pbCerts = append(pbCerts, &pb.Certificate{
			Type:                    certType,
			DomainName:              certs[0].Subject.CommonName,
			SubjectAlternativeNames: certs[0].DNSNames,
			CreationDate:            timestamppb.New(creationDate),
			ExpirationDate:          timestamppb.New(expirationDate),
			Issuer:                  certs[0].Issuer.CommonName,
			SignatureAlgorithm:      certs[0].SignatureAlgorithm.String(),
			PemCertificateChain:     cert.Certificate,
		})
	}
	usedSslPolicy, err := getSslPolicyForService(
		service,
		proxiesByScope,
		urlMapsByScope,
		sslPoliciesByScope,
	)
	if err != nil {
		return nil, err
	}
	return &pb.LoadBalancer{
		Type:         loadBalancerType,
		Certificates: pbCerts,
		SslPolicy:    usedSslPolicy,
	}, nil
}

// TODO: Retrieve certificates for TCP/SSL LBs as well. This will require retrieving `TargetSslProxies`.
func (collector *GCPCollector) ListLoadBalancers(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	targetHttpsProxies, err := collector.api.ListTargetHttpsProxies(ctx, resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	urlMaps, err := collector.api.ListUrlMaps(ctx, resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	certs, err := collector.api.ListCertificates(ctx, resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	certsByUrl := make(map[string]*compute.SslCertificate)
	for _, cert := range certs {
		certsByUrl[cert.SelfLink] = cert
	}
	backendServices, err := collector.api.ListBackendServices(ctx, resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	sslPolicies, err := collector.api.ListSslPolicies(ctx, resourceGroup.Name)
	if err != nil {
		return nil, err
	}

	loadBalancers := []*pb.Resource{}
	for _, backendService := range backendServices {
		if lb, err := loadBalancerFromBackendService(
			backendService,
			targetHttpsProxies,
			certsByUrl,
			urlMaps,
			sslPolicies,
		); err != nil {
			glog.Errorf("no LB for backend service %s: %v", backendService.Name, err)
		} else {
			loadBalancers = append(loadBalancers, &pb.Resource{
				Uid:               common.GetUUID(3),
				ResourceGroupName: resourceGroup.Name,
				Name:              formatResourceName(backendService.Name, backendService.Id),
				Parent:            resourceGroup.Name,
				Type: &pb.Resource_LoadBalancer{
					LoadBalancer: lb,
				},
			})
		}
	}
	return loadBalancers, nil
}
