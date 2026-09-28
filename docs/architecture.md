# Client architecture

`pkg/alexa` is the public orchestration layer. It exposes the client, options, and event connection. `pkg/dependencies/rest` and `pkg/dependencies/graphql` own their corresponding HTTP request work. `pkg/alexaapimodels` contains public request, response, event, and endpoint models; `pkg/dependencymodels` contains provider wire models.

## Endpoint enumeration

`Client.ListEndpoints` queries GraphQL for endpoint information and REST for device metadata, then merges the available fields. REST metadata is optional: if that request fails, the method returns GraphQL endpoints with whatever data is available. Request `EndpointIncludeFields` to include features and state properties.

The checked-in GraphQL schema in `pkg/schemas/graphql/endpoints-schema.graphql` is the subset used for the operations in this repository. It is not a complete Alexa schema or a promise that every provider field or operation is stable. `genqlient.yaml` and `make generate` keep the query models tied to the checked-in schema and query documents.

## Controls and events

`Client.Control` maps a feature namespace and operation to the provider control request. A successful request means the API accepted or dispatched the request; it does not confirm the physical endpoint reached the requested state. Responses may include operation-level errors.

For events, call `Subscribe` before `ConnectEvents`. `ConnectEvents` returns an HTTP/2 connection whose `Receive` method reads events until an error or closure. The caller owns the returned connection and should close it when the stream is no longer needed.

## Authentication and transport

The client supports access-token and refresh-token options, code-based linking, region selection, and an injected `*http.Client` for REST and GraphQL requests. The event stream establishes a separate HTTP/2 connection. Keep tokens and account data in the consuming application's credential store; do not put them in fixtures, logs, or source control.

## Evidence limits

The automated tests use synthetic JSON responses and test transports. They check request construction and model conversion, but they do not establish that undocumented provider behavior remains available. Live verification is opt-in and should use a disposable account where possible.