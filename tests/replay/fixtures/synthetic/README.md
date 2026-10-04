# Synthetic Alexa replay pairs

These files are hand-authored test scenarios frozen from explicit synthetic inputs. They are **not** captured Amazon traffic or proof of current provider behavior.

- `alexa-rest.json`: 18 dispatched REST routes, including both registration modes and all eight media command variants (26 ordered pairs).
- `alexa-graphql.json`: all seven generated GraphQL documents sent through the GraphQL HTTP route.
- `alexa-events.json`: ordered stream-open response with a directive frame, then the initial keepalive ping and response.
- `alexa-events-http2.json`: the same stream and ping through the pinned HTTP/2 codec over `net.Pipe`, including its default user-agent header. This tests connection injection and application framing without network access.

Each pair records method, origin, escaped path, the complete query multimap (including repeated values), full request headers and body, response status, headers, and body. The replay transport compares the request before returning its paired response and asserts complete consumption. All IDs, timestamps, and credential strings in these examples are fixed, visibly synthetic values; exact matching is their rule. No volatile field is skipped. If future fixtures require variable fields, add an explicit format or decoded-meaning matcher for that field, test a malformed value, and record the rule in this note before using the fixture. Never add real token or cookie values.

The OpenAPI `openAuthorizationPage` operation is a caller-facing URL for browser navigation. The library does not dispatch it, so it has no outbound HTTP pair. Existing `pkg/alexa/testdata` JSON files remain response-only model examples and do not count toward this replay inventory.
