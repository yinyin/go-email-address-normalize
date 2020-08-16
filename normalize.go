package emailaddressnormalize

import (
	"unicode"
)

const domainLengthLimit = 255

var charactersNeedQuote = ([]rune)("\"(),:;<>@[\\]")
var charactersNotVerySafe = ([]rune)("%|!#$*/\\")

func isNeedQuote(ch rune) bool {
	for _, c := range charactersNeedQuote {
		if c == ch {
			return true
		}
	}
	return false
}

func isNotVerySafeCharacter(ch rune) bool {
	for _, c := range charactersNotVerySafe {
		if c == ch {
			return true
		}
	}
	return false
}

func runesIndexRune(s []rune, ch rune) int {
	for idx, elem := range s {
		if elem == ch {
			return idx
		}
	}
	return -1
}

type normalizeStateCallable func(ch rune) (nextState normalizeStateCallable)

type normalizeLocalPartInstance struct {
	localPart             []rune
	lastCommitedCharacter rune

	needQuote bool

	stateCallable normalizeStateCallable
	shouldStop    bool

	hasUnsafeCharacter   bool
	hasNonASCIICharacter bool
}

// commitToLocalPart append given character `ch` into normalized local part.
func (n *normalizeLocalPartInstance) commitToLocalPart(ch rune) {
	if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
		ch = unicode.ToLower(ch)
	} else if !unicode.IsPrint(ch) {
		return // skip non-printables.
	} else if unicode.IsSpace(ch) || isNeedQuote(ch) {
		n.needQuote = true
	} else if (ch == '.') && (n.lastCommitedCharacter == '.') {
		n.needQuote = true
	} else if isNotVerySafeCharacter(ch) {
		n.hasUnsafeCharacter = true
	}
	if ch > unicode.MaxASCII {
		n.hasNonASCIICharacter = true
	}
	n.localPart = append(n.localPart, ch)
	n.lastCommitedCharacter = ch
}

func (n *normalizeLocalPartInstance) stateQuotedLocalPartInEscape(ch rune) (nextState normalizeStateCallable) {
	n.commitToLocalPart(ch)
	return n.stateQuotedLocalPart
}

func (n *normalizeLocalPartInstance) stateQuotedLocalPart(ch rune) (nextState normalizeStateCallable) {
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

func (n *normalizeLocalPartInstance) stateLocalPartCommentQuotedInEscape(ch rune) (nextState normalizeStateCallable) {
	return n.stateLocalPartCommentQuotedText
}

func (n *normalizeLocalPartInstance) stateLocalPartCommentQuotedText(ch rune) (nextState normalizeStateCallable) {
	switch ch {
	case '"':
		return n.stateLocalPartComment
	case '\\':
		return n.stateLocalPartCommentQuotedInEscape
	}
	return nil
}

func (n *normalizeLocalPartInstance) stateLocalPartComment(ch rune) (nextState normalizeStateCallable) {
	switch ch {
	case '"':
		return n.stateLocalPartCommentQuotedText
	case ')':
		return n.stateSimpleLocalPart
	}
	return nil
}

func (n *normalizeLocalPartInstance) stateSimpleLocalPart(ch rune) (nextState normalizeStateCallable) {
	switch ch {
	case '@':
		n.stopCheck()
		n.shouldStop = true
		return n.stateStart
	case '(':
		return n.stateLocalPartComment
	default:
		n.commitToLocalPart(ch)
	}
	return nil
}

func (n *normalizeLocalPartInstance) stateStart(ch rune) (nextState normalizeStateCallable) {
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
		n.shouldStop = true
		return n.stateStart
	default:
		n.commitToLocalPart(ch)
		return n.stateSimpleLocalPart
	}
}

func (n *normalizeLocalPartInstance) stopCheck() {
	if n.lastCommitedCharacter == '.' {
		n.needQuote = true
	}
}

func (n *normalizeLocalPartInstance) putCharacter(ch rune) (shouldStop bool) {
	if n.stateCallable == nil {
		n.stateCallable = n.stateStart
	}
	if nextStateCallable := n.stateCallable(ch); nextStateCallable != nil {
		n.stateCallable = nextStateCallable
	}
	return n.shouldStop
}

func (n *normalizeLocalPartInstance) resultLocalPart() string {
	if !n.needQuote {
		return string(n.localPart)
	}
	buf := make([]rune, 0, len(n.localPart)+2)
	buf = append(buf, '"')
	for _, ch := range n.localPart {
		if ch == '\\' || ch == '"' {
			buf = append(buf, '\\')
		}
		buf = append(buf, ch)
	}
	buf = append(buf, '"')
	return string(buf)
}

