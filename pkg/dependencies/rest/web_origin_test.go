//nolint:testpackage // Exercises private cookie-auth header construction.
package rest

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type expectedWebOrigin struct {
	region          alexaapimodels.Region
	origin          string
	referer         string
	cookieDomain    string
	cookieSubdomain string
}

func regionalWebOrigins() []expectedWebOrigin {
	return []expectedWebOrigin{
		{
			region:          alexaapimodels.RegionUS,
			origin:          "https://alexa.amazon.com",
			referer:         "https://alexa.amazon.com/spa/index.html",
			cookieDomain:    "alexa.amazon.com",
			cookieSubdomain: ".alexa.amazon.com",
		},
		{
			region:          alexaapimodels.RegionEU,
			origin:          "https://alexa.amazon.co.uk",
			referer:         "https://alexa.amazon.co.uk/spa/index.html",
			cookieDomain:    "alexa.amazon.co.uk",
			cookieSubdomain: ".alexa.amazon.co.uk",
		},
		{
			region:          alexaapimodels.RegionJP,
			origin:          "https://alexa.amazon.co.jp",
			referer:         "https://alexa.amazon.co.jp/spa/index.html",
			cookieDomain:    "alexa.amazon.co.jp",
			cookieSubdomain: ".alexa.amazon.co.jp",
		},
	}
}

func assertWebOriginHeaders(t *testing.T, request *http.Request, want expectedWebOrigin) {
	t.Helper()

	if got := request.Header.Get("Origin"); got != want.origin {
		t.Errorf("%s %s: Origin = %q, want %q", want.region, request.URL.Path, got, want.origin)
	}

	if got := request.Header.Get("Referer"); got != want.referer {
		t.Errorf("%s %s: Referer = %q, want %q", want.region, request.URL.Path, got, want.referer)
	}
}

func TestCookieAuthHeadersUseRegionalWebOrigin(t *testing.T) {
	t.Parallel()

	for _, want := range regionalWebOrigins() {
		t.Run(string(want.region), func(t *testing.T) {
			t.Parallel()

			calls := 0
			client := NewClient(
				WithRegion(want.region),
				WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
					calls++
					assertWebOriginHeaders(t, request, want)

					return syntheticResponse(request, http.StatusOK, `{"id":"synthetic-user"}`), nil
				})}),
				WithCookies(map[string]*http.Cookie{"session-id": {Name: "session-id", Value: "synthetic-session"}}),
				WithCSRFToken("synthetic-csrf"),
			)

			_, err := client.GetUserInfo(context.Background(), nil)
			if err != nil {
				t.Fatalf("GetUserInfo returned an error: %v", err)
			}

			if calls != 1 {
				t.Fatalf("expected one request, got %d", calls)
			}
		})
	}
}

func TestCSRFFetchUsesRegionalWebOrigin(t *testing.T) {
	t.Parallel()

	for _, want := range regionalWebOrigins() {
		t.Run(string(want.region), func(t *testing.T) {
			t.Parallel()

			client := NewClient(
				WithRegion(want.region),
				WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
					if request.URL.Host != want.cookieDomain {
						t.Errorf("CSRF fetch host = %q, want %q", request.URL.Host, want.cookieDomain)
					}

					assertWebOriginHeaders(t, request, want)

					response := syntheticResponse(request, http.StatusOK, `{}`)
					response.Header.Add("Set-Cookie", "csrf=fetched-"+string(want.region)+"; Path=/")

					return response, nil
				})}),
				WithCookies(map[string]*http.Cookie{"session-id": {Name: "session-id", Value: "synthetic-session"}}),
			)

			csrf, err := client.GetCSRFToken(context.Background())
			if err != nil || csrf != "fetched-"+string(want.region) {
				t.Fatalf("GetCSRFToken returned %q, %v", csrf, err)
			}
		})
	}
}

// csrfKeyedCookies stores the CSRF value only under a domain-qualified key, so the
// lookup must match the key against the client's regional cookie domain.
func csrfKeyedCookies(key, value string) map[string]*http.Cookie {
	return map[string]*http.Cookie{
		"session-id": {Name: "session-id", Value: "synthetic-session"},
		key:          {Name: "", Value: value},
	}
}

