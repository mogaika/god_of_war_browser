package targets

import (
	"encoding/binary"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad/scr/store"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Entities struct {
	Array []*Entity
}

type Entity struct {
	Matrix mgl32.Mat4

	Field_0x40 uint16
	Field_0x44 uint16
	HandlerId  uint16
	Field_0x48 uint16
	Field_0x4a uint16
	Field_0x4c uint16
	Field_0x4e uint16
	Field_0x50 uint16
	Field_0x52 uint16

	Name string

	Strings []string
}

func EntityFromBytes(b []byte) *Entity {
	e := &Entity{
		Field_0x40: binary.LittleEndian.Uint16(b[0x40:]),
		Field_0x44: binary.LittleEndian.Uint16(b[0x44:]),
		HandlerId:  binary.LittleEndian.Uint16(b[0x46:]),
		Field_0x48: binary.LittleEndian.Uint16(b[0x48:]),
		Field_0x4a: binary.LittleEndian.Uint16(b[0x4a:]),
		Field_0x4c: binary.LittleEndian.Uint16(b[0x4c:]),
		Field_0x4e: binary.LittleEndian.Uint16(b[0x4e:]),
		Field_0x50: binary.LittleEndian.Uint16(b[0x50:]),
		Field_0x52: binary.LittleEndian.Uint16(b[0x52:]),
	}

	textStart := 0x54 + e.Field_0x52 + e.Field_0x4e*4 + e.Field_0x50*2
	e.Name = utils.BytesToString(b[textStart:])

	utils.ReadBytes(&e.Matrix, b[:0x40])

	utils.Dump(e.Name, e.HandlerId, b[0x54:textStart])

	return e
}

func SCR_Entities(b []byte) interface{} {
	entities := &Entities{Array: make([]*Entity, 0)}

	for start := 0; start < len(b); {
		e := EntityFromBytes(b[start:])
		start += int(e.Field_0x44)

		entities.Array = append(entities.Array, e)
	}

	return entities
}

func init() {
	store.AddScriptLoader("SCR_Entities", SCR_Entities)
}