type normalizeInstance struct {
	emailAddress []rune

	localPartNormalizer normalizeLocalPartInstance
	domainPart          []rune

	lastCommitedCharacter rune

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
		localPartNormalizer: normalizeLocalPartInstance{
			localPart: make([]rune, 0, l-1),
		},
		domainPart: make([]rune, 0, l-1),
	}
	return
}

// runNormalize perform normalize on given emailAddress.
func (n *normalizeInstance) runNormalize() (err error) {
	if len(n.emailAddress) < 3 {
		err = ErrGivenAddressTooShort
		return
	}
	stateCallable := n.stateLocalPart
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

func (n *normalizeInstance) stateLocalPart(ch rune) (nextState normalizeStateCallable) {
	if (ch | 0xF) == 0x2F {
		offsetIdx := ch & 0xF
		if 0 == n.subaddressOffsets[offsetIdx] {
			n.subaddressOffsets[offsetIdx] = len(n.localPartNormalizer.localPart)
		}
	}
	if shouldStop := n.localPartNormalizer.putCharacter(ch); shouldStop {
		return n.stateSimpleDomainPart
	}
	return nil
}

func (n *normalizeInstance) isIPLiteralDomain() (bool, error) {
	if (n.idnaDomain || n.dnHasOtherCharacters) && (!n.dnHasColon) {
		return false, nil
	}
	if n.dnHasDecimal && n.dnHasDot && (!n.dnHasHex) && (!n.dnHasColon) {
		return true, nil
	}
	if (n.dnHasDecimal || n.dnHasHex) && n.dnHasColon && (!n.dnHasDot) {
		return true, nil
	}
	err := &ErrUnknownDomainCharacterCombination{
		idnaDomain:           n.idnaDomain,
		dnHasDecimal:         n.dnHasDecimal,
		dnHasHex:             n.dnHasHex,
		dnHasDot:             n.dnHasDot,
		dnHasColon:           n.dnHasColon,
		dnHasOtherCharacters: n.dnHasOtherCharacters,
	}
	return false, err
}

func (n *normalizeInstance) check(opt *NormalizeOption) (err error) {
	if !opt.AllowIPLiteral {
		var isIPLiteral bool
		if isIPLiteral, err = n.isIPLiteralDomain(); nil != err {
			return
		} else if isIPLiteral {
			err = ErrGivenAddressHasIPLiteral
			return
		}
	}
	if (!opt.AllowQuotedLocalPart) && n.localPartNormalizer.needQuote {
		err = ErrGivenAddressNeedQuote
		return
	}
	if (!opt.AllowLocalPartSpecialChars) && n.localPartNormalizer.hasUnsafeCharacter {
		err = ErrGivenAddressContainSpecialCharacter
		return
	}
	if (!opt.AllowLocalPartInternationalChars) && n.localPartNormalizer.hasNonASCIICharacter {
		err = ErrGivenAddressLocalPartContainI18NCharacter
		return
	}
	return
}

func (n *normalizeInstance) normalizeLocalPart(opt *NormalizeOption) (resultLocalPart string) {
	buf := n.localPartNormalizer.localPart
	if opt.RemoveSubAddressingWith != nil {
		subAddrChars := opt.RemoveSubAddressingWith(string(n.domainPart))
		for _, ch := range subAddrChars {
			if (ch | 0xF) == 0x2F {
				offsetIdx := ch & 0xF
				if ofst := n.subaddressOffsets[offsetIdx]; (ofst > 0) && (ofst < len(buf)) {
					buf = buf[:ofst]
				}
			} else if ofst := runesIndexRune(buf, ch); ofst >= 0 {
				buf = buf[:ofst]
			}

		}
	}
	if len(buf) == 0 {
		return
	}
	if opt.RemoveLocalPartDots {
		n2 := normalizeLocalPartInstance{
			localPart: make([]rune, 0, len(buf)),
		}
		for _, ch := range buf {
			if ch == '.' {
				continue
			}
			n2.putCharacter(ch)
		}
		n2.stopCheck()
		resultLocalPart = n2.resultLocalPart()
	} else {
		resultLocalPart = string(buf)
	}
	return
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
	if err = normalizeInst.check(opt); nil != err {
		return
	}
	checkedLocalPart := normalizeInst.localPartNormalizer.resultLocalPart()
	normalizedLocalPart := normalizeInst.normalizeLocalPart(opt)
	domainPart := string(normalizeInst.domainPart)
	checkedEmailAddress = checkedLocalPart + "@" + domainPart
	normalizedEmailAddress = normalizedLocalPart + "@" + domainPart
	return
}
