# Independent paired-replay review

Initial reviewed commit: `46ce820044aa17ff5adb0f9f2e8eb891e886e8a6`.
Re-audited implementation commit: `12d713d0a1e8a159c8d1957ce79dbb66943257d3`.
Standard: item 15 of `go-third-party-template/docs/library-standards.md`
and the paired-fixture detail in its `docs/verification.md`. This review is
independent of the Alexa implementation. The earlier itemized verification
report predates item 15.

| Item | Verdict | Evidence |
| --- | --- | --- |
| 15. Paired replay | **Verified at `12d713d`** | Checked-in synthetic pairs cover every dispatched REST route, seven GraphQL documents, and the directive stream plus ping. The strict replay transport validates each request before its response and asserts ordered exhaustion. See the re-audit below. |
| 14. Independent verification | **Open** | Checklist items 4 and 13 and the exact-tag release gate remain open. The full 15-item final-commit signoff must follow those checks. |

## Initial findings at `46ce820`

1. **The capture replay helper is not used by the operation tests.**
   `pkg/testing/replay_capture.go` loads ordered `CapturePair` files, checks
   the method, complete URL, headers, and body before returning the next
   response, rejects calls beyond the transcript, and provides
   `AssertConsumed`. `pkg/testing/replay_capture_test.go` exercises a query,
   header, or body mismatch and verifies exhaustion and a repeated call.
   A source search found no other caller of
   `NewReplayCaptureRoundTripper`; the test creates both pairs in a temporary
   directory. There are no checked-in `capture_*.json` pairs.
2. **The HTTP operation inventory has no stored paired coverage.** The
   OpenAPI inventory contains 22 operations, including one caller-facing
   OAuth navigation operation and 21 dispatched HTTP operations. One
   dispatched route is the GraphQL POST, through which seven generated
   GraphQL operations execute. All 21 dispatched routes and their supported
   outcomes need paired request/response replay. The six JSON files in
   `pkg/alexa/testdata/` are labeled synthetic but contain responses only.
   `pkg/alexa/client_enumeration_test.go` chooses one of these by decoded
   path or substrings in a GraphQL body, then returns it without checking a
   complete outbound request. REST and GraphQL transport tests inspect
   selected request fields and construct responses inline; none stores and
   consumes a complete pair. Their existing assertions are useful unit
   tests but do not meet item 15's fixture and replay requirement.
3. **The directive stream lacks an ordered bidirectional transcript.**
   `pkg/alexa/http2_synthetic_test.go` returns inline directive and `/ping`
   responses based on decoded path. It checks a token and some authority
   behavior, but does not store the stream-open request, response status and
   headers, ordered incoming directive frames, ping exchange, and terminal
   behavior in a transcript that rejects unexpected or duplicate messages
   and asserts full consumption. Event parser tests also feed inline JSON
   directly to the parser. The AsyncAPI directive channel therefore has no
   item-15 replay evidence.
4. **Volatile matching and provenance need a defined policy.** The helper
   compares the full URL, header map, and body byte for byte. It has no
   explicit match rules for volatile request IDs, timestamps, signatures,
   or redacted credentials, so real sanitized exchanges cannot be replayed
   with format or decoded-meaning checks. New synthetic pairs should be
   identified as synthetic; any sanitized capture requires separate source,
   date, and redaction provenance. No present fixture claims to be a live
   account capture.

Fresh `go test -count=1 -race ./pkg/testing ./pkg/alexa
./pkg/dependencies/rest ./pkg/dependencies/graphql` passed at the reviewed
commit. This confirms the current unit tests, not paired replay across the
wire inventory.

## Re-audit at `12d713d`

| Initial finding | Disposition and independent evidence |
| --- | --- |
| Helper not used by operation tests | **Resolved.** `pkg/testing.SyntheticReplay` is loaded directly by the REST, GraphQL, and event operation suites from three checked-in files. Each suite calls `AssertConsumed`; the capture helper remains separate for possible future captured exchanges. |
| HTTP operation inventory | **Resolved.** `alexa-rest.json` contains 26 pairs for all 18 dispatched REST routes, including two registration modes and eight media commands. `alexa-graphql.json` contains seven pairs, one per generated GraphQL document. `tests/replay/inventory_test.go` cross-checks OpenAPI operation IDs and the seven GraphQL documents. The 22nd OpenAPI operation, `openAuthorizationPage`, yields a browser URL for caller navigation and is not dispatched by the library. The six `pkg/alexa/testdata` files remain labeled response-only model inputs and are excluded from replay claims. |
| Directive stream | **Resolved for the built-in HTTP/2 edge.** `alexa-events.json` stores an ordered GET for the stream, a multipart response carrying the synthetic directive frame, and the keepalive `/ping` request/response. The event suite checks the parsed public event, ping, order, exhaustion, duplicate, and out-of-order rejection. |
| Volatile matching and provenance | **Resolved for these synthetic pairs.** All IDs, timestamps, and credential placeholders in these suites are fixed synthetic values and are compared exactly, including complete request headers and body. The fixture README states the exact-match rule and requires a constrained matcher if future dynamic fields are introduced. Every pair has `source: synthetic`; no live capture is claimed. |

`SyntheticReplay.RoundTrip` matches method, origin, escaped path, complete
repeated query values, full request headers, and body before serving the
paired status, headers, and body. Its negative test mutates every request
field and verifies failure before response; the operation suites use ordered
consumption and reject duplicate calls. I independently counted 26 REST,
seven GraphQL, and two event pairs, all with synthetic provenance and
response headers. Fresh `-count=1 -race` tests of `pkg/testing`, REST,
GraphQL, `pkg/alexa`, and `tests/replay` passed. `make lint`, `make check`,
`go run ./tools/apiroutes --check`, and its negative tests passed. A fresh
non-generated race profile passed the 80% floor at 81.5% (2,100/2,576).

**Item 15 verdict: verified at `12d713d`.** The broader 15-item review
cannot be signed yet: the repository checklist still leaves item 4 pending
published regeneration/gate and exact-tag checks, and item 13 pending a
fresh rendered-site copy/link audit. Item 14 remains open until those gates
pass and every item is rechecked at the final release commit.
