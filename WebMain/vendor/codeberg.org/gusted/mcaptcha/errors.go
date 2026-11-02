package mcaptcha

import "errors"

var (
	// ErrMissingSecret errors when no secret was provided by the code.
	ErrMissingSecret = errors.New("no secret was provided")

	// ErrMissingSitekey errors when no sitekey was provided by the code.
	ErrMissingSitekey = errors.New("no sitekey was provided")

	// ErrMissingToken errors when no token was provided by the code.
	ErrMissingToken = errors.New("no token was provided")
)
