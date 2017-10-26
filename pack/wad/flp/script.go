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
	//utils.LogDump("SCRIPT ", realBuf)

	buf := realBuf
	for len(buf) != 0 {
		op := buf[0]
		if op&0x80 != 0 {
			opLen := binary.LittleEndian.Uint16(buf[1:])
			buf = buf[3:]

			var stringRepr string
			switch op {
			case 0x96:
				stringRepr = "push "
				pos := uint16(0)
				for pos < opLen {
					if buf[pos] == 0 {
						l := uint16(utils.BytesStringLength(buf[pos+1:]))
						stringRepr += fmt.Sprintf(" s'%s'", utils.DumpToOneLineString(buf[pos+1:pos+1+l]))
						pos += uint16(l) + 2
					} else {
						stringRepr += fmt.Sprintf(" f'%v'", math.Float32frombits(binary.LittleEndian.Uint32(buf[pos+1:])))
						pos += 5
					}
				}
			case 0x9e:
				stringRepr = "CallFrame @pop_string"
			default:
				stringRepr = fmt.Sprintf("dump{%s}", utils.DumpToOneLineString(buf[:opLen]))
			}

			s.Opcodes = append(s.Opcodes, fmt.Sprintf("OP 0x%x: %s", op, stringRepr))

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
			case 0x12:
				stringRepr = "@push_bool = convert_to_bool @pop_any"
			case 0x1c:
				stringRepr = "@push vfs get @pop_string1"
			case 0x1d:
				stringRepr = "vfs set @pop_string2 = @pop_string1"
			}
			s.Opcodes = append(s.Opcodes, fmt.Sprintf("OP 0x%x: %v", op, stringRepr))
		}
	}

	return s
}
