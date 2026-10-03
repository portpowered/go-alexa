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

## Reviewer 1 — preliminary audit at `f22de7d`

Reviewed source commit: `f22de7d9757eb5ad486d294ff420eb7878e7f63e`.
This is preliminary, not final sign-off. Exact-commit blocking CI is failing;
schema-derived wire constants and their gates are being corrected; final
publication and CLI installation proof are pending. The linked shared
standards were read from template commit
`843ec3f2ef2d28920c39ddbfef63d6a857c6fe05`. The working tree contained
uncommitted credential-permission edits during this review; verdicts below
describe the committed SHA.

### Item verdicts

1. **PASS (preliminary).** `pkg/alexa` is a reusable provider API. The README
   and `docs/guides/` contain no consuming-application adapter or rollout plan.

2. **OPEN.** The README and MDX guides cover configuration, authentication,
   enumeration/control, events, player state, and quality of service;
   `docs/verification.md` distinguishes synthetic fixtures and maintainer
   history from provider verification. But the rendered API landing page and
   several operations (`queryRestEndpoints`, `sendEndpointInterfaceMessage`,
   `getUserInfo`, `deregisterEndpoint`) omit an explicit implementation-derived
   / not-provider-verified label. Fix and inspect the generated pages.

3. **PASS (preliminary).** README badges cover Go version, CI, coverage,
   release, Go Reference, license, and documentation; targets are live
   repository reports, and no template repository value remains in badge URLs.

4. **FAIL.** OpenAPI/AsyncAPI/GraphQL schemas and generation cover the
   discovered REST, GraphQL, event, and HTTP/2 contracts. Independent source
   search nevertheless found library-defined wire values missing from
   schema-generated constants. `pkg/alexa/client_control.go` constructs
   playback feature and operation strings (play/pause/resume/next/previous/
   stop); the OpenAPI route parameters are open strings. The speaker
   `setVolume`/`adjustVolume` route has the same concern. `pkg/alexaapimodels/
   constants.go` defines fixed registration device type IDs while auth schemas
   leave `device_type` open; `pkg/alexaapimodels/requests.go` defines provider
   IDs used in a library-built behavior payload without schema enum constants.
   These are library-selected values, not genuinely caller-open fields.
   Disposition: add schema sources and generated constants/typed aliases,
   extend negative gates, regenerate the inventory, and re-review.

5. **FAIL.** Exact-SHA CI run `37109445660` failed. Nested CLI lint reported
   G304 at `cmd/go-alexa/internal/cli/credential_permissions_other.go:8`.
   The Windows CLI job failed
   `TestSavedCredentialFileHasProtectedOwnerOnlyDACL` because the actual DACL
   did not grant the current user SID. Later build/race/coverage and Windows
   module checks were not proven. A reported root-local check does not override
   blocking CI. Fix both failures and require complete passing CI on final SHA.

6. **OPEN.** `tools/coverage` excludes generated code/test support, reports
   package and aggregate coverage, and enforces the 80% combined minimum; CI
   config uses race-enabled coverage. Reported SDK coverage was 83.9%, under
   the 90% target. But exact-SHA CI failed before coverage ran, so this review
   has no successful blocking coverage result or final remaining-behavior
   report. Recheck and report package/combined totals and exclusions at final
   SHA.

7. **FAIL.** Public API, generated wire models, and transport are located in
   `pkg/alexa`, `pkg/dependencymodels`, and `pkg/dependencies/` respectively;
   schemas and generated files are split by responsibility. The independent
   type scan found no additional standalone handwritten production wire struct.
   A follow-up raw-object scan did find `LegacyAppliance.capabilities` modeled
   as `[JSON!]` in the GraphQL SDL although production code reads its
   `interfaceName` key; this unmodeled nested shape is recorded as F6 below.
   Primitive/library constants listed under item 4 also remain manually
   declared or absent from schema-generated definitions. Resolve those gaps
   and verify imports from the separate consumer module.

8. **PASS (preliminary).** `pkg/alexa/client.go` uses validated functional
   options and sensible defaults. Account credentials are passed through
   explicit session/request paths, not reusable service configuration.

