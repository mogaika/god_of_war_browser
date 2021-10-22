package utils

import (
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

func SDump(a ...interface{}) string {
	return spewConfig.Sdump(a...)
}

func LogDump(a ...interface{}) {
	log.Println(spewConfig.Sdump(a...))
}
