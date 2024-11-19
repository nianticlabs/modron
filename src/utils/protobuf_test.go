package utils

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func TestTypeFromResource(t *testing.T) {
	tests := []struct {
		res  *pb.Resource
		want string
	}{
		{
			res:  &pb.Resource{Type: &pb.Resource_VmInstance{}},
			want: "VmInstance",
		},
		{
			res:  &pb.Resource{Type: &pb.Resource_ApiKey{}},
			want: "APIKey",
		},
		{
			res:  &pb.Resource{Type: &pb.Resource_ServiceAccount{}},
			want: "ServiceAccount",
		},
		{
			res:  &pb.Resource{Type: &pb.Resource_KubernetesCluster{}},
			want: "KubernetesCluster",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got, err := TypeFromResource(tt.res)
			if err != nil {
				t.Errorf("TypeFromResource() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("TypeFromResource() gotTy = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeFromResourceFail(t *testing.T) {
	tests := []struct {
		res     *pb.Resource
		wantErr string
	}{
		{
			res:     &pb.Resource{},
			wantErr: "cannot find field in oneof",
		},
		{
			res:     nil,
			wantErr: "resource must not be nil",
		},
	}
	for _, tt := range tests {
		t.Run(tt.wantErr, func(t *testing.T) {
			_, err := TypeFromResource(tt.res)
			if diff := cmp.Diff(tt.wantErr, err.Error()); diff != "" {
				t.Errorf("TypeFromResource() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestToAcceptedTypes(t *testing.T) {
	type args struct {
		types []proto.Message
	}
	tests := []struct {
		name    string
		args    args
		wantRes []string
	}{
		{
			name: "Three elements",
			args: args{
				types: []proto.Message{
					&pb.APIKey{},
					&pb.ServiceAccount{},
					&pb.Bucket{},
				},
			},
			wantRes: []string{
				"APIKey",
				"ServiceAccount",
				"Bucket",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := ProtoAcceptsTypes(tt.args.types); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("ProtoAcceptsTypes() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
