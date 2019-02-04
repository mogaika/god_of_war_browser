package flp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/Pallinder/go-randomdata"
	"github.com/mogaika/god_of_war_browser/utils"
)

var scriptPushRefFiller []ScriptOpcodeStringPushReference
var scriptPushRefLocker sync.Mutex

type Script struct {
	Opcodes         []*ScriptOpcode `json:"-"`
	Decompiled      string
	labels          map[int16]string
	marshaled       []byte
	staticStringRef map[string][]uint16 // string ot offset
}

type ScriptOpcode struct {
	Offset int16
	Data   []byte
	String string
	Code   byte
}

type ScriptOpcodeStringPushReference struct {
	Opcode *ScriptOpcode `json:"-"`
	String []byte
}

func (s *Script) parseOpcodes(buf []byte, stringsSector []byte) {
	s.Opcodes = make([]*ScriptOpcode, 0)
	s.labels = make(map[int16]string)

	originalBufLen := len(buf)
	for len(buf) != 0 {
		op := &ScriptOpcode{
			Code:   buf[0],
			Offset: int16(originalBufLen - len(buf)),
			Data:   buf,
		}

		jmpOffsetToStr := func(jmpoff uint16, opoff int16) string {
			targetOffset := int16(op.Offset + int16(jmpoff) + opoff)

			var targetLabel string
			if lbl, ok := s.labels[targetOffset]; !ok {
				targetLabel = strings.ToLower(randomdata.SillyName())
				s.labels[targetOffset] = targetLabel
			} else {
				targetLabel = lbl
			}

			return fmt.Sprintf("$%s(%+x=%.4x)", targetLabel, int16(jmpoff), targetOffset)
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

		var stringRepr string = fmt.Sprintf("unknown opcode 0x%x", op.Code)
		buf = buf[1:]
		if op.Code&0x80 != 0 {
			if config.GetPlayStationVersion() == config.PS2 {
				if len(buf) < 2 {
					log.Printf("Error parsing script: op code parameter missed")
				}
				opLen := binary.LittleEndian.Uint16(buf)
				buf = buf[2:]
				op.Data = buf[:opLen]

				switch op.Code {
				case 0x81:
					stringRepr = fmt.Sprintf("GotoFrame %d", binary.LittleEndian.Uint16(buf))
				case 0x83:
					stringRepr = fmt.Sprintf("Fs queue '%s' command '%s', or response result",
						utils.BytesToString(buf), utils.BytesToString(buf[1+utils.BytesStringLength(buf):]))
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

							if l != 0 {
								reff := ScriptOpcodeStringPushReference{
									Opcode: op,
									String: buf[pos+1 : pos+1+l],
								}

								if reff.String[len(reff.String)-1] == 0 {
									reff.String = reff.String[:len(reff.String)-1]
								}
								scriptPushRefFiller = append(scriptPushRefFiller, reff)
							}
							stringRepr += fmt.Sprintf("_string '%s' ", utils.DumpToOneLineString(buf[pos+1:pos+1+l]))
							pos += uint16(l) + 2
						} else {
							stringRepr += fmt.Sprintf("_float %v ", math.Float32frombits(binary.LittleEndian.Uint32(buf[pos+1:])))
							pos += 5
						}
					}
				case 0x99:
					stringRepr = fmt.Sprintf("jump %s", jmpOffsetToStr(binary.LittleEndian.Uint16(buf), 5))
				case 0x9e:
					stringRepr = "CallFrame @pop_string"
				case 0x9d:
					stringRepr = fmt.Sprintf("jump %s if @pop_bool == true", jmpOffsetToStr(binary.LittleEndian.Uint16(buf), 5))
				case 0x9f:
					state := "PLAY"
					if buf[0] == 0 {
						state = "STOP"
					}
					stringRepr = fmt.Sprintf("GotoExpression @pop_string (%s)", state)
				default:
					stringRepr = fmt.Sprintf(" unknown opcode  << dump{%s} >>", utils.DumpToOneLineString(buf[:opLen]))
				}
				buf = buf[opLen:]
			} else if config.GetPlayStationVersion() == config.PSVita {
				opLen := 0
				op.Data = buf
				switch op.Code {
				case 0x81:
					stringRepr = fmt.Sprintf("GotoFrame %d", binary.LittleEndian.Uint16(buf))
					opLen = 2
				case 0x83:
					stringRepr = fmt.Sprintf("Fs queue '%s' command '%s' or response result",
						strFromOffset(0), strFromOffset(2))
					opLen = 4
				case 0x8a:
					stringRepr = fmt.Sprintf("unused opcode 0x%x", op.Code)
					opLen = 3
				case 0x8b:
					stringRepr = fmt.Sprintf("SetTarget '%s'", strFromOffset(0))
					opLen = 2
				case 0x8c:
					stringRepr = fmt.Sprintf("GotoLabel '%s'", strFromOffset(0))
					opLen = 2
				case 0x96:
					if buf[0] == 1 {
						opLen = 5
						stringRepr = fmt.Sprintf("push_float %v ", math.Float32frombits(binary.LittleEndian.Uint32(buf[1:])))
					} else {
						if buf[0] != 0 {
							opLen = 2
							stringRepr = fmt.Sprintf("push_string '%s'", strFromOffset(0))
						} else {
							opLen = 3
							stringRepr = fmt.Sprintf("push_string '%s' ", strFromOffset(1))
						}
					}
				case 0x99:
					stringRepr = fmt.Sprintf("jump %s", jmpOffsetToStr(binary.LittleEndian.Uint16(buf), 3))
					opLen = 2
				case 0x9a:
					stringRepr = fmt.Sprintf("unused opcode 0x%x", op.Code)
					opLen = 1
				case 0x9d:
					stringRepr = fmt.Sprintf("jump %s if @pop_bool == true", jmpOffsetToStr(binary.LittleEndian.Uint16(buf), 3))
					opLen = 2
				case 0x9e:
					stringRepr = "CallFrame @pop_string"
					opLen = 0
				case 0x9f:
					state := "PLAY"
					if buf[0] == 0 {
						state = "STOP"
					}
					stringRepr = fmt.Sprintf("GotoExpression @pop_string (%s)", state)
					opLen = 1
				default:
					log.Panicf("Unknown variable-length opcode 0x%x", op.Code)
				}
				op.Data = buf[:opLen]
				buf = buf[opLen:]
			} else {
				log.Panicf("Unsupported version of ps")
			}
		} else {
			switch op.Code {
			case 0:
				stringRepr = "end"
			case 6:
				stringRepr = "Play"
			case 7:
				stringRepr = "Stop"
			case 0xa:
				stringRepr = "@push_float = @pop_float2 + @pop_float1"
			case 0xb:
				stringRepr = "@push_float = @pop_float2 - @pop_float1"
			case 0xc:
				stringRepr = "@push_float = @pop_float2 * @pop_float1"
			case 0xd:
				stringRepr = "@push_float = @pop_float2 / @pop_float1"
			case 0xe:
				stringRepr = "@push_bool = @pop_float1 == close to == @pop_float2"
			case 0xf:
				stringRepr = "@push_bool = @pop_float2 < @pop_float1"
			case 0x10:
				stringRepr = "@push_bool = @pop_bool1 AND @pop_bool2"
			case 0x11:
				stringRepr = "@push_bool = @pop_bool1 OR @pop_bool2"
			case 0x12:
				stringRepr = "@push_bool = convert_to_bool @pop_any"
			case 0x13:
				stringRepr = "@push_bool = strcmp(@pop_string2, @pop_string1) <= 0"
			case 0x17:
				stringRepr = " @pop_any to nothing"
			case 0x18:
				stringRepr = "@push_float = round @pop_float"
			case 0x1c:
				stringRepr = "@push_any vfs get @pop_string1"
			case 0x1d:
				stringRepr = "vfs set @pop_string2 = @pop_string1"
			case 0x20:
				stringRepr = "SetTarget @pop_string1"
			case 0x21:
				stringRepr = "@push_string = @pop_string2 append to @pop_string1"
			case 0x34:
				stringRepr = "@push_float  current timer value"
			default:
				stringRepr = " unknown opcode "
			}
		}
		op.String = stringRepr
		s.Opcodes = append(s.Opcodes, op)
	}
}

