package emailaddressnormalize

import (
	"errors"
)

// ErrGivenAddressTooShort indicate given email address is too short.
var ErrGivenAddressTooShort = errors.New("given email address is too short")

// ErrGivenAddressHasIPLiteral indicate given email address has IP literal as domain part.
var ErrGivenAddressHasIPLiteral = errors.New("given email address has IP literal as domain part")

// ErrGivenAddressNeedQuote indicate given email address needs quote.
var ErrGivenAddressNeedQuote = errors.New("given email address have to be quoted")

// ErrGivenAddressContainSpecialCharacter indicate given email address contain special characters may harmful to MTA.
var ErrGivenAddressContainSpecialCharacter = errors.New("given email address have special character")

// ErrGivenAddressLocalPartContainI18NCharacter indicate local part of given email address contain international character.
var ErrGivenAddressLocalPartContainI18NCharacter = errors.New("local part of given email address have i18n character")

// ErrEmptyDomainAfterCheck indicate domain part of given address become empty after check process.
var ErrEmptyDomainAfterCheck = errors.New("domain part become empty")

// ErrEmptyLocalPartAfterCheck indicate local part of given address become empty after check process.
var ErrEmptyLocalPartAfterCheck = errors.New("local part become empty after check")

// ErrEmptyLocalPartAfterNormalize indicate local part of given address become empty after normalize process.
var ErrEmptyLocalPartAfterNormalize = errors.New("domain part become empty after normalize")

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
