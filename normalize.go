package emailaddressnormalize

import (
	"unicode"
)

const domainLengthLimit = 255

var charactersNeedQuote = ([]rune)("\"(),:;<>@[\\]")

func isNeedQuote(ch rune) bool {
	for _, c := range charactersNeedQuote {
		if c == ch {
			return true
		}
	}
	return false
}

type normalizeStateCallable func(ch rune) (nextState normalizeStateCallable)

type normalizeInstance struct {
	emailAddress []rune

	localPart  []rune
	domainPart []rune

	lastCommitedCharacter rune

	needQuote            bool
	subaddressOffsets    [16]int
	idnaDomain           bool
	dnHasDecimal         bool
	dnHasHex             bool
	dnHasDot             bool
	dnHasColon           bool
	dnHasOtherCharacters bool
}

func newNormalizeInstance(emailAddress string) (instance *normalizeInstance) {
	aux := ([]rune)(emailAddress)
	l := len(aux)
	instance = &normalizeInstance{
		emailAddress: aux,
		localPart:    make([]rune, 0, l-1),
		domainPart:   make([]rune, 0, l-1),
	}
	return
}

// runNormalize perform normalize on given emailAddress.
func (n *normalizeInstance) runNormalize() (err error) {
	if len(n.emailAddress) < 3 {
		err = ErrGivenAddressTooShort
		return
	}
	stateCallable := n.stateStart
	for _, ch := range n.emailAddress {
		if nextStateCallable := stateCallable(ch); nil != nextStateCallable {
			stateCallable = nextStateCallable
		}
	}
	return
}

// commitToDomainPart append guven character `ch` into normalized domain part.
func (n *normalizeInstance) commitToDomainPart(ch rune) {
	if len(n.domainPart) >= domainLengthLimit {
		return
	}
	if unicode.IsSpace(ch) || (!unicode.IsPrint(ch)) {
		return // skip spaces and non-printables.
	}
	if ch > unicode.MaxASCII {
		if (ch == 0x3002) || (ch == 0xFF0E) || (ch == 0xFF61) {
			ch = '.'
		} else {
			n.idnaDomain = true
		}
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			ch = unicode.ToLower(ch)
		}
	} else if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
		ch = unicode.ToLower(ch)
	} else if (ch == '-') || (ch == ':') {
	} else if ch == '.' {
		if (n.lastCommitedCharacter == '.') || (0 == len(n.domainPart)) {
			return
		}
	} else {
		return // skip non-(Letter, Digit, Hyphen) ASCII characters.
	}
	switch {
	case (ch >= '0') && (ch <= '9'):
		n.dnHasDecimal = true
	case (ch >= 'a') && (ch <= 'f'):
		n.dnHasHex = true
	case ch == '.':
		n.dnHasDot = true
	case ch == ':':
		n.dnHasColon = true
	default:
		n.dnHasOtherCharacters = true
	}
	n.domainPart = append(n.domainPart, ch)
	n.lastCommitedCharacter = ch
}

// commitToLocalPart append given character `ch` into normalized local part.
func (n *normalizeInstance) commitToLocalPart(ch rune) {
	if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
		ch = unicode.ToLower(ch)
	} else if !unicode.IsPrint(ch) {
		return // skip non-printables.
	} else if unicode.IsSpace(ch) || isNeedQuote(ch) {
		n.needQuote = true
	} else if (ch == '.') && (n.lastCommitedCharacter == '.') {
		n.needQuote = true
	}
	if (ch | 0xF) == 0x2F {
		offsetIdx := ch & 0xF
		if 0 == n.subaddressOffsets[offsetIdx] {
			n.subaddressOffsets[offsetIdx] = len(n.localPart)
		}
	}
	n.localPart = append(n.localPart, ch)
	n.lastCommitedCharacter = ch
}

func (n *normalizeInstance) stateIPLiteralDomainPart(ch rune) (nextState normalizeStateCallable) {
	if ch == ']' {
		return n.stateSimpleDomainPart
	}
	n.commitToDomainPart(ch)
	return nil
}

func (n *normalizeInstance) stateSimpleDomainPart(ch rune) (nextState normalizeStateCallable) {
	if ch == '[' {
		return n.stateIPLiteralDomainPart
	}
	n.commitToDomainPart(ch)
	return nil
}

func (n *normalizeInstance) stateQuotedLocalPartInEscape(ch rune) (nextState normalizeStateCallable) {
	n.commitToLocalPart(ch)
	return n.stateQuotedLocalPart
}

func (n *normalizeInstance) stateQuotedLocalPart(ch rune) (nextState normalizeStateCallable) {
	switch ch {
	case '"':
		return n.stateSimpleLocalPart
	case '\\':
		return n.stateQuotedLocalPartInEscape
	default:
		n.commitToLocalPart(ch)
	}
	return nil
}

func (n *normalizeInstance) stateLocalPartCommentQuotedInEscape(ch rune) (nextState normalizeStateCallable) {
	return n.stateLocalPartCommentQuotedText
}

func (n *normalizeInstance) stateLocalPartCommentQuotedText(ch rune) (nextState normalizeStateCallable) {
	switch ch {
	case '"':
		return n.stateLocalPartComment
	case '\\':
		return n.stateLocalPartCommentQuotedInEscape
	}
	return nil
}

func (n *normalizeInstance) stateLocalPartComment(ch rune) (nextState normalizeStateCallable) {
	switch ch {
	case '"':
		return n.stateLocalPartCommentQuotedText
	case ')':
		return n.stateSimpleLocalPart
	}
	return nil
}

func (n *normalizeInstance) stateSimpleLocalPart(ch rune) (nextState normalizeStateCallable) {
	switch ch {
	case '@':
		if n.lastCommitedCharacter == '.' {
			n.needQuote = true
		}
		return n.stateSimpleDomainPart
	case '(':
		return n.stateLocalPartComment
	default:
		n.commitToLocalPart(ch)
	}
	return nil
}

func (n *normalizeInstance) stateStart(ch rune) (nextState normalizeStateCallable) {
	switch ch {
	case '"':
		return n.stateQuotedLocalPart
	case '(':
		return n.stateLocalPartComment
	case '.':
		n.needQuote = true
		n.commitToLocalPart(ch)
		return n.stateSimpleLocalPart
	case '@':
		n.needQuote = true
		return n.stateSimpleDomainPart
	default:
		n.commitToLocalPart(ch)
		return n.stateSimpleLocalPart
	}
}

// NormalizeEmailAddress normalize given email adderss and return checked and normalized
// email addresses.
func NormalizeEmailAddress(emailAddress string, opt *NormalizeOption) (checkedEmailAddress, normalizedEmailAddress string, err error) {
	if opt == nil {
		opt = defaultNormalizeOption
	}
	normalizeInst := newNormalizeInstance(emailAddress)
	if err = normalizeInst.runNormalize(); nil != err {
		return
	}
	// TODO: check normalized result.
	return
}
