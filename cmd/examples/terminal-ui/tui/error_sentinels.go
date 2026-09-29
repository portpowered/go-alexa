package tui

type staticError string

func (err staticError) Error() string { return string(err) }

const (
	errClientNotInitialized staticError = "client not initialized"
	errEmptyInputField      staticError = "field cannot be empty"
)
