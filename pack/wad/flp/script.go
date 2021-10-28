package flp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/scriptlang"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Script struct {
	Data       []scriptlang.Instruction `json:"-"`
	Decompiled []string
}

func (s *Script) parseOpcodes(buf []byte, stringsSector []byte) {
	s.Data = make([]scriptlang.Instruction, 0)
	labels := make(map[int16]*scriptlang.Label)
	opOffsets := make(map[int16]*scriptlang.Opcode)
	labelNamesGenerator := utils.RandomNameGenerator{}

	originalBufLen := len(buf)
	for len(buf) != 0 {
		op := &scriptlang.Opcode{
			Code: buf[0],
		}
		opOffset := int16(originalBufLen - len(buf))
		opOffsets[opOffset] = op
		s.Data = append(s.Data, op)

		buf = buf[1:]

		jmpOffsetToLabel := func(jmpOff uint16, jmpOpShift int16) *scriptlang.Label {
			targetOffset := int16(opOffset + int16(jmpOff) + jmpOpShift)

			if lbl, ok := labels[targetOffset]; !ok {
				name := labelNamesGenerator.RandomName()
				targetLabel := &scriptlang.Label{
					Name:    strings.ToLower(name),
					Comment: fmt.Sprintf("0x%.4x", targetOffset),
				}
				labels[targetOffset] = targetLabel
				return targetLabel
			} else {
				return lbl
			}
		}

		strFromOffset := func(buf []byte, dataoff uint16) string {
			stringSecOff := binary.LittleEndian.Uint16(buf[dataoff:])
			return utils.BytesToString(stringsSector[stringSecOff:])
		}

		var stringRepr string = fmt.Sprintf("unknown opcode 0x%x", op.Code)
		//log.Printf("0x%x", op.Code)
		if op.Code&0x80 != 0 {
			if config.GetGOWVersion() == config.GOW1 && config.GetPlayStationVersion() == config.PS2 {
				if len(buf) < 2 {
					log.Printf("Error parsing script: op code parameter missed")
				}

				opLen := binary.LittleEndian.Uint16(buf)
				buf = buf[2:]

				switch op.Code {
				case 0x81:
					target := binary.LittleEndian.Uint16(buf)
					op.AddParameters(int32(target))
					stringRepr = fmt.Sprintf("GotoFrame %d", target)
				case 0x83:
					s1, s2 := utils.BytesToString(buf), utils.BytesToString(buf[1+utils.BytesStringLength(buf):])
					op.AddParameters(s1, s2)
					stringRepr = fmt.Sprintf("Fs queue '%s' command '%s', or response result", s1, s2)
				case 0x8b:
					s1 := utils.BytesToString(buf)
					op.AddParameters(s1)
					stringRepr = fmt.Sprintf("SetTarget '%s'", s1)
				case 0x8c:
					s1 := utils.BytesToString(buf)
					op.AddParameters(s1)
					stringRepr = fmt.Sprintf("GotoLabel '%s'", s1)
				case 0x96:
					pos := uint16(0)
					for pos < opLen {
						stringRepr = "@push"
						if buf[pos] == 0 {
							l := uint16(utils.BytesStringLength(buf[pos+1:]))
							s := utils.BytesToString(buf[pos+1 : pos+1+l])
							stringRepr += fmt.Sprintf("_string '%s' ", s)
							op.AddParameters(s)
							pos += uint16(l) + 2
						} else {
							f := math.Float32frombits(binary.LittleEndian.Uint32(buf[pos+1:]))
							stringRepr += fmt.Sprintf("_float %v ", f)
							op.AddParameters(f)
							pos += 5
						}
					}
				case 0x99:
					label := jmpOffsetToLabel(binary.LittleEndian.Uint16(buf), 5)
					op.AddParameters(label)
					stringRepr = fmt.Sprintf("jump %s", label)
				case 0x9e:
					stringRepr = "CallFrame @pop_string"
				case 0x9d:
					label := jmpOffsetToLabel(binary.LittleEndian.Uint16(buf), 5)
					op.AddParameters(label)
					stringRepr = fmt.Sprintf("jump %s if @pop_bool == true", label)
				case 0x9f:
					op.AddParameters(int32(buf[0]))
					state := "PLAY"
					if buf[0] == 0 {
						state = "STOP"
					}
					stringRepr = fmt.Sprintf("GotoExpression @pop_string (%s)", state)
				default:
					log.Panicf("unknown opcode %v", buf[:opLen])
				}
				buf = buf[opLen:]
			} else {
				opLen := 0
				switch op.Code {
				case 0x81:
					target := binary.LittleEndian.Uint16(buf)
					op.AddParameters(int32(target))
					stringRepr = fmt.Sprintf("GotoFrame %d", target)
					opLen = 2
				case 0x83:
					s1, s2 := strFromOffset(buf, 0), strFromOffset(buf, 2)
					op.AddParameters(s1, s2)
					stringRepr = fmt.Sprintf("Fs queue '%s' command '%s' or response result", s1, s2)
					opLen = 4
				// case 0x8a:
				//     stringRepr = fmt.Sprintf("unused opcode 0x%x", op.Code)
				//	   opLen = 3
				case 0x8b:
					s := strFromOffset(buf, 0)
					op.AddParameters(s)
					stringRepr = fmt.Sprintf("SetTarget '%s'", s)
					opLen = 2
				case 0x8c:
					s := strFromOffset(buf, 0)
					op.AddParameters(s)
					stringRepr = fmt.Sprintf("GotoLabel '%s'", s)
					opLen = 2
				case 0x96:
					if buf[0] == 1 {
						opLen = 5
						f := math.Float32frombits(binary.LittleEndian.Uint32(buf[1:]))
						op.AddParameters(f)
						stringRepr = fmt.Sprintf("push_float %v ", f)
					} else {
						var s string
						if buf[0] != 0 {
							opLen = 2
							s = strFromOffset(buf, 0)
						} else {
							s = strFromOffset(buf, 1)
							opLen = 3
						}
						op.AddParameters(s)
						stringRepr = fmt.Sprintf("push_string '%s'", s)
					}
				case 0x99:
					label := jmpOffsetToLabel(binary.LittleEndian.Uint16(buf), 3)
					op.AddParameters(label)
					stringRepr = fmt.Sprintf("jump %s", label)
					opLen = 2
				//case 0x9a:
				//	stringRepr = fmt.Sprintf("unused opcode 0x%x", op.Code)
				//	opLen = 1
				case 0x9d:
					label := jmpOffsetToLabel(binary.LittleEndian.Uint16(buf), 3)
					op.AddParameters(label)
					stringRepr = fmt.Sprintf("jump %s if @pop_bool == true", label)
					opLen = 2
				case 0x9e:
					stringRepr = "CallFrame @pop_string"
					opLen = 0
				case 0x9f:
					op.AddParameters(int32(buf[0]))
					state := "PLAY"
					if buf[0] == 0 {
						state = "STOP"
					}
					stringRepr = fmt.Sprintf("GotoExpression @pop_string (%s)", state)
					opLen = 1
				default:
					log.Panicf("Unknown variable-length opcode 0x%x", op.Code)
				}
				buf = buf[opLen:]
			}
		} else {
			switch op.Code {
			case 0:
				stringRepr = "end"
			// case 4:
			// advanced version of forced Stop (current target)
			// case 5:
			// advanced version of forced Stop (current target)
			case 6:
				stringRepr = "Play (current target)"
			case 7:
				stringRepr = "Stop (current target)"
			// case 8:
			// switch of some script flag which is not used anywhere?
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
				stringRepr = "@push_bool = strcmp(@pop_string2, @pop_string1) == 0"
			case 0x14:
				stringRepr = "@push_float = strlen(@pop_string1)"
			case 0x15:
				stringRepr = "@push_string = @pop_float1 @pop_float2 @pop_string3 slice string?"
			case 0x17:
				stringRepr = " @pop_any except string discard"
			case 0x18:
				stringRepr = "@push_float = round @pop_float"
			case 0x1c:
				stringRepr = "@push_any vfs get @pop_string1 (relative to current target)"
			case 0x1d:
				stringRepr = "vfs set @pop_string2 = @pop_string1 (relative to current target)"
			case 0x20:
				stringRepr = "SetTarget @pop_string1"
			case 0x21:
				stringRepr = "@push_string = @pop_string2 append to @pop_string1"
			case 0x22:
				stringRepr = "@push_any get @pop_float1 target @pop_string2"
			case 0x23:
				stringRepr = " @pop_string1 @pop_float2 target @pop_string3"
			case 0x26:
				stringRepr = " @pop_string discard"
			case 0x29:
				stringRepr = "@push_bool = strcmp(@pop_string2, @pop_string1) != 0"
			case 0x34:
				stringRepr = "@push_float  current timer value"
			}
		}
		op.Comment = stringRepr
	}

	for labelOffset, label := range labels {
		labelOp, ok := opOffsets[labelOffset]
		if !ok {
			panic("failed to find label for opcode")
		}

		s.Data = label.InsertBeforeOpcode(s.Data, labelOp)
	}
}