9. **PASS (preliminary).** Shared `Client` configuration/transports are
   separate from account tokens and event connections owned by `Session`.
   Session close is explicit and idempotent; errors and token access are
   visible. I found no silent shared-client account/connection mutation.

10. **PASS (preliminary).** REST, GraphQL, and event HTTP clients have
    injection seams; event transport accepts a `RoundTripper`. The pinned
    `golang.org/x/net` HTTP/2 contract is in `api/external/http2.yaml`.
    `TestEventNativeHTTP2PairedReplay` injects `net.Pipe` beneath
    `http2.Transport` and exercises framed request/response, directive decode,
    keepalive, consumption, and close offline. This does not prove provider
    TLS behavior; recheck network-edge inventory at final SHA.

11. **PASS (preliminary).** Token refresh/exchange are explicit operations
    returning credentials. Refresh does not mutate session state; cookie
    exchange and token assignment are explicit, request auth does not silently
    refresh, and authentication docs assign storage/renewal to the caller.

12. **OPEN.** Customer guides are MDX in `docs/guides/`, link to generated
    reference pages, and Docs run `37109445520` succeeded with its rendered
    link check. The inspected build with matching docs content had 249 HTML
    pages and 22 operation pages. The PR workflow builds but does not deploy;
    final-commit Pages publication is unproven. The evidence-label issue in
    item 2 also remains. Inspect final Docs artifact and publication.

13. **OPEN.** I reviewed README and every tracked Markdown/MDX document,
    including contributor/release notes, guides, checklist, model inventory,
    fixture provenance, and verification notes. Customer navigation excludes
    internal audit reports; README is focused, and guides have clear audience
    and next actions. Generated reference pages are inconsistent about
    evidence status, and final rendered content is not yet reviewed. Fix and
    re-review the final artifact and release-note destinations.

14. **OPEN.** This is one independent preliminary review. Reviewer 2 has not
    appended an all-16 verdict, findings are open, and neither reviewer has
    verified the eventual final commit. Obtain two separate reviews at the same
    final SHA and close all findings before checking this item.

15. **PASS (preliminary).** `pkg/testing.SyntheticReplay` checks method,
    origin, escaped path, repeated query values, headers, and request body
    before returning paired status/headers/body; it rejects mismatch,
    duplicate, unexpected, and unconsumed calls. REST, GraphQL, event, and
    HTTP/2 tests assert complete consumption. The fixture README labels replay
    pairs synthetic; `pkg/alexa/testdata/` is separately labeled response-only
    model input, not replay evidence. No captures are presented as current
    provider proof.

16. **OPEN.** `cmd/go-alexa` is a separate module consuming `pkg/alexa`, with
    offline command tests and CI configuration for all-linter/build/test/module
    checks. The CLI guide documents local invocation while the module uses a
    development `replace`. The exact-SHA lint and Windows ACL checks failed
    (item 5), and separate installation from published CLI/SDK module tags
    remains pending until SDK v0.4.0 is public. Fix CI, then verify module tags
    and a separate consumer install.

### Findings and disposition at this SHA

- **F1 — open (items 4, 7):** library-built playback/speaker values,
  registration device IDs, behavior provider IDs, recognized legacy
  capability aliases, and supported event namespace/name values are not yet
  fully represented by schema-generated constants. Root was notified; planned
  changes are not counted as resolved.
- **F2 — open (items 2, 12, 13):** rendered API pages do not consistently say
  that the contracts are implementation-derived, not provider-issued. Add
  visible labels and inspect final output.
- **F3 — open (items 5, 6, 16):** CI `37109445660` failed nested CLI G304 lint
  and the Windows protected-DACL regression; downstream checks did not run.
  Fix and obtain complete passing final-SHA CI.
- **F4 — open (items 12, 16):** final Pages publication and published standalone
  CLI consumer installation remain pending.
- **F5 — open (item 14):** second reviewer and final-SHA joint verification
  remain outstanding.
