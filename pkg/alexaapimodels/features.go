// Package alexaapimodels provides unified API models for Alexa services.
package alexaapimodels

// FeatureName represents the namespace of a control operation.
type FeatureName string

// String returns the string representation of the namespace.
func (n FeatureName) String() string {
	return string(n)
}

// EventNamespace returns the event namespace format for this control namespace
// Event namespaces use the format "alexa.endpoint.feature.{namespace}".
func (n FeatureName) EventNamespace() string {
	return FeatureEventNamespacePrefix + string(n)
}
