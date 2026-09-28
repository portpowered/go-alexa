# Authentication

`Client` contains service and transport configuration. `Session` contains account credentials and can be closed independently. Supply credentials obtained by your application, keep them in a secret store, and never commit or log access tokens, refresh tokens, private code-pair values, cookies, or customer identifiers.

## Access token

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

`WithBearerToken` accepts an access token. The session uses it for REST, GraphQL, and event-stream requests. `session.Credentials()` returns a copy of the currently configured credentials.

## Refresh token

Store a refresh token on the session, then call the explicit refresh method. It returns new credentials without replacing session state; store the result in your own credential store and install the access token with `SetAccessToken`:

```go
refreshClient, err := alexa.NewClient()
if err != nil {
    return err
}
refreshSession, err := refreshClient.NewSession(alexa.WithRefreshToken(os.Getenv("ALEXA_REFRESH_TOKEN")))
if err != nil {
    return err
}
defer refreshSession.Close()

config := alexaapimodels.DefaultDeviceRegistrationConfig("your-device-serial", "your-device-name")
result, err := refreshSession.RefreshAccessToken(ctx, config)
if err != nil {
    return err
}
if err := refreshSession.SetAccessToken(result.AccessToken); err != nil {
    return err
}
```

The configuration fields identify the client device session. Use values appropriate to your integration and protect the returned tokens like passwords.

For cookie authentication, call `ExchangeRefreshTokenForCookies` explicitly, store its result, and create a session with `WithCookies`. If a CSRF token is needed, retrieve it explicitly with `GetCSRFToken`; requests never exchange refresh tokens or fetch missing CSRF material implicitly.

## Code-based linking

`GenerateCodePair` starts a linking flow. Show the public code to the user and direct them to Amazon's code page; keep the private code private. After the user approves the link, pass both values to `RegisterWithCodePair`. The returned access and refresh tokens are credentials. The `cmd/examples/cbl-auth` program demonstrates this interactive sequence without printing private codes or token values.

## Cookies and CSRF

`WithCookies` and `WithCSRFToken` are available for callers that already obtained a session. Cookie and CSRF values are credentials and must be redacted from logs and any network captures.

## Regions and errors

`WithRegion` accepts `RegionUS`, `RegionEU`, or `RegionJP`. Methods return typed errors from `alexaapimodels`, including authentication, token, network, connection, bad-request, and HTTP errors. Wrapped errors preserve their cause where supported; use `errors.As` when you need to inspect a typed error.
