package targets

import (
	"encoding/binary"
	"fmt"
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

type Entities struct {
	Array []*Entity
}

func TypeIdToString(typeid uint8) string {
	if typeid < 4 {
		return []string{"float", "int", "bool", "string"}[typeid]
	} else {
		return fmt.Sprintf("*unknown type 0x%x*", typeid)
	}
}

type Entity struct {
	Matrix mgl32.Mat4

	Field_0x40        uint16
	Size              uint16
	EntityType        EntityType
	EntityUniqueID    uint16
	VariableOffset    uint16
	TargetEntitiesIds []uint16
	Name              string
	Variables         []entitycontext.Variable
	Handlers          []EntityHandler

	DebugTargetEntitiesNames []string
	DebugReferencedBy        []uint16
	DebugReferencedByNames   []string
}

func EntityFromBytes(b []byte, ec *entitycontext.EntityLevelContext) (*Entity, error) {
	e := &Entity{
		Field_0x40:     binary.LittleEndian.Uint16(b[0x40:]),
		Size:           binary.LittleEndian.Uint16(b[0x44:]),
		EntityType:     EntityType(binary.LittleEndian.Uint16(b[0x46:])),
		EntityUniqueID: binary.LittleEndian.Uint16(b[0x48:]),
		VariableOffset: binary.LittleEndian.Uint16(b[0x4a:]),
		Handlers:       make([]EntityHandler, 0),
	}

	variablesCount := binary.LittleEndian.Uint16(b[0x4c:])
	handlersCount := binary.LittleEndian.Uint16(b[0x4e:])
	targetEntitiesCount := binary.LittleEndian.Uint16(b[0x50:])
	opcodesStreamsSize := binary.LittleEndian.Uint16(b[0x52:])

	b = b[:e.Size]

	if false { // e.Field_0x40 != 0 {
		return nil, errors.Errorf("Field Field_0x40 = %v", e.Field_0x40)
	}

	textStart := int(0x54 + opcodesStreamsSize + handlersCount*4 + targetEntitiesCount*2)
	textPos := textStart

	readEntityString := func() string {
		result := utils.BytesToString(b[textPos:])
		textPos += utils.BytesStringLength(b[textPos:]) + 1
		return result
	}
	e.Name = readEntityString()

	switch e.EntityType {
	case ENTITY_TYPE_LEVEL_DATA, ENTITY_TYPE_GLOBAL_DATA:
		e.Variables = make([]entitycontext.Variable, variablesCount)
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

	e.TargetEntitiesIds = make([]uint16, targetEntitiesCount)
	e.DebugTargetEntitiesNames = make([]string, targetEntitiesCount)
	for i := range e.TargetEntitiesIds {
		id := binary.LittleEndian.Uint16(b[opcodesDescrStart+handlersCount*4+uint16(i*2):])
		e.TargetEntitiesIds[i] = id
		e.DebugTargetEntitiesNames[i] = ec.EntityIdNameMap[id]
	}

	opcodesStreamStart := int(opcodesDescrStart + handlersCount*4 + targetEntitiesCount*2 + 2)

	opcodesStream := b[opcodesStreamStart:]

	for i := uint16(0); i < handlersCount; i++ {
		start := binary.LittleEndian.Uint16(b[opcodesDescrStart+i*4+2:])

		id := binary.LittleEndian.Uint16(b[opcodesDescrStart+i*4:])

		h := EntityHandler{Id: id}
		h.parseOpcodes(opcodesStream, int(start), ec)
		e.Handlers = append(e.Handlers, h)

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
		e.EntityType, e.EntityUniqueID, e.Name, len(e.Handlers), len(e.TargetEntitiesIds))*/

		entities.Array = append(entities.Array, e)
		if _, ok := ec.EntityIdNameMap[e.EntityUniqueID]; !ok {
			ec.EntityIdNameMap[e.EntityUniqueID] = e.Name
			for _, id := range e.TargetEntitiesIds {
				ec.EntityReferences[id] = append(ec.EntityReferences[id], e.EntityUniqueID)
			}
		}

		e.DebugReferencedBy = ec.EntityReferences[e.EntityUniqueID]
		if e.DebugReferencedBy == nil {
			e.DebugReferencedBy = make([]uint16, 0)
		}
		e.DebugReferencedByNames = make([]string, len(e.DebugReferencedBy))
		for i, id := range e.DebugReferencedBy {
			e.DebugReferencedByNames[i] = ec.EntityIdNameMap[id]
		}
	}

	return entities, nil
}

func init() {
	store.AddScriptLoader("SCR_Entities", SCR_Entities)
}
