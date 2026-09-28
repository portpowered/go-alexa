---
title: 'Authentication'
---

# Authentication

`Client` stores endpoint and network configuration. `Session` stores credentials for one account. Load credentials from a secret store or environment and never commit or log them.

## Use an access token

```go
client, err := alexa.NewClient(alexa.WithRegion(alexaapimodels.RegionUS))
if err != nil {
    return err
}
session, err := client.NewSession(alexa.WithBearerToken(os.Getenv("ALEXA_BEARER_TOKEN")))
if err != nil {
    return err
}
defer session.Close()
```

The token must be an access token. `WithRegion` can select the configured US, EU, or JP endpoints. The sections below describe token refresh, cookies, and code-based linking.

## Refresh an access token

Create a session with `WithRefreshToken` and call `Session.RefreshAccessToken` explicitly. It returns a token response and does not update the session. Store the result in your credential store, then call `Session.SetAccessToken` or create a new session with `WithBearerToken`. `Session.Credentials` returns a copy of the configured credentials.

## Code-based linking

`GenerateCodePair` returns a public code and a private code. Show the public code to the user, keep the private code private, and call `RegisterWithCodePair` after user approval. The returned access and refresh tokens are credentials. This flow creates a linked device session and should only be started as an explicit user action.

For cookie-based requests, explicitly call `Session.ExchangeRefreshTokenForCookies`, store the returned cookies, and use `WithCookies`. Call `Session.GetCSRFToken` to retrieve CSRF material when needed. Requests do not exchange refresh tokens or fetch missing CSRF values implicitly. `WithCSRFToken` accepts a CSRF token already managed by your application.
