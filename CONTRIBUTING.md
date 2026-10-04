# Contributing

Follow the shared [library standards](https://github.com/portpowered/go-third-party-template/blob/main/docs/library-standards.md)
and the repository [checklist](docs/template-checklist.md). Keep the current
[independent review record](docs/independent-review.md) in contributor material.

Before submitting a change, run:

```sh
make generate
make lint
make check
```

Change schemas or generator configuration rather than generated Go files.
Edit the responsibility-specific OpenAPI sources under `api/openapi/sources/`;
`api/openapi.yaml` is bundled from those sources by `make generate-api`.
Standalone behavior, feature-event, and compatibility specs remain in their
own files and are validated by `make check`.
Group provider wire definitions by API responsibility in `pkg/dependencymodels`;
keep conversion and transport behavior separate. The model inventory must cover
exported compatibility types and anonymous nested objects as well as active
wire-boundary types. Generated markers and route-gate results alone do not
prove that every model is generated.

Use deterministic paired request/response fixtures for transport behavior.
Keep synthetic examples separate from captured evidence and never commit
credentials or private captures. See [verification](docs/verification.md) for
fixture rules and opt-in integration tests. Run the API compatibility check
before moving exported types; generated compatibility projections must preserve
existing caller shapes when a patch release is intended.

Keep README content useful to callers. Publish operation guides as MDX under
`docs/guides/`; put maintenance details here or in the contributor notes.
Review every tracked document for audience, duplication, and incoming links.
Use the [release instructions](docs/releasing.md) after independent sign-off.
