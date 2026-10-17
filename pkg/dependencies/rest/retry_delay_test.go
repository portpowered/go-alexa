//nolint:testpackage // Exercises the private retry loop.
package rest

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

func TestRetryDelayIsAskedOncePerFailedAttemptBeforeTheLast(t *testing.T) {
	t.Parallel()

	var delays []int

	attempts := 0

	client := NewClient(WithBearerToken("synthetic-token"),
		WithRetryDelay(func(attempt int) time.Duration {
			delays = append(delays, attempt)

			return 0
		}),
		WithHTTPClient(&http.Client{
			Transport: syntheticRoundTripper(func(*http.Request) (*http.Response, error) {
				attempts++

				return nil, syntheticFailureError("synthetic transport failure")
			}),
		}))

	started := time.Now()

	response, err := client.doRequest(context.Background(), apiroutes.MethodGetUserInfo, apiroutes.PathGetUserInfo, nil)
	if response != nil {
		_ = response.Body.Close()
	}

	if err == nil {
		t.Fatal("expected the final transport failure")
	}

	if attempts != 3 || len(delays) != 2 || delays[0] != 0 || delays[1] != 1 {
		t.Fatalf("attempts=%d delays=%v, want 3 attempts and delays for attempts 0 and 1", attempts, delays)
	}

	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("zero retry delay still waited %v", elapsed)
	}
}

func TestDefaultRetryDelayWaitsOneThenTwoSeconds(t *testing.T) {
	t.Parallel()

	if defaultRetryDelay(0) != time.Second || defaultRetryDelay(1) != 2*time.Second {
		t.Fatalf("default retry delays changed: %v %v", defaultRetryDelay(0), defaultRetryDelay(1))
	}
}
