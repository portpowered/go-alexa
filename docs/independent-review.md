# Independent review record

The checklist remains open. This record retains the latest completed independent
review; earlier audits are available in Git history. Fixes after the reviewed
commit require both reviewers to inspect the same final implementation commit
and verify its CI, publication, and public installation evidence.

## Reviewer 1 - independent audit at cb8ee3e617deeffe08c5ed87c4d52046b278fbe0

I did not implement the Alexa SDK or CLI. I reviewed a read-only Git archive at `C:\Users\andre\AppData\Local\Temp\alexa-r1-cb8-d4d42731ba934dae8d32161886d59581`, pinned against template `987b9c34a6b927472c21604617b6842a4238746b` and its client-design, website, and library-standards documents. The reviewed PR is #1, based on `main`, still draft at the time of review.

Exact-SHA CI run [37178111349](https://github.com/portpowered/go-alexa/actions/runs/37178111349) passed Linux `verify` and Windows `cli-windows`. The Linux job regenerated/checks GraphQL and wire artifacts, checked module metadata and formatting, ran pinned golangci-lint v2.3.0, then `make check` and non-generated coverage. Exact logs report 84.6% combined coverage (2,397/2,834), with public SDK 83.0%, API models 86.5%, GraphQL transport 94.2%, REST transport 83.8%, and generated dependency models 100% (3 statements).

Exact-SHA Documentation run [37178111384](https://github.com/portpowered/go-alexa/actions/runs/37178111384) passed its build and checked 44,910 internal links over 250 rendered pages. I downloaded artifact 11293199706 and inspected the 250 HTML files: all contained the implementation-derived label, including the root page. The Pages deploy job and Pages artifact upload were skipped because this was a pull request. Publication remains pending.

## Item verdicts

1. **PASS.** README, public packages, examples, and guides are Alexa-specific reusable library material; they contain no consuming-application adapter or rollout plan.

2. **PASS.** README and customer MDX guides cover operations, authentication, errors, configuration, transport injection, endpoint workflows, player state, events, and CLI usage. The examples match the public Client/Session API. The authentication guide explicitly assigns browser callback, state, and PKCE to the consuming app, because this SDK exposes code-pair registration rather than that browser flow. The site labels its checked-in contract implementation-derived, consistent with docs/verification.md.

3. **PASS.** README has Go version, CI, coverage, release, Go Reference, license, and Pages badges; URLs point to this repository or its reports. There are no template-repository values. The public latest release is still v0.3.2, so these badges do not establish publication of this candidate.

4. **OPEN - R1-F1.** The checked-in responsibility schemas, generated models/routes, GraphQL selection/decoder checks, model inventory, external HTTP/2 inventory, and route call-site tests cover the listed production contracts. The inventory contains 1,507 lines, including semantic projections and the two explicitly excluded legacy dependency configs. The exact CI generation/inventory/route checks pass. A new temporary probe against the exact archive exposed a named-result bypass in the handwritten wire-value gate:

   ```go
   func inventedValue() (result string) {
       result = "brandNewUnregisteredValue"
       return
   }
   func request() { _ = wire.Payload{Value: inventedValue()} }
   ```

   The gate returned nil. A second probe also returned nil:

   ```go
   func expose(input *wire.Payload) (result map[string]any) {
       result = input.Extra
       return
   }
   func request(input *wire.Payload) { unknown(expose(input)) }
   ```

   This lets a fixed library wire value or generated map provenance pass through a named result and escape inspection. In `tools/modelinventory/wire_returned_literals.go`, `inspectWireReturnedLiterals` examines only explicit `ReturnStmt.Results`; `tools/modelinventory/wire_aggregate_sources.go`, `generatedWireReturnedValue`, follows explicit return expressions only. Both probes ran in temporary test code inside the detached archive; no candidate source was changed. Add named-result assignment analysis and negative tests for both cases, retaining caller-open positive cases. This blocks full item-4 wire-model signoff.

5. **PASS.** Exact CI run 37178111349 is green on the reviewed SHA for Linux and Windows. The workflows pin Go 1.24.2 and golangci-lint v2.3.0; both lint configs specify literal `linters.default: all`. `make check` covers formatting, lint, vet, build, race tests, model inventory, OpenAPI bundle, and both Go module metadata checks. The Windows job runs CLI permission/lifecycle race tests, build, vet, and module checks. No lint findings are reported by CI.

6. **PASS - floor met; target short.** CI's non-generated coverage gate reports 84.6% (2,397/2,834), above the required 80% and below the 90% target. Package results are 83.0% `pkg/alexa`, 86.5% `pkg/alexaapimodels`, 94.2% GraphQL, 83.8% REST, and 100% for 3 dependency-model statements. The generated-code filter is implemented in `tools/coverage`.

7. **OPEN - R1-F1; public consumer pending.** The SDK and CLI use `pkg/alexa`, `pkg/alexaapimodels`, `pkg/dependencymodels`, transport-specific `pkg/dependencies/*`, and separate responsibility schemas. The named-result wire gate gap in item 4 prevents complete inventory signoff. The CLI is a nested module, but `cmd/go-alexa/go.mod` requires SDK v0.4.0 and retains `replace github.com/portpowered/go-alexa => ../..`. Public tags/releases expose only SDK v0.3.2 and earlier; `go list -m -versions` returned no CLI versions. Public-proxy SDK/CLI consumer and install checks remain unrun until release.

8. **PASS.** `alexa.NewClient(opts ...Option)` applies validated options and defaults. Account-specific credentials are passed through `NewSession`, not reusable client configuration. Tests cover nil/conflicting transport options and session option validation.

9. **PASS.** The reusable Client holds endpoint/transport configuration; credentials and event connections belong to Session. Session/connection close cancels owned stream work, and stream errors are returned through the connection API. Tests cover session isolation and event cleanup.

10. **PASS.** REST, GraphQL, and event traffic each accept injected HTTP clients/transports. `WithEventTransport` allows an injected `http2.Transport` whose `DialTLSContext` can return an offline `net.Conn`. `pkg/alexa/http2_native_replay_test.go:20` exercises the real HTTP/2 codec through `net.Pipe`, matches paired requests/responses, reads an event, completes keepalive, asserts replay consumption, calls Close, and waits for server shutdown.

11. **PASS.** Refresh and cookie exchange are explicit operations. `Session.RefreshAccessToken` returns the response without mutating the access or refresh token; the authentication guide tells callers to store the returned credentials and explicitly install an access token with `SetAccessToken`.

12. **PENDING PUBLICATION.** All customer guides are MDX under `docs/guides/`, the generated reference and guides render, and the exact Docs workflow passed the 44,910-link check. The reviewed PR run skipped the main-only Pages deployment, so the candidate pages are not yet live.

13. **OPEN - review-record consolidation and release-note publication.** I reviewed every tracked Markdown/MDX file, the README, guide navigation, and rendered artifact. README is caller-focused; contributor-only schemas, inventory, verification, release notes, fixture notes, and checklist are outside guide navigation. Guides are operation-focused and distinguish synthetic/implementation-derived behavior. However, `docs/independent-review.md` contains only an R1 placeholder and an R2 review of stale commit c1d459e; it is not a current consolidated record for cb8. The v0.4 release note and its external guide destinations have not been generated or published yet.

14. **OPEN.** There is no R1 all-16 review for cb8 in the repository record; its R1 heading says review is pending. The only completed R2 section reviews c1d459e, not cb8. R1-F1 and publication/install items remain open, so two-reviewer signoff is unavailable. This report is a temporary reviewer artifact for root consolidation, not a change to the reviewed tree.

15. **PASS for the implemented non-OAuth callback flows; caller-managed OAuth callback is N/A.** `pkg/testing.SyntheticReplay` matches method, origin, escaped path, repeated query, headers, and body before returning the paired status/headers/body; tests reject mismatches, duplicate/unexpected calls, and unconsumed exchanges. REST, GraphQL, CLI, and event tests use paired ordered exchanges. The synthetic REST fixture has full code-pair, CSRF, and OTP registration request expectations, including the paired email/OTP challenge; CLI auth has paired requests. The native HTTP/2 test uses the frame codec and verifies close as noted above. All current volatile values are fixed synthetic values. This SDK does not implement the browser OAuth callback/PKCE exchange; the guide assigns that responsibility to the consuming app.

16. **OPEN - public CLI install pending.** The standalone CLI is a separate `cmd/go-alexa` module using the SDK. It has help, JSON output, nonzero failures, credential file/stdin/environment inputs, explicit export, code-based linking and refresh, endpoint discovery/control, player state, and cancellable event cleanup; offline paired CLI tests and Windows/Linux CI pass. Its local replace and required unpublished SDK v0.4.0 prevent a clean public `go install`/consumer check. The tagged release workflow is present but has not run for this candidate.

## Finding and disposition

- **R1-F1 - OPEN at cb8ee3e (items 4 and 7).** Named-result helper assignments evade fixed-wire-string validation and generated-map provenance. Exact accepted probes are recorded under item 4. The earlier c1 package-global alias and explicit helper-return probes now have dedicated coverage in `tools/modelinventory/wire_global_sources_test.go`, but named results remain unhandled at this SHA. Re-run the focused negative tests and full gates on the next frozen commit, then have both reviewers verify that exact commit.


## Reviewer 2 — independent audit at c1d459e225809f3064b27aa0fe3e39ad60cc45b0

I did not implement this SDK or CLI. I reviewed a git archive of the exact c1d459e225809f3064b27aa0fe3e39ad60cc45b0 tree against the checklist pinned to template 987b9c34a6b927472c21604617b6842a4238746b and its linked standards. Exact-SHA CI [run 37175399314](https://github.com/portpowered/go-alexa/actions/runs/37175399314) completed successfully for Linux verify and Windows CLI. Exact-SHA Documentation [run 37175399237](https://github.com/portpowered/go-alexa/actions/runs/37175399237) built 250 pages and checked 44,910 internal links; I downloaded artifact 11292867058 and independently confirmed the implementation-derived label on all 250 HTML pages. The Pages deploy job did not run for this review build, so live publication remains open. From the archive, root and CLI race tests, model-inventory check, route check, and OpenAPI bundle check also passed. Two blind model-gate probes ran through a temporary Go overlay; no candidate source files were modified.

### Item verdicts

1. **PASS.** The public package is a reusable Alexa client. The README, examples, and customer guides describe the provider library and do not add a consuming-application adapter or rollout plan.

2. **PASS.** The README and customer MDX guides cover supported operations, code-based linking, token refresh, errors, transport injection, endpoint discovery and control, player state, events, and CLI use. They match the exported Client, Session, and credential APIs. The authentication guide assigns browser callback state and PKCE to the caller because the SDK does not implement that OAuth callback.

3. **PASS.** The README shows Go version, CI, coverage, release, Go Reference, license, and documentation badges. The badge destinations use this repository or its live reports; I found no template repository values.

4. **OPEN — R2-F10 and R2-F11 (items 4, 7).** The checked-in responsibility schemas, generated routes/models, model inventory, GraphQL evidence, and source checks cover the enumerated HTTP, GraphQL, event, and external HTTP/2 edges. The exact-tree inventory, route, and bundle checks passed, and the test suite contains negative model/key, decoder, enum-binding, mutation, and route controls. Two independent wire-construction bypasses remain:

       var saved map[string]any
       func mutate(caller string) { saved["brandNewCrossFunctionKey"] = caller }
       func retain(input *wire.Payload) { saved = input.Extra }
       func send(input *wire.Payload, caller string) {
           retain(input)
           mutate(caller)
           _, _ = json.Marshal(input)
       }

   A temporary overlay test showed the gate accepts this package-global map escape and later fixed key, even though the generated map is retained from and serialized with a generated model. Also, a helper-returned fixed value is not followed into a generated wire field: inventedValue returns "brandNewUnregisteredValue", and constructing wire.Payload{Value: inventedValue()} passes the gate. These are library-selected wire values, not caller-defined open input. Both gaps require negative tests and keep items 4 and 7 open.

5. **PASS.** Exact-SHA CI run 37175399314 passed the Linux verify and Windows CLI jobs. The Linux job regenerated and checked artifacts, checked module metadata and formatting, ran pinned golangci-lint v2.3.0 with literal linters.default: all, then build, race tests, and coverage. The Windows job passed CLI credential-permission and lifecycle tests, build, vet, and module checks.

6. **PASS — floor met; target short.** CI reports 84.6% combined non-generated coverage (2,397/2,834 statements), above the required 80% floor and below the 90% target. Package coverage is 83.0% for pkg/alexa, 86.5% for pkg/alexaapimodels, 94.2% for GraphQL, 83.8% for REST, and 100% over three statements for dependency models.

7. **OPEN — R2-F10 and R2-F11; public consumer pending.** The public SDK, generated dependency models, and transport packages use the required package boundaries and responsibility-specific schemas. The two model-construction gaps in item 4 prevent complete wire-model signoff. The CLI module requires SDK v0.4.0 but its go.mod still contains a local replace to the repository root. The public Go proxy and remote tags did not contain v0.4.0 or a cmd/go-alexa module tag; go list returned unknown revision for both requested public versions. A clean public-proxy SDK/CLI consumer build is therefore not yet proven.

8. **PASS.** Client construction uses validated functional options and sensible defaults. Account credentials are passed when creating a Session rather than stored in reusable client configuration.

9. **PASS.** Account tokens and event connections belong to explicit Session and connection objects. Callers can read and update tokens explicitly, observe close errors, and close streams; tests cover isolation and cleanup.

10. **PASS.** REST and GraphQL use injected HTTP clients; event traffic accepts injected HTTP/event transports. The native HTTP/2 replay substitutes a net.Pipe connection beneath the pinned transport and exercises paired framed traffic, keepalive, consumption, and close without live provider access.

11. **PASS.** Token refresh and cookie exchange are explicit operations that return credentials to the caller. Refresh does not silently replace session state; callers choose when to install and persist refreshed values.

12. **OPEN.** Documentation run 37175399237 built and checked the artifact, but this review run did not deploy to GitHub Pages. The exact candidate's live documentation publication and destinations are unverified.

13. **OPEN — current review record needs consolidation.** I reviewed the README, all tracked Markdown/MDX files, guide navigation, and the rendered artifact. Customer guides are concise, operation-focused, distinguish implementation-derived evidence, and remain separated from maintainer inventories and verification details. However, the contributor review record still contains all-16 sections for earlier SHAs and no Reviewer 1 section for c1; those stale review records need consolidation into the current exact-SHA record before release.

14. **OPEN.** This section is Reviewer 2's all-16 audit at c1d459e. Reviewer 1 has not reviewed this exact candidate. R2-F10 and R2-F11, the public SDK/CLI consumer and install proof, and Pages publication remain open; a two-reviewer final signoff is not available.

15. **PASS.** SyntheticReplay matches method, origin, escaped path, repeated query values, full headers, and body before releasing the paired response; it rejects mismatches, duplicates, unexpected calls, and unconsumed pairs. REST, GraphQL, HTTP/2, and CLI tests use offline paired exchanges or ordered frame transcripts. The authentication flow here is code-pair registration and explicit token operations; callback state and PKCE are caller-managed because this SDK does not implement that callback or verifier exchange.

16. **OPEN.** The standalone CLI is in cmd/go-alexa as a separate module and consumes the public SDK. It provides help, JSON results, nonzero failures, credential files/stdin/environment inputs, explicit export, linking and refresh, endpoint discovery, player read/control, and cancellable event listening with cleanup; offline paired command tests and Linux/Windows CI are present. The nested go.mod still has a local replace, the required SDK v0.4.0 is not published, no CLI module tag exists, and no public go install at a tag or separate public consumer install has been demonstrated.

### Findings and disposition at c1d459e

- **R2-F5 — RESOLVED at 572f2a1 (item 4).** Local declarations and prior assignments are indexed by lexical declaration and name; negative probes cover aliases reassigned to unregistered values.
- **R2-F6 — RESOLVED in this candidate (item 4).** Direct and aliased indexed writes, zero-value map aliases, collection receivers, parenthesized receivers, and intermediate index paths have source checks and negative tests.
- **R2-F7 — RESOLVED at cc35d238 (items 4, 7).** SDL bindings validate control-field enums, and production control values use schema-generated constants.
- **R2-F8 — RESOLVED at cc35d238 (item 4).** The video search phrase format has a behavior schema owner and a checked construction site.
- **R2-F9 — RESOLVED in this candidate (items 4, 7).** Recursive receiver-key inspection and tests reject unregistered intermediate keys through direct and aliased indexed paths.
- **R2-F10 — OPEN at c1d459e (items 4, 7).** Generated map provenance stored in a package-scope variable is not followed across functions; a later fixed-key mutation before serialization is accepted. Add a package-scope variable-object provenance gate and the exact negative probe above.
- **R2-F11 — OPEN at c1d459e (items 4, 7).** Fixed string values returned by local helpers are not traced when used in generated wire fields. Add return-value provenance checks and a negative probe for the helper-return construction above.
