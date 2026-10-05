# Independent review record

Status: **open; fixes and two final independent approvals are required**.

The reports below audit `f125426a5f2c242d8100edd08814827ae18c9028`
against shared template `05e93ff08899414207e9335717e7d7b0190ebd09`.
They do not approve later changes or close the current
[checklist](template-checklist.md).

The current checklist pins shared template `62cc3cb5a1308dae8700f92052f99b1c455a1d98`.

## Findings being resolved

- Unresolved imported helper output can reach schema-owned wire fields.
- The route scanner excludes an added CLI network call.
- A request body's backing byte slice can change after construction.
- An injected HTTP cookie jar can share account state between sessions.
- The upgrade example should use the generated friendly-name default.
- The review record needs one consistent current set of verdicts.

The earlier reader-consumption probe was withdrawn: `net/http` clones that
reader, so the probe did not change the emitted body. It is not an open finding.

Earlier reviews and publication evidence remain in Git history. SDK `v0.4.0`
and CLI `cmd/go-alexa/v0.4.0` are published; publication does not approve the
new fixes. Final reviewers must independently check the fixed source and
publication state. The initial reviewers disclosed reading historical review
excerpts; their reproduced findings are retained here, and a fresh pair will
perform the final blind audit.

## Reviewer 1 — independently authored report

# Independent go-alexa checklist audit (Reviewer 1)

## Review basis

- Repository: `github.com/portpowered/go-alexa`
- Exact reviewed source commit: `f125426a5f2c242d8100edd08814827ae18c9028` (detached clean clone)
- Shared template revision: `05e93ff08899414207e9335717e7d7b0190ebd09`
- Scope: independent audit of all 16 requirements in `docs/template-checklist.md`; source, generated output, tests, CI runs, release artifacts, and published docs were inspected. No repository files were edited. All negative controls below used temporary overlays in the clean clone and were restored.
- Standards read from the shared template: `docs/library-standards.md`, `docs/verification.md`, `docs/client-design.md`, `docs/website.md`, and `docs/releasing.md`; also read the library checklist and `docs/verification.md`.

**Blind-review disclosure:** While searching narrowly for authentication coverage, one broad `rg` command also printed matching excerpts from `docs/independent-review.md`. I did not intentionally open or inspect that report and did not use its contents or claims as evidence. I did not read it further. Because checklist item 13 requires every tracked documentation file to be reviewed, that item remains open. This report's findings and reproductions are my own.

## Verdicts

1. **PASS — reusable library remains application-independent.** The SDK exposes the provider package under `pkg/alexa`, generated provider models under `pkg/dependencymodels`, and transport packages under `pkg/dependencies`. The public README and guides address SDK callers, not a named consuming application. Application-side callback/PKCE responsibility is explicitly left to the consuming app.

2. **PASS, with an unverified-value note — public API examples match exported shapes.** The README and customer guides show exported options, client methods, model shapes, authentication, errors, and transport injection. The upgrading guide (`docs/guides/upgrading.mdx:41-44`) uses `Type: "text"` for `FriendlyNameValue`; the schema marks this property as an open string with default `PLAIN` (`api/openapi/sources/endpoints.yaml:245-251`), and the generated constant/default outbound implementation use `PLAIN` (`pkg/dependencymodels/wire_constants.gen.go:52`, `pkg/dependencies/rest/endpoints.go:179-181`). The field is string-typed, so the example is API-shape-valid, but I found no provider evidence that `text` is a supported provider value. This is not evidence of a rejected request; the sample value should be treated as unverified rather than provider-confirmed. Customer docs consistently frame the contracts as implementation-derived, and synthetic material is separated from captures/history.

3. **PASS — README badges and repository identity are live.** README badges point to the current repository's Go version, CI, coverage, release, Go Reference, license, and documentation. Public Pages, release, coverage, pkg.go.dev, and repository destinations returned expected content/status in the checks performed.

