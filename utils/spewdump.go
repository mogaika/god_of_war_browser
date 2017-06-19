package utils

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

func Dump(a ...interface{}) {
	// workaround for sped buffered output
	fmt.Println(spew.Sdump(a...))
}
