package targets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/pkg/errors"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/scr/entitycontext"
	"github.com/mogaika/god_of_war_browser/pack/wad/scr/store"
	"github.com/mogaika/god_of_war_browser/utils"
)

type EntityType uint8

const (
	ENTITY_TYPE_ENTRY_SENSOR       EntityType = 0
	ENTITY_TYPE_EXIT_SENSOR        EntityType = 1
	ENTITY_TYPE_CREATION_SENSOR    EntityType = 2
	ENTITY_TYPE_DESTRUCTION_SENSOR EntityType = 3
	ENTITY_TYPE_EVENT_SENSOR       EntityType = 4
	ENTITY_TYPE_ANIMATOR           EntityType = 5
	ENTITY_TYPE_UNKNOWN_6          EntityType = 6
	ENTITY_TYPE_VIS                EntityType = 7
	ENTITY_TYPE_EVENT_TRANSMITTER  EntityType = 8
	ENTITY_TYPE_START              EntityType = 9
	ENTITY_TYPE_SPAWN_ENEMY        EntityType = 10
	ENTITY_TYPE_CREATOR            EntityType = 11
	ENTITY_TYPE_GLOBAL_DATA        EntityType = 12
	ENTITY_TYPE_LEVEL_DATA         EntityType = 13
	ENTITY_TYPE_MARKER             EntityType = 14
	ENTITY_TYPE_SOUND_CONTROLLER   EntityType = 15
)

const (
	VAR_TYPE_FLOAT = 0
	VAR_TYPE_INT   = 1
	VAR_TYPE_BOOL  = 2
)

const (
	SCOPE_ENTITY     = 0
	SCOPE_INTERNAL   = 1
	SCOPE_GLOBALDATA = 2
	SCOPE_LEVELDATA  = 3
)

type Entities struct {
	Array []*Entity
}

type EntityHandler struct {
	Stream     []byte `json:"-"`
	Start      int
	Decompiled string
}