4. **FAIL — source gates still have two concrete provenance gaps.** Existing evidence is strong: `go run ./tools/apiroutes --check` and `go run ./tools/modelinventory -check` pass at baseline; the inventory contains 1,492 model rows. Compile-valid temporary controls rejected novel fixed values/keys in generated payloads, later map writes, named-result/bare-return helpers, long helper chains, recursive fixed fallbacks, unreferenced handwritten wire structs, forged generated markers, map escapes, and invalid generated map mutations; caller-defined open input through `SendSequence` remained accepted. A named-result helper returning a generated map to an unverified callable was also rejected.

   Two required cases were not rejected:

   - **Cross-package unresolved helper provenance:** a temporary helper in `pkg/alexaapimodels` returned `time.Now().Format(time.RFC3339Nano)`; a helper in `pkg/alexa` passed that value through the imported package into `dependencymodels.Sequence.Type` (a generated `SequenceType` enum). `go test -run '^$' ./pkg/alexa ./pkg/alexaapimodels` passed, and root-default `go run ./tools/modelinventory -write` and `-check` both exited 0. This is library-produced unresolved provenance, not caller input. The enum's `Valid` method does not make the unchecked conversion safe at the construction boundary.
   - **CLI direct network primitive outside the route inventory:** a compile-valid overlay in `cmd/go-alexa/internal/cli` called `http.Get` on an unlisted URL. `go test -run '^$' ./internal/cli` passed. From the repository root, `go run ./tools/apiroutes --check`, `go run ./tools/modelinventory -write`, and `-check` all exited 0. `tools/apiroutes/main.go:697-711` walks only `root/pkg`, so direct network imports/calls added to the CLI are not gated by this inventory.

   Request-object controls that I tested were correctly rejected: post-construction `req.URL.User` assignment, `req.GetBody` mutation, and `req.TransferEncoding`/`req.ContentLength` mutations all compiled and then failed the exact root-default route check with “Client.Do request was reassigned or mutated after its schema-bound constructor.” A first body-reader probe was withdrawn: `http.NewRequestWithContext` clones a `*bytes.Reader`, so consuming the original reader did not change the request body. No finding is based on that probe.

5. **PASS — exact-commit blocking lint/check CI is green.** GitHub CI run `37192237178` is successful for exact `f125426…`; its verify job includes pinned golangci-lint v2.3.0, blocking `make lint`, `make check`, race tests, generation/inventory checks, and module checks. `.golangci.yml` has literal `linters.default: all`; no `issues-exit-code=0` or new-issues-only mode was found. Exceptions are path/rule-scoped with reasons. The independent review also passed the request-object negative controls on this exact source revision; checklist item 5 itself is marked only for completed CI at this revision.

6. **PASS — offline synthetic replay and coverage gates meet the stated floor.** CI's combined non-generated production coverage is 84.6%, above the 80% floor and below the 90% target. Coverage exclusions are reported. REST, GraphQL, event, and native HTTP/2 tests use exact synthetic paired requests/responses; the replay matcher checks method, origin, escaped path, repeated query values, headers and body, rejects unexpected/duplicate calls, and asserts consumption. `TestEventNativeHTTP2PairedReplay` injects `net.Pipe` and waits for server completion after close. Repository documentation explicitly records that these are synthetic and not evidence of current provider behavior.

7. **PASS — package boundaries and model inventory are in place.** Provider APIs, dependency wire models, and transport implementations have the requested package separation and responsibility schemas. Generated models are linked in `docs/model-inventory.md` to schema, generator, and use/conversion sites. The inventory covers legacy exports, anonymous wire objects, nested payloads, enums/constants, and custom decoders. Independent negative controls rejected unreferenced handwritten wire models, anonymous wire objects, and forged generated markers. Separate consumer module verification succeeded for SDK v0.4.0 (see item 16).

8. **PASS — functional options and configuration are explicit.** `pkg/alexa` clients are initialized with options and validated defaults. Base URL and HTTP client are injected; reusable client configuration does not retain account credentials. CLI credentials are supplied by documented environment/stdin/file paths rather than ordinary secret-bearing flags.

9. **PASS — account/session state is visible and separated.** The public client is reusable across accounts; session/token/stream lifecycles are represented explicitly. Session ownership and close/error state are caller-visible rather than hidden mutable account state on the client.

