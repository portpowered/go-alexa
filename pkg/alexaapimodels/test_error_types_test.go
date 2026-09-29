package alexaapimodels_test

type syntheticFailureError string

func (err syntheticFailureError) Error() string { return string(err) }
