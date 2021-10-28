package targets

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/mogaika/god_of_war_browser/pack/wad/scr/entitycontext"
	"github.com/mogaika/god_of_war_browser/scriptlang"
	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	SCOPE_ENTITY     = 0
	SCOPE_INTERNAL   = 1
	SCOPE_GLOBALDATA = 2
	SCOPE_LEVELDATA  = 3
)

type EntityHandler struct {
	Id         uint16
	Data       []scriptlang.Instruction `json:"-"`
	Decompiled []string
}

func ScopeToString(scope uint16) string {
	if scope < 4 {
		return []string{"Entity", "Internal", "GlobalData", "LevelData"}[scope]
	} else {
		return fmt.Sprintf("*unknown scope 0x%x*", scope)
	}
}

var scopeInernalVariables = map[uint16]entitycontext.Variable{
	44: entitycontext.Variable{Type: VAR_TYPE_INT, Name: "PlayerHealthShardsCount"},
	45: entitycontext.Variable{Type: VAR_TYPE_INT, Name: "PlayerMagicShardsCount"},
	48: entitycontext.Variable{Type: VAR_TYPE_INT, Name: "DifficultyLevel"},
	50: entitycontext.Variable{Type: VAR_TYPE_INT, Name: "BladesLevel"},
	57: entitycontext.Variable{Type: VAR_TYPE_INT, Name: "PlayerSirenPiecesCount"},
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
			//log.Printf("Unknown entity scope func %d", fid)
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
			//log.Printf("Unknown internal scope func %d", fid)
		}
	case SCOPE_GLOBALDATA:
		switch fid {
		default:
			//log.Printf("Unknown global scope func %d", fid)
		}
	case SCOPE_LEVELDATA:
		switch fid {
		default:
			//log.Printf("Unknown level scope func %d", fid)
		}
	}
	return ""
}

func argsParseScopeFunc(b []byte) (scope, fid uint16) {
	args := binary.LittleEndian.Uint16(b)
	return args >> 12, args & 0xfff
}

