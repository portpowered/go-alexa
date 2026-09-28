# GraphQL schema and fixture status

`endpoints-schema.graphql` is the checked-in schema subset consumed by genqlient for the query and mutation documents in this directory. It is not an authoritative or complete Alexa schema. Do not add provider operations based only on synthetic examples; verify the operation separately and label its evidence.

The JSON files under `pkg/alexa/testdata/` are synthetic test inputs. They contain no account captures and do not prove live provider behavior. Real recordings, if ever added, must be sanitized, dated, described in a neighboring provenance note, and stored separately from synthetic fixtures.