10. **PASS — active network edges are replaceable offline.** REST and GraphQL accept injected HTTP transports, events expose the HTTP/2 connection/dial seam, and native HTTP/2 framing is exercised over `net.Pipe`. The pinned `golang.org/x/net/http2` dependency is versioned in module metadata and described in `api/external/http2.yaml` with source-matched exchange/call-site evidence. No WebSocket/MQTT/RTC edge is claimed as supported.

11. **PASS — token changes are explicit.** Token exchange and refresh operations return the resulting credentials to the caller. The reusable client does not silently refresh and retain new tokens; customer documentation describes caller storage/renewal responsibility.

12. **PASS — customer guides render and reference links resolve.** The exact Documentation workflow `37192237182` succeeded for `f125426…`, including Fumadocs build, rendered link check, artifact upload, and Pages deploy. On the downloaded artifact, `python tools/check_site_links.py` checked 44,910 internal links across 251 rendered pages. The published guide routes and expected headings/content were checked, including CLI, authentication, events, endpoint operations, and upgrading; release guide and v0.4.0 verification links were also checked. No OpenAPI `externalDocs` entries are present to supply hidden runtime links. The live docs routes served the expected content rather than a generic HTTP-200 fallback.

13. **OPEN — full tracked-doc audit cannot be signed off under the blind constraint.** README, CONTRIBUTING, all nine MDX guides, checklist, verification/release/model-inventory docs, schema note, and fixture READMEs were reviewed for audience and purpose; the rendered Pages artifact was also checked. I have not inspected `docs/independent-review.md` due the blind-review instruction and the contamination disclosure above. The requirement explicitly says every tracked documentation file, including excluded-from-site files, must be reviewed. The `Type: "text"` upgrade example is a low-confidence value note, not a proven API defect.

14. **OPEN — no independent sign-off.** This report is one separate reviewer record only. Item 4 has two unresolved gate gaps, item 13 remains open, and any remediation must be rechecked by both independent reviewers at the final commit. No checklist sign-off is given.

15. **PASS — paired synthetic replay covers supported wire lifecycles.** Replay fixtures classify themselves as synthetic, match full request/response pairs or ordered event frames, and assert consumption/cleanup. Native HTTP/2 close completes through the injected connection. Auth transport fixtures cover credential exchange, token refresh, cookies/CSRF and code-pair forms; OTP challenge handling has negative coverage. There are no sanitized captures in the repository, which is stated plainly. The SDK does not implement an OAuth callback or PKCE client flow; callback-state/PKCE responsibilities belong to the consuming application, so I did not infer an SDK implementation gap from that out-of-scope flow. Maintainer-reported real-account success is documented as unrecorded history, not reproducible evidence.

16. **PASS — published standalone CLI consumes the public SDK.** The separate module under `cmd/go-alexa` is tagged at the CLI v0.4.0 release and requires public SDK v0.4.0 without a local `replace`. A fresh consumer successfully installed the CLI using `go install github.com/portpowered/go-alexa/cmd/go-alexa@v0.4.0`; installed `--help` listed the documented commands. A separate SDK consumer successfully resolved v0.4.0 from the public Go proxy, imported public packages, tidied, and passed `go test ./...`. Exact-commit CI also passed CLI Windows lifecycle/build/vet/module checks. Release workflow `37183039192` succeeded for the implementation/release tag; the SDK release was published 2026-10-04. No release or merge was performed as part of this audit.

## CI and publication evidence

- Exact source SHA `f125426…`: CI run `37192237178` SUCCESS; Documentation run `37192237182` SUCCESS.
- SDK v0.4.0 release workflow run `37183039192` SUCCESS; tagged implementation CI run `37183020394` SUCCESS.
- PR1 merge commit: `e95cdaf612041487dd9f024dd1681a56203aad36`.
- SDK source remained unchanged after `8a22252` through the reviewed documentation/CLI heads; the CLI module at its v0.4.0 tag matches current CLI source.
- Published docs artifact internal-link check: 44,910 links / 251 pages.
- Public SDK v0.4.0 consumer import/test and CLI v0.4.0 `go install` both passed in fresh temporary directories.

## Final result

Do not treat this audit as a release sign-off. Item 4 has two confirmed source-gate bypasses, item 13 remains open because one tracked report was not inspected under the blind-review constraint, and item 14 must remain open until those findings are resolved and independently verified at the final commit.

