package emailaddressnormalize

// SubAddressingCharactersFunc represent callable return sub-addressing characters of given domain part.
type SubAddressingCharactersFunc func(domainPart string) (subAddressChars []rune)

// NormalizeOption contain parameters for normalize function.
type NormalizeOption struct {
	AllowQuotedLocalPart             bool
	AllowLocalPartSpecialChars       bool
	AllowLocalPartInternationalChars bool
	AllowIPLiteral                   bool

	RemoveSubAddressingWith SubAddressingCharactersFunc
	RemoveLocalPartDots     bool
}

var defaultSubAddressChars = ([]rune)("+%")

func defaultSubAddressingCharactersFunc(domainPart string) (subAddressChars []rune) {
	return defaultSubAddressChars
}

var defaultNormalizeOption = &NormalizeOption{
	RemoveSubAddressingWith: defaultSubAddressingCharactersFunc,
	RemoveLocalPartDots:     true,
}
