---
title: 'Subscribe to endpoint events'
---

# Subscribe to endpoint events

Call `Subscribe` before connecting. Check the response's `Errors` field, then open the event connection and close it when the consumer stops.

```go
subscription, err := client.Subscribe(ctx, alexaapimodels.SubscribeRequest{
    Entities: []alexaapimodels.SubscribeEntity{{
        EntityType: alexaapimodels.EntityTypeEndpoint,
    }},
    DurationInMinutes: 4,
})
if err != nil {
    return err
}
if len(subscription.Errors) != 0 {
    return fmt.Errorf("subscription returned %d errors", len(subscription.Errors))
}

connection, err := client.ConnectEvents(ctx)
if err != nil {
    return err
}
defer connection.Close()

for {
    event, err := connection.Receive()
    if err != nil {
        return err
    }
    fmt.Printf("%s/%s: %+v\n", event.Namespace, event.Name, event.Payload)
}
```

`Receive` blocks. Close the connection to stop reading; the caller owns its lifecycle. Known event payloads are decoded into typed models. The synthetic event tests do not establish which events Amazon currently emits.