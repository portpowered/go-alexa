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
   schemas and generated files are split by responsibility. Independent struct
   search did not find another handwritten production wire struct or anonymous
   provider object missing from the inventory. However, primitive/library
   constants listed under item 4 remain manually declared or absent from
   schema-generated definitions. Resolve that gap and verify imports from the
   separate consumer module.

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
  registration device IDs, and behavior provider IDs are not yet represented
  by schema-generated constants. Root was notified; planned changes are not
  counted as resolved.
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
