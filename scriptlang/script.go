package scriptlang

import (
	"fmt"
)

type Opcode struct {
	Code       byte
	Parameters []interface{}
	Comment    string
}

func (op *Opcode) String() string {
	s := fmt.Sprintf("%.2X:", op.Code)
	for _, p := range op.Parameters {
		switch p.(type) {
		case string:
			s += fmt.Sprintf(" %q", p)
		default:
			s += fmt.Sprint(" ", p)
		}
	}
	return s
}

type Label struct {
	Name    string
	Comment string
}

func (l *Label) String() string {
	return "$" + l.Name
}

func (l *Label) GoString() string {
	return fmt.Sprintf("label %q", l.String())
}
