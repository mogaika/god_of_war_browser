package flp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"html"
	"log"
	"math"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/Pallinder/go-randomdata"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/utils"
)

var scriptPushRefFiller []ScriptOpcodeStringPushReference
var scriptPushRefLocker sync.Mutex

type Script struct {
	Opcodes    []*ScriptOpcode `json:"-"`
	Decompiled string

	labels          map[int16]string
	labelToOffset   map[string]int16
	marshaled       []byte
	staticStringRef map[string][]uint16 // string to offset
}

type ScriptOpcode struct {
	Offset     int16
	Data       []byte
	Code       byte
	Parameters []interface{}
}

func (s *Script) OffsetToLabel(offset int16) string {
	var lName string
	if existingName, ok := s.labels[offset]; !ok {
		lName = fmt.Sprintf("$%s(%.4x)", strings.ToLower(randomdata.SillyName()), offset)
		s.labels[offset] = lName
		s.labelToOffset[lName] = offset
	} else {
		lName = existingName
	}

	return lName
}

func (op *ScriptOpcode) String() string {
	switch op.Code {
	case 0:
		return "end"
	case 6:
		return "Play"
	case 7:
		return "Stop"
	case 0xa:
		return "@push_float = @pop_float2 + @pop_float1"
	case 0xb:
		return "@push_float = @pop_float2 - @pop_float1"
	case 0xc:
		return "@push_float = @pop_float2 * @pop_float1"
	case 0xd:
		return "@push_float = @pop_float2 / @pop_float1"
	case 0xe:
		return "@push_bool = @pop_float1 == close to == @pop_float2"
	case 0xf:
		return "@push_bool = @pop_float2 < @pop_float1"
	case 0x10:
		return "@push_bool = @pop_bool1 AND @pop_bool2"
	case 0x11:
		return "@push_bool = @pop_bool1 OR @pop_bool2"
	case 0x12:
		return "@push_bool = convert_to_bool @pop_any"
	case 0x13:
		return "@push_bool = strcmp(@pop_string2, @pop_string1) <= 0"
	case 0x17:
		return " @pop_any to nothing"
	case 0x18:
		return "@push_float = round @pop_float"
	case 0x1c:
		return "@push_any vfs get @pop_string1"
	case 0x1d:
		return "vfs set @pop_string2 = @pop_string1"
	case 0x20:
		return "SetTarget @pop_string1"
	case 0x21:
		return "@push_string = @pop_string2 append to @pop_string1"
	case 0x34:
		return "@push_float  current timer value"
	case 0x81:
		return fmt.Sprintf("GotoFrame %d", op.Parameters...)
	case 0x83:
		return fmt.Sprintf("Fs queue '%s' command '%s', or response result", op.Parameters...)
	case 0x8a:
		return fmt.Sprintf("unused opcode 0x8a %+#v", op.Parameters)
	case 0x8b:
		return fmt.Sprintf("SetTarget '%s'", op.Parameters...)
	case 0x8c:
		return fmt.Sprintf("GotoLabel '%s'", op.Parameters...)
	case 0x96:
		return fmt.Sprintf("push %#+v", op.Parameters)
	case 0x99:
		return fmt.Sprintf("jump %s", op.Parameters...)
	case 0x9a:
		return fmt.Sprintf("unused opcode 0x9a %+#v", op.Parameters)
	case 0x9d:
		return fmt.Sprintf("jump %s if @pop_bool == true", op.Parameters...)
	case 0x9e:
		return fmt.Sprintf("CallFrame @pop_string")
	case 0x9f:
		return fmt.Sprintf("GotoExpression @pop_string (PLAY: %v)", op.Parameters...)
	}
	return fmt.Sprintf(" unknown opcode 0x%.2x << dump{%s} >>", op.Code, op.Data)
}

type ScriptOpcodeStringPushReference struct {
	Opcode *ScriptOpcode `json:"-"`
	String []byte
}

