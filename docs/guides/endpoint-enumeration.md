---
title: 'Enumerate endpoints'
---

# Enumerate endpoints

`ListEndpoints` combines GraphQL endpoint data with optional REST device metadata. Set `Properties` to request state values and `Features` to request supported feature metadata.

```go
response, err := client.ListEndpoints(ctx, alexaapimodels.EndpointQuery{
    IncludeFields: &alexaapimodels.EndpointIncludeFields{
        Properties: true,
        Features:   true,
    },
})
if err != nil {
    return err
}
for _, endpoint := range response.Results {
    name := endpoint.EndpointID
    if endpoint.FriendlyName != nil {
        name = endpoint.FriendlyName.Value
    }
    fmt.Printf("%s: %s (%d features)\n", name, endpoint.EndpointID, len(endpoint.Features))
}
```

The merged result can omit REST-only fields when the secondary REST request fails. Feature names describe the operations exposed in the response; they do not guarantee that a command will succeed on the physical endpoint.

The schema reference covers the checked-in GraphQL subset. Synthetic fixtures exercise conversions and are not live-service evidence.
