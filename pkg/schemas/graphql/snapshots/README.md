# Live schema provenance

`live-schema.graphql` was rendered from an authenticated North America GraphQL
introspection response on 2026-10-10 at
`https://alexa.amazon.com/nexus/v1/graphql`, using the maintainer's saved CBL
login. The capture contains type definitions, fields, arguments, input defaults,
enum values, interfaces and possible types. It excludes endpoint data,
credentials, headers and account identifiers. Descriptions and deprecation
annotations are not retained in the SDL projection.

The snapshot is evidence for the schema visible to that authenticated account
at that time. It does not prove authorization or execution of device commands.
A separate read-only ListEndpointsWithStates query succeeded for 115 endpoints;
only aggregate verification results are documented, not private endpoint data.

The runtime compatibility subset remains `../endpoints-schema.graphql`.
`make live-schema-check` validates shipped operation documents against this
snapshot. Refresh instructions are in `../../NOTE.md`; the snapshot tool writes
schema metadata only. Synthetic request/response replay pairs remain explicitly
synthetic and have been updated for the new aliased volume selection.
