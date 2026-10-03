# Releasing

Work through [the checklist](template-checklist.md) and obtain two independent
reviews of every item at the exact implementation commit. Run `make generate`,
`make lint`, and `make check`; verify generation drift, the complete wire-model
and endpoint inventories, non-generated coverage, and documentation links.

The [release workflow](../.github/workflows/release.yml) runs on version tags.
It repeats checks, compares the public API with the previous stable tag,
compiles a separate consumer fetched through the public Go proxy, and publishes
a GitHub Release. Verify the exact tag's workflow and published module before
recording the release as complete.

The standalone CLI is a nested Go module with its own tags in the
`cmd/go-alexa/vX.Y.Z` form. Publish the SDK release it requires first. Before
tagging the CLI module, remove development-only `replace` directives and verify
the nested module through the public Go proxy; the CLI release workflow rejects
replacements and checks `go install` for the tag.

Use a semantic version that permits the reported API changes. Patch releases
must preserve compatibility. Before v1, an intentional breaking change needs
a minor version increase. Review `pkg/alexa`, `pkg/alexaapimodels`, and
`pkg/dependencymodels` when model definitions move.

Release notes must not present synthetic fixtures as observed provider
exchanges. Link to [verification guidance](verification.md#fixture-provenance)
for the evidence available with the release.
