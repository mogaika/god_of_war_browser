package utils

import (
	"bytes"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
)

var spewConfig *spew.ConfigState

func init() {
	spewConfig = spew.NewDefaultConfig()
	spewConfig.DisableCapacities = true
}

func Dump(a ...interface{}) {
	fmt.Println(spewConfig.Sdump(a...))
}

func DumpToOneLineString(buf []byte) string {
	var out bytes.Buffer

	for _, b := range buf {
		if b >= 0x20 && b <= 0x7f {
			out.WriteRune(rune(b))
		} else {
			out.WriteString(fmt.Sprintf("\\x%.2x", b))
		}
	}

	return out.String()
}

func SDump(a ...interface{}) string {
	return spewConfig.Sdump(a...)
}

func LogDump(a ...interface{}) {
	log.Println(spewConfig.Sdump(a...))
}
