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

Files under `pkg/alexa/testdata/` are synthetic inputs for offline conversion tests. They are response-only model examples, not captured traffic or paired replay evidence. The checked-in request/response pairs live under `tests/replay/fixtures/synthetic/` with a provenance README. None establish current provider behavior.

`pkg/testing.SyntheticReplay` matches each outbound method, origin, escaped path, complete repeated query multimap, full headers, and body before releasing the paired status, headers, and body. REST, GraphQL, and event tests call `AssertConsumed`; schema inventory tests fail when a dispatched OpenAPI or generated GraphQL operation lacks a pair. The event script orders the directive stream response frame before the keepalive ping. Mismatch, duplicate, unexpected, and unconsumed exchanges have negative tests. All synthetic IDs, timestamps, and credentials are fixed values and match exactly; no volatile field is skipped. If a future captured exchange needs a variable ID, timestamp, signature, or redacted credential, add a field-specific format or decoded-value matcher and a malformed-value test before admitting it. `ReplayCaptureRoundTripper` remains for sanitized real captures when provenance is available.

The maintainer reports that the APIs covered by this library worked when tested with their own real Alexa accounts. Those tests were not recorded here with an operation list, dates, results, or sanitized request/response captures. Treat this as maintainer-reported history; it does not provide reproducible or independently auditable evidence of current provider behavior. There are no real captured request/response fixtures in this repository. If sanitized captures are added later, keep them in an explicitly named `captured` directory, record the operation, UTC capture date, source category, redactions, and supported behavior in a neighboring provenance note, and remove tokens, cookies, personal data, customer IDs, and device IDs before adding them. Do not use a capture as proof of current behavior if its source or collection date is unknown.

The live integration tests in `test/integration` run only when `ALEXA_REFRESH_TOKEN` is set; otherwise they skip. They are not part of credential-free CI. Use a disposable account for live testing and do not commit its credentials or output.

## Documentation and schema

The pinned external codec contract is inventoried in `api/external/http2.yaml`.
`TestEventNativeHTTP2PairedReplay` injects a `net.Pipe` connection beneath
`http2.Transport`, serves paired exchanges through `http2.Server`, and verifies
HTTP/2 requests, directive decoding, keepalive, complete consumption, and close.
The codec owns standard HTTP/2 control frames; generated provider routes and
message models remain the application's contract. This is offline synthetic
evidence and does not verify provider TLS negotiation.

The `Documentation` workflow generates a static API reference from the checked-in GraphQL SDL, OpenAPI document, AsyncAPI document, and authored guides. The GraphQL SDL is the subset used by this client. The HTTP schemas describe route and parser shapes read from the source; they are not provider-issued specifications or verified observations. The reported account tests are not documented at the operation or protocol level, and there are no sanitized live REST or HTTP/2 captures in this repository. Synthetic tests do not prove current service behavior.

`make generate` runs genqlient against the SDL and operation documents,
oapi-codegen against the HTTP and API-specific component schemas, and Modelina
against `api/asyncapi.yaml`. Provider field definitions belong in
`pkg/dependencymodels`, with separate generated files for authentication,
endpoints, behaviors, feature events, media, and stream payloads. Compatibility
projections preserve old exported Go shapes when those differ from runtime wire
nullability. `tools/oapiallOfix` restores Go embedding from schema `allOf`
references and rejects inconsistent compositions. Handwritten companions implement behavior or conversion; they do
not duplicate wire fields. CI checks regeneration, untracked generated output,
endpoint call sites, and the complete model inventory.

`AuthConfig` and `DeviceAuthConfig` are retained legacy caller-side configuration
shapes. No SDK operation accepts, encodes, or decodes them; they have no JSON
fields or codecs and are excluded from the wire-model population. Their deprecated
comments make the caller-owned OAuth scope explicit.

Edit the API responsibility files listed in `api/openapi/sources.yaml`;
`tools/openapibundle` builds `api/openapi.yaml` for code generation and the site.
Requests, responses, and their nested components stay in their owning API file.
`make openapi-bundle-check` rejects bundle drift and mismatched operation owners.

The inventory must include active models, exported compatibility definitions,
anonymous nested serialization objects, and custom encoders or decoders. Link
each type to its schema component, generated definition, generator command,
and use. Independently search for missing entries. Negative tests must reject
an unreferenced exported handwritten JSON struct and an anonymous nested wire
object. A generated file list or route-gate pass does not prove completeness.

Include the concrete behavior payloads and wire constants used by library
builders, not just their outer sequence envelope. Distinguish those known
shapes from caller-defined open payload input. For GraphQL, check actual emitted
operation selections, generator-added discriminators, and the exact decoder
branches against SDL possible types. Label type-only implementations without
selected subtype fields explicitly. Negative controls must reject unregistered
nested payloads or wire keys, mismatched decoder branches, possible-type drift,
and missing discriminators.

The primitive inventory records generated wire constants, SDK constant projections,
and their semantic types. SDK and GraphQL uses are resolved through exact imports.
Shared values are bound to the SDL, with negative tests for drift and missing
bindings. Open feature-state strings list known SDL values and preserve future
values. Legacy capability JSON is bound to its known component while retaining the
original scalar selection. Wire-construction tests reject new literal values and
keys, local aliases, and later field mutations; separate tests reject forged
constant output. String templates, authentication choices, and cookie names also
have schema owners. Named event components bind enum fields to the SDL, with
negative checks for dropped bindings and inline nested objects. The flat speaker
volume/muted event is implementation-derived and has no direct SDL counterpart. SDK constants are generated in their semantic package because
compatibility wire types import that package; both projections share schema values.

The generic GraphQL `Execute`, `Query`, and `Mutate` methods accept only the seven generated operation documents. The allowlist is generated from genqlient output; an unknown document returns an error before any request. `RunBehavior` accepts a JSON sequence string for compatibility and validates its sequence and node types before sending. Caller-supplied feature names and operation payload fields remain open within the schema's documented path parameters and object fields; they cannot create an unlisted HTTP method or route.

## Documentation review

Review every tracked Markdown and MDX file, including documents excluded from
the site. Record its audience and purpose, remove redundant internal reports,
and check incoming links after deletion. Keep the README useful to callers and
the customer guide navigation free of maintenance audits. Retain one current
checklist and [independent review](independent-review.md); run the full rendered
site link check after source review.