- **F6 — open (items 4, 7):** GraphQL
  `LegacyAppliance.capabilities: [JSON!]` contains the fixed nested
  `interfaceName` key consumed by
  `pkg/alexa/client_enumeration.go#extractLegacyCapabilityInterfaces`; the
  current schema/generator does not type or generate that shape/key. The same
  source contains raw map lookups for `timeOfSample` and `timeOfLastChange`
  instead of generated accessors. Model the known nested object while
  retaining explicitly open additional fields, and route property access
  through generated types/accessors before re-review.
- **F7 — open (items 4, 7):** The exported legacy `ParseEvent` path in
  `pkg/alexa/client_event.go` walks the known `directive`, `header`, `payload`,
  `renderingUpdates`, and `resourceMetadata` structure through nested
  `map[string]interface{}` values from the compatibility `Message.Data` field.
  The key constants are generated from AsyncAPI, and the HTTP/2 path uses the
  generated `DirectiveMessage`, but this compatibility parser still does not
  decode that recognized event shape through the generated type. Convert it
  through `DirectiveMessage` or document and gate a narrower compatibility
  boundary.
- **F8 — open (items 4, 7):** A supplemental scan found additional known
  primitive values and manual projections not tied to generated constants:
  `client_enumeration.go` builds `NameValueObject` with raw `"PLAIN"` although
  `DefaultFriendlyNameType` is generated; GraphQL feature control compares
  states to raw `"ON"`/`"LOCKED"`; `feature-controls.yaml` models known lock,
  toggle, thermostat mode, and temperature scale values as plain strings
  despite corresponding GraphQL enums. The public compatibility enums for
  feature names/operations, endpoint display categories, QoS experience, and
  subscription entity types also need schema-backed values or explicit
  documented open-string treatment. Add call-site checks that reject raw
  library-selected wire values; generating constants alone does not prove
  their use.

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

## Reviewer 1 — independent audit at `f2ac3103b106d5618930e9fad42e7d0a56770961`

