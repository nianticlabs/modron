package common

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/api/compute/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

const (
	ResourceAPIKey              = "ApiKey"
	ResourceBucket              = "Bucket"
	ResourceDatabase            = "Database"
	ResourceExportedCredentials = "ExportedCredentials" //nolint:gosec
	ResourceGroup               = "Group"
	ResourceKubernetesCluster   = "KubernetesCluster"
	ResourceLoadBalancer        = "LoadBalancer"
	ResourceNamespace           = "Namespace"
	ResourceNetwork             = "Network"
	ResourcePod                 = "Pod"
	ResourceResourceGroup       = "ResourceGroup"
	ResourceServiceAccount      = "ServiceAccount"
	ResourceVMInstance          = "VmInstance"
)

// See `Ssl`
const (
	CertificateManaged     = "MANAGED"
	CertificateSelfManaged = "SELF_MANAGED"
	CertificateUnknown     = "TYPE_UNSPECIFIED"
)

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
	valueJSON, err := protojson.Marshal(value)
	if err != nil {
		return nil, err
	}

	rsrc := &pb.Resource{}
	if err := protojson.Unmarshal(valueJSON, rsrc); err != nil {
		return nil, err
	}

	return rsrc, nil
}

// TODO: Cast without (un)marshaling if possible.
func StructValueFromResource(rsrc *pb.Resource) (*structpb.Value, error) {
	rsrcJSON, err := protojson.Marshal(rsrc)
	if err != nil {
		return nil, err
	}

	value := &structpb.Value{}
	if err := protojson.Unmarshal(rsrcJSON, value); err != nil {
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
