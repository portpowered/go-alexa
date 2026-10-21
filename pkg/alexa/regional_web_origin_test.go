//nolint:testpackage // Shares the package's synthetic transport helpers.
package alexa

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestConcurrentRegionalSessionsSendTheirOwnWebOrigin(t *testing.T) {
	t.Parallel()

	regions := []struct {
		region alexaapimodels.Region
		host   string
	}{
		{alexaapimodels.RegionUS, "alexa.amazon.com"},
		{alexaapimodels.RegionEU, "alexa.amazon.co.uk"},
		{alexaapimodels.RegionJP, "alexa.amazon.co.jp"},
	}

	const requestsPerSession = 20

	var requests sync.WaitGroup

	for _, want := range regions {
		transport := responseTransport(func(request *http.Request) (*http.Response, error) {
			origin := "https://" + want.host
			if request.URL.Host != want.host ||
				request.Header.Get("Origin") != origin ||
				request.Header.Get("Referer") != origin+"/spa/index.html" {
				t.Errorf("%s request to %s sent Origin %q and Referer %q", want.region, request.URL.Host,
					request.Header.Get("Origin"), request.Header.Get("Referer"))
			}

			return syntheticResponse(request, `{"id":"synthetic-user"}`, nil), nil
		})

		client, err := NewClient(WithRegion(want.region), WithRESTHTTPClient(&http.Client{Transport: transport}))
		if err != nil {
			t.Fatalf("NewClient(%s) returned an error: %v", want.region, err)
		}

		session, err := client.NewSession(
			WithCookies(map[string]*http.Cookie{"session-id": {Name: "session-id", Value: "synthetic-session"}}),
			WithCSRFToken("synthetic-csrf"),
		)
		if err != nil {
			t.Fatalf("NewSession(%s) returned an error: %v", want.region, err)
		}

		t.Cleanup(func() { closeTestSession(t, session) })

		requests.Add(1)

		go func() {
			defer requests.Done()

			request := alexaapimodels.UserInfoRequest{Platform: "synthetic-app", Version: "1.2.3", CSRFToken: ""}
			for range requestsPerSession {
				_, err := session.GetUserInfo(context.Background(), request)
				if err != nil {
					t.Errorf("%s GetUserInfo returned an error: %v", want.region, err)
				}
			}
		}()
	}

	requests.Wait()
}
