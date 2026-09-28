# Alexa API coverage

This library implements a subset of provider-specific account and endpoint operations through the public `alexa.SessionInterface`. Stateless code-based linking and token exchange methods are available on `alexa.Client`:

- code-based linking and access-token refresh;
- endpoint enumeration and feature-state conversion;
- feature control;
- quality-of-service requests;
- event subscription and an HTTP/2 event connection;
- user information and player-state requests.

The documentation site includes three protocol views: the GraphQL subset in `pkg/schemas/graphql`, the implementation-derived HTTP routes in `api/openapi.yaml`, and the directive input accepted by the HTTP/2 stream parser in `api/asyncapi.yaml`. The GraphQL SDL and the HTTP schemas describe this client's implementation; none is an Amazon-issued, complete, or currently verified provider contract. The maintainer reports that these APIs worked with their own Alexa accounts, but those tests were not documented with an operation list, dates, results, or sanitized captures. Reproducible account-test evidence is absent from this repository. There are no sanitized live REST or event-stream captures here, and synthetic tests do not establish provider behavior.

The REST and event-stream schemas generate internal Go wire models used by the client. `make generate` refreshes those models along with the GraphQL client, and CI checks for generated-file drift. See [verification and fixture guidance](verification.md) for provenance and local checks.

The provider may change undocumented endpoints or payloads. The library reports transport, HTTP, authentication, and model errors where it can; operation responses may contain their own error fields. See the package reference and the operation guides for the exact method inputs and result types. Synthetic test fixtures are labeled as such and must not be cited as observations of live service behavior.
