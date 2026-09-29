package testing

type staticError string

func (err staticError) Error() string { return string(err) }

const (
	errReplayRequestsExhausted    staticError = "unexpected replay request: all"
	errReplayRequestMismatch      staticError = "replay request"
	errReplayPairsUnconsumed      staticError = "replay consumed"
	errSyntheticReplayInvalid     staticError = "synthetic replay"
	errSyntheticRequestsExhausted staticError = "unexpected synthetic replay request after"
	errSyntheticRequestMismatch   staticError = "synthetic replay request"
	errSyntheticPairsUnconsumed   staticError = "synthetic replay consumed"
)
