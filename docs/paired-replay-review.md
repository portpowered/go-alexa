# Independent paired-replay review

Reviewed commit: `46ce820044aa17ff5adb0f9f2e8eb891e886e8a6`.
Standard: item 15 of `go-third-party-template/docs/library-standards.md`
and the paired-fixture detail in its `docs/verification.md`. This review is
independent of the Alexa implementation. The earlier itemized verification
report predates item 15.

| Item | Verdict | Evidence |
| --- | --- | --- |
| 15. Paired replay | **Open** | The replay helper works on temporary pairs in its own unit test, but no supported Alexa operation is tested from a checked-in request/response pair or ordered directive transcript. |
| 14. Independent verification | **Open** | Item 15 and the existing release/schema/site gates remain open. Recheck every checklist item after the fixes at the final commit. |

## Findings

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

**Signoff:** neither item 15 nor renewed item 14 can be checked at this
commit.
