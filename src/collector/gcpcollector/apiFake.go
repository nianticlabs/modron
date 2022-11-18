package gcpcollector

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/apikeys/v2"
	cloudasset "google.golang.org/api/cloudasset/v1p1beta1"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/spanner/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"google.golang.org/api/storage/v1"

	"github.com/nianticlabs/modron/src/model"
)

func NewFake(ctx context.Context, storage model.Storage) *GCPCollector {
	fakeApi := GCPApiFake{}
	return &GCPCollector{
		api:     &fakeApi,
		storage: storage,
	}
}

type GCPApiFake struct {
}

func (api *GCPApiFake) ListApiKeys(name string) (*apikeys.V2ListKeysResponse, error) {
	return &apikeys.V2ListKeysResponse{
		Keys: []*apikeys.V2Key{
			{
				Name:         "api-key-unrestricted-0",
				Restrictions: nil,
			},
			{
				Name: "api-key-unrestricted-1",
				Restrictions: &apikeys.V2Restrictions{
					ApiTargets: nil,
				},
			},
			{
				Name: "api-key-with-overbroad-scope-1",
				Restrictions: &apikeys.V2Restrictions{
					ApiTargets: []*apikeys.V2ApiTarget{
						{
							Service: "iamcredentials.googleapis.com",
						},
						{
							Service: "storage_api",
						},
						{
							Service: "apikeys",
						},
					},
				},
			},
			{
				Name: "api-key-without-overbroad-scope",
				Restrictions: &apikeys.V2Restrictions{
					ApiTargets: []*apikeys.V2ApiTarget{
						{
							Service: "bigquerystorage.googleapis.com",
						},
					},
				},
			},
		},
	}, nil
}

func (api *GCPApiFake) ListBuckets(name string) (*storage.Buckets, error) {
	creationTimestamp := time.Now().Format(time.RFC3339)
	return &storage.Buckets{
		Items: []*storage.Bucket{
			{
				Name:            "bucket-1",
				Id:              "bucket-1",
				TimeCreated:     creationTimestamp,
				Encryption:      &storage.BucketEncryption{},
				RetentionPolicy: &storage.BucketRetentionPolicy{},
				IamConfiguration: &storage.BucketIamConfiguration{
					UniformBucketLevelAccess: &storage.BucketIamConfigurationUniformBucketLevelAccess{},
				},
			},
			{
				Name:            "bucket-2",
				Id:              "bucket-2",
				TimeCreated:     creationTimestamp,
				Encryption:      &storage.BucketEncryption{},
				RetentionPolicy: &storage.BucketRetentionPolicy{},
				IamConfiguration: &storage.BucketIamConfiguration{
					UniformBucketLevelAccess: &storage.BucketIamConfigurationUniformBucketLevelAccess{},
				},
			},
			{
				Name:            "bucket-3",
				Id:              "bucket-3",
				TimeCreated:     creationTimestamp,
				Encryption:      &storage.BucketEncryption{},
				RetentionPolicy: &storage.BucketRetentionPolicy{},
				IamConfiguration: &storage.BucketIamConfiguration{
					UniformBucketLevelAccess: &storage.BucketIamConfigurationUniformBucketLevelAccess{},
				},
			},
		},
	}, nil
}

func (api *GCPApiFake) ListAllResourceGroups(ctx context.Context) ([]*cloudresourcemanager.Project, error) {
	return []*cloudresourcemanager.Project{
		{
			ProjectId:      "modron-test",
			LifecycleState: "ACTIVE",
		},
		{
			ProjectId:      "pending-deletion",
			LifecycleState: "DELETE_REQUESTED",
		},
	}, nil
}

func (api *GCPApiFake) ListBucketsIamPolicy(bucketId string) (*storage.Policy, error) {
	iamPolicies := map[string]*storage.Policy{
		"bucket-1": {
			Bindings: []*storage.PolicyBindings{
				{
					Role: "roles/storage.objectViewer",
					Members: []string{
						"allAuthenticatedUsers",
					},
				},
			},
		},
		"bucket-2": {
			Bindings: []*storage.PolicyBindings{
				{
					Role: "roles/storage.objectViewer",
					Members: []string{
						"account-1",
					},
				},
				{
					Role: "roles/storage.objectViewer",
					Members: []string{
						"account-2",
					},
				},
			},
		},
		"bucket-3": {
			Bindings: []*storage.PolicyBindings{
				{
					Role: "roles/storage.objectViewer",
					Members: []string{
						"allUsers",
					},
				},
			},
		},
	}
	if iamPolicy, ok := iamPolicies[bucketId]; ok {
		return iamPolicy, nil
	} else {
		return nil, fmt.Errorf("invalid bucket id %q", bucketId)
	}
}

func (api *GCPApiFake) ListProjectIamPolicy(name string) (*cloudresourcemanager.Policy, error) {
	return &cloudresourcemanager.Policy{
		Bindings: []*cloudresourcemanager.Binding{
			{
				Role: "roles/owner",
				Members: []string{
					"user:account-1@example.com",
					"user:account-2@example.com",
				},
			},
			{
				Role: "roles/test2",
				Members: []string{
					"account-2",
				},
			},
			{
				Role: "roles/iam.serviceAccountAdmin",
				Members: []string{
					"account-1",
				},
			},
			{
				Role: "roles/dataflow.admin",
				Members: []string{
					"account-1",
				},
			},
			{
				Role: "roles/viewer",
				Members: []string{
					"account-2",
				},
			},
		},
	}, nil
}

