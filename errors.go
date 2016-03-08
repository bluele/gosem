package gosem

import (
	"errors"
)

var (
	TimeoutError        = errors.New("Timeout error")
	TooManyReleaseError = errors.New("Too many release resources")
)
