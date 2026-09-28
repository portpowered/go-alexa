# Client architecture

`pkg/alexa` is the public orchestration layer. `Client` holds validated region, endpoint, timeout, and network configuration. `Session` holds one account's credentials and account operations. `pkg/dependencies/rest` and `pkg/dependencies/graphql` own their corresponding HTTP request work. `pkg/alexaapimodels` contains public request, response, event, and endpoint models; `pkg/dependencymodels` contains provider wire models.

## Endpoint enumeration

`Client.ListEndpoints` queries GraphQL for endpoint information and REST for device metadata, then merges the available fields. REST metadata is optional: if that request fails, the method returns GraphQL endpoints with whatever data is available. Request `EndpointIncludeFields` to include features and state properties.

The checked-in GraphQL schema in `pkg/schemas/graphql/endpoints-schema.graphql` is the subset used for the operations in this repository. `api/openapi.yaml` and `api/asyncapi.yaml` describe REST requests/responses and the directive input accepted by the event parser. All three schemas describe shapes used by this implementation, not complete or provider-verified contracts. `make generate` keeps the GraphQL client and internal REST/event wire models tied to those schema files.

## Controls and events

`Client.Control` maps a feature namespace and operation to the provider control request. A successful request means the API accepted or dispatched the request; it does not confirm the physical endpoint reached the requested state. Responses may include operation-level errors.

For events, call `Session.Subscribe` before `Session.ConnectEvents`. The returned HTTP/2 connection has a caller-owned lifecycle. Close the connection when the stream is no longer needed; closing its session also closes any streams still open.

## Authentication and transport

Create clients with functional options such as `WithRegion`, base URL overrides, `WithTimeout`, and independent REST, GraphQL, and event HTTP clients or transports. Create a session with `WithBearerToken`, `WithRefreshToken`, cookies, and other account data. Refresh and cookie exchange methods run only when called, return the resulting credentials, and leave storage and token rotation to the caller. Missing CSRF material is returned as an error; call `Session.GetCSRFToken` explicitly to fetch it. Keep credentials in the consuming application's credential store; do not put them in fixtures, logs, or source control.

## Evidence limits

The automated tests use synthetic JSON responses and test transports. They check request construction and model conversion, but they do not establish that undocumented provider behavior remains available. The maintainer reports successful API testing with their own Alexa accounts, but did not record the operations, dates, results, or sanitized captures; reproducible account-test evidence is absent from this repository. Live verification is opt-in and should use a disposable account where possible.