func (api *GCPApiFake) SearchIamPolicy(ctx context.Context, scope string, query string) ([]*cloudasset.IamPolicySearchResult, error) {
	return []*cloudasset.IamPolicySearchResult{
		{
			Policy: &cloudasset.Policy{
				Bindings: []*cloudasset.Binding{
					{
						Members: []string{"owner@example.com"},
						Role:    "roles/owner",
					},
				},
			},
		},
	}, nil
}

func (api *GCPApiFake) ListZones(name string) (*compute.ZoneList, error) {
	return &compute.ZoneList{
		Items: []*compute.Zone{
			{Name: "zone-1"},
			{Name: "zone-2"},
			{Name: "zone-3"},
		},
	}, nil
}

func (api *GCPApiFake) ListClustersByZone(name string, zone string) (*container.ListClustersResponse, error) {
	return &container.ListClustersResponse{
		Clusters: []*container.Cluster{},
	}, nil
}

func (api *GCPApiFake) ListCertificates(name string) (*compute.SslCertificateAggregatedList, error) {
	creationTimestamp := time.Now()
	sslCertificatesScopedList := map[string]compute.SslCertificatesScopedList{
		"scope-0": {
			SslCertificates: []*compute.SslCertificate{
				{
					Name:              "cert-0",
					Type:              "MANAGED",
					CreationTimestamp: creationTimestamp.Format(time.RFC3339),
					ExpireTime:        creationTimestamp.Add(time.Hour * 24 * 365).Format(time.RFC3339),
					SelfLink:          "/links/cert-0",
					Certificate: strings.ReplaceAll(`
					-----BEGIN CERTIFICATE-----
					MIIFTTCCAzUCCQDBvVCMwjjyWjANBgkqhkiG9w0BAQUFADBVMRAwDgYDVQQLDAdV
					bmtub3duMRAwDgYDVQQKDAdVbmtub3duMRAwDgYDVQQHDAdVbmtub3duMRAwDgYD
					VQQIDAd1bmtub3duMQswCQYDVQQGEwJDSDAeFw0yMjA3MTkwOTEwMDRaFw0yMzA3
					MTkwOTEwMDRaMHwxJTAjBgNVBAMMHGRvbWFpbi0xLm1vZHJvbi5uaWFudGljLnRl
					YW0xEDAOBgNVBAsMB1Vua25vd24xEDAOBgNVBAoMB1Vua25vd24xEDAOBgNVBAcM
					B1Vua25vd24xEDAOBgNVBAgMB3Vua25vd24xCzAJBgNVBAYTAkNIMIICIjANBgkq
					hkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA7b8cRiT7Wy7/uyY3sowwEBpWd8V6tzQP
					00ISr6K8Yi5boIbJ8sO17DQttOckfmRl6/itzDrZnaaJ/0Jxh7WfifVvxACyDdEZ
					QxVUf/FOoyw9qUeI4PH7tPogW6Zkc10L2i7v6fvz6FzT1QTRRlX74X8AD3F/9I8X
					w5SGfvfCnWJHGiY3cp29hezUkqhfuPUDhN+vRnQuyoxjib2BBtiWTwr4+t2kfWep
					db6LuIZ6fLcfFN9CCow6YfK2Q0kw9VdPsVr8YkDMdCEKoyJKQcKIB4B5BfMGTH0F
					8op9nxIgNJ6K6LgpUtOWQBvAKIzRpnJQfq7wfqWLEp8P4F7VWq/ysLP7tJnMc643
					c397n7y+DpGGHCB/jfrg/Uu6rzpLiDwZaFeNSU9MQ074oZ+bEpJRFb40FEKK+Qov
					ytXE/f7oC+5hPnDKPN1DDYZAMw2cMzyL0W3/lOi+X3HuxWCDieNgVbvfWea9ejfC
					NuA86OrzELyHqqTXw7jr1rIdNlPjcU4G0mAuqsfHBD42wc406OBL45zKe+Icu7dt
					3ps/dx58ZroYOVqWEo+lfAG9F9hktX9HJUhVGzTLFjsd0UzeGvPHLgL2Y/GlHyK1
					kym4tDFzDLuk4jG7G20ctaIdjbhh0UDp0uVmCZY5r78h1mQzObXFkeup2VdI+yIe
					bN1o6Po8nAkCAwEAATANBgkqhkiG9w0BAQUFAAOCAgEAtPEBotLk5ucJ70wpnPX2
					agRWJ8MpJvqnUP5iEVv9iJlD2EnUSU+E9YuuaMipw+F7g6BUFx39/ZQmzqR3Jh1c
					gPaNU5YdVWqHPnukCMXKWfvN8WJPLyrZaJenjn/nFwFnEBsle6JtCQJ6okEXwD1V
					LQoopVfqkXyYVICupOZhcTa/6MB59tUOUzOy5LnBZj3XGEQXE67eA+eDg1vfivDS
					ux1H1eopE978RtGArmnZCkuUxxv39aEDWbN58tFb7MRcy43GuK3GdPlP9bUh02+d
					f9dmpLWgrnxyub8tqK9bV3A/POHk3KLY1bUc5ZZFJVM3rR0Y87P38bPcOfcvb/H2
					SI9H7UjWMI5+K+DwZNL00h0N9EgHxcewslav7JTWAm1SSmMrUOLLHWhlAOsKWpAt
					f67dWGWem4df0hrAk4kyyWlBDssDNgh9zN2VXewTZd4j8S5Sr9pTzVMGlTaIpCWn
					bRKfJpVEKgEAzmVBnmLEyKcX32LFeDIt+JfZcIjEzzxkMQhtcrYfDZOJGs9J5rh1
					M0ovQVnQiVfVRIyt/TiRkuRIDAcOcwO1np3IPTz63oO3iEkMqWbh4z6+ho/3j3cm
					gFNrdll08NWC5hmcCIwv6hHk7DlLVXzrDP3ZLNm7JcW2gwygn8BgQSu1SLAlaM3c
					R1UzrGiyiwwbtyAKYwrn+A0=
					-----END CERTIFICATE-----
					-----BEGIN CERTIFICATE-----
					MIIFJjCCAw4CCQCEKQM+pKbyBTANBgkqhkiG9w0BAQsFADBVMRAwDgYDVQQLDAdV
					bmtub3duMRAwDgYDVQQKDAdVbmtub3duMRAwDgYDVQQHDAdVbmtub3duMRAwDgYD
					VQQIDAd1bmtub3duMQswCQYDVQQGEwJDSDAeFw0yMjA3MTkwOTA5MzZaFw0yMzA3
					MTkwOTA5MzZaMFUxEDAOBgNVBAsMB1Vua25vd24xEDAOBgNVBAoMB1Vua25vd24x
					EDAOBgNVBAcMB1Vua25vd24xEDAOBgNVBAgMB3Vua25vd24xCzAJBgNVBAYTAkNI
					MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA8fqY2Tdvu+fBAlUhMvPu
					3KPSjUMS2Oe1rfhRGHT/oIIp+oQenAKkUnlkJoct+UKiDijg4ovFKUjenxWqu66n
					mCEUFGtBWZrIPlczhMm+vuQipMy6VTV+8wgMFd40wkupxwPNhTN1tk7wsOqu2bQy
					FMquqJGU4LAUCuUuu95p2OaKxBzKcK2zhJKUEPuqZJED9UPdurwUfoNMbekh4YlO
					WDmbVGJDOcs8o9fxeAgS7uGL7Y1G53BZ3ZVj3kZ4puAbvRd7g3mxiFKgY0V/6IOb
					Cj4x3L4HJry2KNnaRj1/MxELAv5STRo5Sa7CTXUmKaxyEsgK2+ek5ispo8w1Efvk
					Q7oQiKoCmtmgRWIroRHPOahBldk82CCiQHGJV5gCEL+n62OJb51uM51Amwr1pBLm
					6p3PflLUp8nIfUGUw845T8KfdZPz7nOif+IGfch1nZZJN1tuw9YcfKbjsh62xi+/
					nN4B2KFQl89MgCjWm+TQDoP4ToIS1+BlB+DXPy8zLwa+sUSNMawXf9LIYV7wWjD+
					Snk/8IceCaxF3SW5EjausKQ+cYXN6LecOlL5Aw9I1++PuZ4VTfbtC+BFvRUP/apb
					FzzLziMwxlhC2LV6lMJN6V6T+pM1HnNDPv0SuUCO/lzI/NKC42llmS3xUSMMoZNo
					QrU3ClZk2Fs0z8/qc0ycb/cCAwEAATANBgkqhkiG9w0BAQsFAAOCAgEASL9Td1DU
					NdtfPpW39Uwia5hnG6rnRz37EcqWup+V9zzvzOFc1FnRPItNPRJLmJoQ9CLAWJVG
					yrE/bcODLbyGeGC6vvRhTcpriij99kEjy36cxm+XqkSBUYRqUs87jLvZrdhPaSKq
					P1gj3LK7LE7nCYuP6zVLrnrVxlVbeKXIs5Pcaa9sYR9oi+hGDRn6bcDj4y7qixxW
					LuVYnjyHs/pKb+75DRyaDFF7u4VlXcqGH2t8F+ZipNzMT2mx7sr+xQkpsbJQRVSL
					Cx4ih2TbcyqApLU50JgjtbQYvMOngB+NI2LgJ/VOzokSLGG9YbblfYMPggQYaUXC
					bDQuPvQCG0hQpqBKgluk65AmCba4BCRLLUy01U1i7ScxtmtWyn1HSLyJmxGkCGxc
					OWL4qMDIGtgE32Es2WW8VSfFpH7n85hFx6Z93tVTgjxWQn6t2cAu17qbgVuI81mp
					gpKRYgexWtC/K2bftPGrjajWSsRTIM1JZd6awtjBdbgvLeu02MERQQ5wZ7a9Ee1X
					EjKOG1vj2c95sMiuwebY+evXZ5najnNsdYwfSXyX1hULt1R59hxPcuYVig1qM+ch
					oRU0QKlNW30K+RQPb/ZGMFODsaYNOxvgAQSSeOQjyoVVHm5ZBZoU3LY98M9X2kFx
					FbGm99HuLTXv1ReyURGzjxZIAqHd6hnX5wk=
					-----END CERTIFICATE-----
					`, "\t", ""),
				},
				{
					Name:              "cert-1",
					Type:              "SELF_MANAGED",
					CreationTimestamp: creationTimestamp.Format(time.RFC3339),
					ExpireTime:        creationTimestamp.Add(time.Hour * 24 * 365).Format(time.RFC3339),
					SelfLink:          "/links/cert-1",
					Certificate: strings.ReplaceAll(`
					-----BEGIN CERTIFICATE-----
					MIIFTTCCAzUCCQD9AMCeW12GEDANBgkqhkiG9w0BAQUFADBVMRAwDgYDVQQLDAdV
					bmtub3duMRAwDgYDVQQKDAdVbmtub3duMRAwDgYDVQQHDAdVbmtub3duMRAwDgYD
					VQQIDAd1bmtub3duMQswCQYDVQQGEwJDSDAeFw0yMjA3MTkwOTAwNDBaFw0yMzA3
					MTkwOTAwNDBaMHwxJTAjBgNVBAMMHGRvbWFpbi0wLm1vZHJvbi5uaWFudGljLnRl
					YW0xEDAOBgNVBAsMB1Vua25vd24xEDAOBgNVBAoMB1Vua25vd24xEDAOBgNVBAcM
					B1Vua25vd24xEDAOBgNVBAgMB3Vua25vd24xCzAJBgNVBAYTAkNIMIICIjANBgkq
					hkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA0no57IbQwzRNJ61/FoBPknTqa24oNATG
					DU/oY4bTQCOPC7bEc/5IAPQ8Fhd2u3HsI5rkkUGAeQCAnEymB5z+RMG4uGZk7CKS
					mozKFhf2asSSehWENYCORc2sM5DRMC+7mOHl1+Gs3uIpZhq0C2L+HAbyXi9TE8bu
					lyyD/bk65ocGfSsajHJr6VzlKI/YZFzty6VXlSafxcEVbUrZ//YqtogCHQHN5d9Y
					sw+5tYehGqi0O1XPQ+KkJ4PmP9yD46MAtPtiQwBRClA7G68jJ3j50stG/A59Otvj
					COUApdQeL9VwHhNh9t4TQZaZcGLFb9qyNKW2hhJP63cQTPQ7asTagKemdSB0zwoT
					LKZia67bG4PXExLwPuxI12oFug/g0AGdrmlJiIjXGvZPyEJ7bc/6hUxWfEpfo77v
					brkndn2aJTS98HI3R+eBCbCkYK6yWBJmAV1RjcNNGlPOIHxWeJtnG90L/n7vNIL1
					LtzK/W9cTLAH1PkexeSmOEJqYC0FihjkhXZM0KUZFjjT5z5wdZsDX1f+q/6Yvo/R
					2OphVI5GfY/W5Jc2mzsGzkiHi3epow0MEtbEJ4c12afxKuUGqwKyO8T92DOVZ2BS
					iew/DZ673hbek3C4sVq+c4fCIxYOLL/fqeL5r+yh7a+tAGlIxU1fpSodanu+ksQU
					T7TqJFH0LXECAwEAATANBgkqhkiG9w0BAQUFAAOCAgEANm6bCAju+jm9sI/Pry1u
					KS6SRMNX0ne0knjDRXO0lHjMz095xmEQA/Wu6dpPkhMo7QdYsZ0bntIgt4Go/1qx
					3rZUnP9B77BqPsLxqgBfMxryZVADQMEvsUkr0yK9g6crEQ/aSu5h9hEiSX1FA+9y
					Jy00TwxS2NcTPd1AuFa5/lXZaw2Iu9nwZ4/+2QuZrjmZfE28gcUGb1GDcBMzcqZM
					8O4J4Xogi5DSzQLucPkBX8uD1ibEn1Cs16Kzq9X+45M3zPWwNnV4yM+38ZxU20Do
					gDTKR2Md5JByhKt+8TSe1S+fEg5cZwGj8P3LCUFUAPjloHu59sjmHcc21MHhS3JQ
					TpqQJgLpo3bdwbhUsvenXSUsk08e0PnvaIym6ALgDku/ZWYLmkGKHDSWE4otFDkF
					BUbDHHxSuH8Pk5eNSOf7rfFmDk7r9Hj3ryqMZf8xq+kIHmzNESAQskFScBPj3iY3
					mCb7p2/gEmSddYR7TtDG+J1au4+sVDkFd9dIrMVhwZuY19m0S6TqpJ3pp9p6OBoq
					d8ZyTuiP6LTehRiFBQrFA7LGhU14pPIVbOS5PuP2W91DzL9ZfKwsQ/Tr08ZhOH58
					ocnbZWQWQ6NEZzsnrwuyNa7DxLUfDc3Itfg94oYy4YSO7SdrifJWAvHlqV69CZ8K
					G67Se2laEbw8sNaehw/0mpE=
					-----END CERTIFICATE-----
					-----BEGIN CERTIFICATE-----
					MIIFJjCCAw4CCQDlQR4bAQ4wBzANBgkqhkiG9w0BAQsFADBVMRAwDgYDVQQLDAdV
					bmtub3duMRAwDgYDVQQKDAdVbmtub3duMRAwDgYDVQQHDAdVbmtub3duMRAwDgYD
					VQQIDAd1bmtub3duMQswCQYDVQQGEwJDSDAeFw0yMjA3MTkwOTAwMTZaFw0yMzA3
					MTkwOTAwMTZaMFUxEDAOBgNVBAsMB1Vua25vd24xEDAOBgNVBAoMB1Vua25vd24x
					EDAOBgNVBAcMB1Vua25vd24xEDAOBgNVBAgMB3Vua25vd24xCzAJBgNVBAYTAkNI
					MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAx1GS+PP1If+ShSuJeIHm
					jCTAr7d8nExf3hZ1x41+MaJfDx6XSFsyqDi9rad2KGnI7x3OFmLP7AfIhVfN+mYB
					HcCST5YfPMuyaLa2roLZHaz0ZzFYhexMsK5Lq/QnaW/Exoo/lY2Rxu8uYLHOX28i
					6lQPka52Wi8D8sWHfUMUMxjoaV97LYFYsz8ngqeN11bvXkDoo5Tp0CWRb1o4LoBt
					cf8ub5PJH8YW5LBglC6rTR1Q9H9YX6gqnSrP3vqqfkVdqPLTlSILqCgqxvKqbjIb
					RkMlkzaiqNf3rz5iC8br0ZfKWgs/Jzvhea+K3J5Y3YbwExjttUekhXIUZqbzAfNQ
					TIAUKfLlmmJfywbmI+xgUfSxNAPwgASjnxmkbAXsVW8SjUvnnLsbla6SyS1a+bwe
					1OP0gmwjtfeGN7QSCtU+GSZ/3RP1mfluog2acHR7HfRi2dzVyabGPKe1FbU/QFIu
					dtl6YSXvKUFM1mC5lIj0s05vTaszw7JKAEQaVizaDCFt/d3xI9oaQFi1Z6W/DxmB
					6GS98iQ9ydLatEXCipfmTrJhxf9mbRE8Z4NVTg4kEllNwcV9W1yRqdVnfLcGe6TB
					lpfk63PYBsLfuB6KVEyu3hGwOfP4UVNE/A/BfyYYVKWobR7L4GzOxmm24ikSo6fq
					HhNOTKogByaSoXBfdm8g1WcCAwEAATANBgkqhkiG9w0BAQsFAAOCAgEAWrVESQoE
					UXRxoUOxEKhS9UgsQrf70k8wHgwgEGFUATPFTfMInWhrQ+VMj4ImSxDOT5tDLADq
					0hU/h0oQ2XkC14YVpcF835Vt2mCPaRugPzjxzU/Ky9Ie39izZeBvG2orCthEglof
					FtVZGc3vCmWEXs7zPhSwx2BsZNw0xVMLg6lTok7wVcf66lW/1PWQp8dKwwZlSvHg
					VgRLfmH2yisEak0euw1zMYRs4GdwMxUC69ImremBG1MrAQdyvp1ZM7XamyLFZivg
					UTnVMDLduHub+IpITnl2IYgqvMdATpL9h4n036WvYvu2qP765j0ZW0yvBMFSVS3/
					eKVEn3NqK3nFZGlo4W/i3mElbFtd+q8mxQiI1S9hF1W1yTuVPDfsVYzO+wWOQHdk
					b4XoWikC2eq98zMp9wWPFrnbNFiTLTllWKUYWZQgx7UrkA+wmtKOAwxiY0tADMHg
					IwLHGUhHpIG5ErJX7AKFUShb3jSqujOkU8Bmtr0W2jd+uYGp8MWT8d5drrO2aVAW
					CdBMmRly672Sy1Y7MTZjLykrMEdsnmXosvIvzPWzbqAjsTJQR3OKSFMaBO+lxCXs
					n+WngS5fO6hiTKqf1fjDSeBhOlpywVV8h0ONMNF0TIHyydJEYbIlZBajER3dUIZs
					muOKyutYtJqW5tqke8N7Yy9oDUlqtt6gnFE=
					-----END CERTIFICATE-----
					`, "\t", ""),
				},
				{
					Name:              "cert-2",
					Type:              "MANAGED",
					CreationTimestamp: creationTimestamp.Format(time.RFC3339),
					ExpireTime:        creationTimestamp.Add(time.Hour * 24 * 365).Format(time.RFC3339),
					SelfLink:          "/links/cert-2",
					Certificate: strings.ReplaceAll(`
					-----BEGIN CERTIFICATE-----
					MIIFTTCCAzUCCQCUhTr1JbteOjANBgkqhkiG9w0BAQUFADBVMRAwDgYDVQQLDAdV
					bmtub3duMRAwDgYDVQQKDAdVbmtub3duMRAwDgYDVQQHDAdVbmtub3duMRAwDgYD
					VQQIDAd1bmtub3duMQswCQYDVQQGEwJDSDAeFw0yMjA3MTkwOTExMjlaFw0yMzA3
					MTkwOTExMjlaMHwxJTAjBgNVBAMMHGRvbWFpbi0yLm1vZHJvbi5uaWFudGljLnRl
					YW0xEDAOBgNVBAsMB1Vua25vd24xEDAOBgNVBAoMB1Vua25vd24xEDAOBgNVBAcM
					B1Vua25vd24xEDAOBgNVBAgMB3Vua25vd24xCzAJBgNVBAYTAkNIMIICIjANBgkq
					hkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAuiFSKID/n0ozLBdVrxCjyyHq9jEsbOWw
					CmcJ7nbrsiP9w1qceljMg1/IvqaqHySoPvxurp8Vxq9OXgWVJzmCuMyoCUMsXEyt
					K9+CcCz0AMoaxY5uO2+ehpm2SPnx+p8zfKqH6n9Heto+YU2IQYJTkbAIEKXSJib4
					iDutY7AlCAzlMqFXZpxyPBQfgFvM744WE/FOhv3ASdJ7TTqK8vuulSk46zgYe/hz
					jBevt8MUIyE9dsQ/eIL4HsS+OofkDU4onsTJVrMpetZD0KwB6lBrdVLkIbn5mWHD
					2ONRLuuncQPLkyc1LEjY5+RYj2KUrXX+3jByco7e4pFDDJjFKXDl/MujCNtWtb9z
					SaR362Ic6/93CkvmCiuS9IeqrMv8ZniM7HOtSs4Zq/e9Ym4/YJ5BbzrbSK8pMvhE
					dsSENgxlSGCVBrs1DLBSmJ89qX3Q+Y4ejJq97Tzs9yR6MEycFYKOGX4FdwkfcZb6
					Fi+v13D+9x8WUtehTOcap9jXbiACSzGbkVbD0Q/MlEPhj9erkyQQFwhaLTZCoEo9
					SpMH/pJt8n4AYtfl1Xrw8yLjv93n54MNjMDOELMJluPAP4GfkAOtS/ck3gT/Cm53
					SW76ocPTXrN0cvybl5ShM0coC+jukbTbBUQ7eiv37LEcsZCj5CNW4/ee/Vrn1XlI
					TtBrN9lZVjUCAwEAATANBgkqhkiG9w0BAQUFAAOCAgEAfKrlr72uUmD0CNmQctP4
					dAN0vuIkNC5aRW8P3p8Ba7urgWo53r2f25gxbpW/ESRT0vgMChHhoKzgP65o67n2
					8neOIedfJPTxZRlKqsc1Yhlp+FtVgnrgk0oTZupYqIHT4L98uFzNkhZo7FFHOLzA
					rrbMABXgRs8umW7OMsjBw2akbRwxbPwalM8NGvgzH1zAUz5oMq9s1AqFHITdLDvu
					/CWM/vBJGXqkQ/+uHYrOP1E8hrAZtQFEQ9rJVR3fKLAqAcSLBmB0q94zbXoBq/W8
					uZyXdrItA1COb8oeHSzJejcGEPl6VhfsgKTHoSwWCvplPZ8SOP4g8XohqXT9aakR
					ism2Fmc9yewA7GOzU4vw/6Oqaka6MwoyIUUbb7Mt6rdcE+gJiqfujwyilt+DIkx+
					D7Ex+meotSTP+xWIJbaeoNxJ/M19gGH0M9FxlIoYr/flYCSkUNlGy5EFSclI7VTz
					GlOCJmBICrj3VDP3Q4iHv8O7DErAv+oHYf7j53/jg2mIDeIVJRjihzwYBaZN6dLd
					sVmCwzd/ZntJlu2II2PnrR5UxsfmpevMSgMrChSOKh/mfPGVNF78r7QXawkrndZa
					nymkGGSoll+Shlv5MpKB9PR4XfMM14dyuE568AmSbMnGPhYqSmauXDSSXnNi7KtT
					kNOsaHc3Uaw1jIi2BOwpJfY=
					-----END CERTIFICATE-----
					-----BEGIN CERTIFICATE-----
					MIIFJjCCAw4CCQDCXGx9MFOr3zANBgkqhkiG9w0BAQsFADBVMRAwDgYDVQQLDAdV
					bmtub3duMRAwDgYDVQQKDAdVbmtub3duMRAwDgYDVQQHDAdVbmtub3duMRAwDgYD
					VQQIDAd1bmtub3duMQswCQYDVQQGEwJDSDAeFw0yMjA3MTkwOTExMjBaFw0yMzA3
					MTkwOTExMjBaMFUxEDAOBgNVBAsMB1Vua25vd24xEDAOBgNVBAoMB1Vua25vd24x
					EDAOBgNVBAcMB1Vua25vd24xEDAOBgNVBAgMB3Vua25vd24xCzAJBgNVBAYTAkNI
					MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA6HkB1jv4AjGmAgkwlozD
					zmkTQiomPS6jWiQBec58j8s0WhLlWwvPKkv090kpY13lGk8FF8OR+wb7J7x6AZl6
					03+PzacBGsXDiRkQ08HYSZ0pW/hRZFQ7UEq00luUZut7xzebbnCDqvSz/C8sGuZk
					V4o9wWiOkDRD2szHRqmiHjZeHePCy/3xEPDb/OMuwXmjnS9qkMYtsCLtub54QmJZ
					3fdol86wSobkydhQ8vvIJqfxmNX3bhJUx6PDNAjoyOgxX+YBMLyDDHRuXFH7mj5H
					Nv0ZDmuVgMGrgzAmWoGIfFRP/R8rGvpscPX0GrhQdoroyLAkvDZiAYnK/Fx8N8Kz
					AHpfxQQIR62X9vuLWtUCdvV0qo0GG2+QCAws4n3BM6TwXkCkQcpg6pIJgtvk/avN
					hVAQcLlul5ohnAVPlMV6/cs+UnOTn8pkCvE3G1JuaSHjsHELropXMeYbrUX5Q+Kd
					9JGxtC/sUIAmAL2YrFxI6tC9RFCCK2HJxZMh012wcwz/HSSrT4yXv+P61OUK0mSo
					cvttUjgpGE5Z0hvyRMEq/UIwuWjNymcO+8f62Cn0v4EM1bh5XS5F0d8ILZjKU4JX
					Bqi7eOWd5MFCS3Q8pdCOvWtxefMDgZb1MYwLyVcsuaHIdrVtzeAEieqD/rbVoXe7
					0oQ8bfjDyWOGbpusqwCxCnECAwEAATANBgkqhkiG9w0BAQsFAAOCAgEAvGLv5q7w
					vZl/UPxIhs1HWEbGaNiOXsiRXYbs6tgViABcsjjZkErlv8oo3KSkf6yfXI55y6tn
					wKqP2RgKuTrWaTp63h6EDGHnLl0X+Nq6FGnxwiRaJ2iYzwvC2mEDVtEgod7DtRdI
					7xRqufLdHJEm8uUm+EfCjWaamnAqc2RCGU50Dezn7Tif88eAMCMMu0Nqg+HCr3zM
					G+/a2OyFxUNZRWGRkthQjQTjPcTu/uBuxtLMzG8BqfpAM8XCH26Q0zxnD4g4NLuo
					03932aEaNFDHfyQJ4sFbH65+1EzudAt2emWS3g519+/0UU9Lthb6Y/aMvnR37Q9x
					dh9GLl+PkDpE8GOZGuavwkNCyvKTGNwYRpQrK8fal6e2sTyKObzmn+s0tkvputy1
					mp/DbMIIXykFRZ8Y2Aps6pgSjJBuI1HBR3nAX+J1fTAjUghEMkbt/N9MheUDQfzh
					hO9Qgo27PltMyUPOuhclLKHIZLsJrgSf8dFfWHpzSYFhtPr2gNTaSWuD6UxLKPHr
					bz9GjrScMtDzB5n8alcySomoATWP5wnPArb6wJyg8pfyrb/43VZeaWCDhPYDVj1i
					A/klwbs/a2YF0w72ZTd1aydFct5ONPcYhcY/4Zip5JZT5SCzWNaNTp8UL5TTv4zn
					y4rzKfl2JQSqXBbOdR4KUDN0uhXFqPDEyK4=
					-----END CERTIFICATE-----
					`, "\t", ""),
				},
			},
		},
	}
	return &compute.SslCertificateAggregatedList{
		Items: sslCertificatesScopedList,
	}, nil
}

