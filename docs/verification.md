# Verification and fixture guidance

## Local checks

Run the formatter on changed Go files, then run the repository checks:

```sh
gofmt -w path/to/changed.go
make lint
make check
make generate
```

`make check` runs `go vet`, builds the packages and examples, and runs tests with the race detector. CI regenerates the GraphQL client and the OpenAPI/AsyncAPI wire models, then fails if generated output or module metadata is stale. These checks do not require Amazon credentials.

## Fixture provenance

Files under `pkg/alexa/testdata/` are synthetic inputs for offline conversion tests. They are not captured traffic and are not evidence of current service behavior. The request and response transports used by tests also return synthetic data. Keep synthetic inputs in that directory and mark their synthetic status in the neighboring README.

The maintainer reports that the APIs covered by this library worked when tested with their own real Alexa accounts. Those tests were not recorded here with an operation list, dates, results, or sanitized request/response captures. Treat this as maintainer-reported history; it does not provide reproducible or independently auditable evidence of current provider behavior. There are no real captured request/response fixtures in this repository. If sanitized captures are added later, keep them in an explicitly named `captured` directory, record the operation, UTC capture date, source category, redactions, and supported behavior in a neighboring provenance note, and remove tokens, cookies, personal data, customer IDs, and device IDs before adding them. Do not use a capture as proof of current behavior if its source or collection date is unknown.

The live integration tests in `test/integration` run only when `ALEXA_REFRESH_TOKEN` is set; otherwise they skip. They are not part of credential-free CI. Use a disposable account for live testing and do not commit its credentials or output.

## Documentation and schema

The `Documentation` workflow generates a static API reference from the checked-in GraphQL SDL, OpenAPI document, AsyncAPI document, and authored guides. The GraphQL SDL is the subset used by this client. The HTTP schemas describe route and parser shapes read from the source; they are not provider-issued specifications or verified observations. The reported account tests are not documented at the operation or protocol level, and there are no sanitized live REST or HTTP/2 captures in this repository. Synthetic tests do not prove current service behavior.

`make generate` runs genqlient against the checked-in SDL and operation documents, oapi-codegen against `api/openapi.yaml` and `api/embedded-wire.yaml`, and Modelina against `api/asyncapi.yaml`. The embedded schema describes the JSON text within `sequenceJson` and metric-specific event properties; these models retain historical exported names in `pkg/dependencymodels`. They are exported compatibility types, while private REST and directive models live under `internal`. CI compares every generated output to the checked-in files and checks method/path and query-key callsites.

The generic GraphQL `Execute`, `Query`, and `Mutate` methods accept only the seven generated operation documents. The allowlist is generated from genqlient output; an unknown document returns an error before any request. `RunBehavior` accepts a JSON sequence string for compatibility, then validates its sequence and node types against `api/embedded-wire.yaml` before sending. Caller-supplied feature names and operation payload fields remain open within the schema's documented path parameters and object fields; they cannot create an unlisted HTTP method or route.
