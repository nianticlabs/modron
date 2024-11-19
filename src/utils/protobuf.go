package utils

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func TypeFromResource(rsrc *pb.Resource) (ty string, err error) {
	if rsrc == nil {
		return "", fmt.Errorf("resource must not be nil")
	}
	reflectMsg := rsrc.ProtoReflect()
	if reflectMsg == nil {
		return "", fmt.Errorf("ProtoReflect() returned nil")
	}
	typeField := reflectMsg.Descriptor().Oneofs().ByName("type")
	if typeField == nil {
		return "", fmt.Errorf("cannot find field \"type\"")
	}
	field := reflectMsg.WhichOneof(typeField)
	if field == nil {
		return "", fmt.Errorf("cannot find field in oneof")
	}
	fieldMessage := reflectMsg.Get(field).Message()
	if fieldMessage == nil {
		return "", fmt.Errorf("field message is nil")
	}
	ty = string(fieldMessage.Descriptor().FullName())
	return
}

func ProtoAcceptsTypes(types []proto.Message) (res []string) {
	for _, t := range types {
		res = append(res, string(t.ProtoReflect().Descriptor().FullName()))
	}
	return
}
