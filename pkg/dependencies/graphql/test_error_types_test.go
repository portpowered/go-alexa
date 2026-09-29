//nolint:testpackage // Shares a private synthetic error type with white-box transport tests.
package graphql

type syntheticFailureError string

func (err syntheticFailureError) Error() string { return string(err) }
