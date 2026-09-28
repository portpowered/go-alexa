# Releasing

The [release workflow](../.github/workflows/release.yml) runs when a `vMAJOR.MINOR.PATCH` tag is pushed. It validates the tag even when no prior release exists, checks the public API against the previous stable tag when one exists, runs race tests, vet, and build, verifies the tagged module from a fresh consumer using `proxy.golang.org`, then creates a GitHub Release with generated notes. Only the publishing job has `contents: write` permission.

Before pushing the first clean-history tag, complete the provider behavior review in [the migration checklist](template-checklist.md), item 4. In particular, the REST and HTTP/2 schemas still need verified provider evidence. Synthetic fixtures and implementation-derived schemas do not satisfy that release sign-off. The first release has no prior stable tag, so its API comparison reports that there is no baseline.

After that sign-off, choose a semantic version tag. Breaking API changes are allowed with a major version increase, and before v1 also with a minor version increase. Patch releases and stable minor releases must preserve compatibility. The workflow does not create or push tags.
