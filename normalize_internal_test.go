package emailaddressnormalize

import (
	"testing"
)

func doLocalPartNormalize(t *testing.T, givenValue, expectValue string, doStopCheck bool) {
	var n normalizeLocalPartInstance
	var lastStopSig bool
	for _, ch := range ([]rune)(givenValue) {
		lastStopSig = n.putCharacter(ch)
	}
	if doStopCheck {
		if lastStopSig {
			t.Errorf("not expecting stop sign: \"%s\"", givenValue)
		}
		n.stopCheck()
	} else if !lastStopSig {
		t.Errorf("expecting stop sign: \"%s\"", givenValue)
	}
	if resultValue := n.resultLocalPart(); resultValue != expectValue {
		t.Errorf("unexpect result for \"%s\": [%s], expect: [%s]", givenValue, resultValue, expectValue)
	}
}

func TestNormalizeLocalPartInstance_WithoutStop(t *testing.T) {
	doLocalPartNormalize(t, "localpart", "localpart", true)
	doLocalPartNormalize(t, "LocalPart", "localpart", true)
	doLocalPartNormalize(t, "local.part", "local.part", true)
	doLocalPartNormalize(t, ".local", "\".local\"", true)
	doLocalPartNormalize(t, "local.", "\"local.\"", true)
	doLocalPartNormalize(t, "\"local.\"", "\"local.\"", true)
	doLocalPartNormalize(t, "local..part", "\"local..part\"", true)
	doLocalPartNormalize(t, "\"local..part\"", "\"local..part\"", true)
	doLocalPartNormalize(t, "local part", "\"local part\"", true)
	doLocalPartNormalize(t, "\"local part\"", "\"local part\"", true)
	doLocalPartNormalize(t, "\"local_part.addr\"", "local_part.addr", true)
	doLocalPartNormalize(t, "\"local_part\\\\.addr\"", "\"local_part\\\\.addr\"", true)
}

func TestNormalizeLocalPartInstance_WithStop(t *testing.T) {
	doLocalPartNormalize(t, "localpart@", "localpart", false)
	doLocalPartNormalize(t, "LocalPart@", "localpart", false)
	doLocalPartNormalize(t, "local.part@", "local.part", false)
	doLocalPartNormalize(t, ".local@", "\".local\"", false)
	doLocalPartNormalize(t, "local.@", "\"local.\"", false)
	doLocalPartNormalize(t, "\"local.\"@", "\"local.\"", false)
	doLocalPartNormalize(t, "local..part@", "\"local..part\"", false)
	doLocalPartNormalize(t, "\"local..part\"@", "\"local..part\"", false)
	doLocalPartNormalize(t, "local part@", "\"local part\"", false)
	doLocalPartNormalize(t, "\"local part\"@", "\"local part\"", false)
	doLocalPartNormalize(t, "\"local_part.addr\"@", "local_part.addr", false)
	doLocalPartNormalize(t, "\"local_part\\\\.addr\"@", "\"local_part\\\\.addr\"", false)
}
