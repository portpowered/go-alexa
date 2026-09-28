# Schema and fixture status

`graphql/endpoints-schema.graphql` is the checked-in schema subset consumed by genqlient for the query and mutation documents in that directory. It is not an authoritative or complete Alexa schema. The REST and directive parser shapes live in `api/openapi.yaml` and `api/asyncapi.yaml`; both are implementation-derived. The maintainer reports successful API tests with their own accounts, but did not document which operations were exercised or provide sanitized captures and provenance. Reproducible provider evidence is absent from this repository. Do not add provider operations based only on synthetic examples; verify the operation separately and label its evidence.

The JSON files under `pkg/alexa/testdata/` are synthetic test inputs. They contain no account captures and do not prove live provider behavior. Real recordings, if ever added, must be sanitized, dated, described in a neighboring provenance note, and stored separately from synthetic fixtures.
