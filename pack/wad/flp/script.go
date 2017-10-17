package flp

import (
	"encoding/binary"
	"fmt"

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
			s.Opcodes = append(s.Opcodes, fmt.Sprintf("OP 0x%x: %v", op, utils.DumpToOneLineString(buf[:3+opLen])))

			buf = buf[3+opLen:]
		} else {
			s.Opcodes = append(s.Opcodes, fmt.Sprintf("OP 0x%x", op))

			buf = buf[1:]
		}
	}

	return s
}
