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
