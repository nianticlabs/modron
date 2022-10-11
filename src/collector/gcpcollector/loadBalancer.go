package gcpcollector

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/pb"
)

var (
	externalMatch = regexp.MustCompile("INTERNAL")
	internalMatch = regexp.MustCompile("EXTERNAL")
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
				err = fmt.Errorf("unable to parse X.509 certificate from DER block")
			} else {
				certs = append(certs, cert)
			}
		}
		next = rest
	}
	return
}

func getBackendServiceCerts(
	service *compute.BackendService,
	proxiesByScope map[string]compute.TargetHttpsProxiesScopedList,
	certs map[string]*compute.SslCertificate,
	urlMapsByScope map[string]compute.UrlMapsScopedList,
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
		for _, urlMaps := range urlMapsByScope {
			for _, urlMap := range urlMaps.UrlMaps {
				newCerts = append(newCerts, getCertsForUrlMap(proxy, urlMap)...)
			}
		}
		return newCerts
	}
	serviceCerts := []*compute.SslCertificate{}
	for _, proxies := range proxiesByScope {
		for _, proxy := range proxies.TargetHttpsProxies {
			serviceCerts = append(serviceCerts, getCertsForProxy(proxy)...)
		}
	}
	return serviceCerts, nil
}

func loadBalancerFromBackendService(
	service *compute.BackendService,
	proxiesByScope map[string]compute.TargetHttpsProxiesScopedList,
	certs map[string]*compute.SslCertificate,
	urlMapsByScope map[string]compute.UrlMapsScopedList,
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
		return nil, fmt.Errorf("unable to retrieve certificates for backend service %q", service.Name)
	} else {
		serviceCerts = append(serviceCerts, newServiceCerts...)
	}

	pbCerts := []*pb.Certificate{}
	for _, cert := range serviceCerts {
		certType, err := common.TypeFromSslCertificate(cert)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve type of certificate %q", cert.Name)
		}
		creationDate, err := time.Parse(time.RFC3339, cert.CreationTimestamp)
		if err != nil {
			return nil, fmt.Errorf("unable to parse creation timestamp of certificate %q", cert.Name)
		}
		expirationDate, err := time.Parse(time.RFC3339, cert.ExpireTime)
		if err != nil {
			return nil, fmt.Errorf("unable to parse expiration timestamp of certificate %q", cert.Name)
		}
		// Parse the certificate chain. The certificate at index 0 is the leaf certificate.
		certs, err := certsFromPemChain(cert.Certificate)
		if err != nil {
			return nil, fmt.Errorf("error parsing chain of certificate %q: %v", cert.Name, err)
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
	return &pb.LoadBalancer{
		Type:         loadBalancerType,
		Certificates: pbCerts,
	}, nil
}

// TODO: Retrieve certificates for TCP/SSL LBs as well. This will require retrieving `TargetSslProxies`.
func (collector *GCPCollector) ListLoadBalancers(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	targetHttpsProxies, err := collector.api.ListTargetHttpsProxies(resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	urlMaps, err := collector.api.ListUrlMaps(resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	certs, err := collector.api.ListCertificates(resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	certsByUrl := make(map[string]*compute.SslCertificate)
	for _, certsByScope := range certs.Items {
		for _, cert := range certsByScope.SslCertificates {
			certsByUrl[cert.SelfLink] = cert
		}
	}
	backendServices, err := collector.api.ListBackendServices(resourceGroup.Name)
	if err != nil {
		return nil, err
	}

	loadBalancers := []*pb.Resource{}
	for _, backendServiceList := range backendServices.Items {
		for _, backendService := range backendServiceList.BackendServices {
			if lb, err := loadBalancerFromBackendService(
				backendService,
				targetHttpsProxies.Items,
				certsByUrl,
				urlMaps.Items,
			); err != nil {
				return nil, err
			} else {
				loadBalancers = append(loadBalancers, &pb.Resource{
					Uid:               collector.getNewUid(),
					ResourceGroupName: resourceGroup.Name,
					Name:              formatResourceName(backendService.Name, backendService.Id),
					Parent:            resourceGroup.Name,
					Type: &pb.Resource_LoadBalancer{
						LoadBalancer: lb,
					},
				})
			}
		}
	}
	return loadBalancers, nil
}
