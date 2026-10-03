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

## Reviewer 2 — preliminary independent audit at `9a6edceecc02163766d54fbc05e0a136ce487561`

I did not implement this SDK or CLI. I independently reviewed this commit against checklist items 1–16 and the shared standards pinned at `843ec3f2ef2d28920c39ddbfef63d6a857c6fe05`. This is a preliminary review of that source snapshot; findings remain open until root’s fixes and the final SHA are reviewed.

### Item verdicts

1. **PASS (preliminary).** The public packages are provider-focused under `pkg/alexa`; README and customer guides do not depend on a consuming application. Application secret stores are referenced generically.

2. **OPEN.** README and MDX guides document supported operations, session authentication, errors, transport injection, control behavior, and synthetic evidence. Generated reference pages still lack visible evidence labels on many operations; see R2-F3.

3. **PASS.** README shows Go version, CI, coverage, release, Go Reference, license, and GitHub Pages badges, all pointing to project reports or destinations.

4. **FAIL.** Route/schema generation and the route/model negative gates run in CI, and REST, GraphQL, behavior, event, and pinned HTTP/2 responsibilities have separate checked-in schemas. My source scan nevertheless found library-selected values and recognized wire shapes that remain handwritten, open strings, or raw maps; see R2-F1 and R2-F2.

