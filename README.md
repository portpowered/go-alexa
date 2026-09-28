# go-alexa

`go-alexa` is a Go client for a subset of Amazon Alexa account and endpoint APIs. It provides code-based linking and token refresh, endpoint enumeration, endpoint control, player state, quality-of-service requests, subscriptions, and an HTTP/2 event connection. These are provider-specific APIs and may change without notice.

[![Go version](https://img.shields.io/github/go-mod/go-version/portpowered/go-alexa)](go.mod)
[![CI](https://github.com/portpowered/go-alexa/actions/workflows/ci.yml/badge.svg)](https://github.com/portpowered/go-alexa/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https%3A%2F%2Fportpowered.github.io%2Fgo-alexa%2Fcoverage.json)](https://portpowered.github.io/go-alexa/coverage.html)
[![Latest release](https://img.shields.io/github/v/release/portpowered/go-alexa)](https://github.com/portpowered/go-alexa/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/portpowered/go-alexa/pkg/alexa.svg)](https://pkg.go.dev/github.com/portpowered/go-alexa/pkg/alexa)
[![License](https://img.shields.io/github/license/portpowered/go-alexa)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-GitHub%20Pages-blue)](https://portpowered.github.io/go-alexa/)

## Install

```sh
go get github.com/portpowered/go-alexa@v0.2.0
```

`v0.1.0` was the first release from the cleaned history. The v0.2.0 API moves
account operations from `Client` to `Session`; see the [authentication guide](https://portpowered.github.io/go-alexa/docs/guides/authentication/)
for the migration pattern.

## Quick start

Provide an access token through your application's secret store or environment. Do not commit tokens or print them to logs.

```go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func main() {
	token := os.Getenv("ALEXA_BEARER_TOKEN")
	if token == "" {
		log.Fatal("set ALEXA_BEARER_TOKEN")
	}

	client, err := alexa.NewClient(alexa.WithRegion(alexaapimodels.RegionUS))
	if err != nil {
		log.Fatal(err)
	}
	session, err := client.NewSession(alexa.WithBearerToken(token))
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := session.ListEndpoints(ctx, alexaapimodels.EndpointQuery{
		IncludeFields: &alexaapimodels.EndpointIncludeFields{
			Properties: true,
			Features:   true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("found %d endpoints", len(response.Results))
}
```

`Client` holds reusable endpoint and transport configuration. Create an account `Session` with credentials before calling API methods, then close the session when finished. Every network operation accepts a context. See the [authentication guide](https://portpowered.github.io/go-alexa/docs/guides/authentication/) for token refresh and code-based linking.

## Supported operations

| Operation | Method | Notes |
|---|---|---|
| Code-based linking | `GenerateCodePair`, `RegisterWithCodePair` | User approval is required to complete linking. Treat the private code and returned tokens as credentials. |
| Token refresh | `RefreshAccessToken` | Uses a refresh token and device registration configuration. |
| Endpoint enumeration | `ListEndpoints` | Combines a GraphQL endpoint result with REST device metadata; some fields may be absent. |
| Endpoint control | `Control` | Submits one supported feature operation. The response is not a guarantee that the physical device changed state. |
| Quality of service | `RequestEndpointQualityOfService` | Requests cloud-side state polling; returned state can still be delayed or inaccurate. |
| Event subscription and stream | `Subscribe`, `ConnectEvents` | Create a subscription before opening the HTTP/2 connection. |
| Account profile | `GetUserInfo` | Returns personal account information; handle the result as sensitive data. |
| Player state | `GetPlayerState` | Provider- and device-dependent. |

The package exposes endpoint, feature, event, request, response, and error models under `pkg/alexaapimodels`. Its GraphQL, OpenAPI, and AsyncAPI files describe implementation-derived subsets, not authoritative Amazon contracts. The maintainer reports account testing, but operation-level results and sanitized captures are not recorded. See the [generated API reference](https://portpowered.github.io/go-alexa/docs/) and [verification guidance](docs/verification.md).

## Authentication, errors, and transports

Create a reusable client with `WithRegion`, `WithTimeout`, endpoint overrides, and network options. Put account credentials on `client.NewSession(alexa.WithBearerToken(...))`. `WithRefreshToken` only stores a caller-managed token: refresh it explicitly and install the returned access token with `Session.SetAccessToken`, or explicitly exchange it for cookies. `WithRESTHTTPClient`, `WithGraphQLHTTPClient`, and `WithEventHTTPClient` inject separate network edges; `WithEventTransport` accepts an event-stream `RoundTripper` such as an HTTP/2 transport.

Errors use typed values such as `AuthenticationError`, `NetworkError`, `TokenError`, `BadRequestError`, and `HTTPError` in `alexaapimodels`; many wrapped errors preserve their cause with `Unwrap`. `ControlResponse`, `SubscribeResponse`, and `QualityOfServiceResponse` can also contain operation-level errors even when the HTTP request succeeded.

## Examples

Runnable examples are in `cmd/examples/`:

- `cbl-auth` demonstrates code-based linking and token refresh. It creates a linked device and requires interactive approval.
- `endpoint-control` lists endpoints and sends a power-off request to the first endpoint with a power feature.
- `events` creates an endpoint subscription and listens for events.

The examples use environment variables where credentials are needed. No account tokens or endpoint identifiers are hard-coded.

## Verification

Normal checks are offline and use synthetic fixtures; no Amazon credentials are required:

```sh
make lint
make check
make generate
```

`make check` runs vet, build, and race-enabled tests. The synthetic fixtures define test inputs and are not proof of current service behavior. Live integration tests in `test/integration` run only when `ALEXA_REFRESH_TOKEN` is set; otherwise they skip:

```sh
go test ./test/integration/...
```

Read [fixture guidance](docs/verification.md) before adding any captured payload and the [release guide](docs/releasing.md) before tagging. Never check in raw captures, tokens, cookies, customer identifiers, or device identifiers.

## License

Apache-2.0. See [LICENSE](LICENSE).