## Reviewer 2 — independently authored report

# Independent all-16 audit — go-alexa

Reviewed revision: `f125426a5f2c242d8100edd08814827ae18c9028` (`docs: require diagnostics for unresolved wire provenance`, 2026-10-04). The audit used a clean detached clone at `C:\Users\andre\AppData\Local\Temp\go-alexa-review2-f125426`; it was clean at completion. No changes remain in the repository clone: compile-valid probes and a regenerated inventory were made only in that temporary detached clone, then removed/restored. The original checkout and its local changes were not touched.

Standards basis: target `docs/template-checklist.md` plus `go-third-party-template` at `05e93ff08899414207e9335717e7d7b0190ebd09` (`docs/library-standards.md`, `docs/verification.md`, `docs/client-design.md`, `docs/website.md`, and `docs/releasing.md`). The exact f125 source tree contains no SDK production-code changes beyond the already-reviewed implementation; its delta from `8a22252` is README, standalone CLI module metadata, customer CLI guide, checklist, and review-record documentation.

## Verdict summary

| Item | Verdict | Summary |
|---:|---|---|
| 1 | PASS | Library content remains application-independent. |
| 2 | PASS | Customer instructions match the public API and state evidence limits. |
| 3 | PASS | Required README badges point to live project reports. |
| 4 | FAIL / OPEN | Documentation generation works, but a compile-valid unknown wire value passes the exact root `make check`; request-body backing aliases also escape the source gate. |
| 5 | PASS | Full pinned all-linter, build, vet, race, replay, and generation checks pass. |
| 6 | PASS WITH TARGET GAP | Combined non-generated coverage is 84.6%, above the 80% CI floor and below the stated 90% target. |
| 7 | PASS | Package/schema boundaries are appropriate; public SDK and CLI install from public modules. |
| 8 | PASS | Validated functional options configure endpoints and network clients. |
| 9 | FAIL / OPEN | A caller-injected cookie jar is shared across account sessions and forwards one session's cookie to another. |
| 10 | PASS | REST, GraphQL, and native HTTP/2 edges are injectable and tested offline. |
| 11 | PASS | Refresh and cookie exchange are explicit caller-managed operations. |
| 12 | PASS | Customer MDX guides render and link to generated references; the published site and link checks pass. |
| 13 | FAIL / OPEN | Post-delivery review found overlapping reports for older commits/standards and no reconciled report for f125. |
| 14 | FAIL / OPEN | Open source/session findings and the stale review record preclude final sign-off. |
| 15 | PASS, scoped | Paired replay is complete for implemented REST, GraphQL, and event transports. OAuth callback state and PKCE are caller-owned because the SDK does not implement a browser callback flow. |
| 16 | PASS | CLI is separate, public-installable, backed by the SDK, and has offline lifecycle/security tests. |

**No release sign-off:** items 4, 9, 13, and 14 remain open. Coverage is below the 90% target but meets the explicit CI minimum.

## Findings requiring disposition

### F1 — Unknown library output is accepted as a wire value (item 4)

In the clean exact-SHA clone, I added a compile-valid exported probe in `pkg/dependencies/rest` whose `Scopes` field on generated `WireCodePairRequest` receives `uuid.NewString()`. The value is produced by an imported library, so its provenance is neither a schema constant nor caller input. `go test -run '^$' ./pkg/dependencies/rest` compiled it. After inventory regeneration for the overlay, the default root `go run ./tools/modelinventory -check` passed. Most importantly, the exact root `make check GOLANGCI_LINT=C:\Users\andre\go\bin\golangci-lint.exe` also exited 0 with the overlay present (Go toolchain pinned through `GOTOOLCHAIN=go1.24.2`). The probe and regenerated inventory were then removed/restored.

This contradicts checklist item 4's fail-closed instruction: unresolved provenance must not be classified as caller-owned. `tools/modelinventory/wire_callable_escapes.go:25-31` rejects unknown calls when a generated wire value is passed *as an argument*; it does not reject an unknown call's return value when that value flows into a generated field. A two-file named-result helper returning `uuid.NewString()` showed the same acceptance in the focused compiler and inventory gates.

