package flp

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/mogaika/god_of_war_browser/utils"
)

type Script struct {
	Opcodes []string
}

func NewScriptFromData(realBuf []byte) *Script {
	s := &Script{Opcodes: make([]string, 0)}

	buf := realBuf
	for len(buf) != 0 {
		op := buf[0]
		if op&0x80 != 0 {
			opLen := binary.LittleEndian.Uint16(buf[1:])
			buf = buf[3:]

			var stringRepr string
			switch op {
			case 0x81:
				stringRepr = fmt.Sprintf("GotoFrame %d", binary.LittleEndian.Uint16(buf))
			case 0x8b:
				stringRepr = fmt.Sprintf("SetTarget '%s'", utils.BytesToString(buf))
			case 0x8c:
				stringRepr = fmt.Sprintf("GotoLabel '%s'", utils.BytesToString(buf))
			case 0x96:
				pos := uint16(0)
				for pos < opLen {
					stringRepr = "@push"
					if buf[pos] == 0 {
						l := uint16(utils.BytesStringLength(buf[pos+1:]))
						stringRepr += fmt.Sprintf("_string '%s' ", utils.DumpToOneLineString(buf[pos+1:pos+1+l]))
						pos += uint16(l) + 2
					} else {
						stringRepr += fmt.Sprintf("_float %v ", math.Float32frombits(binary.LittleEndian.Uint32(buf[pos+1:])))
						pos += 5
					}
				}
			case 0x99:
				stringRepr = fmt.Sprintf("jump %+d", binary.LittleEndian.Uint16(buf))
			case 0x9e:
				stringRepr = "CallFrame @pop_string"
			case 0x9d:
				stringRepr = fmt.Sprintf("jump %+d if @pop_bool == true", binary.LittleEndian.Uint16(buf))
			case 0x9f:
				state := "PLAY"
				if buf[0] == 0 {
					state = "STOP"
				}
				stringRepr = fmt.Sprintf("GotoExpression '%s' (%s)", utils.BytesToString(buf[1:]), state)
			default:
				stringRepr = fmt.Sprintf("   << dump{%s} >>", utils.DumpToOneLineString(buf[:opLen]))
			}

			s.Opcodes = append(s.Opcodes, fmt.Sprintf("OP 0x%.2x: %s", op, stringRepr))

			buf = buf[opLen:]
		} else {
			buf = buf[1:]
			var stringRepr string
			switch op {
			case 6:
				stringRepr = "Play"
			case 7:
				stringRepr = "Stop"
			case 0xe:
				stringRepr = "@push_bool = @pop_float1 is equal @pop_float2"
			case 0x11:
				stringRepr = "@push_bool = @pop_bool1 AND @pop_bool2"
			case 0x12:
				stringRepr = "@push_bool = convert_to_bool @pop_any"
			case 0x1c:
				stringRepr = "@push_any vfs get @pop_string1"
			case 0x1d:
				stringRepr = "vfs set @pop_string2 = @pop_string1"
			case 0x20:
				stringRepr = "SetTarget @pop_string1"
			}
			s.Opcodes = append(s.Opcodes, fmt.Sprintf("OP 0x%.2x: %v", op, stringRepr))
		}
	}

	return s
}
