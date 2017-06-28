package targets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad/scr/store"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Entities struct {
	Array []*Entity
}

type OpcodesStream struct {
	Stream     []byte `json:"-"`
	Start      int
	Decompiled string
}

func (o *OpcodesStream) Decompile() {
	if o.Decompiled == "" {
		o.Decompiled = DecompileEscScript(o.Stream, o.Start)
	}
}

func ScopeToString(scope uint16) string {
	if scope < 4 {
		return []string{"Entity", "Internal", "GlobalData", "LevelData"}[scope]
	} else {
		return fmt.Sprintf("*unknown scope 0x%x*", scope)
	}
}

func TypeIdToString(typeid uint8) string {
	if typeid < 4 {
		return []string{"float", "int", "bool", "string"}[typeid]
	} else {
		return fmt.Sprintf("*unknown type 0x%x*", typeid)
	}
}

const (
	SCOPE_ENTITY     = 0
	SCOPE_INTERNAL   = 1
	SCOPE_GLOBALDATA = 2
	SCOPE_LEVELDATA  = 3
)

func GetEscFunc(scope, fid uint16) string {
	switch scope {
	case SCOPE_INTERNAL:
		switch fid {
		case 0xc:
			return "Idle (bool needIdle)"
		case 0xa:
			return "Print text on screen(type, messageID)"
		case 0x1f:
			return "Print text on screen(unkn, type, messageID)"
		}
	}
	return ""
}

func argsParseScopeFunc(b []byte) (scope, fid uint16) {
	args := binary.LittleEndian.Uint16(b)
	return args >> 12, args & 0xfff
}

func DecompileEscScript(b []byte, pointer int) string {
	fail := false
	var output bytes.Buffer

	wrline := func(format string, args ...interface{}) {
		output.WriteString(fmt.Sprintf(format, args...) + "\n")
	}

	defer func() {
		if r := recover(); r != nil {
			wrline("%v", r)
		}
	}()

	for !fail {
		opcode := b[pointer]
		output.WriteString(fmt.Sprintf("0x%.4x: %.2X:  ", pointer, opcode))

		if opcode >= 0x3a {
			wrline("exit;")
			break
		}

		pointer++
		opcodeBuf := b[pointer:]

		switch opcode {
		case 0x00:
			wrline("push_float %f;", math.Float32frombits(binary.LittleEndian.Uint32(opcodeBuf)))
			pointer += 4
		case 0x01:
			wrline("push_int %d;", int32(binary.LittleEndian.Uint32(opcodeBuf)))
			pointer += 4
		case 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09:
			scope, fid := argsParseScopeFunc(opcodeBuf)
			var target string
			if scope < 8 {
				target = ScopeToString(scope)
			} else if scope == 8 {
				target = "Special local scope? Currnet obj scope?"
			} else if scope > 8 {
				target = "Array at 0x334A98 word[0x256]"
			}
			if opcode < 0x06 {
				wrline("get_scope_%s from (0x%x)'%s' val 0x%x;", TypeIdToString(opcode-2), scope, target, fid)
			} else {
				wrline("set_scope_%s from (0x%x)'%s' val 0x%x;", TypeIdToString(opcode-2), scope, target, fid)
			}
		case 0x0b:
			scope, fid := argsParseScopeFunc(opcodeBuf)
			wrline("call function scope(0x%x), func(0x%x) '%s' => int;", scope, fid, GetEscFunc(scope, fid))
			pointer += 2
		case 0x0c:
			scope, fid := argsParseScopeFunc(opcodeBuf)
			wrline("call function scope(0x%x), func(0x%x) '%s' => float;", scope, fid, GetEscFunc(scope, fid))
			pointer += 2

		case 0x0e:
			off := binary.LittleEndian.Uint16(opcodeBuf)
			wrline("push_string '%s';", utils.BytesToString(b[off:]))
			pointer += 2

		case 0x11:
			wrline("push_bool TRUE;")
		case 0x12:
			wrline("push_bool FALSE;")

		case 0x23:
			wrline("pop_int2float_push;")

		case 0x38:
			wrline("pop_result;")
		default:
			wrline("unknown opcode 0x%x;", opcode)
			fail = true
		}
	}
	return output.String()
}

type Entity struct {
	Matrix mgl32.Mat4

	Field_0x40         uint16
	Size               uint16
	EntityType         uint16
	EntityUniqueID     uint16
	Field_0x4a         uint16
	Field_0x4c         uint16
	HandlersCount      uint16
	Field_0x50         uint16
	OpcodesStreamsSize uint16

	Name     string
	Handlers map[uint16]OpcodesStream
}

func EntityFromBytes(b []byte) *Entity {
	e := &Entity{
		Field_0x40:         binary.LittleEndian.Uint16(b[0x40:]),
		Size:               binary.LittleEndian.Uint16(b[0x44:]),
		EntityType:         binary.LittleEndian.Uint16(b[0x46:]),
		EntityUniqueID:     binary.LittleEndian.Uint16(b[0x48:]),
		Field_0x4a:         binary.LittleEndian.Uint16(b[0x4a:]),
		Field_0x4c:         binary.LittleEndian.Uint16(b[0x4c:]),
		HandlersCount:      binary.LittleEndian.Uint16(b[0x4e:]),
		Field_0x50:         binary.LittleEndian.Uint16(b[0x50:]),
		OpcodesStreamsSize: binary.LittleEndian.Uint16(b[0x52:]),
		Handlers:           make(map[uint16]OpcodesStream),
	}

	textStart := 0x54 + e.OpcodesStreamsSize + e.HandlersCount*4 + e.Field_0x50*2
	e.Name = utils.BytesToString(b[textStart:])

	utils.Dump(e.Name, e.EntityType)

	opcodesDescrStart := uint16(0x54)
	opcodesStreamStart := int(opcodesDescrStart + e.HandlersCount*4 + e.Field_0x50*2 + 2)

	opcodesStream := b[opcodesStreamStart:]

	for i := uint16(0); i < e.HandlersCount; i++ {
		start := binary.LittleEndian.Uint16(b[opcodesDescrStart+i*4+2:])

		id := binary.LittleEndian.Uint16(b[opcodesDescrStart+i*4:])

		function := OpcodesStream{Stream: opcodesStream, Start: int(start)}
		function.Decompile()
		e.Handlers[id] = function

		fmt.Printf("============ handler %d ============ \n%s", id, function.Decompiled)
	}

	utils.ReadBytes(&e.Matrix, b[:0x40])
	return e
}

func SCR_Entities(b []byte) interface{} {
	entities := &Entities{Array: make([]*Entity, 0)}

	for start := 0; start < len(b); {
		e := EntityFromBytes(b[start:])
		start += int(e.Size)

		entities.Array = append(entities.Array, e)
	}

	return entities
}

func init() {
	store.AddScriptLoader("SCR_Entities", SCR_Entities)
}
