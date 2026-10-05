# Independent review record

At the user's request on 2026-10-04, final dual review and checklist closure
are deferred while known customer-facing fixes, required CI, and releases
are completed. These reports audit earlier snapshots and do not approve
the final release. Unverified checklist items remain open.

Status: **open; findings and final dual approval remain outstanding**.

The two reports below independently audit all 16 checklist items at source
`713031f0267ffccc18b4d34e849f108f9c65ed04`, against shared template
`62cc3cb5a1308dae8700f92052f99b1c455a1d98`.
They were delivered before either reviewer read this record or the other report.
Later implementation changes require verification on the final commit.
See the [current checklist](template-checklist.md). Historical reports remain
in Git history. Passing CI does not resolve the findings below.

## Reviewer 1 — initial independent audit

### Independent reviewer 1 — initial 16-item verdict

Reviewed source: `github.com/portpowered/go-alexa` at `713031f0267ffccc18b4d34e849f108f9c65ed04`.
Review date: 2026-10-04.
Scope: independent source, test, CI, documentation, and published-page audit. No source files were changed.

#### Summary

This SHA is **not ready for checklist sign-off**. Four implementation blockers are established: item 2 lacks schema-owned examples and validation; item 4 fails to surface known feature payload variants in the generated reference and omits their schema from the docs build; item 15 has four reproducible HTTP request-matching bypasses; and item 16 has no CLI logout command. Item 7 still needs consumer proof for the planned new SDK/CLI release, item 13 awaits inspection of the deliberately deferred review record, and item 14 cannot pass while findings remain open and two final reviews are not complete.

`make lint` and `make check` both pass in a fresh isolated checkout. GitHub CI run `37251625662` and Documentation run `37251625675` both completed successfully on the exact reviewed SHA. The public SDK/CLI v0.4.0 release exists; its evidence does not establish consumer installation for the planned v0.5.0 release.

#### Review setup and evidence

The checkout was the isolated archive at `C:/Users/andre/AppData/Local/Temp/go-ring-blind-review-lfed1dfa/checkout`; its `go.mod` identifies the Alexa module. The parent-directory manifest confirms the archive matches the requested source SHA, except for a neutral placeholder replacing `docs/independent-review.md`. The checkout has a fresh root commit with no parent or remote; no old Git objects were fetched or inspected. I read `docs/template-checklist.md` from this source, which identifies the exact shared template commit `62cc3cb5a1308dae8700f92052f99b1c455a1d98`, and consulted the isolated shared library/client standards.

Using fresh temporary Go/lint caches, `make lint` passed with zero issues in both the SDK and nested CLI modules (golangci-lint 2.3.0). `make check` passed formatting, lint, vet, build, race tests, model inventory, OpenAPI bundle, module tidy, and module verification. The combined non-generated production coverage is 84.6% (2,409/2,849 statements), above the enforced 80% floor and below the stated 90% target; package figures include `pkg/alexa` 83.1%, `alexaapimodels` 85.6%, GraphQL 94.2%, REST 83.7%, and generated-model package 100% over three non-generated statements.

I downloaded the exact-SHA documentation artifact. Its internal-link checker reported 44,914 internal links across 251 rendered pages. An independent content check found title and H1 content on all 246 generated documentation pages and substantive content on all 10 guides. I also inspected the published CLI and upgrade guides and a generated account-linking reference page for expected content; each returned the expected page content. The published v0.4.0 SDK and `cmd/go-alexa/v0.4.0` release notes point to those pages and the verification guidance.

#### Item verdicts

1. **PASS.** The reusable package, README, examples, and guides describe the Alexa client and CLI without a consuming-application adapter or rollout plan. The README and customer examples are provider/library oriented.

2. **OPEN — source blocker.** Customer guides cover operations, auth, errors, and transport injection with examples, and synthetic replay provenance is explicitly separated. However, `api/` has no schema `example`/`examples` entries, and CI does not validate sanitized request, response, and event examples against their owning schemas. In particular, there are no canonical full-envelope examples covering nested feature payloads, resource updates, relation changes, and failures as required by the checklist. Replay fixtures are identified as synthetic, which is good provenance practice but does not fill the schema-example requirement.

3. **PASS.** `README.md` includes Go version, CI, coverage, release, Go Reference, license, and documentation badges. The targets use the live repository, workflow, Pages coverage artifact, Go module, and license locations.

