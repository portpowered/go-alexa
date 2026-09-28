# Verification and fixture guidance

## Local checks

Run the formatter on changed Go files, then run the repository checks:

```sh
gofmt -w path/to/changed.go
make lint
make check
make generate
```

`make check` runs `go vet`, builds the packages and examples, and runs tests with the race detector. CI regenerates the GraphQL client and fails if generated output or module metadata is stale. These checks do not require Amazon credentials.

## Fixture provenance

Files under `pkg/alexa/testdata/` are synthetic inputs for offline conversion tests. They are not captured traffic and are not evidence of current service behavior. The request and response transports used by tests also return synthetic data. Keep synthetic inputs in that directory and mark their synthetic status in the neighboring README.

There are no real captured request/response fixtures in this repository. If sanitized captures are added later, keep them in an explicitly named `captured` directory, record the operation, UTC capture date, source category, redactions, and supported behavior in a neighboring provenance note, and remove tokens, cookies, personal data, customer IDs, and device IDs before adding them. Do not use a capture as proof of current behavior if its source or collection date is unknown.

The live integration tests in `test/integration` run only when `ALEXA_REFRESH_TOKEN` is set; otherwise they skip. They are not part of credential-free CI. Use a disposable account for live testing and do not commit its credentials or output.

## Documentation and schema

The `Documentation` workflow generates a static API reference from the checked-in GraphQL SDL and builds the authored guides. The SDL is the subset used by this client, not a complete Alexa contract. `make generate` runs genqlient against the checked-in SDL and operation documents; CI compares its result to the checked-in generated client.