func (s *Script) parseOpcodes(buf []byte, stringsSector []byte) {
	s.Opcodes = make([]*ScriptOpcode, 0, 8)
	s.labels = make(map[int16]string)
	s.labelToOffset = make(map[string]int16)

	originalBufLen := len(buf)
	for len(buf) != 0 {
		op := &ScriptOpcode{
			Code:   buf[0],
			Offset: int16(originalBufLen - len(buf)),
			Data:   buf,
		}

		labelFromJmp := func(jmpoff uint16, opoff int16) string {
			targetOffset := int16(op.Offset + int16(jmpoff) + opoff)
			return s.OffsetToLabel(targetOffset)
		}

		strFromOffset := func(dataoff uint16) string {
			stringSecOff := binary.LittleEndian.Uint16(op.Data[dataoff:])

			str := utils.BytesToString(stringsSector[stringSecOff:])
			if s.staticStringRef == nil {
				s.staticStringRef = make(map[string][]uint16)
			}
			if _, exists := s.staticStringRef[str]; !exists {
				s.staticStringRef[str] = make([]uint16, 0)
			}
			s.staticStringRef[str] = append(s.staticStringRef[str], uint16(op.Offset)+1+dataoff)
			return str
		}

		buf = buf[1:]
		if op.Code&0x80 != 0 {
			op.Parameters = make([]interface{}, 0)

			switch config.GetPlayStationVersion() {
			case config.PS2:
				if len(buf) < 2 {
					log.Printf("Error parsing script: op code parameter missed")
				}
				opLen := binary.LittleEndian.Uint16(buf)
				buf = buf[2:]
				op.Data = buf[:opLen]

				switch op.Code {
				case 0x81:
					op.Parameters = append(op.Parameters, binary.LittleEndian.Uint16(buf))
				case 0x83:
					op.Parameters = append(op.Parameters,
						utils.BytesToString(buf), utils.BytesToString(buf[1+utils.BytesStringLength(buf):]))
				case 0x8b, 0x8c:
					op.Parameters = append(op.Parameters, utils.BytesToString(buf))
				case 0x96:
					pos := uint16(0)
					for pos < opLen {
						if buf[pos] == 0 {
							l := uint16(utils.BytesStringLength(buf[pos+1:]))
							op.Parameters = append(op.Parameters, utils.DumpToOneLineString(buf[pos+1:pos+1+l]))
							pos += uint16(l) + 2
						} else {
							op.Parameters = append(op.Parameters, math.Float32frombits(binary.LittleEndian.Uint32(buf[pos+1:])))
							pos += 5
						}
					}
				case 0x99, 0x9d:
					op.Parameters = append(op.Parameters, labelFromJmp(binary.LittleEndian.Uint16(buf), 5))
				case 0x9f:
					op.Parameters = append(op.Parameters, buf[0] != 0)
				}
				buf = buf[opLen:]
			case config.PSVita:
				opLen := 0
				op.Data = buf
				switch op.Code {
				case 0x81:
					op.Parameters = append(op.Parameters, binary.LittleEndian.Uint16(buf))
					opLen = 2
				case 0x83:
					op.Parameters = append(op.Parameters, strFromOffset(0), strFromOffset(2))
					opLen = 4
				case 0x8a:
					opLen = 3
				case 0x8b, 0x8c:
					op.Parameters = append(op.Parameters, strFromOffset(0))
					opLen = 2
				case 0x96:
					if buf[0] == 1 {
						opLen = 5
						op.Parameters = append(op.Parameters, math.Float32frombits(binary.LittleEndian.Uint32(buf[1:])))
					} else {
						if buf[0] != 0 {
							opLen = 2
							op.Parameters = append(op.Parameters, strFromOffset(0))
						} else {
							opLen = 3
							op.Parameters = append(op.Parameters, strFromOffset(1))
						}
					}
				case 0x99, 0x9d:
					op.Parameters = append(op.Parameters, labelFromJmp(binary.LittleEndian.Uint16(buf), 3))
					opLen = 2
				case 0x9a:
					opLen = 1
				case 0x9e:
					opLen = 0
				case 0x9f:
					op.Parameters = append(op.Parameters, buf[0] != 0)
					opLen = 1
				default:
					log.Panicf("Unknown variable-length opcode 0x%x", op.Code)
				}
				op.Data = buf[:opLen]
				buf = buf[opLen:]
			default:
				log.Panicf("Unsupported version of ps")
			}

		}
		s.Opcodes = append(s.Opcodes, op)
	}
}

func (s *Script) dissasembleToString() string {
	text := NewDecompiler(s).Decompile()
	return text

	/*
		strs := make([]string, 0)
		pos := int16(0)
		ops := s.Opcodes

		strs = append(strs, text)

		for len(ops) != 0 {
			if label, ex := s.labels[pos]; ex {
				strs = append(strs, fmt.Sprintf("%.4x: $%s", pos, label))
			}

			op := ops[0]
			if op.Offset == pos {
				strs = append(strs, fmt.Sprintf("%.4x: %.2x: %s", op.Offset, op.Code, op.String()))
				ops = ops[1:]
			}

			pos++
		}

		return strings.Join(strs, "\n")
	*/
}

func (s *Script) Marshal() []byte {
	var r bytes.Buffer
	for _, op := range s.Opcodes {
		r.WriteByte(op.Code)
		if op.Code&0x80 != 0 {
			switch config.GetPlayStationVersion() {
			case config.PSVita:
				var lenBuf [2]byte
				binary.LittleEndian.PutUint16(lenBuf[:], uint16(len(op.Data)))
				r.Write(lenBuf[:])
				r.Write(op.Data)
			case config.PS2:
				r.Write(op.Data)
			default:
				panic("Unsupported ps version for flp script marshaling")
			}
		}
	}
	s.marshaled = r.Bytes()
	return s.marshaled
}

func NewScriptFromData(buf []byte, stringsSector []byte) (s *Script) {
	s = new(Script)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error parsing script: %v\n%s", r, debug.Stack())
		}
	}()
	s.parseOpcodes(buf, stringsSector)
	s.Decompiled = strings.Replace(html.EscapeString(s.dissasembleToString()), "\n", "<br>", -1)
	return s
}

func (oref *ScriptOpcodeStringPushReference) ChangeString(data []byte) {
	var buf bytes.Buffer
	buf.WriteByte(0)
	buf.Write(data)
	if len(data) == 0 || data[len(data)-1] != 0 {
		buf.WriteByte(0)
	}
	oref.Opcode.Data = buf.Bytes()
}

func enterScriptPushRefFiller() {
	scriptPushRefFiller = make([]ScriptOpcodeStringPushReference, 0)
	scriptPushRefLocker.Lock()
}

func exitScriptPushRefFiller() []ScriptOpcodeStringPushReference {
	defer func() {
		scriptPushRefFiller = nil
		scriptPushRefLocker.Unlock()
	}()
	return scriptPushRefFiller
}