func (api *GCPApiFake) ListTargetHttpsProxies(name string) (*compute.TargetHttpsProxyAggregatedList, error) {
	targetHttpsProxiesScopedList := map[string]compute.TargetHttpsProxiesScopedList{
		"scope-0": {
			TargetHttpsProxies: []*compute.TargetHttpsProxy{
				{
					Name:            "proxy-0",
					SslCertificates: []string{"/links/cert-0"},
					UrlMap:          "/links/url-map-0",
				},
				{
					Name:            "proxy-1",
					SslCertificates: []string{"/links/cert-1"},
					UrlMap:          "/links/url-map-1",
				},
				{
					Name:            "proxy-2",
					SslCertificates: []string{"/links/cert-1", "/links/cert-2"},
					UrlMap:          "/links/url-map-2",
				},
				{
					Name:            "proxy-3",
					SslCertificates: []string{},
					UrlMap:          "/links/url-map-3",
				},
			},
		},
	}
	return &compute.TargetHttpsProxyAggregatedList{
		Items: targetHttpsProxiesScopedList,
	}, nil
}

func (api *GCPApiFake) ListUrlMaps(name string) (*compute.UrlMapsAggregatedList, error) {
	urlMapsScopedList := map[string]compute.UrlMapsScopedList{
		"scope-0": {
			UrlMaps: []*compute.UrlMap{
				{
					Name:           "url-map-0",
					DefaultService: "/links/backend-svc-1",
					SelfLink:       "/links/url-map-0",
				},
				{
					Name:           "url-map-1",
					DefaultService: "/links/backend-svc-2",
					SelfLink:       "/links/url-map-1",
				},
				{
					Name:           "url-map-2",
					DefaultService: "/links/backend-svc-3",
					SelfLink:       "/links/url-map-2",
				},
				{
					Name:           "url-map-3",
					DefaultService: "/links/backend-svc-4",
					SelfLink:       "/links/url-map-3",
				},
			},
		},
	}
	return &compute.UrlMapsAggregatedList{
		Items: urlMapsScopedList,
	}, nil
}

