# Alexa API coverage

This library implements a subset of provider-specific account and endpoint operations through the public `alexa.ClientInterface`:

- code-based linking and access-token refresh;
- endpoint enumeration and feature-state conversion;
- feature control;
- quality-of-service requests;
- event subscription and an HTTP/2 event connection;
- user information and player-state requests.

The GraphQL reference on the documentation site is generated from the checked-in SDL and query documents in `pkg/schemas/graphql`. That schema describes the selections used by this client, not all Alexa operations or a stable public provider contract. The REST operations and HTTP/2 event protocol are not described by that GraphQL schema.

The provider may change undocumented endpoints or payloads. The library reports transport, HTTP, authentication, and model errors where it can; operation responses may contain their own error fields. See the package reference and the operation guides for the exact method inputs and result types. Synthetic test fixtures are labeled as such and must not be cited as observations of live service behavior.