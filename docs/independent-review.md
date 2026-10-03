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

## Reviewer 1 — independent audit at 572f2a1da74f187c781d011c2191ec710ae403e0

I did not implement this SDK or CLI. I reviewed the committed tree at this exact SHA against all 16 checklist items and the shared standards pinned at 8fd452025259f9ad724498e15787948edbd06784. The review record also notes the remaining Reviewer 2 section cleanup; findings below describe the committed source at 572f2a1.

Exact-SHA evidence is [CI run 37120299881](https://github.com/portpowered/go-alexa/actions/runs/37120299881) and [Documentation run 37120299912](https://github.com/portpowered/go-alexa/actions/runs/37120299912), both successful. I downloaded and inspected the documentation artifact. Item 4 remains open because the generated-wire gate misses a nested indexed key in a generated wire map; publication and consumer-release proof are also pending.

### Item verdicts

1. **PASS.** pkg/alexa is a reusable provider client. README, examples, and customer guides describe the library and do not include a consuming-application adapter or rollout plan.

2. **PASS.** README and the customer MDX guides document exported client/session operations, authentication, errors, transport injection, endpoint workflows, player state, events, and CLI use with SDK types. The exact-SHA Docs artifact has 250 rendered HTML pages, each visibly labeled “implementation-derived.” docs/verification.md distinguishes implementation-derived behavior from provider-issued specifications, synthetic fixtures, and historical maintainer reports.

3. **PASS.** README badges cover Go version, CI, coverage, release, Go Reference, license, and documentation. Their targets point to this module, its reports, release page, reference, license, and site; I found no template repository badge values.

4. **OPEN — R1-F15.** REST, GraphQL, behavior, event, and external HTTP/2 contracts are in checked-in schemas and feed generation and the model/route gates. I independently searched exported dependency models, anonymous/nested objects, primitive constants and projections, custom decoders, and actual construction/consumption sites, and cross-checked docs/model-inventory.md and the inventory tooling. The 572 update closes the prior zero-value map-alias and parenthesized/indexed-receiver cases: generatedWireIdentifier follows earlier assignments by declaration/name, and negative probes cover zero-value map/pointer aliases, slice/index receivers, nested indexed mutation, and parenthesized receivers. A nested-index case still passes the gate:

       input.Extra["brandNewIntermediate"][alexamodels.RegisteredValue] = caller

   wireAssignmentParts extracts only the outer IndexExpr key, while generatedWireReceiver follows the inner IndexExpr.X for provenance without checking its Index. The raw intermediate map key is therefore never inspected. The same gap applies when code first assigns alias := input.Extra["brandNewIntermediate"] and later writes alias[caller] = caller: the alias provenance is followed, but the initializer index is not checked. I independently confirmed these paths from the committed traversal. Add recursive key validation for every IndexExpr in the receiver chain and negative tests for direct nesting and aliases. Caller-defined open values and generated keys must continue to pass. Until then, complete generation and negative-control coverage are not proven.

5. **PASS.** Exact-SHA CI run 37120299881 passed verify and cli-windows. Verify passed generated client/model checks, module metadata, formatting, pinned golangci-lint v2.3.0 with all linters enabled, build, race tests, and coverage. The Windows job passed credential ACL/lifecycle tests, CLI build/vet, and nested module metadata.

6. **PASS — 80% floor met; 90% target not met.** CI reports combined non-generated coverage of **84.6% (2397/2834 statements)**. Per-package totals are pkg/alexa 83.0%, pkg/alexaapimodels 86.5%, GraphQL 94.2%, REST 83.8%, and dependency models 100.0% over its three non-generated statements. Generated code is excluded and the coverage report identifies remaining uncovered behavior. The stated 90% target is still short.

7. **OPEN.** Public API, dependency models, and transport implementations are separated under pkg/alexa, pkg/dependencymodels, and pkg/dependencies/; API responsibilities have separate schemas and docs/model-inventory.md records generated definitions and uses. R1-F15 leaves a source-gate hole in open-map mutation coverage. A separate public-proxy consumer compile is also not evidenced: cmd/go-alexa/go.mod still requires unpublished SDK v0.4.0 via a development replace to the repository root.

8. **PASS.** alexa.NewClient uses validated functional options with sensible defaults (pkg/alexa/client.go, client_options_session_test.go). Account tokens are supplied to NewSession, not stored in reusable client configuration.

9. **PASS.** Client holds reusable endpoint/transport configuration, while account credentials and event connections belong to explicit Session objects. Session close is explicit and idempotent; token access/update are visible operations, with isolation tests for multiple sessions.

10. **PASS.** REST, GraphQL, and event HTTP transports are injectable; the event API accepts an injected RoundTripper. The pinned golang.org/x/net HTTP/2 contract is in api/external/http2.yaml; TestEventNativeHTTP2PairedReplay exercises framed requests and responses through net.Pipe without provider network or credentials. I found no WebSocket, MQTT, RTC, or other non-HTTP socket edge used by this library.

11. **PASS.** Refresh and cookie exchange are explicit operations that return caller-visible credentials. RefreshAccessToken does not mutate session state; SetAccessToken is separate. The auth guide and tests leave token storage and renewal with the caller.

12. **OPEN.** Documentation run 37120299912 built the site, checked 44,910 internal links across 250 rendered pages, and uploaded a reviewable artifact. I inspected its root, CLI and upgrading guides, sendEndpointInterfaceMessage and queryRestEndpoints REST operations, and listendpoints GraphQL operation. The PR deploy job was skipped, so successful Pages publication and live destinations are not proven.

13. **PASS — reviewed build artifact, pre-publication.** I reviewed every tracked Markdown/MDX file and the exact-SHA rendered artifact. Customer navigation is guide-focused; contributor audits stay out of customer navigation; synthetic fixtures and historical reports are distinguished; release guide destinations match the rendered pages. The artifact carries the implementation-derived label across all pages and passed the all-page link check. The record now contains one R1 and one R2 all-16 section at 572. Live publication remains item 12.

14. **OPEN.** Both reviewers have now written separate all-16 sections at 572f2a1. R1-F15/R2-F9 and the public consumer, CLI publication, and Pages publication work remain open. Both reviewers must recheck the same final frozen SHA after those findings are fixed and release evidence is available.

15. **PASS.** pkg/testing.SyntheticReplay matches method, origin, escaped path, repeated query values, headers, and body before returning the paired response status/headers/body. Tests reject mismatches, duplicate and unexpected calls, and unconsumed exchanges, and assert ordered consumption where it matters. Replay pairs are labeled synthetic; response-only model examples are separately identified and are not presented as replay evidence.

16. **OPEN.** cmd/go-alexa is a separate module with offline paired command/lifecycle tests, explicit credential inputs/export, machine-readable output, cancellation, cleanup, and passing Linux/Windows CI. The guide correctly makes installation conditional on publication. SDK v0.4.0 and a CLI module tag remain unpublished; the nested module still uses a local replace, and no clean public go install ...@version has been demonstrated.

### Findings and disposition at 572f2a1

- **R1-F9 — RESOLVED at 572f2a1 (items 4, 7).** The prior zero-value local map alias sequence is now covered by negative tests. The gate indexes assignments by local declaration and name and follows earlier assignments through the generated map source.
- **R1-F10 — RESOLVED at cc35 (items 4, 7).** tools/modelinventory/control_field_bindings.go validates the SDL binding for LockPayload.lockState, TogglePayload.toggleState, ThermostatSetpoint.scale, and ThermostatModePayload.thermostatMode. Negative tests remove each field reference and require rejection.
- **R1-F11 — RESOLVED at cc35 (items 4, 7).** The SDK primitive-use inventory scans cmd as well as pkg; terminal UI wire construction uses generated lock/toggle state constants, with tests retaining ordinary UI labels and caller inputs as valid.
- **R1-F12 — OPEN (items 12, 13).** The Docs build and rendered artifact are verified, but PR Pages deployment was skipped. Verify live publication and destinations after deploy.
- **R1-F13 — OPEN (items 7, 16).** cmd/go-alexa/go.mod still consumes the local SDK through a development replace and requires v0.4.0. Verify the public SDK consumer check, then the separately tagged CLI module through the public proxy.
- **R1-F14 — OPEN (item 14).** Reviewer 2's current all-16 section is now present at 572f2a1, but neither reviewer has rechecked the eventual fixed commit. Same-final-SHA verification remains outstanding while findings and release evidence are open.
- **R1-F15 — OPEN (items 4, 7).** Later nested index keys are not checked when a generated map is indexed more than once, and a local alias initialized from a nested indexed map can discard the intermediate raw key. Add negative probes for direct nesting and the alias sequence, then recursively inspect each indexed receiver before considering this resolved.

Disposition of earlier Reviewer 1 findings: provider/SDK schema constants and bindings, visible implementation-derived labels, exact-SHA lint/Windows ACL failures, the legacy capability object shape, and typed ParseEvent decoding were addressed before cc35. The cc35 zero-value map alias and parenthesized/indexed receiver concerns are covered by 572; the nested raw intermediate map key is the remaining source-gate issue. Pages publication, public SDK/CLI installation, and final same-SHA agreement remain open as R1-F12–F15.

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
