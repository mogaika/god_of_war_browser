package mesh

import (
	"fmt"
	"io"
)

type Logger struct {
	io.Writer
}

func (l *Logger) Println(a ...interface{}) {
	if l != nil {
		fmt.Fprintln(l, a...)
	}
}
func (l *Logger) Printf(format string, a ...interface{}) {
	if l != nil {
		fmt.Fprintf(l, format+"\n", a...)
	}
}
