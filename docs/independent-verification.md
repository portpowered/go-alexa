# Independent standards review: go-alexa

## Fourth pass: paired replay at `12d713d0a1e8a159c8d1957ce79dbb66943257d3`

An independent reviewer who did not implement the new replay suite inspected
all three checked-in synthetic pair files, their REST, GraphQL, and event
callers, the strict matcher, inventory gate, and negative tests. [The
paired-replay review](paired-replay-review.md) records a separate disposition
for each prior item-15 finding. **Item 15 is verified**: 26 REST pairs cover
18 dispatched routes, seven GraphQL pairs cover the generated documents, and
two ordered HTTP/2 event pairs cover the directive frame and ping. Fresh race
tests, `make lint`, `make check`, the route gate and its negative tests passed;
non-generated coverage was 81.5% (2,100/2,576).

This pass does not renew item 14. Items 1–3 and 5–12 retain their prior
source verdicts; item 6's 80% floor was remeasured above. Item 4 still needs
published generation/gate verification and an exact-tag release run. Item 13
still needs the final rendered-site copy and link audit. Item 14 requires a
fresh full 15-item review of the final commit after those gates, including
release, site, badge and consumer evidence. The repository checklist keeps
those items open.

## Third pass: `af155791cc2ec4b80d25bf4883ffbfc861c398b7`

The previously reported path-reassignment and `Header.Add` bypasses are fixed.
Their negative tests pass, as do `go run ./tools/apiroutes --check`, `make lint`,
and `make check`. The GraphQL guide link now uses the SDL field slug
`query/endpoint/`. CI and Documentation runs for this commit were still in
progress when this pass was written; the exact next release tag is untested.

**Item 4 remains open:** the gate recognizes selected request constructors and
repository wrappers, but a new `http.Client.Post(...)` call with a generated
path can evade its method/path check. A direct `req.Header["X-Undeclared"]`
write can evade its header-name check. The network boundary gate must enumerate
all production outbound construction/send callsites, reject any unregistered
primitive, and cover direct header-map writes. Add negative tests for both
cases. The production inventory currently consists of `http.NewRequestWithContext`
followed by injected `*http.Client.Do` in REST, GraphQL, and the HTTP/2 event
stream; `pkg/testing` contains separate response-capture transport helpers.
There are no production WebSocket, MQTT, RTC, or direct dial callsites.

The current verdict for items 1–3 and 5–11 remains as recorded below. Item 12
needs a passing Documentation deployment. Item 13 needs final rendered-page
review. Item 14 and the release tag remain open.

## Second pass: `61f9ab51261d299704993aa24e6c8e778c09fda0`

The first-pass table below remains as historical review evidence. At the
second-pass commit, `make lint`, `make check`, and CI run `36488310634` pass.
Independent race-enabled, non-generated coverage is 81.1% combined
(2,088/2,576): `pkg/alexa` 77.8%, `pkg/alexaapimodels` 83.9%, GraphQL 91.0%,
REST 81.5%, and `pkg/dependencymodels` 100%. The checklist's 81.2% figure
needs reconciliation with a final profile. Nine handwritten REST wire request
calls now use schema-generated types. Request headers now use generated names,
and negative gate tests cover unschematized routes, method changes, channels,
and headers.

Items 1, 2, 5, and 7–11 retain their first-pass source verdicts. Item 3 still
needs the final deployed badge check. Item 6 passes the 80% floor at this
commit, with exact-tag recheck pending. **Item 4 remains open:** the callsite
gate uses only the first assignment to a path variable, allowing a later
reassignment to change the actual route without detection; it checks
`Header.Set` but not `Header.Add`. Add negative tests and fix both bypasses.
**Item 12 remains open:** Documentation run `36488310636` failed because the
wire-contract guide links to `query/getendpoint/` while the SDL-generated page
uses `query/endpoint/`. Item 13 needs a fresh rendered-site copy/link review.
Item 14 cannot pass until every finding is fixed, the final commit is reviewed,
and exact-tag release checks succeed.

Reviewer: Codex `alexa_reviewer` (independent of implementation). First-pass
reviewed commit: `e250e5e7f16e496375014dbae666d34089310e3e` (28 September 2026).
This is an **open review**, not a migration or release sign-off. The working
tree subsequently changed; every affected item needs a final-commit recheck.

The contracts for private Alexa services are implementation-derived. The
maintainer reports successful account tests, but no operation-level, sanitized
account evidence is present. Synthetic tests establish offline behavior only.

