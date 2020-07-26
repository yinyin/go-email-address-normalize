package emailaddressnormalize

var defaultNormalizeOption = &NormalizeOption{}

// NormalizeOption contain parameters for normalize function.
type NormalizeOption struct {
	AllowQuotedLocalPart bool
	AllowIPLiteral       bool
}