func (s *Script) dissasembleToString() string {
	strs := make([]string, 0)
	pos := int16(0)
	ops := s.Opcodes
	for len(ops) != 0 {
		if label, ex := s.labels[pos]; ex {
			strs = append(strs, fmt.Sprintf("%.4x: $%s", pos, label))
		}

		op := ops[0]
		if op.Offset == pos {
			strs = append(strs, fmt.Sprintf("%.4x: %.2x: %s", op.Offset, op.Code, op.String))
			ops = ops[1:]
		}

		pos++
	}

	return strings.Join(strs, "\n")
}

func (s *Script) Marshal() []byte {
	var r bytes.Buffer
	for _, op := range s.Opcodes {
		r.WriteByte(op.Code)
		if op.Code&0x80 != 0 {
			if config.GetPlayStationVersion() != config.PSVita {
				var lenBuf [2]byte
				binary.LittleEndian.PutUint16(lenBuf[:], uint16(len(op.Data)))
				r.Write(lenBuf[:])
			}
			r.Write(op.Data)
		}
	}
	s.marshaled = r.Bytes()
	return s.marshaled
}

func NewScriptFromData(buf []byte, stringsSector []byte) *Script {
	s := new(Script)
	s.parseOpcodes(buf, stringsSector)
	s.Decompiled = strings.Replace("\n"+s.dissasembleToString(), "\n", "<br>", -1)
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
