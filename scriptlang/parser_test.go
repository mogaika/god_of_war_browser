package scriptlang_test

import (
	"testing"

	"github.com/mogaika/god_of_war_browser/scriptlang"
)

func TestParser(t *testing.T) {
	const test = `
0A: $mesevo 123 123.124 // awdawdad
$mesevo //1231
0A: -123  -123.123
9E: "TEST_STRING1" "TEST STRING 2" 0 //
0A: "TEST_TEST\b\r\n\t\\\" "
0B: // (()) () (((( ))))

AB:
FF:
AA:
AA: //
`
	if _, err := scriptlang.ParseScript([]byte(test)); err != nil {
		t.Error(err)
	}
}
