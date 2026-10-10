# Schema and fixture status

`pkg/schemas/graphql/endpoints-schema.graphql` is the checked-in schema subset consumed by genqlient for the query and mutation documents in that directory. It is not an authoritative or complete Alexa schema. REST wire shapes are grouped in `api/openapi.yaml`; directive frames are modeled by `api/asyncapi.yaml`; behavior sequence and feature payload JSON are split into `api/behaviors.yaml` and `api/feature-events.yaml`. Generated REST, GraphQL, directive, and embedded payload definitions live in `pkg/dependencymodels`, grouped by API responsibility. `tools/graphqlmodels` moves genqlient model declarations and their custom JSON methods from the generated operation file into that package while leaving compatibility aliases for existing imports. The runtime schemas are implementation-derived. Captured CBL authentication and a dated live GraphQL schema snapshot have provider provenance; other device-control paths remain without sanitized live captures. Do not add provider operations based only on synthetic examples; verify the operation separately and label its evidence.

See [fixture guidance](../../docs/verification.md#fixture-provenance) for synthetic examples and capture requirements.

`graphql/snapshots/live-schema.graphql` is a complete North America introspection snapshot
returned by Alexa on 2026-10-10 using a saved CBL login. It contains schema
metadata only. It is kept separately from the compatibility subset used for
generation, and `make live-schema-check` validates all seven shipped operation
documents against it. Introspection establishes accepted fields and types; it
does not establish account entitlement or successful device commands. A live
read-only state query also succeeded for 115 endpoints, including three present
and two unavailable speaker volume values. No endpoint identities or credentials
are included in the snapshot.

Refresh the saved login with `alexa auth refresh`, then run:

```powershell
python tools/graphql-schema-snapshot.py --credentials "$env:APPDATA/go-alexa/credentials.json" --output pkg/schemas/graphql/snapshots/live-schema.graphql
make live-schema-check
```

The volume selection uses `volumeValue: value { value }` to avoid the overlapping
fragment field conflict with floating-point sensor values. Both the outer value
and inner integer preserve absence; present zero remains a measurement.
