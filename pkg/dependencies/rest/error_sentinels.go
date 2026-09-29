package rest

type staticError string

func (err staticError) Error() string { return string(err) }

const (
	errFailedToRetrieveCSRFToken   staticError = "failed to retrieve CSRF token" //nolint:gosec // Fixed validation text is not a credential.
	errUnauthorizedResponse        staticError = "unauthorized"
	errNotFoundResponse            staticError = "not found on API call"
	errBadRequestResponse          staticError = "bad request on API call"
	errServerErrorResponse         staticError = "internal server error on API call"
	errTrailingBehaviorJSON        staticError = "sequence contains trailing JSON"
	errMissingSequenceFields       staticError = "sequence type or start node is missing"
	errBehaviorDepthExceeded       staticError = "behavior node nesting exceeds 16"
	errIncompleteOperationNode     staticError = "opaque operation node is incomplete"
	errSerialNodeWithoutChildren   staticError = "serial node has no children"
	errParallelNodeWithoutChildren staticError = "parallel node has no children"
	errUnsupportedBehaviorNode     staticError = "unsupported behavior node type"
)