Positive negative-control probes behaved as expected: fixed literals returned through bare named results, recursive fixed fallback, a 128-helper sibling-file chain, named-result generated-map escape, forged generated marker, unlisted exported JSON struct, anonymous nested JSON object, and nested unknown map key/value were rejected by the root model inventory gate. Thus this finding is specifically unresolved external output provenance, not a blanket failure of the inventory controls.

### F2 — Mutable body backing storage is outside the route gate (item 4)

A temporary compile-valid request-construction overlay passed the root `go run ./tools/apiroutes --check` when it changed the byte slice backing a `bytes.Reader` after `http.NewRequestWithContext` and before `Client.Do`. The paired replay test observed the modified body and failed its exact request match. The route gate does reject direct request mutations for method, URL path/aliases, URL user info, `GetBody`, `ContentLength`, `TransferEncoding`, and `Trailer`, but does not preserve/check aliases to the body reader's source bytes. This leaves a requested post-construction body-mutation control unimplemented even though the current paired replay catches a changed shipped request.

### F3 — Injected cookie jar crosses session boundaries (item 9)

`pkg/alexa/client.go:171` shallow-copies `http.Client`; `NewSession` and `newRESTClient` reuse that client for every session (`client.go:225-264, 813-816`). An isolated synthetic probe injected an `http.Client` with `cookiejar.Jar`, had session one receive `Set-Cookie: account=first-session`, then called `GetUserInfo` from session two. Go's client automatically sent `Cookie: account=first-session` on the second session request. The probe passed by observing this cross-account cookie.

This occurs only when the caller injects an HTTP client with a non-nil Jar; the default client has no Jar. The public `Client` comment says it is safe to reuse across Alexa accounts, and the options accept a complete `*http.Client` without documenting this shared mutable account state. Until the SDK isolates the jar or clearly constrains this option, item 9 is not proven.

### F4 — The repository review record does not match the current checklist (item 13)

After delivering this report (so the blind-review constraint was honored), I reviewed `docs/independent-review.md`. Its heading describes two reports, but it contains multiple overlapping R2/publication sections for b74 and earlier states. It cites shared-standard pins `25aeb78` and `05e93ff`, says all sixteen items were checked at `d0b091b`, and records “no implementation finding remains open”; the actual target f125 checklist pins `05e93ff` and leaves items 4, 12, and 14 unchecked. The record has no f125 all-item disposition and repeats publication evidence. Item 13 requires one current, consistent review record linked from the checklist; the current record must be reconciled before that requirement can pass.
## Item-by-item evidence

### 1. Application independence — PASS

The README and public SDK describe a reusable provider client, caller-managed credentials, and supported provider operations. Application adapters, rollout logic, and account-specific orchestration are not part of the package or examples. CLI material is separately scoped to `cmd/go-alexa`.

### 2. API-accurate customer guidance — PASS

README and the nine MDX guides under `docs/guides/` show `NewClient`, `NewSession`, API types, explicit token refresh, event cleanup, and error handling consistent with exported methods. `docs/verification.md` calls schemas implementation-derived, labels the fixtures synthetic, and distinguishes unrecorded maintainer account tests from reproducible evidence. No official-provider claim is presented as verified.

### 3. README badges — PASS

README includes Go version, CI, coverage, release, Go Reference, license, and documentation badges, using the actual `portpowered/go-alexa` repository and live endpoints.

### 4. API reference and complete wire inventory — FAIL / OPEN

The Documentation workflow successfully builds and publishes the Fumadocs reference from the checked-in OpenAPI, AsyncAPI, GraphQL SDL, and guides. `docs/model-inventory.md` records schema component, generated Go declaration, generator, and use; `-list-handwritten` independently enumerates 110 caller-input, semantic, and domain records (38, 66, and 6 respectively). Generated route/model/control inventories are checked by the default root gates. The GraphQL gate traces generated response/input models to actual emitted selections and validates abstract-type discriminator/decoder branches against SDL possible types; negative tests cover missing/extra selections, SDL drift, wrong concrete decoders, missing discriminator, and possible-type drift. The pinned external `golang.org/x/net/http2` v0.47.0 traffic is documented in `api/external/http2.yaml` and exercised through its injected `DialTLSContext` over `net.Pipe`.

