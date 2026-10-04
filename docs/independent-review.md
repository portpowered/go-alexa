# Independent review

Implementation reviewed: `8a22252d064857750c053fbd3103f6c79be463af`.
Final published-source and customer documentation state: `b74b322b38f8342c480be9fdce768e7cc1a526fa`.
Standard: shared template `25aeb783126c4049b3bc49286ee7808069db28e9`.

The two reports below are independently authored by `cli_inventory_plan` and
`independent_tplink_tuya`; neither implemented the Alexa SDK or CLI.
Both reviewers confirmed all sixteen requirements and the consolidated record.
SDK `v0.4.0` and CLI `cmd/go-alexa/v0.4.0` are published. Both reviewers verified
public installation and the deployed site. Prior reports remain in Git history.

# Reviewer 2 — final independent audit at main b74b322b38f8342c480be9fdce768e7cc1a526fa

I did not implement the Alexa SDK or CLI. My implementation audit was performed independently against exact code commit `8a22252d064857750c053fbd3103f6c79be463af`; the final main commit `b74b322b38f8342c480be9fdce768e7cc1a526fa` is its descendant. `git diff --name-status 8a22252..b74b322` contains only `README.md`, the CLI module’s `go.mod`/`go.sum`, `docs/guides/cli.mdx`, `docs/template-checklist.md`, and `docs/independent-review.md`. No production SDK/CLI source, schema, generator, model, replay fixture, or test changed after the audited implementation. The final checklist pins shared template `25aeb783126c4049b3bc49286ee7808069db28e9`, including the helper/named-result Item 4 requirements.

## Final-state evidence