func (api *GCPApiFake) ListBackendServices(name string) (*compute.BackendServiceAggregatedList, error) {
	backendServicesScopedList := map[string]compute.BackendServicesScopedList{}
	backendServicesScopedList["scope-0"] = compute.BackendServicesScopedList{
		BackendServices: []*compute.BackendService{
			{
				Name:                "backend-svc-1",
				LoadBalancingScheme: "INTERNAL",
				SelfLink:            "/links/backend-svc-1",
			},
			{
				Name:                "backend-svc-2",
				LoadBalancingScheme: "EXTERNAL",
				SelfLink:            "/links/backend-svc-2",
			},
			{
				Name:                "backend-svc-3",
				LoadBalancingScheme: "EXTERNAL",
				SelfLink:            "/links/backend-svc-3",
			},
			{
				Name:                "backend-svc-4",
				LoadBalancingScheme: "INTERNAL",
				SelfLink:            "/links/backend-svc-4",
			},
		},
	}
	return &compute.BackendServiceAggregatedList{
		Items: backendServicesScopedList,
	}, nil
}

func (api *GCPApiFake) ListRegions(name string) (*compute.RegionList, error) {
	return &compute.RegionList{
		Items: []*compute.Region{
			{Name: "region-1"},
			{Name: "region-2"},
			{Name: "region-3"},
		},
	}, nil
}