The REST route gate correctly rejected the request-object mutations enumerated under F2's positive control, and the model gate rejected the malformed inventory probes listed under F1. However, F1 and F2 show two unfulfilled fail-closed cases in the new requirements; this item cannot pass.

### 5. Blocking lint and offline checks — PASS

`.golangci.yml` has literal `linters.default: all`; CI pins golangci-lint v2.3.0 and runs repository-wide lint as a blocking step. Exceptions in config are narrowly scoped and explained. On the clean exact-SHA clone, `make lint` and `make check` both passed with `GOTOOLCHAIN=go1.24.2` and the installed v2.3.0 binary. `make check` ran root and CLI race tests, vet/build, model inventory, OpenAPI bundle, module tidy/verify. Exact main CI run `37192237178` passed both `verify` and `cli-windows` on f125.

### 6. Coverage — PASS WITH TARGET GAP

Exact CI run `37192237178` reports 84.6% non-generated production coverage, 2,397/2,834 statements, above the required 80% floor. `tools/coverage` reports package and combined statement totals and excludes generated files and `pkg/testing`. Both workflow documentation and `docs/verification.md` disclose the 90% target is not yet met.

### 7. Package boundaries and consumer import — PASS

Reusable public types are in `pkg/alexa` and `pkg/alexaapimodels`; wire models in `pkg/dependencymodels`; transports in `pkg/dependencies/rest`, `pkg/dependencies/graphql`, and the event transport in `pkg/alexa`. Separate schema responsibilities and generated files cover the APIs. Compatibility/semantic types are kept separate from generated wire shapes. I independently created a fresh module with `require github.com/portpowered/go-alexa v0.4.0` and compiled/tested a public `NewClient`/`NewSession`/`ListEndpoints` use. The public SDK v0.4.0 release job also records a successful public-proxy consumer check.

### 8. Client options — PASS

`NewClient` accepts functional options for region, validated HTTP(S) base URLs, event authority, timeout, common or dedicated HTTP clients, and event transport; nil and conflicting event options fail validation. Session options own bearer/refresh tokens, customer ID, cookies, and CSRF values. Invalid region/URL/authority and duplicate conflicting scalar options are tested. The injected-Jar account-isolation issue is separately reported under item 9.

### 9. Stateless client and explicit sessions — FAIL / OPEN

Credentials and event connections are held by `Session`; session closure cancels and closes owned event streams, and tests cover multi-session token separation and cleanup. F3 is a reproducible exception: a mutable injected CookieJar remains shared at `Client` scope and carries response cookies between account sessions.

### 10. Injection at every network edge — PASS

REST and GraphQL requests use injected `http.Client`s; event streams accept `WithEventHTTPClient` or `WithEventTransport`. The event path uses pinned `x/net/http2` and its dial hook; no WebSocket, MQTT, RTC, or separate socket protocol is implemented. Offline native HTTP/2 paired replay verifies actual HTTP/2 requests, framing, parsed directive payload, keepalive, consumption, and teardown. Network call inventory is restricted to named injected-client call sites. Source controls reject uncovered call sites and substituted local/shadow clients.

### 11. Explicit token exchange — PASS

Code-pair registration, refresh, and refresh-token-to-cookie exchange are explicit calls. Refresh returns new credentials and does not silently install them: the authentication guide requires callers to store the returned token and explicitly call `SetAccessToken`. Cookie/CSRF material is likewise caller-requested.

### 12. Customer documentation site — PASS

All nine customer workflows are MDX under `docs/guides/` and link to corresponding generated reference pages. Exact Documentation run `37192237182` passed on f125, including API-reference build, rendered-site link check, and Pages deployment. The artifact contains 251 rendered pages; `tools/check_site_links.py` checked 44,910 internal links successfully. Root, docs root, CLI guide, upgrade guide, generated operation page, and GraphQL page were checked on the published site and returned HTTP 200 with expected content. The scanned API schemas contain no `externalDocs` entries, so runtime schema-supplied external links are absent; the rendered artifact also had no external anchor destinations. Release-note links point to the live upgrade/verification/CLI guidance.

