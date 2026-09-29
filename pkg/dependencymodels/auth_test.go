package alexamodels_test

import (
	"testing"
	"time"

	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func TestToken_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		token    *alexamodels.Token
		expected bool
	}{
		{
			name: "not expired",
			token: &alexamodels.Token{
				ExpiresAt: time.Now().Add(10 * time.Minute),
			},
			expected: false,
		},
		{
			name: "expired",
			token: &alexamodels.Token{
				ExpiresAt: time.Now().Add(-10 * time.Minute),
			},
			expected: true,
		},
		{
			name: "expires soon",
			token: &alexamodels.Token{
				ExpiresAt: time.Now().Add(30 * time.Second),
			},
			expected: true,
		},
		{
			name: "zero time",
			token: &alexamodels.Token{
				ExpiresAt: time.Time{},
			},
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.token.IsExpired()
			if result != testCase.expected {
				t.Errorf("IsExpired() = %v, want %v", result, testCase.expected)
			}
		})
	}
}