// if marshaler == nil, then will not fill string offsets
// useful for size calculation
func (s *Script) Marshal(fm *FlpMarshaler) []byte {
	labelOffsets := make(map[string]int16)
	labelFills := make(map[int]string)

	var buf bytes.Buffer
	for _, instruction := range s.Data {
		switch instruction.(type) {
		case *scriptlang.Label:
			label := instruction.(*scriptlang.Label)
			labelOffsets[label.Name] = int16(buf.Len())
		case *scriptlang.Opcode:
			op := instruction.(*scriptlang.Opcode)
			opOffset := int16(buf.Len())

			writeU16 := func(v uint16) {
				var tmp [2]byte
				binary.LittleEndian.PutUint16(tmp[:], v)
				buf.Write(tmp[:])
			}
			writeU32 := func(v uint32) {
				var tmp [4]byte
				binary.LittleEndian.PutUint32(tmp[:], v)
				buf.Write(tmp[:])
			}
			writeLabelOff := func(labelName string, jmpOpShift int16) {
				labelFills[buf.Len()] = labelName
				// later we substract this value from label offset (currently unknown)
				writeU16(uint16(opOffset + jmpOpShift))
			}
			writeStringOffset := func(s string) {
				if fm != nil {
					fm.sbuffer.Add(s, fm.pos()+buf.Len(), 2).AllowEmptyString = true
				}
				writeU16(0)
			}

			buf.WriteByte(op.Code)
			if op.Code&0x80 != 0 {
				if config.GetGOWVersion() == config.GOW1 && config.GetPlayStationVersion() == config.PS2 {
					switch op.Code {
					case 0x81:
						writeU16(2)
						writeU16(uint16(int16(op.Parameters[0].(int32))))
					case 0x83:
						s1, s2 :=
							utils.StringToBytes(op.Parameters[0].(string), true),
							utils.StringToBytes(op.Parameters[1].(string), true)

						writeU16(uint16(len(s1) + len(s2)))
						buf.Write(s1)
						buf.Write(s2)
					case 0x8b:
						s := utils.StringToBytes(op.Parameters[0].(string), true)
						writeU16(uint16(len(s)))
						buf.Write(s)
					case 0x8c:
						s := utils.StringToBytes(op.Parameters[0].(string), true)
						writeU16(uint16(len(s)))
						buf.Write(s)
					case 0x96:
						// calc op length first
						l := uint16(0)
						for _, p := range op.Parameters {
							switch v := p.(type) {
							case float32, int32:
								l += 5
							case string:
								l += 1 + uint16(len(utils.StringToBytes(v, true)))
							default:
								log.Panicf("%#+v %T", p, p)
							}
						}
						writeU16(l)

						for _, p := range op.Parameters {
							switch v := p.(type) {
							case int32:
								buf.WriteByte(1) // TODO: check this 1
								writeU32(math.Float32bits(float32(v)))
							case float32:
								buf.WriteByte(1) // TODO: check this 1
								writeU32(math.Float32bits(v))
							case string:
								buf.WriteByte(0) // TODO: check this 0
								buf.Write(utils.StringToBytes(v, true))
							}
						}
					case 0x99:
						writeU16(2)
						writeLabelOff(op.Parameters[0].(*scriptlang.Label).Name, 5)
					case 0x9e:
						writeU16(0)
					case 0x9d:
						writeU16(2)
						writeLabelOff(op.Parameters[0].(*scriptlang.Label).Name, 5)
					case 0x9f:
						writeU16(1)
						buf.WriteByte(byte(op.Parameters[0].(int32)))
					default:
						log.Panicf("unknown opcode(gow1) {%+#v}", op)
					}
				} else {
					switch op.Code {
					case 0x81:
						writeU16(uint16(int16(op.Parameters[0].(int32))))
					case 0x83:
						writeStringOffset(op.Parameters[0].(string))
						writeStringOffset(op.Parameters[1].(string))
					case 0x8b:
						writeStringOffset(op.Parameters[0].(string))
					case 0x8c:
						writeStringOffset(op.Parameters[0].(string))
					case 0x96:
						switch v := op.Parameters[0].(type) {
						case int32:
							buf.WriteByte(1)
							writeU32(math.Float32bits(float32(v)))
						case float32:
							buf.WriteByte(1)
							writeU32(math.Float32bits(v))
						case string:
							buf.WriteByte(0)
							writeStringOffset(v)
						default:
							panic(op)
						}
					case 0x99:
						writeLabelOff(op.Parameters[0].(*scriptlang.Label).Name, 3)
					case 0x9d:
						writeLabelOff(op.Parameters[0].(*scriptlang.Label).Name, 3)
					case 0x9e:
					// nothing
					case 0x9f:
						buf.WriteByte(byte(op.Parameters[0].(int32)))
					default:
						log.Panicf("unknown opcode {%+#v}", op)
					}
				}
			}
		}
	}

	result := buf.Bytes()

	// insert label offsets
	for opOffset, labelName := range labelFills {
		labelOffset, exists := labelOffsets[labelName]
		if !exists {
			log.Panicf("unknown label %v", labelName)
		}

		opOffAndJmpShift := binary.LittleEndian.Uint16(result[opOffset:])
		binary.LittleEndian.PutUint16(result[opOffset:], uint16(labelOffset-int16(opOffAndJmpShift)))
	}

	return result
}

func (s *Script) FromDecompiled() error {
	if data, err := scriptlang.ParseScript([]byte(strings.Join(s.Decompiled, "\n"))); err != nil {
		return errors.Wrapf(err, "Failed to decompile script")
	} else {
		//log.Printf("\n%v", s.Decompiled)
		s.Data = data
		//utils.LogDump(s.Decompiled, s.Data)
		return nil
	}
}

func NewScriptFromData(buf []byte, stringsSector []byte) (s *Script) {
	s = new(Script)
	s.parseOpcodes(buf, stringsSector)
	s.Decompiled = strings.Split(scriptlang.RenderScript(s.Data), "\n")

	// testing of decompiler
	/*
		if data, err := scriptlang.ParseScript([]byte(s.Decompiled)); err == nil {
			ns := new(Script)
			ns.Data = data
			newbuf := ns.Marshal(nil)
			ns.parseOpcodes(newbuf, []byte{0})
			ns.Decompiled = ns.dissasembleToString()
			//if len(buf) != len(newbuf) {
			utils.LogDump(buf, newbuf)
			//	log.Panic(len(buf), len(newbuf))
			//}
		} else {
			log.Panic(err)
		}
	*/
	return s
}