func (eh *EntityHandler) parseOpcodes(b []byte, pointer int, ec *entitycontext.EntityLevelContext) {
	startPointer := pointer

	eh.Data = make([]scriptlang.Instruction, 0)
	labels := make(map[int16]*scriptlang.Label)
	opOffsets := make(map[int16]*scriptlang.Opcode)
	labelNamesGenerator := utils.RandomNameGenerator{}

	jmpOffsetToLabel := func(jmpOff int16) *scriptlang.Label {
		// log.Printf("JMP")
		targetOffset := int16(pointer + int(jmpOff) - startPointer)

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

	for {
		op := &scriptlang.Opcode{
			Code: b[pointer],
		}
		opOffsets[int16(pointer-startPointer)] = op
		eh.Data = append(eh.Data, op)

		pointer++

		if op.Code >= 0x3a {
			op.Comment = "exit"
			break
		}

		opcodeBuf := b[pointer:]

		switch op.Code {
		case 0x00:
			a1 := math.Float32frombits(binary.LittleEndian.Uint32(opcodeBuf))
			op.Comment = fmt.Sprintf("push_float %f", a1)
			op.AddParameters(a1)
			pointer += 4
		case 0x01:
			a1 := int32(binary.LittleEndian.Uint32(opcodeBuf))
			op.Comment = fmt.Sprintf("push_int %d", a1)
			op.AddParameters(a1)
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

			ok := false
			var v entitycontext.Variable
			switch scope {
			case SCOPE_LEVELDATA:
				v, ok = ec.LevelData[fid]
			case SCOPE_GLOBALDATA:
				v, ok = ec.GlobalData[fid]
			case SCOPE_INTERNAL:
				v, ok = scopeInernalVariables[fid]
			}
			var fidParam interface{} = int32(fid)
			if ok {
				fidParam = v.Name
			}

			code := op.Code - 0x2
			if code < 0x04 {
				op.Comment = fmt.Sprintf("get_scope_%s from (0x%x)'%s' val 0x%x name %q", TypeIdToString(code), scope, target, fid, v.Name)
			} else {
				op.Comment = fmt.Sprintf("set_scope_%s from (0x%x)'%s' val 0x%x name %q", TypeIdToString(code-4), scope, target, fid, v.Name)
			}
			op.AddParameters(int32(scope), fidParam)
		case 0x0a, 0x0b, 0x0c, 0x0d:
			scope, fid := argsParseScopeFunc(opcodeBuf)
			op.Comment = fmt.Sprintf("call_function scope(0x%x), func(0x%x) '%s' => %s",
				scope, fid, GetEscFunc(scope, fid), TypeIdToString(op.Code-0x0a))
			pointer += 2
			op.AddParameters(int32(scope), int32(fid))
		case 0x0e:
			a1 := utils.BytesToString(b[binary.LittleEndian.Uint16(opcodeBuf):])
			op.Comment = fmt.Sprintf("push_string '%s'", a1)
			pointer += 2
			op.AddParameters(a1)
		case 0x0f:
			off := int16(binary.LittleEndian.Uint16(opcodeBuf))
			pointer += 2
			op.Comment = fmt.Sprintf("pop_jmp_if_not_zero offset(+0x%x)=0x%x", off, pointer+int(off)-startPointer)
			op.AddParameters(jmpOffsetToLabel(off))
		case 0x10:
			off := int16(binary.LittleEndian.Uint16(opcodeBuf))
			pointer += 2
			op.Comment = fmt.Sprintf("jump relative offset(+0x%x)=0x%x", off, pointer+int(off)-startPointer)
			op.AddParameters(jmpOffsetToLabel(off))
		case 0x11:
			op.Comment = fmt.Sprintf("push_bool TRUE")
		case 0x12:
			op.Comment = fmt.Sprintf("push_bool FALSE")
		case 0x13:
			op.Comment = fmt.Sprintf("pop_int_push_change_sign")
		case 0x14:
			op.Comment = fmt.Sprintf("pop_float_push_change_sign")
		case 0x15:
			op.Comment = fmt.Sprintf("pop_pop_int_sum_push")
		case 0x16:
			op.Comment = fmt.Sprintf("pop_pop_float_sum_push")
		case 0x17:
			op.Comment = fmt.Sprintf("pop_pop_int_sub_push")
		case 0x18:
			op.Comment = fmt.Sprintf("pop_pop_float_sub_push")
		case 0x19:
			op.Comment = fmt.Sprintf("pop_pop_int_mul_push")
		case 0x1a:
			op.Comment = fmt.Sprintf("pop_pop_float_mul_push")
		case 0x1b:
			op.Comment = fmt.Sprintf("pop_pop_int_div_push")
		case 0x1c:
			op.Comment = fmt.Sprintf("pop_pop_float_div_push")
		case 0x1d:
			op.Comment = fmt.Sprintf("pop_int_pop_int_mod_push")
		case 0x1e:
			op.Comment = fmt.Sprintf("pop_bool2float_push") // if input != 0 ? 1.0 : 0.0
		case 0x1f:
			op.Comment = fmt.Sprintf("pop_bool2int_push") // if input != 0 ? 1 : 0
		case 0x20:
			op.Comment = fmt.Sprintf("pop_float2bool_push")
		case 0x21:
			op.Comment = fmt.Sprintf("pop_float2int_push")
		case 0x22:
			op.Comment = fmt.Sprintf("pop_int2bool_push")
		case 0x23:
			op.Comment = fmt.Sprintf("pop_int2float_push")
		case 0x24:
			op.Comment = fmt.Sprintf("pop_pop_bool_push_bool_logical_not")
		case 0x25:
			op.Comment = fmt.Sprintf("pop_pop_int_push_bool_logical_and")
		case 0x26:
			op.Comment = fmt.Sprintf("pop_pop_int_push_bool_logical_or")
		case 0x27:
			op.Comment = fmt.Sprintf("pop_pop_int_push_bool_logical_xor")
		case 0x28:
			op.Comment = fmt.Sprintf("pop_pop_int_push_bool_if_less")
		case 0x29:
			op.Comment = fmt.Sprintf("pop_pop_float_push_bool_if_less")
		case 0x2a:
			op.Comment = fmt.Sprintf("pop_pop_strcmp_push_bool_if_less")
		case 0x2b:
			op.Comment = fmt.Sprintf("pop_pop_int_push_bool_if_less_or_equal")
		case 0x2c:
			op.Comment = fmt.Sprintf("pop_pop_float_push_bool_if_less_or_equal")
		case 0x2d:
			op.Comment = fmt.Sprintf("pop_pop_strcmp_push_bool_if_less_or_equal")
		case 0x2e:
			op.Comment = fmt.Sprintf("pop_pop_int_push_bool_if_bigger")
		case 0x2f:
			op.Comment = fmt.Sprintf("pop_pop_float_push_bool_if_bigger")
		case 0x30:
			op.Comment = fmt.Sprintf("pop_pop_strcmp_push_bool_if_bigger")
		case 0x31:
			op.Comment = fmt.Sprintf("pop_pop_int_push_bool_if_not_bigger_or_equal")
		case 0x32:
			op.Comment = fmt.Sprintf("pop_pop_float_push_bool_if_not_bigger_or_equal")
		case 0x33:
			op.Comment = fmt.Sprintf("pop_pop_strcmp_push_bool_if_not_bigger_or_equal")
		case 0x34:
			op.Comment = fmt.Sprintf("pop_pop_int_push_bool_if_equal")
		case 0x35:
			op.Comment = fmt.Sprintf("pop_pop_float_push_bool_if_equal")
		case 0x36:
			op.Comment = fmt.Sprintf("pop_pop_bool_push_bool_if_equal")
		case 0x37:
			op.Comment = fmt.Sprintf("pop_pop_strcmp_push_bool_if_equal")
		case 0x38:
			op.Comment = fmt.Sprintf("pop_result")
		default:
			log.Panicf("Unknown opcode 0x%x", op.Code)
		}
	}

	for labelOffset, label := range labels {
		labelOp, ok := opOffsets[labelOffset]
		if !ok {
			panic("failed to find label for opcode")
		}

		eh.Data = label.InsertBeforeOpcode(eh.Data, labelOp)
	}

	eh.Decompiled = scriptlang.RenderScriptLines(eh.Data)

}