- Exact-main [CI run 37185198117](https://github.com/portpowered/go-alexa/actions/runs/37185198117) succeeded at b74b322. Linux verification passed generation, module metadata, formatting, pinned all-linter checks, builds, race tests, and non-generated coverage. Windows CLI credential-permission, lifecycle, build, vet, and module checks passed. Coverage was 84.6% (2,397/2,834): above the 80% blocking floor and below the 90% target; SDK 83.0%, API models 86.5%, GraphQL 94.2%, REST 83.8%, dependency models 100% (3/3 statements).
- Exact-main [Documentation run 37185198209](https://github.com/portpowered/go-alexa/actions/runs/37185198209) succeeded, including the main Pages deployment. Its rendered-link check passed. Reviewable artifact 11296188667 contained 251 HTML files; all 251 had the `implementation-derived` marker and none contained U+FFFD. The Pages artifact was 11296253531.
- After deployment, the live homepage, CLI guide, and upgrade guide returned HTTP 200. Each contained the implementation-derived marker and no replacement character. All three reported `Last-Modified: Sun, 04 Oct 2026 07:21:42 GMT`; the live CLI guide includes `cmd/go-alexa@v0.4.0`.
- SDK tag `v0.4.0` resolves to `8fbb9dc1bfe6064594dd500d8a170a74fba9ec42`; its [Release run 37183039192](https://github.com/portpowered/go-alexa/actions/runs/37183039192) succeeded, including generated-artifact, coverage, lint, API compatibility, test/vet/build, and public-proxy consumer checks. CLI tag `cmd/go-alexa/v0.4.0` resolves to `d6bc1d12c75705e0595ae8156ea223c249cd2e91`; its [CLI Release run 37183841537](https://github.com/portpowered/go-alexa/actions/runs/37183841537) succeeded, including tagged-module validation and public-proxy installation.
- I independently tested public CLI installation from a new, initially empty temporary workspace/cache using Go 1.24.2 and `GOPROXY=https://proxy.golang.org,direct`, `GOSUMDB=sum.golang.org`: `go install github.com/portpowered/go-alexa/cmd/go-alexa@v0.4.0` succeeded; `go-alexa --help` exited 0. `go version -m` identified both CLI module v0.4.0 and SDK v0.4.0, with no replacement entries.

## Item verdicts

1. **PASS.** README, exported packages, examples, and customer guides describe a reusable client. They contain no consuming-application adapter or rollout plan.

2. **PASS.** README and customer guides cover supported operations, authentication, errors, transport injection, endpoint workflows, player state, events, and CLI use. Examples match Client/Session APIs. The authentication guide assigns browser callback/state/PKCE to the consuming application because the SDK implements code-pair registration, not the browser callback flow.

3. **PASS.** README badges cover Go version, CI, coverage, release, Go Reference, license, and documentation. They point to the live repository/reports; no template-repository values remain. SDK and CLI v0.4.0 releases now exist.

4. **PASS.** The audited responsibility schemas, generated models/routes, model inventory, source gates, and negative controls cover REST, GraphQL, event, embedded behavior payloads, and external HTTP/2 traffic. The inventory has 1,507 lines and 1,496 rows. Generated dependency wire types/codecs cover active wire structs; handwritten `AuthConfig` and `DeviceAuthConfig` are legacy caller-side configuration with no JSON fields/codecs. Gates reject unlisted/anonymous models, forged generated markers, schema/model/decoder drift, wrong enum/key values, mutations, and helper/global/named-result map or fixed-value escapes. Race-enabled overlay probes rejected prior package-global/helper-return and named-result escapes while preserving caller-defined values/open keys. The final checklist includes template 25a’s named-result and returned-callback language. No implementation finding remains open.

5. **PASS.** Exact-main CI succeeded on Linux and Windows. Root and CLI configurations use literal `linters.default: all`; CI runs the full SDK and nested CLI lint/build/test/vet/module checks, plus generated artifact and inventory gates.

6. **PASS — floor met, target short.** Exact-main coverage is 84.6% (2,397/2,834), above the blocking 80% floor and below the 90% target. Generated code is excluded by the maintained-coverage tool.

7. **PASS.** SDK, API models, dependency models, transports, schemas, and the CLI remain separate modules/packages. `cmd/go-alexa` requires published SDK v0.4.0 and no longer has a local `replace`. The public install and both release workflows passed.

8. **PASS.** `alexa.NewClient(opts ...Option)` uses validated functional options and defaults. Account credentials are supplied to explicit Session creation rather than stored as mutable reusable-client state.

9. **PASS.** Sessions and event connections hold account-specific state; callers own token refresh and stream lifetime. Close/cancellation paths and session isolation have race-tested coverage.

10. **PASS.** REST, GraphQL, and events have independently injectable transports. The native HTTP/2 replay uses an injected offline `net.Pipe`, checks paired frames/consumption, and waits for server shutdown; no live provider call is needed.

11. **PASS.** Refresh and cookie exchange are explicit. Refresh returns credentials without silently replacing session state; documentation directs the caller to store and install tokens.

12. **PASS.** Main Pages deployment succeeded in exact Docs run 37185198209. The live home, CLI, and upgrade pages return 200 and show the implementation-derived marker.

13. **PASS.** The final checklist pins template 25a. The CLI README and guide now give the released `go install ...@v0.4.0` command; rendered HTML contains the same instruction. The exact artifact has an implementation-derived marker on every page and no replacement characters. Contributor inventory/review material is not in customer navigation; customer guidance stays focused on library usage.

14. **PASS for the reviewed implementation and final documentation delta.** Independent R1 and R2 all-16 reports are recorded for the exact implementation SHA 8a22252. The subsequent b74 changes are limited to customer installation instructions, nested-module sums, checklist/review records; I independently rechecked those files, exact-main CI, and the deployed site. No source implementation changed after the reviewed SHA.

15. **PASS.** Offline replay uses complete paired exchanges and ordered frames. The synthetic matcher validates method, origin, escaped path, repeated query values, headers, and body before responses are released; unexpected, duplicate, mismatched, or unconsumed traffic fails. REST, GraphQL, event HTTP/2, and CLI tests cover authentication errors and cancellation/close cleanup. Synthetic behavior is not represented as a sanitized live capture.

16. **PASS.** The nested CLI is separately tagged `cmd/go-alexa/v0.4.0` and imports public SDK v0.4.0. Release workflow and Windows/Linux CI passed. Empty-cache public `go install` and `--help` succeeded with both module versions embedded and no replace entry. Commands provide JSON results, nonzero failures, explicit token export/refresh, secure env/stdin/file credentials, safe output, and cancellable event cleanup.

## Finding disposition

The prior R2 findings for package-global generated-map escape, helper-returned fixed values, and named-result value/map escapes are resolved at 8a22252 with race-tested negative and caller-owned positive controls. No SDK/CLI implementation blocker remains. Main Pages publication and public SDK/CLI releases and install have now been verified independently.

---

# Independent Alexa publication review (R2)

Status: final Alexa R2 review at b74 is complete. The canonical review record contains both independent final reports, is linked from the checklist, and all other items are complete. This R2 signoff for item 14 is PASS; the repository checklist remains open until the responsible maintainer checks it. This report is stored in TEMP, separate from repository source.

- Reviewed implementation: `8a22252d064857750c053fbd3103f6c79be463af`.
- Published CLI-module finalization: `d6bc1d12c75705e0595ae8156ea223c249cd2e91`.
- Current documentation follow-up: `b74b322b38f8342c480be9fdce768e7cc1a526fa` (only README.md and docs/guides/cli.mdx changed from d6).
- Current checklist pin: shared template `25aeb783126c4049b3bc49286ee7808069db28e9`.
- Consolidation: both final independent all-16 reports are present in the current `docs/independent-review.md`, linked from `docs/template-checklist.md`; item 14 is the only checklist box left open pending maintainer signoff.
- I have no Alexa implementation authorship. Archives were read-only; all consumer/install probes and this report are under `%TEMP%`.

## Evidence

- Existing canonical independent records in `docs/independent-review.md` contain separate R1 and R2 all-16 reviews of implementation commit 8a. Both identify the named-result source-gate fixes at 8a and find no remaining SDK/CLI implementation blocker. I independently checked the final commit delta and publication evidence instead of treating those older verdicts as final signoff.
- Main CI run `37183250064` succeeded at exact d6; Documentation run `37183250051` succeeded at exact d6, including build, rendered-link check, Pages artifact upload, and Pages deploy. Its artifact has 251 HTML files, generated OpenAPI/GraphQL/AsyncAPI reference routes, a visible `implementation-derived` label, and `coverage.json` = 84.6%. Live root, docs root, CLI guide, and upgrade guide returned HTTP 200 at the published Pages site.
- SDK release tag `v0.4.0` points to `8fbb9dc1bfe6064594dd500d8a170a74fba9ec42`; release workflow `37183039192` succeeded. The nested CLI tag `cmd/go-alexa/v0.4.0` points to d6 and its CLI Release run `37183841537` succeeded. SDK release notes link to the upgrade guide and pinned verification doc; CLI release notes give `go install github.com/portpowered/go-alexa/cmd/go-alexa@v0.4.0` and link the CLI guide.
- Fresh isolated Go 1.24.2 cache: `go install github.com/portpowered/go-alexa/cmd/go-alexa@v0.4.0` succeeded from `proxy.golang.org`; installed binary `--help` exited successfully and shows credential-safe command usage.
- Separate clean consumer module: `go get github.com/portpowered/go-alexa@v0.4.0`, `go mod tidy`, and `go test ./...` passed after compile-checking public SDK, model, and dependency-model import paths and an SDK method signature. The consumer had no local replace and used a fresh module cache.
- b74 is a two-file docs-only delta from d6. It replaces stale “first standalone CLI release is pending” copy in both README and CLI MDX with the released `@v0.4.0` install command. I verified the source diff and no stale “pending/until then” statement remains in those customer-facing pages.
- Exact b74 CI `37185198117` succeeded: Linux verify and Windows CLI jobs passed. Coverage on that exact SHA is 84.6% (2,397/2,834), with SDK 83.0%, API models 86.5%, GraphQL 94.2%, REST 83.8%, and dependency models 100% (3 statements). Exact b74 Documentation run `37185198209` succeeded, including rendered-link check, Pages artifact upload, and deploy with pages_build_version b74. Live homepage, docs root, CLI guide, upgrade guide, OpenAPI reference page, and GraphQL reference returned HTTP 200; live `coverage.json` returned 84.6%. The live CLI guide contains the `@v0.4.0` install command and no pre-release pending text.

## Verdicts

1. **PASS.** README, public packages, examples, and customer guides describe a reusable Alexa library. No consuming-application adapter or rollout plan is in this repository.
2. **PASS.** README and customer MDX guides cover supported operations, account authentication/linking, errors, transport configuration/injection, endpoint workflows, player state, events, and CLI use. Examples match the SDK Client/Session API; caller-owned browser callback state and PKCE responsibilities are stated.
3. **PASS.** README badges cover Go version, CI, coverage, release, Go Reference, license, and documentation; repository and report destinations are specific to go-alexa.
4. **PASS at 8a implementation.** The API, GraphQL, event, and external HTTP/2 schemas and generated models/routes are tied to actual uses in the checked inventory. Existing independent source-gate tests and R2 overlays rejected package-global generated-map escape, helper-return fixed-value escape, named-result scalar/map escape, and callback mutations while retaining caller-open positives. Route, model-inventory, and OpenAPI-bundle checks passed in the exact 8a archive. No implementation source changed in d6 or b74.
5. **PASS.** Exact b74 CI `37185198117` succeeded on both Linux verify and Windows CLI jobs. The blocking run includes generation, pinned full-repository all-linter, module checks, build/race, CLI lifecycle, and coverage gates.
6. **PASS — floor met; target short.** Exact b74 CI reports combined non-generated coverage of 84.6% (2,397/2,834), above 80% and below 90%. Packages: SDK 83.0% (1,282/1,544), API models 86.5% (90/104), GraphQL 94.2% (277/294), REST 83.8% (745/889), dependency models 100% (3/3). Generated exclusions are reported in verification docs and the coverage tool.
7. **PASS.** SDK/model/transport boundaries and responsibility-specific schemas meet the requested layout. The nested CLI module imports SDK v0.4.0 without a local replace. A clean public-proxy consumer compiled against the public SDK and model paths. (The SDK release tag points to 8fbb, and the CLI module tag points to d6.)
8. **PASS.** `alexa.NewClient(opts ...Option)` uses validated options and defaults. Credentials are provided per explicit Session, not stored as reusable client account state.
9. **PASS.** Account state and event connections are held in explicit caller-owned sessions/connections; close, cancellation, and cleanup behavior are covered in tests.
10. **PASS.** REST, GraphQL, and events accept injected transports. The native HTTP/2 test uses an offline `net.Pipe`, exercises paired frames through the actual codec, consumes the exchange, closes, and waits for server shutdown.
11. **PASS.** Token/cookie refresh operations are explicit; `Session.RefreshAccessToken` returns credentials without silently replacing session token state. Guides assign persistence and explicit token installation to the caller.
12. **PASS.** Exact b74 Documentation run `37185198209` rendered the site, passed the rendered-link checker, uploaded its Pages artifact, and deployed with the b74 SHA in the deployment payload. The live root, docs root, CLI guide, upgrade guide, OpenAPI reference, and GraphQL reference returned HTTP 200.
13. **PASS.** I reviewed README, all tracked Markdown/MDX, guide navigation, the rendered b74 Pages site, and SDK/CLI release notes. The README and CLI guide now show the published v0.4.0 install command; live HTML has no stale pending-release text. Customer guides keep wire inventories and maintenance mechanics in contributor documentation. SDK release notes link to the live upgrade guide and pinned verification document; CLI release notes link to the live CLI guide.
14. **PASS.** The canonical `docs/independent-review.md` contains separate all-16 reports from two reviewers for implementation 8a and the final b74 documentation/publication state. The checklist links to that record. Items 1–13 and 15–16 are complete; the earlier stale-install-copy finding is resolved and no other finding remains open. Both reviewers independently verified the current publication state.
15. **PASS at 8a implementation.** Existing independent review and tests verify paired offline request/response replay, complete method/origin/escaped-path/repeated-query/header/body matching, mismatch/duplicate/unexpected/unconsumed rejection, and actual HTTP/2 framed exchange teardown. The SDK does not implement browser OAuth callback/PKCE and documents that scope.
16. **PASS.** The CLI is a separate SDK-consuming module with command/help, JSON output, nonzero errors, file/stdin/environment credential input, explicit export, endpoint discovery/control, player state, and cancellable event cleanup. CLI tests and CI cover offline paired command flows. Public nested tag/release exists, and the clean public `go install ...@v0.4.0` test plus installed help succeeded.

## Finding disposition

- **A-R2-1 — RESOLVED in b74 (items 13, 16).** At d6 after CLI publication, README said the first CLI release was pending and suggested running from a checkout; the CLI guide also said publication was pending. b74 replaces both passages with `go install github.com/portpowered/go-alexa/cmd/go-alexa@v0.4.0`; the exact b74 Docs deployment succeeded and the live CLI guide has the install command with no stale pending text.
- **Previous source-gate findings — RESOLVED at 8a.** Canonical R1/R2 evidence and exact implementation source remained unchanged through d6/b74. No current implementation blocker found.
- **Final closeout — PASS.** The two independent all-16 reports are consolidated in `docs/independent-review.md`, linked from the checklist, and verified against final b74. CI, Pages deployment, SDK/CLI releases, public SDK consumer, public CLI install, live guide/reference links, and finding verification are complete.
