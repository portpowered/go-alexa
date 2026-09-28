---
title: 'Request endpoint state polling'
---

# Request endpoint state polling

`RequestEndpointQualityOfService` asks the service to poll endpoint state for a period and use the selected experience profile. The result may include per-operation errors.

```go
result, err := client.RequestEndpointQualityOfService(ctx, alexaapimodels.QualityOfServiceRequest{
    Endpoints: []string{endpoint.EndpointID},
    Configuration: alexaapimodels.QualityOfServiceConfiguration{
        TypeOfExperience: alexaapimodels.QualityOfServiceExperienceBackgroundEphemeral,
        DurationInSeconds: 60,
    },
})
if err != nil {
    return err
}
if len(result.Errors) != 0 {
    return fmt.Errorf("quality-of-service request returned %d errors", len(result.Errors))
}
```

A successful request does not guarantee that cloud state immediately matches the physical endpoint. Check returned errors and treat the resulting state as potentially delayed.