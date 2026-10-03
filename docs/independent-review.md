# Independent review record

Status: pending final implementation and independent verification.

The previous review signed off complete model generation from selected
generated files and transport call sites. It did not inventory every exported
dependency model or anonymous nested object. The current migration must resolve
that gap and verify every item in [the checklist](template-checklist.md) at its
final commit. An earlier CI pass cannot close newly identified omissions.

Record both independent reviewers, the exact implementation SHA, CI and
documentation runs, each reviewer's separate verdicts for all 16 items, and
every finding's disposition here.
Keep the complete model inventory with the generation tooling; it must map
schema components, generated definitions, commands, and actual use, including
compatibility projections. Search independently for definitions missing from
that inventory. Keep the review in contributor material rather than customer
guide navigation.

## Reviewer 1 — independent audit at b0e4a77d8fac48558aa84130d558c51d33ddd531

I did not implement this SDK or CLI. I reviewed the committed tree at this exact SHA against every item in the checklist and the shared client-design and website standards pinned at `987b9c34a6b927472c21604617b6842a4238746b`, plus the repository verification standard. The review record also keeps the prior finding dispositions from earlier snapshots.

Exact-SHA evidence is [CI run 37121074363](https://github.com/portpowered/go-alexa/actions/runs/37121074363) and [Documentation run 37121074352](https://github.com/portpowered/go-alexa/actions/runs/37121074352), both successful. CI passed generation/model/route gates, root and nested module metadata, formatting, pinned golangci-lint v2.3.0 with all linters enabled, build and race tests, the 80% coverage floor, and Windows CLI race/build/vet/module checks. CI reports 84.5% combined non-generated coverage (2396/2834 statements). The Docs run built the site, checked 44,910 internal links across 250 pages, and uploaded artifact `documentation-site` (11273024351). I downloaded and inspected it; all 250 rendered HTML pages carry the implementation-derived label. The PR Pages deploy job was skipped. Items 2, 4, 7, 12, 14, 15, and 16 remain open for the reasons below.

### Item verdicts

1. **PASS.** `pkg/alexa` is a reusable provider client. README, examples, and customer guides describe the library without a consuming-application adapter or rollout plan.

2. **OPEN — R1-F17.** README and customer MDX guides cover the supported code-pair linking, token refresh, errors, transports, endpoints, controls, and event lifecycle. The exported Go package comment in `pkg/alexa/client.go` nevertheless says the client supports “full OAuth flows.” At this SHA, client operations implement code-pair registration and token refresh; `/ap/oa` only appears as a caller-managed regional URL, and its schema says callers add authorization query values and that the path is implementation-derived and not provider-verified. I found no callback, state-binding, or PKCE logic. Qualify the public package claim as caller-managed authorization or provide and verify the claimed full flow before presenting it as supported behavior.

3. **PASS.** README badges cover Go version, CI, coverage, release, Go Reference, license, and documentation; targets point to project reports and public destinations. No template repository values remain.

4. **OPEN — R1-F18, R1-F19.** At b0 the receiver-key fix recursively checks every indexed segment and follows local declaration/earlier-assignment aliases through selectors, parentheses, dereferences, slices, and type assertions. Tests reject direct and aliased intermediate map keys and retain caller/generated-key positive controls. I independently searched the production model declarations, exported dependency types, anonymous/nested shapes, primitive/constants, custom decoders, and call sites, and cross-checked `docs/model-inventory.md`; this b0 diff adds no new production SDK wire model or network edge. Two source-gate gaps remain. First, b0 parses each Go file separately, so a package-level `const rawKey = "brandNewKey"` in a sibling file used as `input.Extra[rawKey] = caller` has no resolved declaration in the use file and no literal for the checker to inspect (R1-F18). Second, b0 does not inspect map mutations performed through helper calls or `maps.Copy`; for example, `mutate(input.Extra, caller)` is not connected to `func mutate(extra map[string]any, caller string) { extra["brandNewKey"] = caller }`, and `maps.Copy(input.Extra, map[string]any{"brandNewKey": caller})` is not a direct assignment/composite on a generated type (R1-F19). Add negative tests for each while preserving open caller values and generated keys.

5. **PASS.** Exact-SHA CI run 37121074363 passed the blocking `verify` job and `cli-windows`. Verify ran generation/drift checks, module metadata, formatting, pinned all-linter lint, build/race tests, and the coverage gate; Windows passed credential permission and CLI lifecycle tests, build/vet, and module metadata. This confirms the exact candidate's blocking CI.

6. **PASS — 80% floor met; 90% target not met.** Combined non-generated coverage is **84.5% (2396/2834 statements)**. Package coverage is `pkg/alexa` 83.0% (1281/1544), `pkg/alexaapimodels` 86.5% (90/104), GraphQL 94.2% (277/294), REST 83.8% (745/889), and dependency models 100% over three non-generated statements. Generated-code exclusion and remaining uncovered behavior are reported; the 90% target remains short.

7. **OPEN.** The public API, generated dependency models, and REST/GraphQL transport implementations are separated under `pkg/alexa`, `pkg/dependencymodels`, and `pkg/dependencies/`; the model inventory records schemas, generated types, generator commands, and uses. R1-F18 and R1-F19 leave source-gate gaps in fixed-key and helper-map mutation coverage. A clean public-proxy consumer check is also pending: `cmd/go-alexa/go.mod` requires unpublished SDK v0.4.0 and retains `replace github.com/portpowered/go-alexa => ../..`.

8. **PASS.** `alexa.NewClient` uses validated functional options and sensible defaults. Credentials are supplied to `NewSession`, not reusable client configuration.

9. **PASS.** Client configuration is reusable while account credentials and event connections belong to explicit sessions. Session close, token access, and token updates are caller-visible; session isolation tests cover multiple accounts.

10. **PASS.** REST, GraphQL, and event HTTP transports are injectable. `TestEventNativeHTTP2PairedReplay` drives the pinned HTTP/2 codec over an injected `net.Pipe`, matches the framed request and response, closes the event connection, and waits for the server to stop. The network inventory shows no WebSocket, MQTT, RTC signaling, or other non-HTTP socket edge used by this library.

11. **PASS.** Token refresh and cookie exchange are explicit operations. `RefreshAccessToken` returns caller-visible credentials without mutating session state; callers install a new access token explicitly, and the authentication guide assigns token storage and renewal to the caller.

12. **OPEN.** Documentation run 37121074352 built the site and checked all 44,910 internal links across 250 rendered pages. I inspected the root, authentication/CLI/upgrading guides, and generated REST and GraphQL pages in artifact 11273024351. The PR deploy job was skipped; Pages publication and live destinations are not yet proven.

13. **PASS — reviewed artifact, pre-publication.** I reviewed README and tracked Markdown/MDX files. Customer navigation stays guide-focused; inventories and maintainer evidence remain contributor material; synthetic fixtures and implementation-derived contracts are labeled; release guide destinations match the rendered site. The artifact visibly labels all 250 pages and passes the site-wide link check. Publication remains item 12.

14. **OPEN.** The review record has a separate R1 section for b0; Reviewer 2 must record and verify its own all-16 section at the same final SHA. Findings R1-F17/R1-F18/R1-F19 and the remaining public consumer, Pages, and CLI release proof are still open, so this item cannot be checked.

15. **OPEN — R1-F16 and R1-F17.** `pkg/testing.SyntheticReplay` matches method, origin, escaped path, repeated query values, full headers, and body before returning the paired status/headers/body; it rejects mismatches, unexpected/duplicate calls, and incomplete ordered consumption. The auth fixtures bind the synthetic device serial across code-pair, registration, and refresh requests; CSRF and token-exchange fields are exact, and no volatile field is wildcard-normalized. Event replay matches the HTTP/2 handshake/frame exchange, and tests assert connection/server or CLI response-body cleanup after close/cancellation. However, `RegisterWithEmailPassword` exposes an OTP-required public challenge, but at b0 its `MissingRequiredAuthenticationData`/`OTP` response is synthesized by a response-only RoundTripper test; that test does not match the complete outbound request, and the challenge pair is absent from `alexa-rest.json`. Add the challenge as an exact paired request/response and assert consumption. OAuth state/PKCE is not implemented; the public claim of full OAuth flows is unresolved under R1-F17. No RTC/signaling exchange is used by this library.

16. **OPEN.** `cmd/go-alexa` is a separate module with offline paired command/auth/event tests, secure credential-file handling, explicit export, JSON output, nonzero failures, cancellation, and cleanup; exact-SHA Linux and Windows CI passed. The CLI still has a development `replace` to the unpublished SDK v0.4.0. No public SDK consumer compile, CLI module tag, or clean `go install ...@version` result is available.

### Findings and disposition at b0e4a77

- **R1-F9 — RESOLVED at 572f2a1 (items 4, 7).** Zero-value local map aliases are covered by negative tests and source-local declaration/assignment tracking.
- **R1-F10 — RESOLVED at cc35 (items 4, 7).** SDL field bindings for lock, toggle, thermostat scale, and thermostat mode are checked with deletion negatives.
- **R1-F11 — RESOLVED at cc35 (items 4, 7).** Primitive-use inventory scans `cmd`; terminal UI wire values use generated constants while ordinary labels/caller inputs remain valid.
- **R1-F12 — OPEN (items 12, 13).** Build artifact and all-page links are reviewed; live Pages deployment is not verified.
- **R1-F13 — OPEN (items 7, 16).** SDK v0.4.0 and the CLI module remain unpublished; verify both public-proxy consumer/install checks after tagging.
- **R1-F14 — OPEN (item 14).** Reviewer 2 has not yet re-reviewed this exact SHA; both reviewers must verify the same final candidate.
- **R1-F15 — RESOLVED at b0e4a77 (items 4, 7).** Recursive receiver-key checks and negative probes now catch the multiply indexed intermediate key and the equivalent alias initializer while preserving caller and generated keys.
- **R1-F16 — OPEN at b0e4a77 (item 15).** The OTP-required email registration response is tested without a request-matched pair or replay-consumption assertion.
- **R1-F17 — OPEN at b0e4a77 (items 2, 15).** The public package comment claims full OAuth flow support, but the client has no callback/state/PKCE implementation and the `/ap/oa` contract is explicitly caller-managed and unverified.
- **R1-F18 — OPEN at b0e4a77 (items 4, 7).** The exact b0 tree has no package-wide wire-source resolver or sibling-file test. Its construction checker parses one file at a time and does not resolve a package-level string constant from another file, so a library-defined raw wire key can evade the fixed-value scan.
- **R1-F19 — OPEN at b0e4a77 (items 4, 7).** The checker does not inspect map mutations passed through helper calls or `maps.Copy`; direct helper and copy probes can add unregistered keys to a generated wire map without a generated-type composite or direct assignment at the call site.

An interim inspection appeared to find a package-wide resolver and a two-file test, but those files were untracked additions in the moving worktree and absent from b0 (`git ls-tree -r b0e4a77 -- tools/modelinventory`); they are not evidence for this exact-SHA review.

Earlier Reviewer 1 findings for generated provider constants/bindings, visible implementation-derived labels, exact-SHA lint/Windows ACL failures, the legacy capability shape, and typed `ParseEvent` decoding were resolved before cc35. Item 12 Pages publication, item 13 post-publication review, item 14 final two-reviewer agreement, public SDK/CLI installation, and the b0 findings above remain open.

## Reviewer 2 — independent audit at `572f2a1da74f187c781d011c2191ec710ae403e0`

I did not implement this SDK or CLI. I reviewed the committed source at this exact SHA against all 16 checklist items and the shared standards pinned at `8fd452025259f9ad724498e15787948edbd06784`. Exact-SHA checks are [CI run 37120299881](https://github.com/portpowered/go-alexa/actions/runs/37120299881) and [Documentation run 37120299912](https://github.com/portpowered/go-alexa/actions/runs/37120299912), both successful. The documentation artifact `documentation-site` (artifact 11272683685) contains 250 rendered HTML pages; I confirmed each visibly says “implementation-derived,” and `tools/check_site_links.py` checked 44,910 internal links. The PR deploy job was skipped, so publication remains open.

### Item verdicts

1. **PASS.** `pkg/alexa` is a provider client. The README, examples, and customer guides explain the reusable SDK without a consuming-application adapter or rollout plan.

2. **PASS.** The README and customer MDX guides cover exported client/session use, authentication, errors, transport injection, endpoint workflows, player state, events, and CLI use. The exact-SHA artifact visibly labels every rendered page “implementation-derived”; `docs/verification.md` distinguishes that evidence from provider-issued specifications, synthetic fixtures, and historical references.

3. **PASS.** The README has Go version, CI, coverage, release, Go Reference, license, and documentation badges. Targets point to the project reports and live destinations; I found no template repository values.

4. **OPEN — R2-F9.** REST, GraphQL, behavior, event, and external HTTP/2 responsibilities have checked-in schemas, generated artifacts, route/model gates, and paired replay checks. I cross-checked the model inventory and the source-side generation gates, including exported dependency models, nested/anonymous shapes, semantic constants, custom decoders, and call sites. At this SHA the wire-construction gate still misses an intermediate key in a multiply indexed open map. For example, `input.Extra["brandNewIntermediate"][alexamodels.RegisteredValue] = caller` resolves `input.Extra` to a generated wire model, but `wireAssignmentParts` extracts only the outermost index key and `generatedWireReceiver` follows the inner receiver without validating that inner index. The raw intermediate key is therefore not rejected. A local alias such as `alias := input.Extra["brandNewIntermediate"]; alias[caller] = caller` has the same gap. This is library-selected wire data; add recursive key checks for every index in the receiver chain and negative tests while retaining caller-provided and generated-key positive controls. Items 4 and 7 remain open.

5. **PASS.** Exact-SHA CI run 37120299881 passed `verify` and `cli-windows`. The blocking verify job passed generation/drift checks, module metadata, formatting, pinned golangci-lint v2.3.0 with literal `linters.default: all`, build, race tests, and coverage. Windows passed CLI credential ACL/lifecycle checks, build/vet, and nested-module metadata.

6. **PASS — 80% floor met; 90% target not met.** The race-enabled CI coverage gate reports 84.6% combined non-generated coverage (2397/2834 statements). Package results are `pkg/alexa` 83.0%, `pkg/alexaapimodels` 86.5%, GraphQL 94.2%, REST 83.8%, and dependency models 100.0% over three non-generated statements. Generated code is excluded and the gate names the remaining uncovered behavior.

7. **OPEN.** The public API, dependency models, and transport code are separated in `pkg/alexa`, `pkg/dependencymodels`, and `pkg/dependencies/`, with responsibility-specific schemas and `docs/model-inventory.md`. R2-F9 prevents complete wire-generation signoff. A clean public-proxy consumer compile is also unproven: `cmd/go-alexa/go.mod` requires unpublished v0.4.0 and still has `replace github.com/portpowered/go-alexa => ../..`.

8. **PASS.** `alexa.NewClient` uses validated functional options and sensible defaults. Account credentials are supplied to `NewSession`, not retained in reusable client configuration.

9. **PASS.** Account tokens and event connections belong to explicit `Session` objects. `Session.Close` is explicit and idempotent, and token reads/updates are caller-visible operations.

10. **PASS.** REST, GraphQL, and event HTTP transports are injectable. `TestEventNativeHTTP2PairedReplay` exercises the pinned HTTP/2 framing over `net.Pipe` without provider access or credentials; the network inventory has no WebSocket, MQTT, RTC, or other non-HTTP socket edge.

11. **PASS.** Refresh and cookie exchange are explicit operations that return caller-visible credentials. `RefreshAccessToken` does not mutate session state; `SetAccessToken` is separate, and the auth guide documents caller-owned token storage and renewal.

12. **OPEN.** Documentation run 37120299912 built the site and checked 44,910 internal links across 250 pages; I inspected the artifact. Its PR Pages deploy job was skipped, so publication and live destinations are not proven.

13. **PASS — reviewed build artifact, pre-publication.** I reviewed the README and tracked Markdown/MDX documentation at the previous exact source snapshot; the only documentation changes since then are the independent-review record, which I reviewed at 572f. Customer navigation remains guide-focused, maintainer audits stay outside it, synthetic fixtures remain distinct from provider evidence, and release guide destinations match the rendered site. The exact-SHA artifact has the site-wide evidence label and passed the all-page link checker. Live publication remains item 12.

14. **OPEN.** Both reviewers have reviewed this SHA, and the review record now has separate all-16 sections. R2-F9, the public consumer/CLI publication work, and Pages publication remain open; both reviewers must recheck the same final frozen SHA after fixes and release evidence.

15. **PASS.** `pkg/testing.SyntheticReplay` matches method, origin, escaped path, repeated query values, headers, and body before returning paired status/headers/body. Its tests reject mismatches, duplicate/unexpected calls, and unconsumed exchanges. The CLI and native HTTP/2 use paired offline request/response or ordered-frame tests; fixtures are labeled synthetic.

16. **OPEN.** `cmd/go-alexa` is a separate module with offline paired command/lifecycle tests, explicit credential inputs/export, JSON output, nonzero failures, cancellation, cleanup, and blocking Linux/Windows CI. However, SDK v0.4.0 and a nested CLI module tag are unpublished; the CLI still uses a local replace and no clean public `go install ...@version` has been shown.

### Findings and disposition at 572f2a1

- **R2-F5 — RESOLVED at 572f (item 4).** Local declarations and prior assignments are indexed by lexical declaration/name. The negative probes cover a caller-valued alias later reassigned to a novel fixed value and a declared local assigned a novel value.
- **R2-F6 — PARTLY RESOLVED at 572f (item 4).** Direct indexed writes, single-index key/value changes, zero-value map aliases, collection receivers, and parenthesized receivers now have source checks and negative probes. The multiply indexed intermediate-key escape remains as R2-F9.
- **R2-F7 — RESOLVED at cc35 (items 4, 7).** Control-field bindings validate the lock, toggle, thermostat scale, and thermostat mode fields against their SDL enums; the primitive-use gate scans `cmd` and production wire construction uses generated state constants.
- **R2-F8 — RESOLVED at cc35 (item 4).** The video search phrase format is represented by behavior schema metadata and checked at its construction site.
- **R2-F9 — OPEN at 572f (items 4, 7).** Intermediate keys in multiply indexed generated open maps, including keys reached through local aliases, are not inspected by the source gate. Recheck the new recursive receiver-key gate and its negative and positive controls at the next frozen SHA.