func (api *GCPApiFake) ListSubNetworksByRegion(name string, region string) (*compute.SubnetworkList, error) {
	if region == "region-1" {
		return &compute.SubnetworkList{
			Items: []*compute.Subnetwork{
				{
					Name:                  "subnetwork-private-access-should-not-be-reported",
					IpCidrRange:           "IpCdrRange1",
					Purpose:               "PRIVATE",
					PrivateIpGoogleAccess: true,
				},
				{
					Name:                  "subnetwork-no-private-access-should-be-reported",
					IpCidrRange:           "IpCdrRange1",
					Purpose:               "PRIVATE",
					PrivateIpGoogleAccess: false,
				},
				{
					Name:                  "psc-network-should-not-be-reported",
					IpCidrRange:           "IpCdrRange1",
					Purpose:               "PRIVATE_SERVICE_CONNECT",
					PrivateIpGoogleAccess: false,
				},
			},
		}, nil
	}

	return &compute.SubnetworkList{
		Items: []*compute.Subnetwork{},
	}, nil
}

func (api *GCPApiFake) ListServiceAccount(name string) (*iam.ListServiceAccountsResponse, error) {
	return &iam.ListServiceAccountsResponse{
		Accounts: []*iam.ServiceAccount{
			{
				Email: "account-1",
			},
			{
				Email: "account-2",
			},
		},
	}, nil
}

