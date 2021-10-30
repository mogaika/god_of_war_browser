package entity

import (
	"bytes"
	"encoding/binary"
	"log"
	"strings"

	"github.com/mogaika/god_of_war_browser/scriptlang"

	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/mogaika/god_of_war_browser/pack/wad/scr/entitycontext"
)

type marshaler struct {
	buf *bytes.Buffer
}

func (m *marshaler) write(data []byte) {
	m.buf.Write(data)
}

func (m *marshaler) w16(v uint16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	m.buf.Write(buf[:])
}

func (m *marshaler) w32(v uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	m.buf.Write(buf[:])
}

func (m *marshaler) pos() int {
	return m.buf.Len()
}

func (m *marshaler) skip(count int) {
	m.buf.Write(make([]byte, count))
}

func (e *Entity) compileScripts() {
	for i := range e.Handlers {
		if data, err := scriptlang.ParseScript([]byte(strings.Join(e.Handlers[i].Decompiled, "\n"))); err != nil {
			log.Panicf("Failed to decompile script: %v", err)
		} else {
			e.Handlers[i].Data = data
		}
	}
}

func (e *Entity) marshalScriptStreams(ec *entitycontext.EntityLevelContext) (result []byte, offsets []uint16) {
	e.compileScripts()

	var stream bytes.Buffer
	stringFills := make(map[int]string)

	for _, eh := range e.Handlers {
		offset := stream.Len()
		data, hStringFills := eh.Compile(ec)

		stream.Write(data)
		for off, s := range hStringFills {
			stringFills[off+offset] = s
		}
		offsets = append(offsets, uint16(offset))
	}

	stringOffsets := make(map[string]uint16)
	for _, s := range stringFills {
		stringOffsets[s] = uint16(stream.Len())
		stream.Write(utils.StringToBytes(s, true))
	}

	result = stream.Bytes()
	for offset, s := range stringFills {
		binary.LittleEndian.PutUint16(result[offset:], stringOffsets[s])
	}

	return result, offsets
}

func (e *Entity) marshalBuffer(ec *entitycontext.EntityLevelContext) []byte {
	m := marshaler{
		buf: &bytes.Buffer{},
	}

	var variablesMap map[uint16]entitycontext.Variable
	switch e.EntityType {
	case ENTITY_TYPE_LEVEL_DATA:
		variablesMap = ec.LevelData
	case ENTITY_TYPE_GLOBAL_DATA:
		variablesMap = ec.GlobalData
	}
	for i, v := range e.Variables {
		variablesMap[e.PhysicsObjectId+uint16(i)] = v
	}

	handlerStream, handlerOffsets := e.marshalScriptStreams(ec)

	m.write(utils.AsBytes(e.Matrix))
	m.w16(e.Field_0x40) // 0x40
	m.w16(e.Field_0x42) // not used
	m.skip(2)           // size, will be written later
	m.w16(uint16(e.EntityType))
	m.w16(e.EntityUniqueID) // 0x48
	m.w16(e.PhysicsObjectId)
	m.w16(e.Field_0x4C)
	m.w16(uint16(len(e.Handlers)))
	m.w16(uint16(len(e.TargetEntitiesIds))) // 0x50
	m.w16(uint16(len(handlerStream)) + 2)   // 0x52 op stream size

	for i, eh := range e.Handlers {
		m.w16(eh.Id)
		m.w16(handlerOffsets[i])
	}

	for _, id := range e.TargetEntitiesIds {
		m.w16(id)
	}
	m.w16(0)
	m.write(handlerStream)

	m.write(utils.StringToBytes(e.Name, true))

	for _, v := range e.Variables {
		m.buf.WriteByte(v.Type)
		m.write(utils.StringToBytes(v.Name, true))
	}

	if m.buf.Len()%4 != 0 {
		m.buf.Write(make([]byte, 4-(m.buf.Len()%4)))
	}

	result := m.buf.Bytes()
	binary.LittleEndian.PutUint16(result[0x44:], uint16(len(result)))

	return result
}
