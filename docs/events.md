# Events

Create an event subscription before connecting. The subscription response may contain operation-level errors, so inspect both the returned error and the response fields.

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
    return fmt.Errorf("event subscription returned %d errors", len(subscription.Errors))
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
    fmt.Printf("%s/%s for endpoint %s\n", event.Namespace, event.Name, event.EndpointID)
}
```

`Receive` blocks until an event arrives or the connection fails. Close the connection when the consumer stops. The connection is separate from the client and the caller owns its lifecycle. Event payloads are decoded into typed values for known event types; unknown or malformed payloads return errors.

The repository's event tests use synthetic data. They do not claim that a particular event type or payload shape is currently emitted by Amazon.