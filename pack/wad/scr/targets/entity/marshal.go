package entity

import (
	"bytes"
	"encoding/binary"

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

func (e *Entity) marshalBuffer(ec *entitycontext.EntityLevelContext) *bytes.Buffer {
	m := marshaler{
		buf: &bytes.Buffer{},
	}

	m.write(utils.AsBytes(e.Matrix))
	//m.w32()

	return m.buf
}
