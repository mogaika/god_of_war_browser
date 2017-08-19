package utils

import (
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
)

func Dump(a ...interface{}) {
	// workaround for sped buffered output
	fmt.Println(spew.Sdump(a...))
}

func LogDump(a ...interface{}) {
	log.Println(spew.Sdump(a...))
}
