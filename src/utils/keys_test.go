package utils_test

import (
	"testing"

	"github.com/nianticlabs/modron/src/utils"
)

func TestGetKeyID(t *testing.T) {
	tests := []struct {
		name     string
		keyRef   string
		expected string
	}{
		{
			name:     "valid key reference",
			keyRef:   "projects/my-project/serviceAccounts/sa-1/keys/abc",
			expected: "abc",
		},
		{
			name:     "invalid key reference",
			keyRef:   "invalid-key-reference",
			expected: "invalid-key-reference",
		},
		{
			name:     "key reference with less than 6 parts",
			keyRef:   "projects/my-project/serviceAccounts/sa-1/keys",
			expected: "projects/my-project/serviceAccounts/sa-1/keys",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.GetKeyID(tt.keyRef); got != tt.expected {
				t.Errorf("GetKeyID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetServiceAccountNameFromKeyRef(t *testing.T) {
	tests := []struct {
		name     string
		keyRef   string
		expected string
	}{
		{
			name:     "valid key reference",
			keyRef:   "projects/my-project/serviceAccounts/sa-1/keys/abc",
			expected: "sa-1",
		},
		{
			name:     "invalid key reference",
			keyRef:   "invalid-key-reference",
			expected: "invalid-key-reference",
		},
		{
			name:     "key reference with less than 6 parts",
			keyRef:   "projects/my-project/serviceAccounts/sa-1/keys",
			expected: "projects/my-project/serviceAccounts/sa-1/keys",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.GetServiceAccountNameFromKeyRef(tt.keyRef); got != tt.expected {
				t.Errorf("GetServiceAccountNameFromKeyRef() = %v, want %v", got, tt.expected)
			}
		})
	}
}