5. **PASS (preliminary).** Exact-commit CI run [37110976620](https://github.com/portpowered/go-alexa/actions/runs/37110976620) passed both `verify` and `cli-windows`. Root and nested `.golangci.yml` set `linters.default: all`; CI pins golangci-lint v2.3.0, and the run reports 0 issues for both modules. The earlier 105c3c3 Windows module check differed only in go.sum line endings; `.gitattributes` fixes LF on 9a, and the exact-9a Windows job passes.

6. **PASS (preliminary, minimum met).** The exact-9a race-enabled coverage gate reports `pkg/alexa` 82.1%, `pkg/alexaapimodels` 86.5%, `pkg/dependencies/graphql` 94.2%, `pkg/dependencies/rest` 83.8%, `pkg/dependencymodels` 100.0% over its 3 non-generated statements, and combined non-generated package coverage 84.1% (2394/2847), above the enforced 80% minimum. The 90% target is not met. `tools/coverage` excludes generated Go files and `pkg/testing` helpers.

7. **OPEN.** Public and dependency packages are separated, but the source/model gaps in R2-F1 and R2-F2 remain. A public consumer compile through the proxy is not yet evidenced: `cmd/go-alexa/go.mod` still has `replace ... => ../..`, and the CLI’s separate install is also pending.

8. **PASS (preliminary).** `alexa.NewClient` uses validated options and defaults; credentials are supplied to `NewSession`, not stored in reusable client configuration.

9. **PASS (preliminary).** Session owns account credentials and open event connections. `Session.Close` is explicit, repeated close is safe, and the reusable client does not silently swap account or connection state.

10. **PASS (preliminary).** REST, GraphQL, and event-stream HTTP clients/transports are injectable. `TestEventNativeHTTP2PairedReplay` injects a `net.Pipe` beneath the pinned HTTP/2 transport and checks framed requests, directive parsing, keepalive, consumption, and close without provider access.

11. **PASS (preliminary).** `RefreshAccessToken` returns caller-visible credentials without installing the new access token; `SetAccessToken` is explicit. Cookie exchange and CSRF retrieval are explicit calls too.

12. **OPEN.** The customer guides are MDX under `docs/guides/` and link to generated pages. Exact-9a Documentation run [37110976619](https://github.com/portpowered/go-alexa/actions/runs/37110976619) built 250 pages and checked 44,910 internal links successfully. Its deploy job was skipped for this pull request; Pages deploys from `main`, so final publication remains unverified.

13. **OPEN.** I reviewed every tracked Markdown/MDX document and the rendered site artifact. The README is focused, customer navigation excludes contributor audits, and provenance notes distinguish synthetic fixtures. However, rendered evidence labels remain inconsistent (R2-F3), so the published pages are not ready for sign-off.

14. **OPEN.** Reviewer 1’s section is preliminary at f22de7d, and this section is preliminary at 9a6edce. Neither review verifies the eventual final commit; both reviewers must recheck the same final SHA and close every finding.

15. **PASS (preliminary).** `pkg/testing.SyntheticReplay` matches each request before returning its paired response and checks method, origin, escaped path, repeated query, headers, and body. REST, GraphQL, event, native HTTP/2, and CLI tests assert consumption; mismatch and unconsumed-pair negative cases are present. Fixtures are explicitly synthetic.

16. **OPEN.** `cmd/go-alexa` is a separate module over the public SDK. It has offline paired command tests for account linking, refresh failures, endpoint listing/control, player state, and event cancellation/close; help, JSON output, nonzero errors, explicit credential export, protected credential files, and all-linter CI are present. But the SDK v0.4.0 dependency is not yet published, the development `replace` remains, and no `cmd/go-alexa/v*` release tag or separate public `go install` proof exists.

### Reviewer 2 findings

- **R2-F1 — open (items 4, 7):** Library-defined wire values are not consistently schema-derived or checked at use sites. `FeatureNamePlayback`/`FeatureNameSpeaker` and the operation string from `FeatureOperationName` flow through `pkg/alexa/client_control.go#sendThirdPartyPlaybackMessage` and `#controlInterfaceVolume` into `pkg/dependencies/rest/endpoints.go#SendInterfaceMessage`, which inserts them into the actual REST path. `api/compat/endpoints.yaml` models these fields as open strings. Public `ProviderID` constants flow from `controlMusic` into `MusicPlaySearchPhrasePayload.musicProviderId`, whose schema is string; `DeviceTypeIphone`/`DeviceTypeSimulator` flow into registration bodies while the auth schema also says string. `api/asyncapi.yaml` leaves directive `namespace`/`name` open strings while `client_event.go` matches manually declared supported names/namespaces. Additional source-selected values include legacy capability aliases, `"PLAIN"` despite generated `DefaultFriendlyNameType`, raw `"ON"`/`"LOCKED"` state comparisons, and feature-control lock/toggle/mode/scale strings despite corresponding GraphQL enums. Put each known library-selected value in its responsibility schema, generate the public semantic projections from those sources without introducing an import cycle, and gate the actual call sites while preserving genuinely open caller values.

- **R2-F2 — open (items 4, 7):** Known nested wire shapes still pass through raw JSON maps. `LegacyAppliance.capabilities` is `[JSON!]` in the GraphQL SDL although `extractLegacyCapabilityInterfaces` reads `interfaceName`; generated `FeatureProperty` implementations carry `timeOfSample` and `timeOfLastChange`, but enumeration marshals them to maps and indexes raw keys; the exported `ParseEvent` compatibility path walks the known directive/header/payload/rendering-update/resource-metadata tree through nested maps even though `DirectiveMessage` is generated. Type these known portions in schemas and route reads through generated types, retaining explicit open extra fields where required.

- **R2-F3 — open (items 2, 12, 13):** In the exact-9a rendered artifact, 16 of 22 OpenAPI operation pages, the AsyncAPI event page, all 212 GraphQL pages, and the site root do not visibly label the contract as implementation-derived and not provider-verified. Examples include `sendEndpointInterfaceMessage`, `sendMediaCommand`, `getUserInfo`, and `deregisterEndpoint`. Add evidence-status copy where customers see each contract, rebuild the site, and inspect the final artifact.

- **R2-F4 — open (items 7, 12, 14, 16):** At this review snapshot, GitHub Pages is configured to publish from `main`, while the exact-9a pull-request run skipped deploy. The GitHub release list has SDK v0.3.2 as the latest release and no `cmd/go-alexa/v*` tag; the nested module still requires unpublished v0.4.0 with a local `replace`. Publish and verify the SDK first, then verify a clean public CLI install and the final Pages artifact. Both reviewers must recheck that same final commit.

## Reviewer 1 — independent audit at cc35d2384dc34b11d186c2f7d51af330e805a934

I did not implement this SDK or CLI. I reviewed the committed tree at this exact SHA against all 16 checklist items and the shared standards pinned at 8fd452025259f9ad724498e15787948edbd06784. The shared checkout now contains later, uncommitted wire-gate edits; they are not included in this review. Those edits do not change the exact-cc35 findings below.

Exact-SHA evidence is [CI run 37119339216](https://github.com/portpowered/go-alexa/actions/runs/37119339216) and [Documentation run 37119339302](https://github.com/portpowered/go-alexa/actions/runs/37119339302), both successful. The documentation artifact 11272771541 was downloaded and inspected. Item 4 remains open because the source gate still misses a valid map alias mutation, and release/publication evidence is still pending.

### Item verdicts

1. **PASS.** pkg/alexa is a reusable provider client. README, examples, and customer guides describe the library and do not include a consuming-application adapter or rollout plan.

2. **PASS.** README and the customer MDX guides document exported client/session operations, authentication, errors, transport injection, endpoint workflows, player state, events, and CLI use with SDK types. I checked the rendered documentation artifact: all 250 HTML pages visibly carry the “implementation-derived” label. docs/verification.md separates implementation-derived behavior from provider-issued specifications, synthetic fixtures, and historical maintainer reports.

3. **PASS.** README badges cover Go version, CI, coverage, release, Go Reference, license, and documentation. Their targets point to this module, its reports, release page, reference, license, and site; I found no template repository badge values.

4. **OPEN — R1-F9.** Checked-in REST, GraphQL, behavior, event, and external HTTP/2 schemas feed generation and the model/route gates. I independently searched the wire-model population, including exported dependency models, anonymous/nested objects, primitive constants and projections, custom decoders, and actual construction/consumption sites, and cross-checked docs/model-inventory.md and the inventory tooling. The generated-wire construction gate now checks direct indexed writes, short-declaration map aliases, and local key aliases. It still misses a zero-value map alias assigned later:

       var alias map[string]any
       alias = input.Extra
       alias["brandNewKey"] = caller

   The alias declaration has no initializer, and the later map write is not traced from wireSourceAssignments, so this library mutation of a generated wire object's open map can evade the negative gate. Caller-defined open fields must remain accepted, but this later fixed-key/value mutation must fail. Reviewer 2 separately reported an indexed/parenthesized receiver traversal bypass at this SHA; that is recorded as Reviewer 2's finding and I have not counted it as my independent reproduction. Because a known source-gate bypass remains, I cannot sign off complete wire/model generation.

5. **PASS.** Exact-SHA CI run 37119339216 passed both verify and cli-windows. The blocking verify job ran generation/drift checks, module metadata, formatting, pinned golangci-lint v2.3.0 with all linters enabled, build, race tests, and coverage. The Windows job passed CLI race tests, credential ACL/lifecycle checks, build/vet, and nested module metadata.

6. **PASS — 80% floor met; 90% target not met.** CI reports combined non-generated coverage of **84.6% (2397/2834 statements)**. Per-package totals are pkg/alexa 83.0%, pkg/alexaapimodels 86.5%, GraphQL 94.2%, REST 83.8%, and dependency models 100.0% over its three non-generated statements. Generated code is excluded and the coverage report identifies remaining uncovered behavior. The stated 90% target is still short.

7. **OPEN.** Public API, dependency models, and transport implementations are separated under pkg/alexa, pkg/dependencymodels, and pkg/dependencies/; API responsibilities have separate schemas and the generated inventory is in docs/model-inventory.md. The map-provenance gap in R1-F9 prevents a complete generation signoff. Also, cmd/go-alexa/go.mod still has a development replace to the repository root and requires unpublished SDK v0.4.0; a clean public-proxy consumer compile has not been shown.

8. **PASS.** alexa.NewClient uses validated functional options with sensible defaults (pkg/alexa/client.go, client_options_session_test.go). Account tokens are supplied to NewSession, not stored in reusable client configuration.

9. **PASS.** Client holds reusable endpoint/transport configuration, while account credentials and event connections belong to explicit Session objects. Session close is explicit and idempotent; token access/update are visible operations, with isolation tests for multiple sessions.

10. **PASS.** REST, GraphQL, and event HTTP transports are injectable; the event API accepts an injected RoundTripper. The pinned golang.org/x/net HTTP/2 contract is in api/external/http2.yaml; TestEventNativeHTTP2PairedReplay exercises framed requests/responses through net.Pipe with no provider network or credentials. The source/network inventory found no WebSocket, MQTT, RTC, or other non-HTTP socket edge used by this library.

11. **PASS.** Refresh and cookie exchange are explicit operations that return caller-visible credentials. RefreshAccessToken does not mutate session state; SetAccessToken is separate. The auth guide and tests leave token storage and renewal with the caller.

12. **OPEN.** Documentation run 37119339302 built the site, checked 44,910 internal links across 250 pages, and uploaded a reviewable documentation-site artifact. I reviewed the root, REST and GraphQL operation/type pages, AsyncAPI page, guide index, and CLI guide. The PR deploy job was skipped, so successful Pages publication and live destinations are not proven.

13. **OPEN.** I reviewed the README and every tracked Markdown/MDX file, including guides, checklist, model inventory, release/verification notes, fixture provenance, and the independent-review record, as well as the rendered artifact. Customer navigation is guide-focused; contributor audits stay out of customer navigation; synthetic fixtures and historical reports are distinguished; release workflow destinations match the rendered guide locations. The current record still contains Reviewer 2's duplicated preliminary and f2ac sections, which must be consolidated into one current all-16 review section. Final published-site review also depends on item 12.

14. **OPEN.** My current all-16 report is at cc35. Reviewer 2 independently found another item-4 traversal gap at cc35, but their current all-16 section has not yet been consolidated here. Both reviewers must verify every fix at the same eventual frozen SHA, and publication/consumer findings must be closed before this item can pass.

15. **PASS.** pkg/testing.SyntheticReplay matches method, origin, escaped path, repeated query values, headers, and body before returning the paired response status/headers/body. Tests reject mismatches, duplicate and unexpected calls, and unconsumed exchanges, and assert ordered consumption where it matters. Replay pairs are labeled synthetic; response-only model examples are separately identified and are not presented as replay evidence.

16. **OPEN.** cmd/go-alexa is a separate module with offline paired command/lifecycle tests, explicit credential inputs/export, machine-readable output, cancellation, cleanup, and passing Linux/Windows CI. The customer guide correctly makes installation conditional on publication. SDK v0.4.0 and a CLI module tag are not published; the nested module still uses a local replace, and no clean public go install ...@version has been demonstrated.

### Findings and disposition at cc35

- **R1-F9 — OPEN (items 4, 7).** tools/modelinventory/wire_construction.go does not follow a zero-value local map declaration through a later assignment from a generated open map and subsequent indexed mutation. Add a negative test for this sequence and verify the gate retains provenance through supported aliases without rejecting caller-defined entries.
- **R1-F10 — RESOLVED at cc35 (items 4, 7).** tools/modelinventory/control_field_bindings.go validates the SDL binding for LockPayload.lockState, TogglePayload.toggleState, ThermostatSetpoint.scale, and ThermostatModePayload.thermostatMode. Negative tests remove each field reference and require rejection; the generated components remain bound to the corresponding GraphQL enums.
- **R1-F11 — RESOLVED at cc35 (items 4, 7).** The SDK primitive-use inventory scans cmd as well as pkg; the terminal UI now uses generated lock/toggle state constants at wire construction, with tests retaining ordinary UI labels and caller inputs as valid.
- **R1-F12 — OPEN (items 12, 13).** Docs build and rendered artifact are verified, but the PR's Pages deployment was skipped. Verify live publication and destinations after deploy.
- **R1-F13 — OPEN (items 7, 16).** cmd/go-alexa/go.mod still consumes the local SDK through a development replace and requires v0.4.0. Verify the public SDK consumer check, then the independently installable/tagged CLI module through the public proxy.
- **R1-F14 — OPEN (item 14).** Reviewer 2's all-16 current section and both reviewers' recheck of the final frozen SHA remain outstanding.

Disposition of earlier Reviewer 1 findings: the generated provider/SDK constants and schema bindings, rendered evidence labels, exact-SHA lint/Windows ACL failures, legacy capability shape, and typed ParseEvent decode were addressed before cc35. The f2ac control-field and command-primitive findings are resolved above. The earlier publication, CLI-release, and same-final-SHA findings remain open as R1-F12–F14; the new zero-value map alias gap is R1-F9.

## Reviewer 2 — independent audit at `f2ac3103b106d5618930e9fad42e7d0a56770961`

I did not implement this SDK or CLI. I reviewed the committed tree at this exact SHA, excluding later uncommitted changes in the shared checkout, and checked the 16-item checklist against its pinned shared standards at `8fd452025259f9ad724498e15787948edbd06784`. Exact-SHA evidence includes [CI run 37117253367](https://github.com/portpowered/go-alexa/actions/runs/37117253367) and [Documentation run 37117253374](https://github.com/portpowered/go-alexa/actions/runs/37117253374). My source and gate review found the following unresolved item-4 gaps; consequently items 4, 7, and 14 remain open even where their other evidence is complete.

### Item verdicts

1. **PASS.** `pkg/alexa` is a provider client. README, examples, and customer guides cover the library itself and contain no consuming-application adapter or rollout plan.

2. **PASS.** README and the nine customer MDX guides document authentication, supported operations, errors, client configuration, endpoint workflows, player state, events, and CLI usage with examples using exported SDK types. The exact Docs artifact visibly labels all 250 rendered HTML pages “implementation-derived”; the verification and schema notes distinguish that contract from provider-issued specifications and synthetic fixtures.

3. **PASS.** README contains Go version, CI, coverage, release, Go Reference, license, and documentation badges. Their targets point to the project module, reports, releases, reference, license, and site; I found no template repository values.

4. **FAIL — findings open.** REST, GraphQL, behavior, event, and pinned HTTP/2 responsibilities have checked-in schemas and generated outputs; route, schema, model, and replay gates pass the exact CI run. However, I independently found source-gate bypasses for reassigned local aliases and later indexed map writes, no field-level gate retaining control payload enum bindings, a source-scope omission for example wire values, and a handwritten library-owned video search string template. See R2-F5–F8 below. These prevent the claimed complete generated wire inventory and negative-control coverage.

5. **PASS.** Exact-SHA CI passed `verify` and `cli-windows`. The root and nested CLI modules configure literal `linters.default: all`; CI pins golangci-lint v2.3.0 and blocks on lint, generation, formatting, module metadata, build, vet, and race tests. The Windows job also passed its credential-file ACL and lifecycle checks.

6. **PASS — 80% floor met; 90% target not met.** The CI coverage report shows combined non-generated coverage of **84.6% (2397/2834 statements)**. Per-package totals are `pkg/alexa` 83.0%, `pkg/alexaapimodels` 86.5%, GraphQL 94.2%, REST 83.8%, and dependency models 100.0% of three non-generated statements. The report excludes generated code and identifies remaining uncovered behavior.

7. **OPEN.** Public, dependency-model, and transport packages are separated, with responsibility-specific schemas and the complete model inventory in `docs/model-inventory.md`. The source-gate omissions in R2-F5–F8 still leave generated-model completeness unproven. Also, `cmd/go-alexa/go.mod` requires unpublished SDK v0.4.0 using `replace github.com/portpowered/go-alexa => ../..`; a public-proxy consumer compile has not been demonstrated.

8. **PASS.** `alexa.NewClient` accepts validated functional options and has sensible defaults. Account tokens are supplied to `NewSession` instead of being retained in reusable client configuration.

9. **PASS.** The client holds reusable endpoint and transport configuration, while session credentials and event connections are owned by explicit `Session` objects. Close is explicit and idempotent; token reads and updates are caller-visible operations.

10. **PASS.** REST, GraphQL, and event transports are injectable. Event streaming can inject an HTTP RoundTripper; the pinned `golang.org/x/net` HTTP/2 framing is tested over `net.Pipe`, including consumed frames, ping, and close without credentials or provider network access. I found no WebSocket, MQTT, or RTC edge used by this library.

11. **PASS.** Refresh and cookie/CSRF operations are explicit. `RefreshAccessToken` returns refreshed credentials without silently installing them; `SetAccessToken` is separate, and caller token-storage responsibility is documented and tested.

12. **OPEN.** The exact-SHA Docs run built the `documentation-site` artifact and checked 44,910 internal links across 250 pages. The PR deploy step was skipped, so successful Pages publication and live destination checks remain unverified.

13. **PASS — reviewed build artifact, pre-publication.** I inspected all tracked Markdown/MDX files at f2ac, including README, contributing/releasing/verification material, all customer guides, checklist and review record, model inventory, schema and fixture notes, and the rendered artifact. Customer navigation is guide-focused; contributor audits stay out of customer navigation; the README remains caller-focused; fixture notes keep synthetic inputs separate from provider evidence. The rendered site-wide evidence label is visible. Publication remains open under item 12.

14. **OPEN.** Reviewer 1 and I have now each audited all 16 items at f2ac, but the candidate still has open source-gate, publication, and consumer-release findings. Both reviewers must recheck the same eventual fixed implementation commit and close every finding before this item can pass.

15. **PASS.** `pkg/testing.SyntheticReplay` compares method, origin, escaped path, query multimap, headers, and body before returning a paired status, headers, and body. Tests reject mismatches, duplicates, unexpected and unconsumed exchanges and verify consumption on REST, GraphQL, event, and HTTP/2 paths. Replay fixtures are explicitly synthetic and separate from response-only model examples.

16. **OPEN.** `cmd/go-alexa` is a separate module over the public SDK, with offline paired command tests for authentication, refresh failures, endpoint listing/control, player state, and event cancellation/close. Help, JSON results, nonzero failures, explicit credential export, protected credential files, and blocking nested-module CI are present. The SDK v0.4.0 dependency is unpublished, the local `replace` remains, and neither a CLI module tag nor clean public-proxy `go install` evidence exists.

### Reviewer 2 findings

- **R2-F5 — OPEN (item 4).** `tools/modelinventory/wire_construction.go:51-67` checks post-construction assignments only when the left side is a selector; `:144-173` follows an identifier's original `ast.Object` declaration but does not account for later assignments to that local. Thus `value := caller; value = "newFixedValue"; _ = wire.Payload{Value: value}` is not rejected: the generated field use follows the first caller-valued declaration, while the later `value = ...` assignment has a bare identifier on the left and is skipped. Existing negatives cover direct literals, declaration aliases, and selector field mutation, but not a caller-valued local later reassigned to a literal or the equivalent local declaration followed by reassignment (`tools/modelinventory/wire_construction_test.go:9-35,67-115`). Track source-local assignments to their later uses and reject these negative cases.

- **R2-F6 — OPEN (item 4).** The same assignment scan accepts only `*ast.SelectorExpr` at `wire_construction.go:52-56`; it skips an indexed write such as `input.Extra["brandNewKey"] = caller` or `input.Extra[caller] = "brandNewValue"` to a generated wire object's open map after construction. The composite-literal checks do not establish provenance for later index mutations. Add negatives for indexed map key/value writes and aliases to the map, while preserving explicitly caller-defined open values.

- **R2-F7 — OPEN (items 4, 7).** The frozen schema binds lock, toggle, thermostat scale, and thermostat mode payload fields to `ControlKnown*` components, but the model inventory has no control-field binding check proving those field references remain linked to the original GraphQL SDL enums; the event-field checker does not cover these controls. Separately, `tools/modelinventory/sdk_primitives.go` scans only `pkg`, while `cmd/examples/terminal-ui/tui/control_view.go:647-650,674-677` constructs generated `ControlLockPayload`/`ControlTogglePayload` wire values with handwritten `"LOCKED"`, `"UNLOCKED"`, `"ON"`, and `"OFF"`. Validate each control schema field against its SDL enum and include production example sources in the primitive use gate.

- **R2-F8 — OPEN (item 4).** `pkg/dependencies/rest/endpoints.go:1142-1154` constructs the video behavior payload using the library-owned format string `"%s on %s"` through `fmt.Sprintf`, then copies the result into both `searchPhrase` and `sanitizedSearchPhrase`. `api/behaviors.yaml` describes these fields as strings but does not own or generate this fixed template. Represent library-built string templates in the behavior schema and have the generator/gate verify the actual construction site.
