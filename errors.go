package emailaddressnormalize

import (
	"errors"
)

// ErrGivenAddressTooShort indicate given email address is too short.
var ErrGivenAddressTooShort = errors.New("given email address is too short")
