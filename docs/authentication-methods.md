# Authentication

The client does not contain credentials. Supply a token obtained by your application, and keep it in a secret store or environment variable. Never commit access tokens, refresh tokens, private code-pair values, cookies, or customer identifiers.

## Access token

```go
client, err := alexa.NewClient(
    alexa.WithBearerToken(os.Getenv("ALEXA_BEARER_TOKEN")),
    alexa.WithRegion(alexaapimodels.RegionUS),
)
if err != nil {
    return err
}
defer client.Close()
```

`WithBearerToken` accepts an access token. It configures REST and GraphQL calls and is also used to authenticate the event stream.

## Refresh token

`WithRefreshToken` configures the cookie-backed REST requests that exchange a refresh token for session cookies. For an access token, call `RefreshAccessToken` with a refresh token and a registration configuration, then create a client with the returned access token:

```go
refreshClient, err := alexa.NewClient()
if err != nil {
    return err
}
defer refreshClient.Close()

config := alexaapimodels.DefaultDeviceRegistrationConfig("your-device-serial", "your-device-name")
result, err := refreshClient.RefreshAccessToken(ctx, alexaapimodels.TokenRefreshRequest{
    RefreshToken: os.Getenv("ALEXA_REFRESH_TOKEN"),
    Config:       config,
})
if err != nil {
    return err
}

apiClient, err := alexa.NewClient(alexa.WithBearerToken(result.AccessToken))
if err != nil {
    return err
}
defer apiClient.Close()
```

The configuration fields identify the client device session. Use values appropriate to your integration and protect the returned tokens like passwords.

## Code-based linking

`GenerateCodePair` starts a linking flow. Show the public code to the user and direct them to Amazon's code page; keep the private code private. After the user approves the link, pass both values to `RegisterWithCodePair`. The returned access and refresh tokens are credentials. The `cmd/examples/cbl-auth` program demonstrates this interactive sequence without printing private codes or token values.

## Cookies and CSRF

`WithCookies` and `WithCSRFToken` are available for callers that already obtained a session. Cookie and CSRF values are credentials and must be redacted from logs and any network captures.

## Regions and errors

`WithRegion` accepts `RegionUS`, `RegionEU`, or `RegionJP`. Methods return typed errors from `alexaapimodels`, including authentication, token, network, connection, bad-request, and HTTP errors. Wrapped errors preserve their cause where supported; use `errors.As` when you need to inspect a typed error.