func TestCSRFCookieLookupUsesRegionalCookieDomain(t *testing.T) {
	t.Parallel()

	for _, want := range regionalWebOrigins() {
		for _, domain := range []string{want.cookieDomain, want.cookieSubdomain} {
			t.Run(string(want.region)+" "+domain, func(t *testing.T) {
				t.Parallel()

				key := domain + ":csrf"

				csrfClient := NewClient(WithRegion(want.region), WithCookies(csrfKeyedCookies(key, "keyed-csrf")),
					WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
						t.Errorf("GetCSRFToken made an unexpected request to %s", request.URL)

						return syntheticResponse(request, http.StatusInternalServerError, `{}`), nil
					})}))

				csrf, err := csrfClient.GetCSRFToken(context.Background())
				if err != nil || csrf != "keyed-csrf" {
					t.Fatalf("GetCSRFToken with %q returned %q, %v", key, csrf, err)
				}

				requestClient := NewClient(WithRegion(want.region), WithCookies(csrfKeyedCookies(key, "keyed-csrf")),
					WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
						if got := request.Header.Get("Csrf"); got != "keyed-csrf" {
							t.Errorf("Csrf header = %q, want keyed-csrf", got)
						}

						return syntheticResponse(request, http.StatusOK, `{"id":"synthetic-user"}`), nil
					})}))

				_, err = requestClient.GetUserInfo(context.Background(), nil)
				if err != nil {
					t.Fatalf("GetUserInfo with %q returned an error: %v", key, err)
				}
			})
		}
	}
}

func TestCSRFCookieLookupIgnoresOtherRegionDomains(t *testing.T) {
	t.Parallel()

	for _, client := range regionalWebOrigins() {
		for _, other := range regionalWebOrigins() {
			if other.region == client.region {
				continue
			}

			key := other.cookieDomain + ":csrf"
			restClient := NewClient(WithRegion(client.region), WithCookies(csrfKeyedCookies(key, "foreign-csrf")))

			_, err := restClient.cachedCSRFToken()
			if !alexaapimodels.IsTokenError(err) {
				t.Errorf("%s client accepted CSRF cookie keyed for %s: %v", client.region, other.cookieDomain, err)
			}
		}
	}
}

func TestConcurrentRegionalClientsKeepTheirOwnWebOrigin(t *testing.T) {
	t.Parallel()

	const requestsPerClient = 25

	var requests sync.WaitGroup

	for _, want := range regionalWebOrigins() {
		client := NewClient(
			WithRegion(want.region),
			WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
				assertWebOriginHeaders(t, request, want)

				if request.URL.Path == "/api/language" {
					response := syntheticResponse(request, http.StatusOK, `{}`)
					response.Header.Add("Set-Cookie", "csrf=fetched; Path=/")

					return response, nil
				}

				return syntheticResponse(request, http.StatusOK, `{"id":"synthetic-user"}`), nil
			})}),
			WithCookies(csrfKeyedCookies(want.cookieDomain+":csrf", "keyed-csrf")),
		)
		fetchClient := NewClient(
			WithRegion(want.region),
			WithHTTPClient(client.httpClient),
			WithCookies(map[string]*http.Cookie{"session-id": {Name: "session-id", Value: "synthetic-session"}}),
		)

		requests.Add(2)

		go func() {
			defer requests.Done()

			for range requestsPerClient {
				_, err := client.GetUserInfo(context.Background(), nil)
				if err != nil {
					t.Errorf("%s GetUserInfo returned an error: %v", want.region, err)
				}
			}
		}()

		go func() {
			defer requests.Done()

			for range requestsPerClient {
				_, err := fetchClient.fetchCSRFTokenFromAPI(context.Background())
				if err != nil {
					t.Errorf("%s CSRF fetch returned an error: %v", want.region, err)
				}
			}
		}()
	}

	requests.Wait()
}
