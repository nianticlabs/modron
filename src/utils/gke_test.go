package utils

import "testing"

func TestGetGKEReference(t *testing.T) {
	tc := [][]string{
		{
			"//container.googleapis.com/projects/project-id/zones/us-central1-b/clusters/gke-cluster-name/k8s/namespaces/kubernetes-ns-name",
			"project-id",
			"us-central1-b",
			"gke-cluster-name",
			"kubernetes-ns-name",
		},
	}

	for _, tt := range tc {
		projectID, location, clusterName, namespace := GetGKEReference(tt[0])
		if projectID != tt[1] {
			t.Errorf("expected projectID %s, got %s", tt[1], projectID)
		}
		if location != tt[2] {
			t.Errorf("expected location %s, got %s", tt[2], location)
		}
		if clusterName != tt[3] {
			t.Errorf("expected clusterName %s, got %s", tt[3], clusterName)
		}
		if namespace != tt[4] {
			t.Errorf("expected namespace %s, got %s", tt[4], namespace)
		}
	}
}
