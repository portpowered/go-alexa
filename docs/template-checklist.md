# Template migration checklist: go-alexa

Source checklist: `go-third-party-template/docs/library-standards.md`, `docs/verification.md`, and `docs/releasing.md`. Sign-offs below describe this repository's evidence and remaining release work.

- [x] **1.** Keep the public client, examples, README, and site independent of any consuming application. Port OS plugin mappings and backend rollout notes were removed from this repository; examples use only this module's public packages.
- [x] **2.** Document supported operations, authentication, errors, and HTTP client injection with the current exported API. Added guides for authentication, endpoint enumeration, control, events, and quality-of-service. Label the fixture suite as synthetic and avoid presenting it as live-service evidence.
- [ ] **3.** Show Go version, CI, coverage, release, Go Reference, license, and documentation badges in the README, with repository-specific links. Badge links are configured, but the coverage Pages site and first GitHub Release are not yet verified for the externalized repository.
- [ ] **4.** Generate the API reference in CI with the shared Fumadocs action and publish it to GitHub Pages. The workflow uses the checked-in GraphQL SDL and authored guides; genqlient successfully regenerates the checked-in client, and CI checks for stale output. The shared site builder produced the GraphQL reference and all five guide pages locally. Keep this item open until the Pages workflow has deployed successfully. The SDL is only the subset used by this client; REST and HTTP/2 operations are not represented by it.
- [x] **5.** Run offline build, lint, race, and synthetic fixture checks. Removed the old captured player-state replay test and its capture files; replaced that coverage with a synthetic response transport test. No real captures remain in the tree. Synthetic fixtures do not establish live Alexa behavior.

## Release and history sign-off

- [x] Inspect the exported API and compile a separate temporary consumer module against the local module path.
- [x] Run `make lint` and `make check`; both passed with `GOWORK=off` and `GOTOOLCHAIN=local`.
- [x] Review the working tree: token-marker scan found no token-shaped values, and no capture/config data files remain. This checks the current tree only; old remote history still requires the parent task rewrite.
- [x] Replace `main` history and remove old tags that retain the old commits. **Sign-off target:** publish this reviewed tree as a new root commit, delete v1.0.0 through v1.0.7, and verify remote refs. The publication audit must confirm this entry.

Unfinished items remain unchecked until verified. This checklist does not claim that the documentation site or a release has been published.
