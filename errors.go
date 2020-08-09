package emailaddressnormalize

import (
	"errors"
)

// ErrGivenAddressTooShort indicate given email address is too short.
var ErrGivenAddressTooShort = errors.New("given email address is too short")

// ErrGivenAddressHasIPLiteral indicate given email address has IP literal as domain part.
var ErrGivenAddressHasIPLiteral = errors.New("given email address has IP literal as domain part")


// ErrUnknownDomainCharacterCombination indicate unknown mix of characters in domain part.
type ErrUnknownDomainCharacterCombination struct {
	idnaDomain           bool
	dnHasDecimal         bool
	dnHasHex             bool
	dnHasDot             bool
	dnHasColon           bool
	dnHasOtherCharacters bool
}

func (e *ErrUnknownDomainCharacterCombination) Error() (result string) {
	result = "[ErrUnknownDomainCharacterCombination:"
	if e.idnaDomain {
		result += "I"
	}
	if e.dnHasDecimal {
		result += "3"
	}
	if e.dnHasHex {
		result += "X"
	}
	if e.dnHasDot {
		result += "."
	}
	if e.dnHasColon {
		result += ":"
	}
	if e.dnHasOtherCharacters {
		result += "C"
	}
	result += "]"
	return
}