Reviewed the committed tree at this exact SHA. The working tree now contains later uncommitted implementation edits; I excluded them and used `git show f2ac...:<path>` for source findings. I did not implement this migration. I read the shared library standards, client design, and website guidance at template commit `843ec3f2ef2d28920c39ddbfef63d6a857c6fe05` and audited all 16 checklist items. The exact PR checks were [CI run 37117253367](https://github.com/portpowered/go-alexa/actions/runs/37117253367) and [Documentation run 37117253374](https://github.com/portpowered/go-alexa/actions/runs/37117253374). They are both tied to this SHA and succeeded, subject to the open findings and publication items below.

### Item verdicts

1. **PASS.** `pkg/alexa` is a reusable Alexa provider client. README, examples, and customer guides describe that package and do not contain a consuming-application adapter or rollout plan.

2. **PASS.** README and the nine MDX guides cover the exported client/session operations, auth, errors, transport injection, endpoint workflows, player state, events, and CLI. I checked the examples against `pkg/alexa` and `pkg/alexaapimodels`. The rendered 250-page artifact visibly brands every HTML page “implementation-derived”; `docs/verification.md` distinguishes that source-derived contract from provider-issued specifications, synthetic fixtures, and undocumented maintainer-reported history. The public CLI guide says its release is pending.

3. **PASS.** README has Go-version, CI, coverage, release, Go Reference, license, and Pages badges. The badge targets point to this module, its workflow, coverage report, release page, package reference, license, and project site; I found no template-repository badge values.

4. **FAIL — finding open.** Protocol schemas and generated inventory cover the discovered REST, GraphQL, behavior, HTTP/2, and event shapes, including named nested event models, legacy capability data, and provider/SDK projections. Independent source and gate review found remaining bypasses: indexed map mutations to a generated wire object's open map are not checked; control payload field bindings can be removed without failing the gate; and terminal UI builders still put fixed lock/toggle wire values in raw strings while primitive usage inventory scans only `pkg`. Details are R1-F9–F11 below. These are known library-selected values/gate omissions, so this item cannot pass at this SHA.

5. **PASS.** Exact-SHA CI run 37117253367 passed `verify` and `cli-windows`. `verify` regenerated models and routes, checked module metadata and formatting, ran pinned golangci-lint v2.3.0 with the blocking all-linter gate, then built and ran race tests and coverage. The Windows job passed race tests, the credential ACL/lifecycle regression, build/vet, and nested module metadata. Both root and CLI configs use `linters.default: all`.

6. **PASS (80% floor met; 90% target not met).** The exact-SHA CI coverage step reports combined non-generated coverage **84.6% (2397/2834 statements)**, above the enforced 80% minimum. Per-package reported totals are `pkg/alexa` 83.0%, `pkg/alexaapimodels` 86.5%, GraphQL 94.2%, REST 83.8%, and dependency models 100.0% of its three non-generated statements. The report generator excludes generated code and calls out remaining uncovered behavior; this is below the 90% target.

7. **OPEN.** Package boundaries and API-responsibility schemas/models are separated under `pkg/alexa`, `pkg/dependencymodels`, and `pkg/dependencies/`, with compatibility projections in their own schema. The complete generated inventory is attached in `docs/model-inventory.md`. However, a public-proxy consumer check runs only on a release tag and no v0.4.0 release is evidenced at this SHA; the separate CLI module still requires v0.4.0 and contains `replace github.com/portpowered/go-alexa => ../..`. The indexed-map/control-binding gaps also prevent a complete model signoff.

8. **PASS.** `alexa.NewClient` uses validated functional options and sensible defaults (`pkg/alexa/client.go`, `client_options_session_test.go`). Account tokens are supplied to `NewSession`, not stored as reusable client configuration.

9. **PASS.** `Client` holds reusable endpoint/transport configuration; account credentials and event connections belong to `Session`. Session close is explicit and idempotent, token reads/updates are explicit, and the tests create two sessions to check isolation (`pkg/alexa/client.go`, `client_options_session_test.go`).

10. **PASS.** REST, GraphQL, and event transports are injectable. The event API accepts an injected `RoundTripper`; the pinned `golang.org/x/net` HTTP/2 contract is separated in `api/external/http2.yaml`. `TestEventNativeHTTP2PairedReplay` drives actual HTTP/2 framing through `net.Pipe` with no provider network or credentials. No WebSocket, MQTT, or RTC edge is used by this library.

11. **PASS.** Refresh and cookie exchange are explicit operations that return caller-visible credentials. `RefreshAccessToken` does not update the session; `SetAccessToken` is a separate call. Token handling responsibilities are documented in the auth guide and README, with tests for explicit refresh/install behavior.

12. **OPEN.** Exact-SHA Documentation CI successfully built the site, checked 44,910 internal links across 250 rendered pages, and uploaded the `documentation-site` artifact. I inspected that artifact, including root, REST operation, GraphQL operation/type, AsyncAPI, and guide pages. The PR deploy job was skipped, however, so current Pages publication and live post-deploy URLs remain unverified.

13. **PASS (reviewed pre-publication artifact).** I reviewed every tracked Markdown/MDX file at f2ac: README, CONTRIBUTING, all nine customer guides, checklist, independent review, model inventory, releasing, verification, both fixture READMEs, and schema note. Their audience/purpose is distinct; customer navigation contains guides rather than contributor audits, and the README remains caller-focused. I reviewed the rendered artifact and the release workflow's v0.4 upgrade-guide, CLI-guide, and verification links against the generated routes. Sitewide rendered labels state “implementation-derived”; the CLI page accurately makes installation conditional on publication. The live deploy remains item 12.

14. **OPEN.** This is Reviewer 1's final-at-this-SHA section. Reviewer 2 has now also appended a separate all-16 section at f2ac; neither reviewer has checked the eventual post-fix SHA. New item-4 findings and publication/consumer proof also remain open. Do not mark this item complete until both reviewers independently verify every fix and every open checklist item at the same final commit.

15. **PASS.** `pkg/testing.SyntheticReplay` matches method, origin, escaped path, repeated query values, headers, and body before it returns the paired response; response status, headers, and body are explicit in the fixtures. Tests reject mismatch, duplicate, unexpected, and unconsumed exchanges, and assert consumption/order on REST, GraphQL, and HTTP/2/event paths. `tests/replay/fixtures/synthetic/README.md` labels these as synthetic. The response-only model examples under `pkg/alexa/testdata/` are separately described as non-capture evidence.

16. **OPEN.** The CLI is separate under `cmd/go-alexa`, has offline paired command/lifecycle tests, explicit credential input/export controls, JSON output, cancellation, and cleanup. Exact-SHA CI passed its Linux and Windows gates. The release is still pending: the nested module requires SDK v0.4.0 via a local `replace`, there is no published CLI module tag, and no clean `go install ...@version` through the public proxy has been demonstrated. The MDX guide correctly describes this as pending.

### Findings and disposition at this SHA

- **R1-F9 — OPEN (items 4, 7).** `tools/modelinventory/wire_construction.go:51-67` checks post-initialization writes only when the left side is `*ast.SelectorExpr`; `input.Extra["brandNewKey"] = caller` and `input.Extra[caller] = "brandNewValue"` are indexed expressions and bypass it. Frozen tests in `wire_construction_test.go:9-35,67-91` cover a map literal and scalar field writes, but not later indexed writes, map aliases, or local key aliases. Extend the lexical map provenance gate and negative tests to cover those cases while preserving caller-defined open map entries.
- **R1-F10 — OPEN (items 4, 7).** At f2ac the control schema currently binds `LockPayload.lockState`, `TogglePayload.toggleState`, `ThermostatSetpoint.scale`, and `ThermostatModePayload.thermostatMode` to their `ControlKnown*` components, and those components list the corresponding SDL enums. The gate checks enum components via `checkOpenGraphQLBindings`, but the frozen tree has no control-field binding check (no `control_field_bindings.go`); unlike event fields, a future edit can replace any of these refs with plain `type: string` without a failing negative test. Add a schema-to-SDL field-reference check and one negative mutation per field.
- **R1-F11 — OPEN (items 4, 7).** `tools/modelinventory/sdk_primitives.go:82-83` walks only `pkg` for semantic primitive uses. The library's terminal UI example therefore escapes the use-site gate, and `cmd/examples/terminal-ui/tui/control_view.go:647-650,674-677` still constructs `ControlLockPayload.State` with `"LOCKED"`/`"UNLOCKED"` and `ControlTogglePayload.State` with `"ON"`/`"OFF"`. Include command/example production sources in the primitive inventory and use the generated known values at those wire construction sites.
- **R1-F12 — OPEN (items 12, 14).** The exact Docs build and 250-page rendered artifact are reviewable and its link gate passed, but the PR deploy job was skipped. Verify successful GitHub Pages publication and the live guide/reference/release-note destinations after publication.
- **R1-F13 — OPEN (items 7, 16).** `cmd/go-alexa/go.mod` still has the local SDK `replace` and requires the unpublished v0.4.0 SDK. Verify the SDK tag with the release workflow's public-proxy consumer, then remove the CLI development replacement, tag the CLI module, and verify a clean public-proxy `go install`.
- **R1-F14 — OPEN (item 14).** Reviewers 1 and 2 have now documented separate all-16 sections at f2ac, but neither has rechecked the eventual fixed commit. Reviewer agreement at the final SHA remains outstanding.

Disposition of Reviewer 1's earlier preliminary findings in this same document: F1's named provider/SDK values and projections are now represented in schemas/generated constants and their listed use sites; F2's rendered evidence-label issue is resolved by the visible sitewide “implementation-derived” label and this artifact review; F3's earlier lint/Windows ACL failures are resolved by exact-SHA CI success; F6's `LegacyAppliance.capabilities.interfaceName` shape is now generated while preserving open extra fields; and F7's known `ParseEvent` directive structure now decodes through `DirectiveMessage`. F4 remains open for Pages publication and published CLI install; F5 remains open until two reviewers verify the same final SHA; F8 remains open in part because the control field-reference gate and raw terminal UI wire values are still missing at f2ac (R1-F10–F11).

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
