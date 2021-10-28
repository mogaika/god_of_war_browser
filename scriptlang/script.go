package scriptlang

import (
	"fmt"
	"strings"
)

type Instruction interface {
	instructionMark()
}

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

func (op *Opcode) AddParameters(params ...interface{}) {
	op.Parameters = append(op.Parameters, params...)
}

func (op *Opcode) instructionMark() {}

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

func (l *Label) InsertBeforeOpcode(instructions []Instruction, targetOp *Opcode) []Instruction {
	index := -1
	for i, instruction := range instructions {
		if op, ok := instruction.(*Opcode); ok {
			if op == targetOp {
				index = i
				break
			}
		}
	}
	if index < 0 {
		panic("failed to find opcode for label")
	}

	// fast insertion via copy
	instructions = append(instructions, nil)
	copy(instructions[index+1:], instructions[index:])
	instructions[index] = l

	return instructions
}

func (l *Label) instructionMark() {}

func RenderScriptLines(instructions []Instruction) []string {
	result := make([]string, 0, len(instructions))
	for _, instruction := range instructions {
		switch v := instruction.(type) {
		case *Opcode:
			if v.Comment == "" {
				result = append(result, v.String())
			} else {
				result = append(result, fmt.Sprintf("%-20s // %s", v.String(), v.Comment))
			}
		case *Label:
			if v.Comment == "" {
				result = append(result, v.String())
			} else {
				result = append(result, fmt.Sprintf("%-20s // %s", v.String(), v.Comment))
			}
		default:
			panic(instruction)
		}
	}
	return result
}

func RenderScript(instructions []Instruction) string {
	return strings.Join(RenderScriptLines(instructions), "\n")
}
