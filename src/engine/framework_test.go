package engine

import (
	"testing"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/pb"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	structpb "google.golang.org/protobuf/types/known/structpb"
)

var resources = []*pb.Resource{
	{
		Name:   "project-0",
		Parent: "",
		IamPolicy: &pb.IamPolicy{
			Resource: nil,
			Permissions: []*pb.Permission{
				{
					Role: "roles/iam.serviceAccountAdmin",
					Principals: []string{
						"serviceAccount:account-0",
					},
				},
				{
					Role: "roles/dataflow.admin",
					Principals: []string{
						"serviceAccount:account-3",
					},
				},
			},
		},
		Type: &pb.Resource_ResourceGroup{
			ResourceGroup: &pb.ResourceGroup{},
		},
	},
	{
		Name:   "project-1",
		Parent: "",
		IamPolicy: &pb.IamPolicy{
			Resource: nil,
			Permissions: []*pb.Permission{
				{
					Role: "roles/iam.serviceAccountUser",
					Principals: []string{
						"serviceAccount:account-1",
						"serviceAccount:account-2",
					},
				},
			},
		},
		Type: &pb.Resource_ResourceGroup{
			ResourceGroup: &pb.ResourceGroup{},
		},
	},
}

func TestResourceFromStructValue(t *testing.T) {
	value1, err := structpb.NewValue(map[string]interface{}{
		"name": "project-0",
		"iamPolicy": map[string]interface{}{
			"resource": map[string]interface{}{},
			"permissions": []interface{}{
				map[string]interface{}{
					"role": "roles/iam.serviceAccountAdmin",
					"principals": []interface{}{
						"serviceAccount:account-0",
					},
				},
				map[string]interface{}{
					"role": "roles/dataflow.admin",
					"principals": []interface{}{
						"serviceAccount:account-3",
					},
				},
			},
		},
		"resourceGroup": map[string]interface{}{},
	})

	if err != nil {
		t.Errorf(`structpb.NewValue unexpected error: "%v"`, err)
	}

	value2, err := structpb.NewValue(map[string]interface{}{
		"name": "project-1",
		"iamPolicy": map[string]interface{}{
			"resource": map[string]interface{}{},
			"permissions": []interface{}{
				map[string]interface{}{
					"role": "roles/iam.serviceAccountUser",
					"principals": []interface{}{
						"serviceAccount:account-1",
						"serviceAccount:account-2",
					},
				},
			},
		},
		"resourceGroup": map[string]interface{}{},
	})

	if err != nil {
		t.Errorf(`structpb.NewValue unexpected error: "%v"`, err)
	}

	values := []*structpb.Value{value1, value2}

	for _, want := range values {
		rsrc, err := common.ResourceFromStructValue(want)
		if err != nil {
			t.Errorf(`ResourceFromStructValue unexpected error: "%v"`, err)
		}

		got, err := common.StructValueFromResource(rsrc)
		if err != nil {
			t.Errorf(`StructValueFromResource unexpected error: "%v"`, err)
		}
		if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
			t.Errorf("CheckRule unexpected diff (-want, +got): %v", diff)
		}
	}
}

func TestStructValueFromResource(t *testing.T) {
	for _, want := range resources {
		value, err := common.StructValueFromResource(want)
		if err != nil {
			t.Errorf(`StructValueFromResource unexpected error: "%v"`, err)
		}

		got, err := common.ResourceFromStructValue(value)
		if err != nil {
			t.Errorf(`GetResourceFromStructValue unexpected error: "%v"`, err)
		}
		if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
			t.Errorf("CheckRule unexpected diff (-want, +got): %v", diff)
		}
	}
}