### 13. Documentation audience and redundancy — FAIL / OPEN

README, contributor/release/verification docs, schema notes, synthetic fixture README, all customer guides, and the rendered Pages site were reviewed for audience, purpose, duplicative content, and stale claims. README stays customer-focused; inventory and verification details remain contributor-facing; no extra navigation for internal audits appears on the site. After delivering this blind report, I inspected the tracked `docs/independent-review.md`. It contains overlapping reviews for older implementation/docs commits, mixes shared-template pins `25aeb78` and `05e93ff`, and reports an all-items completion state at `d0b091b` while the exact target f125 checklist has items 4, 12, and 14 unchecked. It does not include or reconcile a verdict for f125. The file also repeats b74 publication evidence across multiple R2 sections. This is one physical record but not a single current, consistent report for the f125 checklist, so item 13 is not satisfied.
### 14. Independent all-item review — FAIL / OPEN

This report provides a separate verdict/evidence record for every numbered item at the exact f125 SHA. Items 4 and 9 have unresolved findings; post-delivery review of the existing record also found item 13 stale/inconsistent with f125. Per the checklist, item 14 must remain unchecked until the findings are fixed, the record is reconciled, the affected controls are rerun, and two independent reviewers verify the final commit.
### 15. Paired request/response replay — PASS, scoped to implemented transports

`SyntheticReplay.RoundTrip` compares exact method, origin, escaped path, full repeated-query multimap, complete headers, and body before returning the paired status/headers/body; unexpected, duplicate, mismatched, and unconsumed exchanges fail. REST replay contains 26 ordered synthetic exchanges, GraphQL covers all seven generated documents, and event replay covers stream open then ping. Native HTTP/2 replay exercises both sides of the real pinned codec with a directive frame and verifies EOF/close/server-goroutine completion. The existing fixtures are clearly synthetic and `pkg/alexa/testdata` response-only examples are explicitly segregated. No SDK browser callback/redirect exists, so OAuth state and PKCE are explicitly left to the consuming application; this is the scoped N/A for those two clauses. Synthetic auth fixtures reuse the same device-registration identity across code pair, registration, and refresh requests.

### 16. Standalone CLI — PASS

`cmd/go-alexa` is a separate module that imports the SDK. Its offline tests exercise linking, refresh failure/success, export, discovery, explicit power control, player state, and cancellable event listening with stream/session cleanup. Credentials come from environment, stdin, or files; output files are exclusive and owner-restricted (POSIX 0600 / Windows protected current-user ACL). CLI help says secrets are not accepted as arguments or printed in normal output. Exact main CI ran Windows ACL/lifecycle tests and build/vet/module checks; nested CLI release run `37183841537` successfully installed the tagged CLI via public Go proxy and published `cmd/go-alexa/v0.4.0`. I independently ran `go install github.com/portpowered/go-alexa/cmd/go-alexa@v0.4.0` in an isolated GOPATH and confirmed `go-alexa --help`; a separate public-module consumer also passed.

## Publication and evidence scope

- Main CI: [run 37192237178](https://github.com/portpowered/go-alexa/actions/runs/37192237178), success at exact f125 SHA.
- Documentation build/deploy: [run 37192237182](https://github.com/portpowered/go-alexa/actions/runs/37192237182), success at exact f125 SHA; Pages deployment status is success at `https://portpowered.github.io/go-alexa/`.
- SDK public release: `v0.4.0`, commit `8fbb9dc1bfe6064594dd500d8a170a74fba9ec42`; [release record](https://github.com/portpowered/go-alexa/releases/tag/v0.4.0) and release workflow run `37183039192` succeeded.
- CLI public release: `cmd/go-alexa/v0.4.0`, commit `d6bc1d12c75705e0595ae8156ea223c249cd2e91`; [release record](https://github.com/portpowered/go-alexa/releases/tag/cmd/go-alexa/v0.4.0) and CLI release workflow run `37183841537` succeeded.
- The fixtures are synthetic only. `docs/verification.md` records maintainer-reported account testing as historical, without dates, operation list, or sanitized captures. No live account credentials or real provider traffic were used for this audit.