func (api *GCPApiFake) ListInstances(name string) (*compute.InstanceAggregatedList, error) {
	instancesScopedList := map[string]compute.InstancesScopedList{}
	instancesScopedList["scope-0"] = compute.InstancesScopedList{
		Instances: []*compute.Instance{
			{
				Name: "instance-1",
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						NetworkIP: "192.168.0.1",
						AccessConfigs: []*compute.AccessConfig{
							{
								NatIP: "240.241.242.243",
							},
						},
					},
				},
				ServiceAccounts: []*compute.ServiceAccount{
					{
						Email: "account-1",
					},
				},
			},
		},
	}
	return &compute.InstanceAggregatedList{
		Items: instancesScopedList,
	}, nil
}

func (api *GCPApiFake) ListServiceAccountKeys(name string) (*iam.ListServiceAccountKeysResponse, error) {
	return &iam.ListServiceAccountKeysResponse{}, nil
}

func (api *GCPApiFake) ListSpannerDatabases(ctx context.Context, name string) ([]*spanner.Database, error) {
	return []*spanner.Database{
		{
			Name: "spanner-test-db-1",
		},
	}, nil
}

func (api *GCPApiFake) ListUsersInGroup(ctx context.Context, group string) ([]string, error) {
	groups := map[string][]string{
		"emptyGroup": {},
		"group1":     {"groups/group1/memberships/user1", "groups/group1/memberships/group2"},
		"group2":     {"groups/groupd2/memberships/user2"},
	}
	if g, ok := groups[group]; ok {
		return g, nil
	} else {
		return nil, fmt.Errorf("group %q doesn't exist", g)
	}
}

