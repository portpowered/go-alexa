package alexa

type staticError string

func (err staticError) Error() string { return string(err) }

const (
	errClientOptionNil         staticError = "client option must not be nil"
	errConflictingEventOptions staticError = "WithEventHTTPClient and WithEventTransport cannot both be set"
	errClientNil               staticError = "client must not be nil"
	errSessionOptionNil        staticError = "session option must not be nil"
	errAccessCredentialEmpty   staticError = "access token must not be empty"
	errBearerCredentialEmpty   staticError = "bearer token must not be empty"
	errRefreshCredentialEmpty  staticError = "refresh token must not be empty"
	errCustomerIDEmpty         staticError = "customer ID must not be empty"
	errCookiesNil              staticError = "cookies must not be nil"
	errCookieNil               staticError = "cookie"
	errConflictingCookies      staticError = "conflicting values for cookies"
	errCSRFTokenEmpty          staticError = "CSRF token must not be empty" //nolint:gosec // Fixed validation text is not a credential.
	errAbsoluteHTTPURLRequired staticError = "expected an absolute HTTP(S) URL without credentials, query, or fragment"
	errHTTPClientNil           staticError = "HTTP client must not be nil"
	errRESTHTTPClientNil       staticError = "REST HTTP client must not be nil"
	errGraphQLHTTPClientNil    staticError = "GraphQL HTTP client must not be nil"
	errEventHTTPClientNil      staticError = "event HTTP client must not be nil"
	errEventTransportNil       staticError = "event transport must not be nil"
	errTimeoutNotPositive      staticError = "timeout must be positive"
	errUnsupportedAlexaRegion  staticError = "unsupported Alexa region"
	errConflictingValues       staticError = "conflicting values for"
	errInvalidEventAuthority   staticError = "invalid HTTP/2 event authority"
	errEventMessageMissing     staticError = "message or data is nil"
	errDirectiveInvalid        staticError = "directive not found or invalid"
	errHeaderInvalid           staticError = "header not found or invalid"
	errPayloadInvalid          staticError = "payload not found or invalid"
	errRenderingUpdatesEmpty   staticError = "renderingUpdates not found or empty"
	errRenderingUpdateInvalid  staticError = "renderingUpdate is invalid"
	errResourceMetadataInvalid staticError = "resourceMetadata not found or invalid"
	errNoFeaturesFound         staticError = "no features found"
	errNoPropertiesFound       staticError = "no properties found"
)