4. **OPEN — source blocker.** The wire/model inventory, generated models, schema gates, and rendered reference have substantial coverage, but the generated reference does not expose known feature-control variants. `api/feature-controls.yaml` defines known payload shapes, while `.github/workflows/docs.yml` passes only `api/openapi.yaml`, `api/asyncapi.yaml`, the GraphQL SDL, and guides to the site builder; it omits `api/feature-controls.yaml`. In the rendered GraphQL reference, `SetEndpointFeaturesInput` contains only `featureControlRequests`, and `FeatureControlRequest.payload` is rendered as generic JSON, without the known variants' required fields/result shapes. This is a direct violation of the checklist's generated-reference requirement.

5. **PASS.** The config pins golangci-lint v2.3.0, sets `linters.default: all`, and makes repository-wide lint blocking. `make lint`/`make check` passed independently, and exact-SHA blocking GitHub CI passed (`37251625662`). The reviewer was not the implementer. No lint exception or persisted JSON-key regression issue was found in this audit.

6. **PASS.** Deterministic synthetic paired fixtures and session/error cases are present; `make check` ran the race suite. Coverage is measured by package and combined, with generated code excluded from the combined measure. Combined coverage is 84.6%, meeting the 80% gate though below the 90% target.

7. **OPEN — publication evidence pending; source layout passes.** The public package is under `pkg/alexa`; provider models and dependency types are in the requested packages; schemas are split by API responsibility; the checked-in model inventory records schema components, generated declarations, generators, and use sites. A prior v0.4.0 release exists, but this review is for the next planned release. Separate SDK and nested CLI consumer/proxy installation evidence must be produced for the new tags; v0.4 evidence cannot establish v0.5 compatibility.

8. **PASS.** `NewClient` uses functional options with defaults and validation. Account credentials enter explicit `Session` options, not reusable client configuration.

9. **PASS.** Client configuration is reusable while credentials and cookie state live on explicit sessions. A nonnil injected `http.Client.Jar` is rejected with a distinguishable `HTTPClientCookieJarError`; two-session tests exercise isolation through a shared client. Session close and credential state are explicit.

10. **PASS.** The discovered production edges are REST, GraphQL, and the long-lived HTTP/2 event stream. Each has an injectable HTTP client or event transport seam, and tests drive the actual offline transport. No separate MQTT, WebSocket, RTC, or raw-socket production edge was found in the shipped modules.

11. **PASS.** Refresh and refresh-token cookie exchange are explicit operations that return credentials. Refresh does not silently replace session credentials; caller installation/storage obligations are documented in the authentication guide and API comments.

12. **PASS.** Customer guides are MDX under `docs/guides/`, link to generated reference pages, and render expected content. The exact-SHA site artifact passed the repository's link check; the independent scan covered every rendered page. I reviewed the external/release-note targets, including published v0.4.0 release links to the upgrade/CLI guides and verification guidance. The `externalDocs` source search found no schema-supplied external documentation links requiring a separate runtime check.

13. **OPEN — deferred document review.** I inspected the README and all tracked documentation files except `docs/independent-review.md`, which was deliberately withheld until this initial all-16 delivery. The remaining documents have clear contributor, release, customer, or fixture-maintenance purposes; customer pages are substantive and no duplicate internal review report was found outside the withheld file. Final verdict requires inspecting the deferred record and reconciling its links/findings.

14. **OPEN.** This is reviewer 1's independent evidence. Two independent final reviews and dispositions for every finding are still required, and items 2, 4, 7, 13, 15, and 16 remain open.

15. **OPEN — source blocker.** `pkg/testing/synthetic_replay.go` matches method, origin from `req.URL`, escaped path, parsed `URL.Query()`, headers, and body, but does not validate effective authority (`Request.Host`), URL user information, opaque URLs, or malformed raw queries. A compile-valid harness in an isolated temporary copy configured one synthetic `GET https://example.invalid/item` pair and passed each of these mismatches to `SyntheticReplay.RoundTrip`; all four returned the paired response with nil error:
    - `req.Host = "attacker.invalid"`;
    - userinfo `https://secret-user:secret-pass@example.invalid/item`;
    - malformed raw query `?token=%ZZ` (discarded by `URL.Query()`);
    - opaque URL with an unrelated opaque target.

   The harness was outside the reviewed checkout and was compiled against the reviewed replay package. Existing negative tests cover method, URL host, path, repeated query, header, and body mismatches, but not these four cases. Mismatch diagnostics also include request headers/body/query values, so a follow-up should ensure secrets are not echoed.