func (api *GCPApiFake) ListCloudSqlDatabases(ctx context.Context, name string) ([]*sqladmin.DatabaseInstance, error) {
	autoResize := true
	return []*sqladmin.DatabaseInstance{
		{
			Name:            "cloudsql-test-db-ok",
			InstanceType:    "CLOUD_SQL_INSTANCE",
			ConnectionName:  "test-connection",
			DatabaseVersion: "TEST_VERSION",
			Settings: &sqladmin.Settings{
				IpConfiguration: &sqladmin.IpConfiguration{
					RequireSsl: true,
					AuthorizedNetworks: []*sqladmin.AclEntry{
						{
							Value: "127.0.0.1/32",
						},
					},
				},
				StorageAutoResize: &autoResize,
			},
		},
		{
			Name:            "cloudsql-test-db-no-authorized-networks",
			InstanceType:    "CLOUD_SQL_INSTANCE",
			ConnectionName:  "test-connection",
			DatabaseVersion: "TEST_VERSION",
			Settings: &sqladmin.Settings{
				IpConfiguration: &sqladmin.IpConfiguration{
					RequireSsl: true,
				},
				StorageAutoResize: &autoResize,
			},
		},
		{
			Name:            "cloudsql-report-not-enforcing-tls",
			InstanceType:    "CLOUD_SQL_INSTANCE",
			ConnectionName:  "test-connection",
			DatabaseVersion: "TEST_VERSION",
			Settings: &sqladmin.Settings{
				IpConfiguration: &sqladmin.IpConfiguration{
					RequireSsl: false,
					AuthorizedNetworks: []*sqladmin.AclEntry{
						{
							Value: "127.0.0.1/32",
						},
					},
				},
				StorageAutoResize: &autoResize,
			},
		},
	}, nil
}
