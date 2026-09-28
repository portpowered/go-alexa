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
go get github.com/portpowered/go-alexa@main
```

Use `@main` until a new release tag is published after the history cleanup.

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

	client, err := alexa.NewClient(alexa.WithBearerToken(token))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.ListEndpoints(ctx, alexaapimodels.EndpointQuery{
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

`NewClient` accepts options, and every network operation accepts a context. Call `Close` when finished. See [Authentication](docs/guides/authentication.md) for token refresh and code-based linking.

## Supported operations

| Operation | Client method | Notes |
|---|---|---|
| Code-based linking | `GenerateCodePair`, `RegisterWithCodePair` | User approval is required to complete linking. Treat the private code and returned tokens as credentials. |
| Token refresh | `RefreshAccessToken` | Uses a refresh token and device registration configuration. |
| Endpoint enumeration | `ListEndpoints` | Combines a GraphQL endpoint result with REST device metadata; some fields may be absent. |
| Endpoint control | `Control` | Submits one supported feature operation. The response is not a guarantee that the physical device changed state. |
| Quality of service | `RequestEndpointQualityOfService` | Requests cloud-side state polling; returned state can still be delayed or inaccurate. |
| Event subscription and stream | `Subscribe`, `ConnectEvents` | Create a subscription before opening the HTTP/2 connection. |
| Account profile | `GetUserInfo` | Returns personal account information; handle the result as sensitive data. |
| Player state | `GetPlayerState` | Provider- and device-dependent. |

The package also exposes endpoint, feature, event, request, response, and error models under `pkg/alexaapimodels`. The checked-in GraphQL, OpenAPI, and AsyncAPI documents describe the subset used by this client and drive generated client or internal wire models. They are implementation-derived, not complete or authoritative Alexa contracts; REST and event-stream shapes are not supported by sanitized live captures in this repository. See [API coverage](docs/alexa_apis.md) and [verification guidance](docs/verification.md).

## Authentication, errors, and transports

Use `WithBearerToken` with an access token, or `WithRefreshToken` for the cookie-backed REST flows supported by the client. `WithRegion` selects the configured US, EU, or JP service endpoints. `WithHttpClient` injects an `*http.Client` for REST and GraphQL requests. The HTTP/2 event connection has its own transport setup.

Errors use typed values such as `AuthenticationError`, `NetworkError`, `TokenError`, `BadRequestError`, and `HTTPError` in `alexaapimodels`; many wrapped errors preserve their cause with `Unwrap`. `ControlResponse`, `SubscribeResponse`, and `QualityOfServiceResponse` can also contain operation-level errors even when the HTTP request succeeded.

See the [API guide](docs/alexa_apis.md), [authentication guide](docs/guides/authentication.md), and [published reference](https://portpowered.github.io/go-alexa/).

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
