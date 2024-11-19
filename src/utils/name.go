package utils

import "strings"

const (
	containerAPI = "//container.googleapis.com/"
	iamAPI       = "//iam.googleapis.com/"
	computeAPI   = "//compute.googleapis.com/"
)

// GetHumanReadableName returns the human-readable name of the resource - we currently use an allow-list of APIs
// to avoid interpreting the resource name incorrectly.
func GetHumanReadableName(resourceLink string) string {
	for _, p := range []string{containerAPI, iamAPI, computeAPI} {
		if strings.HasPrefix(resourceLink, p) {
			r := strings.TrimPrefix(resourceLink, p)
			split := strings.Split(r, "/")
			return split[len(split)-1]
		}
	}
	return resourceLink
}

func StripProjectsPrefix(prefixedProject string) string {
	return strings.TrimPrefix(prefixedProject, "projects/")
}
