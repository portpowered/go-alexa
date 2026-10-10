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
go get github.com/portpowered/go-alexa@latest
```

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

Install the standalone CLI with `go install github.com/portpowered/go-alexa/cmd/go-alexa@v0.5.0`. See the [CLI guide](https://portpowered.github.io/go-alexa/docs/guides/cli/) for account linking, secure credential files, endpoint commands, player state, and event listening.

The CLI in this checkout also supports `alexa auth login` when built with the
executable name `alexa`. It uses Amazon CBL and stores private credentials under
the OS user configuration directory in `go-alexa/credentials.json`. Later commands
reuse that file automatically; `alexa auth refresh` renews it and `alexa auth logout`
removes it. `GO_ALEXA_CREDENTIALS` overrides the saved credential path. See the
[local CLI guide](docs/guides/cli.mdx) for building against this checkout.

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

The package exposes endpoint, feature, event, request, response, and error models under `pkg/alexaapimodels`. See the [customer guides](https://portpowered.github.io/go-alexa/docs/guides/) and [API reference](https://portpowered.github.io/go-alexa/docs/).

## Authentication, errors, and transports

Create a reusable client with `WithRegion`, `WithTimeout`, endpoint overrides, and network options. Put account credentials on `client.NewSession(alexa.WithBearerToken(...))`. `WithRefreshToken` only stores a caller-managed token: refresh it explicitly and install the returned access token with `Session.SetAccessToken`, or explicitly exchange it for cookies. `WithRESTHTTPClient`, `WithGraphQLHTTPClient`, and `WithEventHTTPClient` inject separate network edges; `WithEventTransport` accepts an event-stream `RoundTripper` such as an HTTP/2 transport. See the [transport configuration example](https://portpowered.github.io/go-alexa/docs/guides/client-configuration/).

Errors use typed values such as `AuthenticationError`, `NetworkError`, `TokenError`, `BadRequestError`, and `HTTPError` in `alexaapimodels`; many wrapped errors preserve their cause with `Unwrap`. `ControlResponse`, `SubscribeResponse`, and `QualityOfServiceResponse` can also contain operation-level errors even when the HTTP request succeeded.

## Examples

Runnable examples are in `cmd/examples/`:

- `cbl-auth` demonstrates code-based linking and token refresh. It creates a linked device and requires interactive approval.
- `endpoint-control` lists endpoints and sends a power-off request to the first endpoint with a power feature.
- `events` creates an endpoint subscription and listens for events.

The examples use environment variables where credentials are needed. No account tokens or endpoint identifiers are hard-coded.

## License

Apache-2.0. See [LICENSE](LICENSE).
