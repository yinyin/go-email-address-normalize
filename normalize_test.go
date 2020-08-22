package emailaddressnormalize_test

import (
	"testing"

	emailaddressnormalize "github.com/yinyin/go-email-address-normalize"
)

func doNormalizeEmailAddressTest(t *testing.T, opt *emailaddressnormalize.NormalizeOption, inputAddr, expectCheckedAddr, expectNormalizedAddr string, expectErr bool) (err error) {
	checkedAddr, normalizedAddr, err := emailaddressnormalize.NormalizeEmailAddress(inputAddr, opt)
	if nil != err {
		if !expectErr {
			t.Errorf("unexpect error (addr: [%s], opt: %#v): %v", inputAddr, opt, err)
		}
		return
	} else {
		if expectErr {
			t.Errorf("expecting error (addr: [%s], opt: %#v)", inputAddr, opt)
			return
		}
	}
	if checkedAddr != expectCheckedAddr {
		t.Errorf("unexpect checked address (addr: [%s], opt: %#v): [%s], expect [%s]", inputAddr, opt, checkedAddr, expectCheckedAddr)
	}
	if normalizedAddr != expectNormalizedAddr {
		t.Errorf("unexpect normalized address (addr: [%s], opt: %#v): [%s], expect [%s]", inputAddr, opt, normalizedAddr, expectNormalizedAddr)
	}
	return
}

func TestNormalizeEmailAddress_DefaultOpt(t *testing.T) {
	doNormalizeEmailAddressTest(t, nil, "User@Example.Net", "user@example.net", "user@example.net", false)
	doNormalizeEmailAddressTest(t, nil, "User+subAddr@Example.Net", "user+subaddr@example.net", "user@example.net", false)
	doNormalizeEmailAddressTest(t, nil, "U.se.r+subAddr@Example.Net", "u.se.r+subaddr@example.net", "user@example.net", false)
	doNormalizeEmailAddressTest(t, nil, "U.se.r_Name+subAddr@Example.Net", "u.se.r_name+subaddr@example.net", "user_name@example.net", false)
}

func TestNormalizeEmailAddress_AllowQuotedLocalPart(t *testing.T) {
	if err := doNormalizeEmailAddressTest(t, nil, "\"User One\"@Example.Net", "", "", true); err != emailaddressnormalize.ErrGivenAddressNeedQuote {
		t.Errorf("unexpect error content for \"User One\"@Example.Net: %v", err)
	}
	opt := &emailaddressnormalize.NormalizeOption{
		AllowQuotedLocalPart: true,
		RemoveSubAddressingWith: func(domainPart string) (subAddressChars []rune) {
			return ([]rune)("+%")
		},
		RemoveLocalPartDots: true,
	}
	doNormalizeEmailAddressTest(t, opt, "\"User One\"@Example.Net", "\"user one\"@example.net", "\"user one\"@example.net", false)
	doNormalizeEmailAddressTest(t, opt, "U.se..r_Name+subAddr@Example.Net", "\"u.se..r_name+subaddr\"@example.net", "user_name@example.net", false)
}

func TestNormalizeEmailAddress_IPLiteral(t *testing.T) {
	if err := doNormalizeEmailAddressTest(t, nil, "user@127.0.0.1", "", "", true); err != emailaddressnormalize.ErrGivenAddressHasIPLiteral {
		t.Errorf("unexpect error content for user@127.0.0.1: %v", err)
	}
	if err := doNormalizeEmailAddressTest(t, nil, "user@2001:db8::ff00:42:8329", "", "", true); err != emailaddressnormalize.ErrGivenAddressHasIPLiteral {
		t.Errorf("unexpect error content for user@2001:db8::ff00:42:8329: %v", err)
	}
	opt := &emailaddressnormalize.NormalizeOption{
		AllowIPLiteral: true,
	}
	doNormalizeEmailAddressTest(t, opt, "user@127.0.0.1", "user@[127.0.0.1]", "user@[127.0.0.1]", false)
	doNormalizeEmailAddressTest(t, opt, "user@[127.0.0.1]", "user@[127.0.0.1]", "user@[127.0.0.1]", false)
	doNormalizeEmailAddressTest(t, opt, "user@2001:db8::ff00:42:8329", "user@[2001:db8::ff00:42:8329]", "user@[2001:db8::ff00:42:8329]", false)
	doNormalizeEmailAddressTest(t, opt, "user@[2001:db8::ff00:42:8329]", "user@[2001:db8::ff00:42:8329]", "user@[2001:db8::ff00:42:8329]", false)
}
