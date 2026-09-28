package alexamodels

import (
	"testing"
	"time"
)

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		token    *Token
		expected bool
	}{
		{
			name: "not expired",
			token: &Token{
				ExpiresAt: time.Now().Add(10 * time.Minute),
			},
			expected: false,
		},
		{
			name: "expired",
			token: &Token{
				ExpiresAt: time.Now().Add(-10 * time.Minute),
			},
			expected: true,
		},
		{
			name: "expires soon",
			token: &Token{
				ExpiresAt: time.Now().Add(30 * time.Second),
			},
			expected: true,
		},
		{
			name: "zero time",
			token: &Token{
				ExpiresAt: time.Time{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsExpired()
			if result != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", result, tt.expected)
			}
		})
	}
}
