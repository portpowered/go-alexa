# Independent review record

The checklist remains open. This record retains the latest completed independent
review; earlier audits are available in Git history. Fixes after the reviewed
commit require both reviewers to inspect the same final implementation commit
and verify its CI, publication, and public installation evidence.

## Reviewer 1

Pending a complete independent review of the final implementation commit.

## Reviewer 2 — independent audit at c1d459e225809f3064b27aa0fe3e39ad60cc45b0

I did not implement this SDK or CLI. I reviewed a git archive of the exact c1d459e225809f3064b27aa0fe3e39ad60cc45b0 tree against the checklist pinned to template 987b9c34a6b927472c21604617b6842a4238746b and its linked standards. Exact-SHA CI [run 37175399314](https://github.com/portpowered/go-alexa/actions/runs/37175399314) completed successfully for Linux verify and Windows CLI. Exact-SHA Documentation [run 37175399237](https://github.com/portpowered/go-alexa/actions/runs/37175399237) built 250 pages and checked 44,910 internal links; I downloaded artifact 11292867058 and independently confirmed the implementation-derived label on all 250 HTML pages. The Pages deploy job did not run for this review build, so live publication remains open. From the archive, root and CLI race tests, model-inventory check, route check, and OpenAPI bundle check also passed. Two blind model-gate probes ran through a temporary Go overlay; no candidate source files were modified.

### Item verdicts

1. **PASS.** The public package is a reusable Alexa client. The README, examples, and customer guides describe the provider library and do not add a consuming-application adapter or rollout plan.

2. **PASS.** The README and customer MDX guides cover supported operations, code-based linking, token refresh, errors, transport injection, endpoint discovery and control, player state, events, and CLI use. They match the exported Client, Session, and credential APIs. The authentication guide assigns browser callback state and PKCE to the caller because the SDK does not implement that OAuth callback.

3. **PASS.** The README shows Go version, CI, coverage, release, Go Reference, license, and documentation badges. The badge destinations use this repository or its live reports; I found no template repository values.

4. **OPEN — R2-F10 and R2-F11 (items 4, 7).** The checked-in responsibility schemas, generated routes/models, model inventory, GraphQL evidence, and source checks cover the enumerated HTTP, GraphQL, event, and external HTTP/2 edges. The exact-tree inventory, route, and bundle checks passed, and the test suite contains negative model/key, decoder, enum-binding, mutation, and route controls. Two independent wire-construction bypasses remain:

       var saved map[string]any
       func mutate(caller string) { saved["brandNewCrossFunctionKey"] = caller }
       func retain(input *wire.Payload) { saved = input.Extra }
       func send(input *wire.Payload, caller string) {
           retain(input)
           mutate(caller)
           _, _ = json.Marshal(input)
       }

   A temporary overlay test showed the gate accepts this package-global map escape and later fixed key, even though the generated map is retained from and serialized with a generated model. Also, a helper-returned fixed value is not followed into a generated wire field: inventedValue returns "brandNewUnregisteredValue", and constructing wire.Payload{Value: inventedValue()} passes the gate. These are library-selected wire values, not caller-defined open input. Both gaps require negative tests and keep items 4 and 7 open.

5. **PASS.** Exact-SHA CI run 37175399314 passed the Linux verify and Windows CLI jobs. The Linux job regenerated and checked artifacts, checked module metadata and formatting, ran pinned golangci-lint v2.3.0 with literal linters.default: all, then build, race tests, and coverage. The Windows job passed CLI credential-permission and lifecycle tests, build, vet, and module checks.

6. **PASS — floor met; target short.** CI reports 84.6% combined non-generated coverage (2,397/2,834 statements), above the required 80% floor and below the 90% target. Package coverage is 83.0% for pkg/alexa, 86.5% for pkg/alexaapimodels, 94.2% for GraphQL, 83.8% for REST, and 100% over three statements for dependency models.

7. **OPEN — R2-F10 and R2-F11; public consumer pending.** The public SDK, generated dependency models, and transport packages use the required package boundaries and responsibility-specific schemas. The two model-construction gaps in item 4 prevent complete wire-model signoff. The CLI module requires SDK v0.4.0 but its go.mod still contains a local replace to the repository root. The public Go proxy and remote tags did not contain v0.4.0 or a cmd/go-alexa module tag; go list returned unknown revision for both requested public versions. A clean public-proxy SDK/CLI consumer build is therefore not yet proven.

8. **PASS.** Client construction uses validated functional options and sensible defaults. Account credentials are passed when creating a Session rather than stored in reusable client configuration.

9. **PASS.** Account tokens and event connections belong to explicit Session and connection objects. Callers can read and update tokens explicitly, observe close errors, and close streams; tests cover isolation and cleanup.

10. **PASS.** REST and GraphQL use injected HTTP clients; event traffic accepts injected HTTP/event transports. The native HTTP/2 replay substitutes a net.Pipe connection beneath the pinned transport and exercises paired framed traffic, keepalive, consumption, and close without live provider access.

11. **PASS.** Token refresh and cookie exchange are explicit operations that return credentials to the caller. Refresh does not silently replace session state; callers choose when to install and persist refreshed values.

12. **OPEN.** Documentation run 37175399237 built and checked the artifact, but this review run did not deploy to GitHub Pages. The exact candidate's live documentation publication and destinations are unverified.

13. **OPEN — current review record needs consolidation.** I reviewed the README, all tracked Markdown/MDX files, guide navigation, and the rendered artifact. Customer guides are concise, operation-focused, distinguish implementation-derived evidence, and remain separated from maintainer inventories and verification details. However, the contributor review record still contains all-16 sections for earlier SHAs and no Reviewer 1 section for c1; those stale review records need consolidation into the current exact-SHA record before release.

14. **OPEN.** This section is Reviewer 2's all-16 audit at c1d459e. Reviewer 1 has not reviewed this exact candidate. R2-F10 and R2-F11, the public SDK/CLI consumer and install proof, and Pages publication remain open; a two-reviewer final signoff is not available.

15. **PASS.** SyntheticReplay matches method, origin, escaped path, repeated query values, full headers, and body before releasing the paired response; it rejects mismatches, duplicates, unexpected calls, and unconsumed pairs. REST, GraphQL, HTTP/2, and CLI tests use offline paired exchanges or ordered frame transcripts. The authentication flow here is code-pair registration and explicit token operations; callback state and PKCE are caller-managed because this SDK does not implement that callback or verifier exchange.

16. **OPEN.** The standalone CLI is in cmd/go-alexa as a separate module and consumes the public SDK. It provides help, JSON results, nonzero failures, credential files/stdin/environment inputs, explicit export, linking and refresh, endpoint discovery, player read/control, and cancellable event listening with cleanup; offline paired command tests and Linux/Windows CI are present. The nested go.mod still has a local replace, the required SDK v0.4.0 is not published, no CLI module tag exists, and no public go install at a tag or separate public consumer install has been demonstrated.

### Findings and disposition at c1d459e

- **R2-F5 — RESOLVED at 572f2a1 (item 4).** Local declarations and prior assignments are indexed by lexical declaration and name; negative probes cover aliases reassigned to unregistered values.
- **R2-F6 — RESOLVED in this candidate (item 4).** Direct and aliased indexed writes, zero-value map aliases, collection receivers, parenthesized receivers, and intermediate index paths have source checks and negative tests.
- **R2-F7 — RESOLVED at cc35d238 (items 4, 7).** SDL bindings validate control-field enums, and production control values use schema-generated constants.
- **R2-F8 — RESOLVED at cc35d238 (item 4).** The video search phrase format has a behavior schema owner and a checked construction site.
- **R2-F9 — RESOLVED in this candidate (items 4, 7).** Recursive receiver-key inspection and tests reject unregistered intermediate keys through direct and aliased indexed paths.
- **R2-F10 — OPEN at c1d459e (items 4, 7).** Generated map provenance stored in a package-scope variable is not followed across functions; a later fixed-key mutation before serialization is accepted. Add a package-scope variable-object provenance gate and the exact negative probe above.
- **R2-F11 — OPEN at c1d459e (items 4, 7).** Fixed string values returned by local helpers are not traced when used in generated wire fields. Add return-value provenance checks and a negative probe for the helper-return construction above.