func (o *EntityHandler) Decompile(ec *entitycontext.EntityLevelContext) {
	if o.Decompiled == "" {
		o.Decompiled = DecompileEscScript(o.Stream, o.Start, ec)
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

func GetEscFunc(scope, fid uint16) string {
	switch scope {
	case SCOPE_ENTITY:
		switch fid {
		case 0x08:
			return "Printf(format, a1,a2,a3,a4)"
		case 0x09, 0x0a, 0x0b, 0x0c:
			return "Get name or index of entity(index in arr[4])"
		case 0xd:
			return "PlayStreamedEntry? (string name)"

		case 0x10:
			return "PreLoadStreamedEntry? (unk, string name)"
		default:
			log.Printf("Unknown entity scope func %d", fid)
		}
	case SCOPE_INTERNAL:
		switch fid {
		case 0x0:
			return "CheckPoint ()"

		case 0x2:
			return "Load? (s1, s2)"
		case 0x3:
			return "LoadWad? (s1)"
		case 0x4:
			return "LoadForWarp? (s1)"
		case 0x5:
			return "Warp? ()"
		case 0x6:
			return "LoadCheck? (s1)"
		case 0x7:
			return "Goto? (s1)"
		case 0x8:
			return "Switch to camera??? (cameraName)"
		case 0x9:
			return "AbortCutscene? (bool)"
		case 0xa:
			return "Print text on screen(type, messageID)"

		case 0xc:
			return "Idle (bool needIdle)"

		case 0x10:
			return "Zone title report ??? (zone_index)"

		case 0x13:
			return "Trigger you have failed screen()"
		case 0x14:
			return "??CreateTimer??(fDuration): timer_id"

		case 0x17:
			return "??DestroyTimer??(timerId)"
		case 0x18:
			return "??PauseTimer_Or_TimerGetElapsedSeconds??(timerId): elapsed_sec"
		case 0x19:
			return "??StartTimer??(timerId)"

		case 0x1b:
			return "Start music group (flags?, soundName)"
		case 0x1c:
			return "Stop music group (flags?)"
		case 0x1d:
			return " music group related (int, int, float)"

		case 0x1f:
			return "Print text on screen(unkn, type, messageID)"

		case 0x2f:
			return "Complete Game"

		case 0x38:
			return "Play music or sound(sndname, unkn(repeat?))"
		default:
			log.Printf("Unknown internal scope func %d", fid)
		}
	case SCOPE_GLOBALDATA:
		switch fid {
		default:
			log.Printf("Unknown global scope func %d", fid)
		}
	case SCOPE_LEVELDATA:
		switch fid {
		default:
			log.Printf("Unknown level scope func %d", fid)
		}
	}
	return ""
}

func argsParseScopeFunc(b []byte) (scope, fid uint16) {
	args := binary.LittleEndian.Uint16(b)
	return args >> 12, args & 0xfff
}

func DecompileEscScript(b []byte, pointer int, ec *entitycontext.EntityLevelContext) string {
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

	startPointer := pointer

	for !fail {
		opcode := b[pointer]
		output.WriteString(fmt.Sprintf("0x%.4x: %.2X:  ", pointer-startPointer, opcode))

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
			pointer += 2
			var target string
			if scope < 8 {
				target = ScopeToString(scope)
			} else if scope == 8 {
				target = "Special local scope? Currnet obj scope? Related to behavior and evt. Using 0x29BCF0"
			} else if scope > 8 {
				target = "Array at 0x334A98 word[0x256]"
			}

			name := "unknown :("
			switch scope {
			case SCOPE_LEVELDATA:
				name = ec.LevelData[fid].Name
			case SCOPE_GLOBALDATA:
				name = ec.GlobalData[fid].Name
			}

			code := opcode - 0x2
			if code < 0x04 {
				wrline("get_scope_%s from (0x%x)'%s' val 0x%x name %q;", TypeIdToString(code), scope, target, fid, name)
			} else {
				wrline("set_scope_%s from (0x%x)'%s' val 0x%x name %q;", TypeIdToString(code-4), scope, target, fid, name)
			}
		case 0x0a, 0x0b, 0x0c, 0x0d:
			scope, fid := argsParseScopeFunc(opcodeBuf)
			wrline("call_function scope(0x%x), func(0x%x) '%s' => %s;", scope, fid, GetEscFunc(scope, fid), TypeIdToString(opcode-0x0a))
			pointer += 2
		case 0x0e:
			off := binary.LittleEndian.Uint16(opcodeBuf)
			wrline("push_string '%s';", utils.BytesToString(b[off:]))
			pointer += 2
		case 0x0f:
			off := int16(binary.LittleEndian.Uint16(opcodeBuf))
			pointer += 2
			wrline("pop_jmp_if_not_zero offset(+0x%x)=0x%x;", off, pointer+int(off)-startPointer)
		case 0x10:
			off := int16(binary.LittleEndian.Uint16(opcodeBuf))
			pointer += 2
			wrline("jump relative offset(+0x%x)=0x%x;", off, pointer+int(off)-startPointer)
		case 0x11:
			wrline("push_bool TRUE;")
		case 0x12:
			wrline("push_bool FALSE;")
		case 0x13:
			wrline("pop_int_push_change_sign;")
		case 0x14:
			wrline("pop_float_push_change_sign;")
		case 0x15:
			wrline("pop_pop_int_sum_push;")
		case 0x16:
			wrline("pop_pop_float_sum_push;")
		case 0x17:
			wrline("pop_pop_int_sub_push;")
		case 0x18:
			wrline("pop_pop_float_sub_push;")
		case 0x19:
			wrline("pop_pop_int_mul_push;")
		case 0x1a:
			wrline("pop_pop_float_mul_push;")
		case 0x1b:
			wrline("pop_pop_int_div_push;")
		case 0x1c:
			wrline("pop_pop_float_div_push;")
		case 0x1d:
			wrline("pop_int_pop_int_mod_push;")
		case 0x1e:
			wrline("pop_bool2float_push;") // if input != 0 ? 1.0 : 0.0
		case 0x1f:
			wrline("pop_bool2int_push;") // if input != 0 ? 1 : 0
		case 0x20:
			wrline("pop_float2bool_push;")
		case 0x21:
			wrline("pop_float2int_push;")
		case 0x22:
			wrline("pop_int2bool_push;")
		case 0x23:
			wrline("pop_int2float_push;")
		case 0x24:
			wrline("pop_pop_bool_push_bool_logical_not;")
		case 0x25:
			wrline("pop_pop_int_push_bool_logical_and;")
		case 0x26:
			wrline("pop_pop_int_push_bool_logical_or;")
		case 0x27:
			wrline("pop_pop_int_push_bool_logical_xor;")
		case 0x28:
			wrline("pop_pop_int_push_bool_if_less;")
		case 0x29:
			wrline("pop_pop_float_push_bool_if_less;")
		case 0x2a:
			wrline("pop_pop_strcmp_push_bool_if_less")
		case 0x2b:
			wrline("pop_pop_int_push_bool_if_less_or_equal;")
		case 0x2c:
			wrline("pop_pop_float_push_bool_if_less_or_equal;")
		case 0x2d:
			wrline("pop_pop_strcmp_push_bool_if_less_or_equal;")
		case 0x2e:
			wrline("pop_pop_int_push_bool_if_bigger;")
		case 0x2f:
			wrline("pop_pop_float_push_bool_if_bigger;")
		case 0x30:
			wrline("pop_pop_strcmp_push_bool_if_bigger;")
		case 0x31:
			wrline("pop_pop_int_push_bool_if_not_bigger_or_equal")
		case 0x32:
			wrline("pop_pop_float_push_bool_if_not_bigger_or_equal")
		case 0x33:
			wrline("pop_pop_strcmp_push_bool_if_not_bigger_or_equal;")
		case 0x34:
			wrline("pop_pop_int_push_bool_if_equal;")
		case 0x35:
			wrline("pop_pop_float_push_bool_if_equal;")
		case 0x36:
			wrline("pop_pop_bool_push_bool_if_equal;")
		case 0x37:
			wrline("pop_pop_strcmp_push_bool_if_equal;")
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

	Field_0x40           uint16
	Size                 uint16
	EntityType           EntityType
	EntityUniqueID       uint16
	VariableOffset       uint16
	variablesCount       uint16
	HandlersCount        uint16
	dependsEntitiesCount uint16
	opcodesStreamsSize   uint16

	DependsEntitiesIds   []uint16
	DependsEntitiesNames []string
	Name                 string
	Variables            []entitycontext.Variable
	Handlers             map[uint16]EntityHandler
}

func EntityFromBytes(b []byte, ec *entitycontext.EntityLevelContext) (*Entity, error) {
	e := &Entity{
		Field_0x40:           binary.LittleEndian.Uint16(b[0x40:]),
		Size:                 binary.LittleEndian.Uint16(b[0x44:]),
		EntityType:           EntityType(binary.LittleEndian.Uint16(b[0x46:])),
		EntityUniqueID:       binary.LittleEndian.Uint16(b[0x48:]),
		VariableOffset:       binary.LittleEndian.Uint16(b[0x4a:]),
		variablesCount:       binary.LittleEndian.Uint16(b[0x4c:]),
		HandlersCount:        binary.LittleEndian.Uint16(b[0x4e:]),
		dependsEntitiesCount: binary.LittleEndian.Uint16(b[0x50:]),
		opcodesStreamsSize:   binary.LittleEndian.Uint16(b[0x52:]),
		Handlers:             make(map[uint16]EntityHandler),
	}
	b = b[:e.Size]

	if false { // e.Field_0x40 != 0 {
		return nil, errors.Errorf("Field Field_0x40 = %v", e.Field_0x40)
	}

	textStart := int(0x54 + e.opcodesStreamsSize + e.HandlersCount*4 + e.dependsEntitiesCount*2)
	textPos := textStart

	readEntityString := func() string {
		result := utils.BytesToString(b[textPos:])
		textPos += utils.BytesStringLength(b[textPos:]) + 1
		return result
	}
	e.Name = readEntityString()

	switch e.EntityType {
	case ENTITY_TYPE_LEVEL_DATA, ENTITY_TYPE_GLOBAL_DATA:
		e.Variables = make([]entitycontext.Variable, e.variablesCount)
		for i := range e.Variables {
			e.Variables[i].Type = b[textPos]
			textPos++
			e.Variables[i].Name = readEntityString()
		}
	}

	switch e.EntityType {
	case ENTITY_TYPE_LEVEL_DATA:
		for i, v := range e.Variables {
			ec.LevelData[uint16(i)+e.VariableOffset] = v
		}
	case ENTITY_TYPE_GLOBAL_DATA:
		for i, v := range e.Variables {
			ec.GlobalData[uint16(i)+e.VariableOffset] = v
		}
	}

	opcodesDescrStart := uint16(0x54)

	e.DependsEntitiesIds = make([]uint16, e.dependsEntitiesCount)
	e.DependsEntitiesNames = make([]string, e.dependsEntitiesCount)
	for i := range e.DependsEntitiesIds {
		id := binary.LittleEndian.Uint16(b[opcodesDescrStart+e.HandlersCount*4+uint16(i*2):])
		e.DependsEntitiesIds[i] = id
		e.DependsEntitiesNames[i] = ec.EntityIdNameMap[id]
	}

	opcodesStreamStart := int(opcodesDescrStart + e.HandlersCount*4 + e.dependsEntitiesCount*2 + 2)

	opcodesStream := b[opcodesStreamStart:]

	for i := uint16(0); i < e.HandlersCount; i++ {
		start := binary.LittleEndian.Uint16(b[opcodesDescrStart+i*4+2:])

		id := binary.LittleEndian.Uint16(b[opcodesDescrStart+i*4:])

		function := EntityHandler{Stream: opcodesStream, Start: int(start)}
		function.Decompile(ec)
		e.Handlers[id] = function

		//fmt.Printf("============ handler %d ============ \n%s", id, function.Decompiled)
	}

	utils.ReadBytes(&e.Matrix, b[:0x40])
	return e, nil
}

var globlock sync.Mutex

func SCR_Entities(b []byte, wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	entities := &Entities{Array: make([]*Entity, 0)}

	globlock.Lock()
	defer globlock.Unlock()

	/*efname, _ := os.OpenFile("entityassoc.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	defer efname.Close()*/

	ec := wrsrc.Wad.GetEntityContext()

	for start := 0; start < len(b); {
		e, err := EntityFromBytes(b[start:], ec)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse entity %d: %v", len(entities.Array), err)
		}
		start += int(e.Size)

		/*fmt.Fprintf(efname, "t%.2d id%.3d %q hc%d dc%d\n",
		e.EntityType, e.EntityUniqueID, e.Name, len(e.Handlers), len(e.DependsEntitiesIds))*/

		entities.Array = append(entities.Array, e)
		ec.EntityIdNameMap[e.EntityUniqueID] = e.Name
	}

	return entities, nil
}

func init() {
	store.AddScriptLoader("SCR_Entities", SCR_Entities)
}
