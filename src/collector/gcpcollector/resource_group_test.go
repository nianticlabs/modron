package gcpcollector

import "testing"

func TestGetResourceGroupLink(t *testing.T) {
	tc := []struct {
		name     string
		expected string
	}{
		{
			name:     "projects/project-1",
			expected: "https://console.cloud.google.com/welcome?project=project-1",
		},
		{
			name:     "folders/1000000",
			expected: "https://console.cloud.google.com/welcome?folder=1000000",
		},
		{
			name:     "organizations/1000000",
			expected: "https://console.cloud.google.com/welcome?organizationId=1000000",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			if got := getResourceGroupLink(tt.name); got != tt.expected {
				t.Errorf("GetResourceGroupLink() = %v, want %v", got, tt.expected)
			}
		})
	}
}
