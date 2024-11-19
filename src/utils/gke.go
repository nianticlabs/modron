package utils

import (
	"fmt"
	"strings"
)

func GetGKEReference(resourceLink string) (projectID string, location string, clusterName string, namespace string) {
	// resourceLink is formatted as follows:
	// //container.googleapis.com/projects/project-id/zones/us-central1-b/clusters/gke-cluster-name/k8s/namespaces/kubernetes-ns-name
	split := strings.Split(resourceLink, "/")
	if len(split) != 12 { //nolint:mnd
		return "", "", "", ""
	}
	projectID = split[4]
	location = split[6]
	clusterName = split[8]
	namespace = split[11]
	return
}

func GetGkePodLink(name string, parent string) string {
	// name is like "my-pod-name"
	// parent is "//container.googleapis.com/projects/project-id/zones/us-central1-b/clusters/gke-cluster-name/k8s/namespaces/kubernetes-ns-name"
	projectID, location, clusterName, namespace := GetGKEReference(parent)
	return fmt.Sprintf("https://console.cloud.google.com/kubernetes/pod/%s/%s/%s/%s/details?project=%s",
		location,
		clusterName,
		namespace,
		name,
		projectID,
	)
}
