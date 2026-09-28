---
title: 'Authentication'
---

# Authentication

The library accepts credentials supplied by your application. Load tokens from a secret store or environment and never commit or log them.

## Use an access token

```go
client, err := alexa.NewClient(alexa.WithBearerToken(os.Getenv("ALEXA_BEARER_TOKEN")))
if err != nil {
    return err
}
defer client.Close()
```

The token must be an access token. `WithRegion` can select the configured US, EU, or JP endpoints. The sections below describe token refresh, cookies, and code-based linking.

## Refresh an access token

Create a client without an API token and call `RefreshAccessToken` with the refresh token and device registration configuration. Pass the returned access token to a new client using `WithBearerToken`. Keep both returned credentials private.

## Code-based linking

`GenerateCodePair` returns a public code and a private code. Show the public code to the user, keep the private code private, and call `RegisterWithCodePair` after user approval. The returned access and refresh tokens are credentials. This flow creates a linked device session and should only be started as an explicit user action.

`WithRefreshToken` is available for the cookie-backed REST requests. `WithCookies` and `WithCSRFToken` can be used when a caller already manages those values.