16. **OPEN — source blocker.** The nested CLI covers code-pair linking, explicit refresh/export, discovery and endpoint operations, safe credential input/storage, JSON output, cancellation, and offline paired transports. The CLI guide documents ownership and storage. However, the auth dispatcher/help expose `link`, `refresh`, and `export` only (`cmd/go-alexa/internal/cli/commands.go`, `cmd/go-alexa/internal/cli/app.go`), and the guide has no logout sequence or command. The checklist expressly requires logout. The code-pair login does perform browser consent and the provider activation/register call; a localhost callback/PKCE listener appears inapplicable to this provider-specific device code flow. The missing logout alone keeps the item open.

#### Process note

One initial parallel command accidentally omitted its `workdir` and read a go-ring file/search result from the active workspace. I discarded that output and used no go-ring source, review record, Git history, or other reviewer report as Alexa evidence. The correct Alexa archive's parent folder name contains `go-ring`; the module declaration and manifest were checked before continuing. This report is written only to the independent-review temporary directory, not the repository.

Signed: /root/alexa_blind_final_1 (independent reviewer 1), 2026-10-04.


## Reviewer 2 — initial independent audit

### Independent review 2 — initial baseline

**Reviewed source commit:** \`713031f0267ffccc18b4d34e849f108f9c65ed04\`

**Review date:** 2026-10-04 (America/Los_Angeles)

**Reviewer:** Independent reviewer 2 (Codex)

**Scope:** All 16 items in \`docs/template-checklist.md\`, compared with shared template commit \`62cc3cb5a1308dae8700f92052f99b1c455a1d98\`.

#### Isolation and method

I inspected only \`C:\Users\andre\AppData\Local\Temp\go-ring-blind-review-sd4__1oi\checkout\`, an isolated source snapshot. Its parent manifest identifies source SHA \`713031f0267ffccc18b4d34e849f108f9c65ed04\`, snapshot SHA \`858d3da450e7d329a486af8e5d73e6c4faabd18f\`, archive SHA-256 \`0b5a756226b00ca134c972fc494842f6778a567554982bacb7e4cad6530dd70b\`, no parent/remote, and 292 file hashes matching the source. The one documented exception is \`docs/independent-review.md\`, replaced with a neutral placeholder; its original contents were not read or hashed. I did not use the original repository, its review record, old Git objects, remote history, or another reviewer's report. No source files were edited or committed.

I read the repository checklist and the shared template's library, client design, verification, and website standards. I inventoried tracked documentation (21 Markdown/MDX files including README) and reviewed the rendered site artifact from Documentation run \`37251625675\`. I checked source code, schema and generated-model inventories, tests, workflows, and the exact commit's CI result.

#### Verification

- Go 1.24.2 and pinned golangci-lint v2.3.0 were used. \`make lint\` passed for the root and CLI modules (0 findings). The reviewed temporary wrapper only added \`--allow-parallel-runners\` after \`run\` to avoid a pre-existing global linter lock; it preserved the pinned executable, configuration, arguments, and exit status.
- \`make check\` passed with temporary Go caches. It covered format, lint, vet, build, race tests, model inventory, OpenAPI bundle, and root/CLI module verification.
- GitHub CI run \`37251625662\` for the exact source SHA completed successfully, including generation checks, blocking lint, build/race tests, and the 80% coverage gate. Documentation run \`37251625675\` completed successful build and deploy.
- The downloaded documentation artifact passed the workflow link checker: 44,914 internal links across 251 rendered pages. I found no external anchor destinations in the rendered site and no schema \`externalDocs\` field. Release-note workflow links target the checked CLI guide.
- A separate temporary consumer module using a local replacement to this exact source passed \`go test ./...\` while importing \`pkg/alexa\`, \`pkg/alexaapimodels\`, and \`pkg/dependencymodels\`. This verifies source import paths only. Publication/consumer-install evidence for the next SDK and CLI tags remains separate and pending; I did not query a public module proxy.

#### Verdicts

| # | Verdict | Evidence and reason |
|---|---|---|
| 1 | **PASS** | The module is \`github.com/portpowered/go-alexa\`; reusable code is in \`pkg/alexa\`, public projections in \`pkg/alexaapimodels\`, and examples/README/guides describe the Alexa client without a consuming application's adapter or rollout assumptions. The independent consumer probe imported the public packages. |
| 2 | **OPEN — F-03** | \`README.md\` and the MDX guides document authentication, operations, errors, configuration, and evidence status; \`docs/verification.md\` distinguishes synthetic pairs from captures. However, \`rg '^\s+(example|examples):' api\` found no canonical schema examples, and CI has no gate validating request/response/event examples against their schemas. Synthetic pairs under \`tests/replay/fixtures/synthetic/\` are separately labeled, but do not meet the checklist's schema-example requirement. |
| 3 | **PASS** | \`README.md\` contains Go version, CI, coverage, release, Go Reference, license, and documentation badges, all using \`portpowered/go-alexa\` destinations. |
| 4 | **OPEN — F-04** | The source has checked-in OpenAPI, AsyncAPI, compatibility, feature-payload, and GraphQL schemas; generated models; \`docs/model-inventory.md\`; \`tools/modelinventory -check\`; \`tools/apiroutes --check\`; network inventories; and CI gates. The model/route checks passed on the reviewed SHA. The rendered reference still hides known nested payloads: the generated \`submitBehaviorPreview\` page displays \`sequenceJson: "string"\` and notes the embedded schema lives in \`api/behaviors.yaml\`, but does not show its fields or variants. The rendered GraphQL \`FeatureControlRequest\` describes \`payload\` only as scalar \`JSON\`; known control payload schemas exist in \`api/feature-controls.yaml\`, but \`.github/workflows/docs.yml\` does not feed that schema to the docs generator. The rendered pages therefore do not expose the required customer-usable nested payload variants/request examples. |
| 5 | **PASS** | Root and CLI linter configs state \`linters.default: all\`; CI pins Go 1.24.2 and golangci-lint v2.3.0 and runs blocking full-module lint. Narrow exceptions name paths/rules and carry explanations. Local \`make lint\` and exact-source CI passed. |
| 6 | **PASS** | Synthetic request/response, error, and session tests exist across REST, GraphQL, events, and CLI. CI runs race tests and filters generated code before enforcing combined non-generated coverage of at least 80%; the coverage page targets green at 90%. Exact-source CI passed. |
| 7 | **PASS** (source) | Package boundaries follow the required \`pkg/alexa\`, \`pkg/alexaapimodels\`, \`pkg/dependencymodels\`, and transport structure. Inventory records generated and compatibility types. The isolated consumer probe passed. Public installation of the next release tag remains a separate publication check. |
| 8 | **PASS** | \`pkg/alexa/client.go\` configures a reusable client through functional options and validates defaults/conflicts. Account credentials are session options, with explicit tests for invalid options and caller-visible token state. |
| 9 | **PASS** | Account credentials/cookies and event state live in explicit sessions. Effective shared HTTP clients with a nonnil cookie jar produce typed \`HTTPClientCookieJarError\`; tests cover isolation of two accounts through one reusable client and full outbound requests (\`pkg/alexa/client_options_session_test.go\`). |
| 10 | **PASS** | Active HTTP/REST, GraphQL, and HTTP/2 event edges accept injected clients/transports. \`WithEventTransport\` and an HTTP/2 \`DialTLSContext\` connection seam are documented; \`pkg/alexa/http2_native_replay_test.go\` exercises framed traffic using an offline connection. The pinned HTTP/2 dependency is inventoried in \`api/external/http2.yaml\`; the route scan covers the shipped modules and examples. |
| 11 | **PASS** | \`GenerateCodePair\`, \`RegisterWithCodePair\`, and \`RefreshAccessToken\` are explicit client/session operations. Refresh returns caller-owned credentials; \`TestSessionRefreshReturnsCredentialsWithoutImplicitlyReplacingThem\` and the authentication guide document that token state/storage remains the caller's responsibility. |
| 12 | **PASS** | All customer guides are MDX under \`docs/guides/\`, link to the generated reference, render in the exact Documentation artifact, and its link check covered the root plus generated pages. The workflow's link result was 44,914/251 with no failures. No runtime \`externalDocs\` links were present. |
| 13 | **OPEN — audit limitation** | I reviewed the 20 readable tracked Markdown/MDX files for purpose, audience, duplication, and incoming links, and reviewed the rendered Pages artifact. The 21st file, \`docs/independent-review.md\`, was deliberately replaced with a neutral placeholder by the isolation manifest. The review-record content is therefore unexamined, so I cannot attest to the checklist requirement that every tracked document, including the current review record, is appropriate. |
| 14 | **OPEN** | This is one independent initial report, on the baseline commit. The required second review, one current repository record containing both reviewers and every finding's final disposition, and re-verification of fixes at a final commit are not complete. This report does not check off item 14. |
| 15 | **OPEN — F-01, F-02** | Paired exchanges and consumption checks exist, but temporary compile-valid probes against the public \`pkg/testing.SyntheticReplay\` API demonstrated the request-identity and diagnostic gaps below. |
| 16 | **OPEN — F-05** | The separate CLI uses the public SDK, has auth link/refresh/export, discovery, power/player reads, event streaming, JSON output, cancellation/cleanup tests, secret redaction, and Windows credential permissions. However, \`cmd/go-alexa/internal/cli/commands.go\` handles only \`link\`, \`refresh\`, and \`export\`; root/group help lists those commands, and there is no logout or credential-removal command. The CLI guide explains private credential storage but provides no logout step. The next public SDK/CLI release-tag consumer installation is also a separate pending publication check. |

#### Findings and dispositions

##### F-01 — Synthetic replay accepts mismatched effective authority and URL identity (item 15)

At \`pkg/testing/synthetic_replay.go:167-173\`, request matching compares method, \`URL.Scheme\` + \`URL.Host\`, escaped path, parsed \`URL.Query()\`, headers, and body. It does not compare \`Request.Host\`, \`URL.User\`, or \`URL.Opaque\`, and it does not reject malformed \`RawQuery\` input that \`URL.Query()\` silently drops.

A compile-valid temporary module using the exported replay API built the same synthetic pair, then asserted rejection after independently mutating each field. Its tests failed because each mismatch was accepted and the paired response returned:
- \`Request.Host = "attacker.test"\`
- URL userinfo
- an opaque URL target
- \`RawQuery = "expand=all&bad=%zz"\`

**Disposition:** Open on source SHA 713031f. Parent has assigned a source fix; this reviewer has not inspected it. Re-test against a fresh post-fix snapshot is required.

##### F-02 — Synthetic replay mismatch diagnostics disclose secrets (item 15)

At \`pkg/testing/synthetic_replay.go:175-183\`, the mismatch error formats the full URL, parsed query, request headers, and request body. The same temporary module asserted that a mismatch diagnostic redact Authorization values; it failed with the exact output:

\`\`\`
synthetic replay request 1 (probe) mismatch: POST https://expected.test/v1?expand=all, query=map[expand:[all]], headers=map[Authorization:[Bearer diagnostic-secret]], body=""
\`\`\`

**Disposition:** Open on source SHA 713031f. Parent has assigned a source fix; this reviewer has not inspected it. Re-test safe diagnostics and secret-bearing query/body/header cases against a fresh post-fix snapshot.

##### F-03 — Canonical schema examples and schema-validation gate are missing (item 2)

The canonical files under \`api/\` contain no OpenAPI/AsyncAPI \`example\` or \`examples\` declarations. CI bundles/checks OpenAPI and builds documentation, but it does not validate every sanitized request, response, and event example against its owning schema. The paired synthetic fixtures are explicitly classified as synthetic and are outside canonical schema components.

**Disposition:** Open. The baseline still lacks the required canonical examples and CI validation.

##### F-04 — Published reference omits known nested payload variants (item 4)

The exact rendered artifact's \`docs/openapi/behaviors/submitBehaviorPreview/index.html\` shows a request snippet with \`sequenceJson: "string"\` and a description that directs readers to the separate \`api/behaviors.yaml\`; the nested sequence's required fields are not in the generated page. \`docs/graphql/types/featurecontrolrequest/index.html\` renders \`payload: JSON\`, and \`docs/graphql/types/json/index.html\` renders only \`scalar JSON\`. Known feature payload definitions and actual helpers exist in \`api/feature-controls.yaml\` and \`pkg/dependencies/graphql/features.go\`, but the docs workflow inputs are OpenAPI, AsyncAPI, GraphQL SDL, and guides; it does not publish the feature-payload schema or map each known operation to its required payload shape.

**Disposition:** Open. Customer-visible reference output needs generated nested variant content and usable request examples.

##### F-05 — CLI does not provide logout (item 16)

\`cmd/go-alexa/internal/cli/commands.go\` has no \`logout\` case/handler. \`cmd/go-alexa/internal/cli/app.go\` lists link, refresh, export, discovery, control, player, and events in root help. \`docs/guides/cli.mdx\` documents secure storage and credential inputs but has no logout instruction or removal command.

**Disposition:** Open. Add a customer-visible logout/removal command and document it; verify the workflow offline and in the fresh CLI snapshot.

#### Signature

I independently attest that the verdicts above describe the isolated source snapshot identified above, with the explicit review-record exception stated in the manifest. No changes were made to that source snapshot.

**/s/ Independent reviewer 2 (Codex)**

**Signed:** 2026-10-04 (America/Los_Angeles)
