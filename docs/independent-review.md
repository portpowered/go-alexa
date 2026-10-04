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

## Reviewer 1 — independent audit at `85cec99c2c70c6773c6d5847ee0ebb3ba64436c6`

I did not implement this SDK or CLI. I checked the committed tree at this exact SHA against all 16 items in the checklist pinned to template `987b9c34a6b927472c21604617b6842a4238746b`, including the linked library, client-design, website, verification, and release standards. Exact-SHA CI is [run 37123180284](https://github.com/portpowered/go-alexa/actions/runs/37123180284); exact-SHA documentation build is [run 37123180285](https://github.com/portpowered/go-alexa/actions/runs/37123180285). CI passed its blocking Linux verify and Windows CLI jobs. Documentation passed its build and checked 44,910 internal links across 250 rendered pages. I downloaded artifact 11274017843 and confirmed all 250 HTML pages visibly carry the `implementation-derived` label; I inspected the rendered site root, guide index, authentication guide, and generated account-linking reference. The Pages deploy job was skipped, so live publication remains unverified.

### Item verdicts

1. **PASS.** `pkg/alexa` is a reusable provider client. The README, examples, and customer guides do not add consuming-application adapters or rollout plans.

2. **PASS.** The README and MDX guides describe the exported client/session API, supported authentication, typed errors, transport injection, endpoint operations, events, and CLI. `pkg/alexa/client.go` now limits the SDK claim to code-pair registration, bearer tokens, and explicit refresh; the authentication guide says browser authorization, callback state, and PKCE are managed by the calling application.

3. **PASS.** The README contains Go version, CI, coverage, release, Go Reference, license, and documentation badges with project destinations; no template repository values remain.

4. **OPEN — R1-F20 and R1-F21.** The checked-in HTTP, GraphQL, event, and external HTTP/2 schemas feed generated routes and models; the inventory maps generated definitions to call sites. I independently checked wire-model and primitive populations, including anonymous/nested event shapes, compatibility projections, custom decoders, semantic constants, and actual package uses. The source gate now resolves sibling-file package constants and checks direct `maps.Copy`/builtin `copy`, local named helper calls, aliases, indexed receiver keys, and the covered later mutations. Two exact-tree gaps remain: (a) the gate only propagates map provenance through recognized local named functions and direct copy operations; it does not fail closed for unverified callable escapes. `wireLocalHelpers` does not resolve an immediately invoked or aliased `*ast.FuncLit`, a same-package method call expressed as a selector, or a function-valued callback parameter. Valid probes include `func(extra map[string]any, v string) { extra["brandNewKey"] = v }(input.Extra, caller)`, `helper{}.mutate(input.Extra, caller)`, and `func request(mutate func(map[string]any, string), input *wire.Payload, caller string) { mutate(input.Extra, caller) }`. (b) exported legacy `AuthConfig` and `DeviceAuthConfig` remain in `pkg/dependencymodels/auth.go` but are absent from both `tools/modelinventory/handwritten.yaml` and `docs/model-inventory.md`. `docs/verification.md` now explains that the types are caller-owned and not used by SDK operations, but the pinned standard still requires unused legacy exports to be represented in the complete inventory. These gaps keep the item open.

5. **PASS.** Exact-SHA CI run 37123180284 passed generated artifact and module checks, formatting, pinned full-repository golangci-lint v2, build/race tests, and coverage gating. The Windows CLI job passed credential permission/lifecycle tests, build, vet, and nested module checks.

6. **PASS — floor met; target short.** CI reports **84.6% combined non-generated coverage (2397/2834 statements)**. Package results: `pkg/alexa` 83.0% (1282/1544), `pkg/alexaapimodels` 86.5% (90/104), GraphQL 94.2% (277/294), REST 83.8% (745/889), and dependency models 100% over three non-generated statements. This exceeds the 80% floor and remains below the 90% target; generated exclusions and remaining uncovered behavior are recorded.

7. **OPEN — R1-F21 and R1-F13.** Public API, dependency models, and transport implementations are organized under `pkg/alexa`, `pkg/dependencymodels`, and `pkg/dependencies/`; schemas are split by responsibility. The omitted legacy config types in F21 need explicit inventory disposition. Also, `cmd/go-alexa/go.mod` requires SDK v0.4.0 and contains `replace github.com/portpowered/go-alexa => ../..`; no separate public-proxy consumer compile is available before publication.

8. **PASS.** `alexa.NewClient` uses validated functional options and defaults. Account credentials are supplied when creating a session, not stored in reusable client configuration.

9. **PASS.** Account credentials and event connections belong to explicit `Session` instances. Token reads/updates, close errors, and connection ownership are visible; tests cover session isolation and cleanup.

10. **PASS.** REST, GraphQL, and HTTP/2 event traffic use injectable HTTP transports. `TestEventNativeHTTP2PairedReplay` exercises the actual framed exchange over an injected `net.Pipe`; the network inventory contains no other WebSocket, MQTT, RTC, or raw socket edge used by this library.

11. **PASS.** Token refresh and cookie exchange are explicit operations returning caller-visible values. Refresh does not replace session state; callers install the access token explicitly and own credential storage and renewal.

12. **OPEN.** Documentation run 37123180285 built the site and checked every rendered page's internal links. The PR deploy job was skipped, so GitHub Pages publication and live destinations are not verified.

13. **PASS — reviewed build artifact, pre-publication.** I reviewed the README and tracked Markdown/MDX files, including this commit's authentication, verification, inventory, and review-record changes. Customer navigation remains guide-focused; contributor inventory and audit material stays out of customer navigation; synthetic fixtures and implementation-derived contracts are identified. The downloaded artifact passes the all-page link check and visibly labels all 250 pages. Publication remains item 12.

14. **OPEN.** Reviewer 2 has not yet independently verified this exact SHA; its current section is still anchored at 572f2a1. F20/F21, public SDK/CLI consumer checks, and Pages publication remain open, so final two-review signoff is not possible.

15. **PASS.** `pkg/testing.SyntheticReplay` matches method, origin, escaped path, repeated query values, full headers, and body before returning a paired response; tests reject mismatch, duplicate/unexpected calls, and unconsumed exchanges. The paired REST suite now consumes full synthetic request/response pairs for the OTP challenge, rejected code-pair registration, and non-challenge email failure. The fixture records each `/auth/register` origin, method, path, headers, body, status, and response body, including the synthetic device serial. Event replay covers HTTP/2 frames and close; CLI tests cover cancellation and response-body cleanup. No OAuth callback/token flow is implemented by the SDK; the `/ap/oa` URL is caller-managed and marked implementation-derived, with callback state and PKCE assigned to the caller.

16. **OPEN — R1-F13.** `cmd/go-alexa` is a separate module with offline paired command tests, explicit credential inputs/export, machine-readable output, nonzero failures, cancellation, and cleanup; Linux and Windows CI passed. The nested module still has a local `replace`, SDK v0.4.0 and CLI module tags are unpublished, and no clean public `go install ...@version` or consumer install has been demonstrated.

### Findings and disposition at 85cec99

- **R1-F9 — RESOLVED at 572f2a1 (items 4, 7).** Negative coverage now follows zero-value local map aliases and prior assignments.
- **R1-F10 — RESOLVED at cc35 (items 4, 7).** SDL bindings for known event/control enum fields are validated, including deletion negatives.
- **R1-F11 — RESOLVED at cc35 (items 4, 7).** Primitive usage inventory includes `cmd`; terminal UI wire choices use generated constants while labels and caller values remain open.
- **R1-F12 — OPEN (items 12, 13).** A successful build artifact and all-page link check do not establish a Pages deployment.
- **R1-F13 — OPEN (items 7, 16).** SDK v0.4.0 and the nested CLI module are unpublished; public-proxy consumer/install checks remain outstanding.
- **R1-F14 — OPEN (item 14).** Reviewer 2 has not verified this exact final candidate.
- **R1-F15 — RESOLVED at b0e4a77 (items 4, 7).** Recursive receiver-key checks include intermediate indexed keys and aliases, with caller/generated-key positives.
- **R1-F16 — RESOLVED at 85cec99 (item 15).** The OTP challenge, rejected code-pair, and non-challenge email failures are full paired replay entries consumed in order.
- **R1-F17 — RESOLVED at 85cec99 (items 2, 15).** The public `full OAuth flows` claim was removed; the SDK and guide state that browser callback/state/PKCE belong to the caller.
- **R1-F18 — RESOLVED at 85cec99 (items 4, 7).** The exact tree now includes `wire_source_packages.go` and a sibling-file negative test for package-level constants and helper mutations. These files were untracked and absent at b0; they are tracked at this SHA.
- **R1-F19 — PARTIALLY RESOLVED at 85cec99 (items 4, 7).** Direct `maps.Copy`/builtin `copy`, generic/function aliases, and recognized local named helper mutations are checked with negative and positive probes. The remaining callable forms are tracked separately as F20.
- **R1-F20 — OPEN (items 4, 7).** The wire mutation gate does not fail closed when a generated map escapes to an unverified callback/helper. Direct or aliased function literals, method selectors, function-valued callbacks, and wrapped/aggregate arguments need negative probes; retain caller/open-key positives.
- **R1-F21 — OPEN (items 4, 7).** `AuthConfig` and `DeviceAuthConfig` need explicit non-wire entries in the complete model inventory and manifest, or removal from the public dependency-model surface.

### Reviewer 1 checkpoint findings at `8a5c84e7acb4277fdf11f13f8925774fa2134a94` (preliminary)

I checked the committed model-inventory and wire-mutation gate code at this exact checkpoint. The probes below ran through a temporary Go test overlay; no source or test files were written to the repository. Documentation run [37170985733](https://github.com/portpowered/go-alexa/actions/runs/37170985733) passed; CI run 37170985737 was still in progress when this note was written. These findings are not final all-16 signoff.

- **R1-F20 — OPEN (items 4, 7).** The new callable analysis catches direct/aliased function literals, ordinary local helpers, and direct generated-map arguments, but misses wrapped provenance. The exact gate accepted `unknown([]any{input.Extra})`, `unknown(expose(input))` where `expose` returns `input.Extra`, and `unknown(func() map[string]any { return input.Extra })`. A local mutator called as `mutate([]any{input.Extra})` also lacks generated provenance. The generic method-expression probe `helper[int].mutate(helper[int]{}, input.Extra)` was accepted with a raw key/value mutation in the method because the receiver argument is not included in parameter mapping.
- **R1-F21 — RESOLVED for the two known compatibility exports at this checkpoint; overall inventory gate remains open.** `AuthConfig` and `DeviceAuthConfig` now have `legacy-nonwire` manifest entries, source markers, and an explicit no-tag reflection test; the scan rejects an unlisted exported no-tag struct in `pkg/dependencymodels`.
- **R1-F22 — OPEN (items 4, 7).** Selector helper resolution uses package-wide method names without proving receiver identity. With a local no-op `func (helper) Mutate(map[string]any) {}`, the unrelated call `external.Mutate(input.Extra)` on an interface parameter was accepted. This can incorrectly treat an unverified external method as a safe local helper.
- **R1-F23 — OPEN (items 4, 7).** The legacy `AuthConfig` scanner accepts a `MarshalJSON() ([]byte, error)` method with no JSON tags or direct codec call. The type is therefore still classifiable as non-wire after acquiring custom wire serialization behavior.
- **R1-F24 — OPEN (items 4, 7).** Per-file model scanning misses an unexported, untagged `privateWire` in `pkg/dependencymodels/private.go` when a different file serializes a typed `privateWire` parameter with `json.Marshal`; neither file produces an inventory entry or diagnostic. The required population includes every production struct encoded in a wire exchange.
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
