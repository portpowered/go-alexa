---
title: 'Control an endpoint'
---

# Control an endpoint

Choose an endpoint that advertises the feature you plan to control, then pass its target, feature namespace, operation name, and a matching payload to `Control`.

```go
var target *alexaapimodels.Endpoint
for _, endpoint := range response.Results {
    if endpoint.HasFeature(alexaapimodels.FeatureNamePower) {
        target = endpoint
        break
    }
}
if target == nil {
    return errors.New("no power-capable endpoint found")
}

result, err := client.Control(ctx, alexaapimodels.ControlRequest{
    Target:    target,
    Namespace: alexaapimodels.FeatureNamePower,
    Name:      alexaapimodels.FeatureOperationNameTurnOff,
    Payload:   alexaapimodels.ControlPowerPayload{},
})
if err != nil {
    return err
}
if len(result.Errors) != 0 {
    return fmt.Errorf("control returned %d operation errors", len(result.Errors))
}
```

A successful method call or empty error list does not confirm the device changed state. Query endpoint state or consume state events when your application needs follow-up information. Control operations and payloads are device-dependent.