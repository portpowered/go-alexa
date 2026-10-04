package alexaapimodels

import "fmt"

// PlaybackDispatchError identifies the selected provider control branch.
// It retains the dependency error for typed HTTP and authentication checks.
//
//modelinventory:semantic Public selected control-path classification with typed cause.
type PlaybackDispatchError struct {
	DispatchPath string
	Operation    FeatureOperationName
	Err          error
}

func (e *PlaybackDispatchError) Error() string {
	return fmt.Sprintf("playback dispatch %s operation %s: %v", e.DispatchPath, e.Operation, e.Err)
}

func (e *PlaybackDispatchError) Unwrap() error { return e.Err }
