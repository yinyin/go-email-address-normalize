package emailaddressnormalize

var defaultNormalizeOption = &NormalizeOption{}

// NormalizeOption contain parameters for normalize function.
type NormalizeOption struct {
	AllowQuotedLocalPart             bool
	AllowLocalPartSpecialChars       bool
	AllowLocalPartInternationalChars bool
	AllowIPLiteral                   bool
}
