package entity

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"

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

	Field_0x40      uint16 // in {0, 1, 2, 3, 4, 8, 16, 40}
	Field_0x42      uint16 // not used
	EntityType      EntityType
	EntityUniqueID  uint16
	PhysicsObjectId uint16

	// == values used at least once depending on entity type:
	// et  2: 3 2 1 5 4
	// et  3: 3 4
	// et  4: 1 2
	// et  5: 1 3
	// et  7: 1 4
	// et  8: 1 3 4 5 2
	// et  9: 2 3 1
	// et 12: used as variables count
	// et 13: used as variables count
	// et 15: 3 1 9
	// other: 1
	// == entity notice
	// 2: WarpFrom, MusicGroupStart, Messages
	// 3: Destruction, Kills, Attack
	// 4: Player/enemy death volume
	Field_0x4C        uint16
	TargetEntitiesIds []uint16
	Name              string
	Variables         []entitycontext.Variable
	Handlers          []EntityHandler

	DebugTargetEntitiesNames []string
	DebugReferencedBy        []uint16
	DebugReferencedByNames   []string
}

func EntityFromBytes(b []byte, ec *entitycontext.EntityLevelContext) (*Entity, int, error) {
	e := &Entity{Handlers: make([]EntityHandler, 0)}

	utils.ReadBytes(&e.Matrix, b[0x00:0x40])
	e.Field_0x40 = binary.LittleEndian.Uint16(b[0x40:])
	e.Field_0x42 = binary.LittleEndian.Uint16(b[0x42:]) // 0x42 - unused
	b = b[:binary.LittleEndian.Uint16(b[0x44:])]
	e.EntityType = EntityType(binary.LittleEndian.Uint16(b[0x46:]))
	e.EntityUniqueID = binary.LittleEndian.Uint16(b[0x48:])
	e.PhysicsObjectId = binary.LittleEndian.Uint16(b[0x4a:])
	e.Field_0x4C = binary.LittleEndian.Uint16(b[0x4c:])
	handlersCount := binary.LittleEndian.Uint16(b[0x4e:])
	targetEntitiesCount := binary.LittleEndian.Uint16(b[0x50:])
	opcodesStreamsSize := binary.LittleEndian.Uint16(b[0x52:])

	const handlersDescrStart = uint16(0x54)
	targetEntitiesStart := handlersDescrStart + handlersCount*4
	opcodesStreamStart := int(targetEntitiesStart + targetEntitiesCount*2 + 2)
	textStart := int(handlersDescrStart + handlersCount*4 + targetEntitiesCount*2 + opcodesStreamsSize)

	textPos := textStart
	readEntityString := func() string {
		result := utils.BytesToString(b[textPos:])
		textPos += utils.BytesStringLength(b[textPos:]) + 1
		return result
	}
	e.Name = readEntityString()

	if v := binary.LittleEndian.Uint16(b[targetEntitiesStart+targetEntitiesCount*2:]); v != 0 {
		// only happens in PAND05A.WAD:ESC_SNDGrpLeverlPull with value 0x01
		log.Printf("targetEntitiesStart+targetEntitiesCount*2 = 0x%x %q", v, e.Name)
	}

	switch e.EntityType {
	case ENTITY_TYPE_LEVEL_DATA, ENTITY_TYPE_GLOBAL_DATA:
		e.Variables = make([]entitycontext.Variable, e.Field_0x4C)
		for i := range e.Variables {
			e.Variables[i].Type = b[textPos]
			textPos++
			e.Variables[i].Name = readEntityString()
		}
	}

	switch e.EntityType {
	case ENTITY_TYPE_LEVEL_DATA:
		for i, v := range e.Variables {
			ec.LevelData[uint16(i)+e.PhysicsObjectId] = v
		}
	case ENTITY_TYPE_GLOBAL_DATA:
		for i, v := range e.Variables {
			ec.GlobalData[uint16(i)+e.PhysicsObjectId] = v
		}
	}

	e.TargetEntitiesIds = make([]uint16, targetEntitiesCount)
	e.DebugTargetEntitiesNames = make([]string, targetEntitiesCount)
	for i := range e.TargetEntitiesIds {
		id := binary.LittleEndian.Uint16(b[targetEntitiesStart+uint16(i*2):])
		e.TargetEntitiesIds[i] = id
		e.DebugTargetEntitiesNames[i] = ec.EntityIdNameMap[id]
	}

	for i := uint16(0); i < handlersCount; i++ {
		id := binary.LittleEndian.Uint16(b[handlersDescrStart+i*4:])
		start := binary.LittleEndian.Uint16(b[handlersDescrStart+i*4+2:])

		h := EntityHandler{Id: id}
		h.parseOpcodes(b[opcodesStreamStart:], int(start), ec)
		e.Handlers = append(e.Handlers, h)
	}

	/*
		if e.EntityType != ENTITY_TYPE_LEVEL_DATA && e.EntityType != ENTITY_TYPE_GLOBAL_DATA {
			if e.Field_0x40 != 0 {
				log.Printf("%d %.2d %s %v", e.Field_0x40, e.EntityType, e.Name, e.TargetEntitiesIds)
			}
		}
	*/

	cmpld := e.marshalBuffer(ec)
	if bytes.Compare(b, cmpld) != 0 {
		utils.LogDump(b)
		utils.LogDump(cmpld)
	}

	return e, len(b), nil
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
		e, size, err := EntityFromBytes(b[start:], ec)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse entity %d: %v", len(entities.Array), err)
		}
		start += size

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

func (ents *Entities) FromJSON(wrsrc *wad.WadNodeRsrc, data []byte) ([]byte, error) {
	ec := wrsrc.Wad.GetEntityContext()

	if err := json.Unmarshal(data, ents); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal")
	}

	var buf bytes.Buffer
	for _, e := range ents.Array {
		buf.Write(e.marshalBuffer(ec))
	}

	return buf.Bytes(), nil
}

func init() {
	store.AddScriptLoader("SCR_Entities", SCR_Entities)
}
