# Independent review record

Both reviewers independently audited implementation commit
`8a22252d064857750c053fbd3103f6c79be463af` and cleared its implementation
findings. It was merged into `main` as
`e95cdaf612041487dd9f024dd1681a56203aad36` after exact CI and documentation
checks passed. The reports below retain their verdicts at the reviewed commit.

The checklist now copies template `25aeb783126c4049b3bc49286ee7808069db28e9`.
Final signoff remains open until both reviewers verify the current records,
main Pages deployment, SDK release, and public standalone CLI installation.
Earlier review records are retained in Git history.

## Reviewer 1 — independent audit at 8a22252d064857750c053fbd3103f6c79be463af

I did not implement the Alexa SDK or CLI. I reviewed the exact commit from a read-only archive at `C:\Users\andre\AppData\Local\Temp\alexa-r1-8a22252d-0ea74142eb2443c3824594c46982f1eb`. The PR is [#1](https://github.com/portpowered/go-alexa/pull/1), based on `main` and open. The candidate checklist in this SHA pins template `987b9c34a6b927472c21604617b6842a4238746b`. I also checked the named-result addition in shared library standard commit `25aeb783126c4049b3bc49286ee7808069db28e9`; the corresponding template/checklist repin is still pending in the candidate documentation.

Exact-SHA [CI run 37182072283](https://github.com/portpowered/go-alexa/actions/runs/37182072283) succeeded. Its Linux `verify` job passed generation, module metadata, formatting, pinned golangci-lint, build, race tests, and the coverage gate; the Windows `cli-windows` job passed CLI credential-permission/lifecycle tests, build, vet, and module checks. Combined non-generated coverage is 84.6% (2,397/2,834), above the 80% floor and below the 90% target: SDK 83.0%, API models 86.5%, GraphQL 94.2%, REST 83.8%, and dependency models 100% (3 statements). Independently, `go test -race ./tools/modelinventory` passed from the exact archive.

Exact-SHA [Documentation run 37182072321](https://github.com/portpowered/go-alexa/actions/runs/37182072321) succeeded: it rendered 250 pages and checked 44,910 internal links. I downloaded artifact 11295876503 and inspected the HTML files: all 250 contain the `implementation-derived` marker. The Pages artifact upload and deploy steps were skipped for this PR run, so live publication remains pending.

## Item verdicts

1. **PASS.** README, packages, examples, and customer guides describe a reusable Alexa library and contain no consuming-application adapter or rollout plan.

2. **PASS.** README and customer MDX guides cover operations, authentication, errors, configuration, injected transports, endpoint workflows, player state, events, and CLI use. The examples match Client/Session APIs. The authentication guide explicitly assigns browser callback, state, and PKCE to the consuming application because this SDK provides code-pair registration rather than that browser flow.

3. **PASS.** README badges cover Go version, CI, coverage, release, Go Reference, license, and Pages, and point to this repository or its reports. I found no template-repository values. The current release badge does not establish publication of this candidate.

4. **PASS at 8a22252; R1-F1 resolved.** The responsibility schemas, generated models/routes, GraphQL checks, model inventory, external HTTP/2 inventory, and route call-site checks cover the documented production contracts. On the exact SHA, `wireReturnExpressions` resolves named results for bare returns; both fixed-string and generated-map analysis consume those expressions. `TestWireConstructionRejectsHelperReturnedFixedValues` rejects direct helper returns, helper chains, local aliases, and named-result fixed values (including a named-result alias). `TestWireConstructionRejectsNamedResultMapEscape` rejects a named-result `map[string]any` returned from `input.Extra` and passed to an unverified helper. `TestWireConstructionAllowsNamedCallerResult` preserves caller-sourced values. Existing helper-mutation tests cover callback/function-literal arguments and caller-open map keys. The returned-literal scanner examines the wire result expression while not treating subsequent error/diagnostic results as wire payloads. The new shared-standard named-result clause is satisfied by these controls; its template/checklist pin remains a documentation task under item 13.

5. **PASS.** Exact-SHA Linux and Windows CI both succeeded. Workflow configuration pins Go and golangci-lint and uses literal `linters.default: all`; the Linux job runs `make check` coverage and the Windows job runs the standalone CLI checks. No CI lint findings were reported.

6. **PASS — floor met, target short.** Exact CI reports 84.6% combined non-generated coverage (2,397/2,834), above 80% and below the 90% target. Package percentages are listed above; generated-code exclusions are implemented in `tools/coverage`.

7. **PASS for code boundaries; OPEN for public consumer proof.** SDK and CLI use separate SDK, API-model, dependency-model, transport, schema, and nested CLI-module boundaries. However, `cmd/go-alexa/go.mod` requires SDK v0.4.0 and still has `replace github.com/portpowered/go-alexa => ../..`. GitHub releases/tags expose SDK through v0.3.2, and no public CLI module tag or clean public SDK/CLI consumer/install check exists yet.

8. **PASS.** `alexa.NewClient(opts ...Option)` applies validated options and defaults. Account credentials are passed to `NewSession`, not stored in reusable client configuration. Tests cover invalid/conflicting transport options and session option validation.

9. **PASS.** Reusable client configuration is separated from session credentials and event connections. Session/connection close cancels owned stream work and returns cleanup errors; tests cover isolation and event cleanup.

10. **PASS.** REST, GraphQL, and events accept injected network clients/transports. The native HTTP/2 replay uses an injected `net.Pipe`, exercises framed paired traffic and keepalive, asserts full replay consumption, closes the connection, and waits for server shutdown without live provider access.

11. **PASS.** Refresh and cookie exchange are explicit operations. `Session.RefreshAccessToken` returns credentials without mutating token state; docs tell callers to persist them and explicitly install an access token with `SetAccessToken`.

12. **OPEN — publication pending.** The exact Docs workflow rendered and checked the candidate and the reviewable artifact is available, but its Pages-artifact and deploy steps were skipped on the PR event. Live Pages publication and live destination checks must follow the merge to `main`.

13. **OPEN — documentation consolidation/pin pending.** I reviewed the README, tracked Markdown/MDX, customer-guide navigation, and rendered artifact. Customer guides are operation-focused and distinguish synthetic or implementation-derived evidence from provider-confirmed behavior; contributor inventories and verification material are outside customer navigation. At this exact SHA, `docs/independent-review.md` has the R1 report for cb8 and the R2 report for c1, not a current consolidated 8a report. The candidate checklist still pins template 987 rather than new standard commit 25aeb78. The v0.4 release note and its published destinations also remain pending.

14. **OPEN.** This is R1’s completed all-16 audit at exact 8a22252. Reviewer 2 has reported that its independent probes pass, but a consolidated R2 all-16 section at this same SHA is not yet in the review record. Two-reviewer final documentation signoff therefore remains pending.

15. **PASS.** SyntheticReplay checks method, origin, escaped path, repeated query values, headers, and body before returning paired status/headers/body; mismatch, duplicate, unexpected, and unconsumed exchanges fail. REST, GraphQL, CLI, and event tests use offline paired exchanges; the native HTTP/2 replay checks frame traffic, consumption, and close. The SDK does not implement the browser OAuth callback/PKCE exchange; docs assign it to the caller.

16. **OPEN — public CLI release/install pending.** `cmd/go-alexa` is a separate CLI module with help, JSON output, nonzero failures, file/stdin/environment credentials, explicit export, code-based linking and refresh, endpoint discovery/control, player state, and cancellable event cleanup. Linux and Windows CLI checks pass. The required SDK v0.4.0 is not yet published, the local `replace` remains, and no public `go install` at a tag or separate public consumer install has been demonstrated.

## Finding disposition

- **R1-F1 — RESOLVED at 8a22252 (items 4 and 7).** The cb8 named-result bypasses are now covered by exact-SHA negative tests for a novel fixed value returned through a named result and a generated map escaping through a named result, plus a caller-sourced positive control. The model-inventory race suite and exact CI pass. No Alexa implementation finding from this R1 audit blocks merge; Pages publication, public SDK/CLI release/install, template repin, and same-SHA R2 record remain open.

## Reviewer 2 — independent audit at 8a22252d064857750c053fbd3103f6c79be463af

I did not implement Alexa SDK or CLI changes. I reviewed a read-only Git archive of exact commit `8a22252d064857750c053fbd3103f6c79be463af` at `C:\Users\andre\AppData\Local\Temp\go-alexa-8a-r2-review-20261003\source`. I checked all 16 requirements against the archive, its checklist pinned to template `987b9c34a6b927472c21604617b6842a4238746b`, the linked standards, and the current Item 4 addendum in template `25aeb783126c4049b3bc49286ee7808069db28e9` (which explicitly covers helper-returned values/maps, named results with bare returns, and returned callbacks).

Exact-SHA CI [37182072283](https://github.com/portpowered/go-alexa/actions/runs/37182072283) and Documentation [37182072321](https://github.com/portpowered/go-alexa/actions/runs/37182072321) both completed successfully with head SHA 8a22252. CI ran Linux verify and Windows CLI; its combined non-generated coverage was 84.6% (2,397/2,834), above the 80% floor and below the 90% target. Documentation built 250 pages and checked 44,910 internal links. I downloaded artifact 11295876503 and inspected it: all 250 HTML pages, including the site root, carry the implementation-derived label; no replacement characters were found. The PR run did not deploy Pages.

I ran these additional checks against the exact archive: `go run ./tools/modelinventory -check`, `go run ./tools/apiroutes --check`, and `go run ./tools/openapibundle -check`; all passed. A race-enabled Go overlay test file outside the archive re-probed the prior cross-function package-global map and helper-returned fixed-value escapes, the named-result fixed-value and generated-map escapes, and returned callbacks that create a fixed wire value or mutate a generated map. Each negative probe was rejected, while caller-owned named values and open map keys were accepted. No archive source was modified.

## Item verdicts

1. **PASS.** The README, public packages, examples, and customer guides describe a reusable Alexa client. I found no consuming-application adapter or rollout instructions.

2. **PASS.** The README and MDX guides cover supported operations, account linking, explicit refresh, errors, transport configuration, endpoint workflows, player state, events, and CLI use. Examples match the public Client/Session API. Guides distinguish synthetic examples and implementation-derived contracts; browser callback state and PKCE are assigned to the consuming application because the SDK does not implement that browser flow.

3. **PASS.** The README has Go version, CI, coverage, release, Go Reference, license, and documentation badges. The links use this repository and its reports; no template repository values remain. The published latest release is still v0.3.2.

4. **PASS at 8a; checklist pin needs refresh.** The wire schemas, generated models/routes, model inventory, source gates, and negative controls cover the inventoried REST, GraphQL, event, behavior-payload, and external HTTP/2 traffic. The model inventory has 1,507 lines and 1,496 rows. Source searches found generated dependency wire types and custom codecs in the generated model package; the handwritten `AuthConfig` and `DeviceAuthConfig` are documented legacy caller-side configs with no JSON fields or codecs. The model gate rejects unlisted/anonymous JSON models and forged generated markers; tests also cover GraphQL selection/decoder drift, enums, keys, aliases, mutations, and generated-map escapes. The exact archive inventory, route, and OpenAPI bundle checks passed. My overlay re-probes rejected the old package-global map and helper-return bypasses, named-result escapes, and returned-callback variants while retaining caller-open positive cases. This resolves the earlier c1 R2-F10/F11 and cb8 R1-F1 code findings at 8a. The 8a checklist still pins template 987; it does not yet include the additional Item 4 paragraph from template 25a, which must be copied when the checklist is repinned.

5. **PASS.** Exact-SHA CI 37182072283 passed its blocking Linux and Windows jobs. The workflow pins Go 1.24.2 and golangci-lint v2.3.0; root and CLI configs both use literal `linters.default: all`. `make lint` runs the full SDK and nested CLI; the workflow also checks generation, module metadata, formatting, build, vet, race tests, inventory, and CLI module checks.

6. **PASS — floor met; target short.** Exact CI reports combined non-generated coverage of 84.6% (2,397/2,834), above 80% and below the 90% target. Package results are SDK 83.0%, public models 86.5%, GraphQL 94.2%, and REST 83.8%. Generated-code exclusions are documented in the coverage tool and verification guide.

7. **OPEN — public consumer/release proof.** The SDK, public model package, generated dependency models, and transport-specific packages follow the required boundaries, with responsibility-specific schemas and a checked model inventory. However, `cmd/go-alexa/go.mod` still requires SDK v0.4.0 through `replace github.com/portpowered/go-alexa => ../..`. Public versions list only SDK v0.1.0–v0.3.2 and no CLI versions. A clean public-proxy consumer check remains pending until publication.

8. **PASS.** Client construction uses validated functional options and service defaults. Account tokens and cookies are supplied on explicit Session creation instead of being kept as mutable reusable-client account state.

9. **PASS.** Accounts use explicit Session values and event connections. Callers own token state and can close a Session or event stream; tests cover isolation, cancellation, and cleanup.

10. **PASS.** REST, GraphQL, and event traffic have separately injectable HTTP clients/transports. The native HTTP/2 replay sets `DialTLSContext` to return an offline `net.Pipe`, exercises the real HTTP/2 codec and paired frames, and waits for server shutdown.

11. **PASS.** Refresh and cookie exchange are explicit operations. Refresh returns credentials without silently replacing the Session token; documentation assigns storage and installation to the caller.

12. **OPEN — main publication pending.** The exact Documentation run built the generated reference and customer guides and checked all 44,910 internal links. The PR build uploaded artifact 11295876503 but did not deploy it to GitHub Pages. Live main-site publication remains a post-merge check.

13. **OPEN — current checklist/review record closeout.** I reviewed the README and all tracked Markdown/MDX documents, the guide navigation, and the exact rendered artifact. Customer copy is caller-focused; inventories and verification detail stay in contributor docs, and the 10 rendered guides link to matching reference pages. However, the 8a checklist still pins template 987 instead of the current 25a Item 4 addendum, and `docs/independent-review.md` at 8a contains R1 at cb8 and R2 at c1, not a consolidated pair for 8a. Root’s pending checklist repin and report consolidation should close these record issues before final checklist signoff.

14. **OPEN.** R1 reviewed cb8 and found the named-result value/map bypass; I re-probed the fix and found it resolved at 8a. The current repository review record does not yet contain both reviewers’ all-16 evidence for exact 8a. Root has R1’s separate 8a report; consolidate it with this R2 report in the canonical review file, then have both reviewers verify the final documentation/checklist commit before checking this item.

15. **PASS.** Synthetic replay pairs match method, origin, escaped path, repeated query values, complete headers, and body before releasing each response; mismatches, duplicates, unexpected calls, and unconsumed pairs fail. REST, GraphQL, event HTTP/2, and CLI tests use offline paired requests/responses or ordered frames, including auth failure and event close/cancellation. No sanitized live captures are claimed; fixtures and docs label examples synthetic. OAuth callback state/PKCE is not an SDK flow and is assigned to the caller.

16. **OPEN — public CLI tag/install pending.** The standalone CLI is a nested module that imports the public SDK. It implements code-based linking, explicit refresh/export, endpoint discovery, player state and explicit power control, JSON output, nonzero failures, environment/stdin/file credentials, redaction, and cancellable event cleanup. Offline paired command tests cover auth success/failure, reads/control, and event cleanup. CI runs its all-linter, build, race, vet, and module checks. The local replace is still present; there is no CLI tag or public `go install`/consumer proof yet. The release workflow is configured to reject replace directives and install the tagged CLI through the public Go proxy.

## Finding dispositions at 8a

- **c1 R2-F10 (package-global generated-map escape): RESOLVED.** A regression test is present and my independent overlay reproduced the rejected cross-file map mutation.
- **c1 R2-F11 (explicit helper-returned fixed value): RESOLVED.** The model gate test and my overlay rejected direct, chained, aliased, and named-result fixed values.
- **cb8 R1-F1 (named-result value/map escapes): RESOLVED at 8a.** `wireReturnExpressions` follows named result identifiers on bare returns; negative tests and my overlay reject fixed-value and generated-map escapes, while caller-owned values pass.
- **R2 release/documentation closeout:** OPEN. Repin the provider checklist to template 25a with its Item 4 addendum, consolidate exact-8a R1/R2 evidence in the canonical review document, publish Pages from main, then publish SDK v0.4.0 and the nested CLI tag and verify public consumer installation. These are release/record tasks; I found no remaining SDK/CLI implementation blocker at 8a.