| Item | Verdict | Independent evidence and remaining work |
| --- | --- | --- |
| 1. Standalone library | Pass | `go.mod`, `README.md`, `cmd/examples/`, and `pkg/alexa` use the `github.com/portpowered/go-alexa` module without a consuming application import. |
| 2. API guidance | Pass | The README documents operations, auth, errors, and injection. `docs/guides/` includes authentication, enumeration, control, events, and quality-of-service workflows. The guides label the private contracts implementation-derived. |
| 3. Badges | Pass for source; deployment recheck required | The README has all seven required badge categories and repository-specific links. Published badge values and destinations need checking again after the final deployment. |
| 4. Complete schema generation | **Fail** | `api/openapi.yaml` declares 22 HTTP method/path pairs (21 dispatched plus caller navigation); `api/asyncapi.yaml` declares the directive stream; seven GraphQL documents are checked in. Generation uses genqlient, oapi-codegen, Modelina, and `tools/apiroutes`. However, `pkg/dependencies/rest/endpoints.go` sends handwritten `alexamodels.MediaCommand` at eight media command callsites and handwritten `alexamodels.BehaviorPreviewRequest` in `SendFireTVSequence`, while other code uses generated wire structs. `pkg/dependencies/rest/client.go` retains handwritten HTTP header names. `tools/apiroutes` checks generated method/path names anywhere in a function, not necessarily the pair passed to the request, and has no effective channel or header callsite check. Its tests cover a changed method/path and query key, but lack the required intentionally unschematized endpoint and channel cases. The source guides' new wire-reference link failed the Documentation run at this commit. Replace wire structs, strengthen the gate, add all negative cases, regenerate, and verify CI and Pages. |
| 5. Offline checks and provenance | Pass for local checks | `make lint` and `make check` passed at the reviewed commit, including race tests. `docs/verification.md` distinguishes synthetic fixtures, unrecorded account testing, and live opt-in tests. No captured fixture is claimed as verified. |
| 6. Coverage | Pending final measurement | `.github/workflows/ci.yml` enforces the `tools/coverage -min 80` non-generated combined floor. Synthetic tests cover client, REST, GraphQL, and models. Re-run and record a clean final-commit profile and package report; 90% remains a target. |
| 7. Package layout | Pass | Reusable API under `pkg/alexa`, transport under `pkg/dependencies`, private wire under `internal`, generated compatibility models under `pkg/dependencymodels`, examples under `cmd/examples`. Separate consumer evidence is recorded in `docs/template-checklist.md`; repeat on final release tag. |
| 8. Options | Pass | `pkg/alexa/client.go` has `NewClient(opts ...Option)` with region, base URL, timeout and transport options. `client_options_session_test.go` exercises invalid and conflicting values. |
| 9. Stateless reusable client | Pass | `Client` holds service configuration and transport only; account credentials, cookies, and event connections belong to `Session`. `Session.Close` and event close/cancellation tests cover lifecycle. |
| 10. Injectable network edges | Pass | Separate REST, GraphQL, and event HTTP client options and event `RoundTripper` are exposed in `pkg/alexa/client.go`. Synthetic transport tests exercise request and response behavior. |
| 11. Caller-visible tokens | Pass | Explicit refresh and cookie exchange return credentials; `Session.Credentials` returns a copy and `SetAccessToken` installs caller-selected rotation. Ordinary requests do not refresh tokens implicitly. |
| 12. MDX site | **Fail at reviewed commit** | Seven customer-facing MDX guides are in `docs/guides/`; contributor and release notes remain Markdown. The new `wire-contracts.mdx` contained a case-sensitive generated-reference link that failed the Documentation workflow. Fix and re-run full-site link checks and deployment. |
| 13. Concise rendered copy | Pending final site | The prior rendered site and links were reviewed according to the checklist. The new inventory guide and regenerated reference require a fresh rendered-site copy and link audit, including root, generated pages, external URLs, and release-note links. |
| 14. Independent verification | **Open** | This document records an independent first pass. Items 4, 12, and 13 remain open, and exact-tag gates have not run on the next release. Re-review every fix at the final commit, attach CI/Documentation/tag evidence, then sign off. |

## Item 4 exchange inventory and gate evidence

The HTTP inventory in `docs/guides/wire-contracts.mdx` lists all 22 currently
declared operations, including OAuth, REST enumeration/control, media, GraphQL,
and directive stream/keepalive. `api/openapi.yaml` maps their method/path pairs;
`api/asyncapi.yaml` maps the directive channel and framing; checked-in GraphQL
SDL and operation documents map the seven supported GraphQL operations;
`api/embedded-wire.yaml` models nested event properties and serialized behavior
sequences. The declared inventory is not yet proven complete by the current
gate because handwritten wire requests and unconstrained callsites remain.

At this commit, `go test ./tools/apiroutes` passes, but its cases do not prove
that a newly introduced unschematized endpoint or channel fails. A method and
its matching path can appear separately in a function and satisfy the scanner
even if the actual request uses another route. A final audit must inspect the
actual generated type and constant at each outbound callsite, then deliberately
mutate a route, method, channel, header, and generated artifact to check that
CI rejects each mutation.

## Release and history disposition

`docs/template-checklist.md` records the standalone repository history rewrite,
old-tag removal, current-tree secret scan, `v0.2.0` publication, prior public
Go proxy consumer check, and the GitHub Support follow-up for cached old SHA
views. This review does not reopen the monolith history. The next tag must run
generation/drift, exchange inventory, coverage, race, compatibility, and
public-proxy consumer checks on its exact commit before item 14 can pass.

## Findings to resolve

1. Replace all handwritten `MediaCommand` and `BehaviorPreviewRequest` values
   at REST wire boundaries with schema-generated types.
2. Generate and use every HTTP header name at wire callsites; gate unschematized
   names, including custom headers.
3. Make the callsite gate match the actual method/path pair and directive
   channel; add negative tests for an unschematized route, changed method,
   unschematized channel, and handwritten header.
4. Fix the wire-contract guide's generated-reference link, run Documentation,
   and review the complete rendered site and external destinations.
5. Record clean final-commit coverage, CI, documentation, separate consumer,
   and exact-tag release gate results, then request reviewer re-verification.
