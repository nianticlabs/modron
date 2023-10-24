package common

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/api/compute/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/pb"
)

const (
	ResourceApiKey              = "ApiKey"
	ResourceBucket              = "Bucket"
	ResourceExportedCredentials = "ExportedCredentials"
	ResourceKubernetesCluster   = "KubernetesCluster"
	ResourceLoadBalancer        = "LoadBalancer"
	ResourceNetwork             = "Network"
	ResourceResourceGroup       = "ResourceGroup"
	ResourceServiceAccount      = "ServiceAccount"
	ResourceVmInstance          = "VmInstance"
	ResourceDatabase            = "Database"
	ResourceGroup               = "Group"
)

// See `Ssl`
const (
	CertificateManaged     = "MANAGED"
	CertificateSelfManaged = "SELF_MANAGED"
	CertificateUnknown     = "TYPE_UNSPECIFIED"
)

func TypeFromResourceAsString(rsrc *pb.Resource) (ty string, err error) {
	if rsrc == nil {
		return "", fmt.Errorf("resource must not be nil")
	}
	switch rsrc.Type.(type) {
	case *pb.Resource_ApiKey:
		ty = ResourceApiKey
	case *pb.Resource_Bucket:
		ty = ResourceBucket
	case *pb.Resource_ExportedCredentials:
		ty = ResourceExportedCredentials
	case *pb.Resource_KubernetesCluster:
		ty = ResourceKubernetesCluster
	case *pb.Resource_LoadBalancer:
		ty = ResourceLoadBalancer
	case *pb.Resource_Network:
		ty = ResourceNetwork
	case *pb.Resource_ResourceGroup:
		ty = ResourceResourceGroup
	case *pb.Resource_ServiceAccount:
		ty = ResourceServiceAccount
	case *pb.Resource_VmInstance:
		ty = ResourceVmInstance
	case *pb.Resource_Database:
		ty = ResourceDatabase
	case *pb.Resource_Group:
		ty = ResourceGroup
	default:
		err = fmt.Errorf("unknown resource type %q", rsrc.Type)
	}
	return
}

func TypeFromSslCertificate(cert *compute.SslCertificate) (ty pb.Certificate_Type, err error) {
	switch cert.Type {
	case CertificateManaged:
		ty = pb.Certificate_MANAGED
	case CertificateSelfManaged:
		ty = pb.Certificate_IMPORTED
	case CertificateUnknown:
		ty = pb.Certificate_UNKNOWN
	default:
		err = fmt.Errorf("unknown certificate type %q", cert.Type)
	}
	return
}

// TODO: Cast without (un)marshaling if possible.
func ResourceFromStructValue(value *structpb.Value) (*pb.Resource, error) {
	valueJson, err := protojson.Marshal(value)
	if err != nil {
		return nil, err
	}

	rsrc := &pb.Resource{}
	if err := protojson.Unmarshal(valueJson[:], rsrc); err != nil {
		return nil, err
	}

	return rsrc, nil
}

// TODO: Cast without (un)marshaling if possible.
func StructValueFromResource(rsrc *pb.Resource) (*structpb.Value, error) {
	rsrcJson, err := protojson.Marshal(rsrc)
	if err != nil {
		return nil, err
	}

	value := &structpb.Value{}
	if err := protojson.Unmarshal(rsrcJson[:], value); err != nil {
		return nil, err
	}

	return value, nil
}

func CloneResource(rsrc *pb.Resource) *pb.Resource {
	return proto.Clone(rsrc).(*pb.Resource)
}

// tries to get a UUID with retries
func GetUUID(retry uint) string {
	retries := retry
	for retries > 0 {
		UUID, err := uuid.NewRandom()
		if err == nil {
			return UUID.String()
		}
		retries--
	}
	// we should not be here so we PANIC
	panic(fmt.Sprintf("Failed getting UUID after %d retries", retry))
}

func init() {
	// We use uuid a lot and without this we get too many collisions.
	uuid.EnableRandPool()
}
