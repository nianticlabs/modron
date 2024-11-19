package utils

import pb "github.com/nianticlabs/modron/src/proto/generated"

func GetResourceRef(rsrc *pb.Resource) *pb.ResourceRef {
	if rsrc == nil {
		return nil
	}
	return &pb.ResourceRef{
		Uid:           &rsrc.Uid,
		GroupName:     rsrc.ResourceGroupName,
		ExternalId:    RefOrNull(rsrc.Name),
		CloudPlatform: pb.CloudPlatform_GCP, // TODO: Change when we have more cloud platforms
	}
}
