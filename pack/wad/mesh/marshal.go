package mesh

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/mogaika/god_of_war_browser/utils"
)

func (o *Object) Marshal() *bytes.Buffer {
	var buf [OBJECT_HEADER_SIZE]byte

	binary.LittleEndian.PutUint16(buf[0:], o.Type)
	binary.LittleEndian.PutUint16(buf[2:], o.Unk02)
	binary.LittleEndian.PutUint32(buf[4:], o.PacketsPerFilter)
	binary.LittleEndian.PutUint16(buf[8:], o.MaterialId)

	if o.JointMapper == nil {
		binary.LittleEndian.PutUint16(buf[0xa:], 0)
	} else {
		binary.LittleEndian.PutUint16(buf[0xa:], uint16(len(o.JointMapper)))
	}

	binary.LittleEndian.PutUint32(buf[0xc:], o.Unk0c)
	binary.LittleEndian.PutUint32(buf[0x10:], o.Unk10)
	binary.LittleEndian.PutUint32(buf[0x14:], o.Unk14)
	buf[0x18] = o.TextureLayersCount
	buf[0x19] = o.Unk19
	binary.LittleEndian.PutUint16(buf[0x1a:], o.NextFreeVUBufferId)
	binary.LittleEndian.PutUint16(buf[0x1c:], o.Unk1c)
	binary.LittleEndian.PutUint16(buf[0x1e:], o.SourceVerticesCount)

	var result bytes.Buffer
	result.Write(buf[:])
	result.Write(o.RawDmaAndJointsData)
	return &result
}

func (g *Group) Marshal(off uint32) *bytes.Buffer {
	var buf [GROUP_HEADER_SIZE]byte
	binary.LittleEndian.PutUint32(buf[0:], g.Unk00)
	binary.LittleEndian.PutUint32(buf[4:], uint32(len(g.Objects)))
	binary.LittleEndian.PutUint32(buf[8:], g.Unk08)

	objectsOffset := uint32(len(g.Objects) * 4)

	needToPad := (objectsOffset + off + GROUP_HEADER_SIZE) % 0x10
	if needToPad != 0 {
		objectsOffset += 0x10 - needToPad
	}

	offsetsBuf := make([]byte, objectsOffset)
	var dataStream bytes.Buffer
	for i := range g.Objects {
		binary.LittleEndian.PutUint32(offsetsBuf[i*4:], GROUP_HEADER_SIZE+objectsOffset+uint32(dataStream.Len()))
		dataStream.Write(g.Objects[i].Marshal().Bytes())

		needToPad := dataStream.Len() % 0x10
		if needToPad != 0 {
			dataStream.Write(make([]byte, 0x10-needToPad))
		}
	}

	var result bytes.Buffer
	result.Write(buf[:])
	result.Write(offsetsBuf)
	result.Write(dataStream.Bytes())
	return &result
}

func (p *Part) Marshal(off uint32) *bytes.Buffer {
	var buf [PART_HEADER_SIZE]byte
	binary.LittleEndian.PutUint16(buf[0:], p.Unk00)
	binary.LittleEndian.PutUint16(buf[2:], uint16(len(p.Groups)))

	groupsOffset := uint32(len(p.Groups)*4) + 4
	offsetsBuf := make([]byte, groupsOffset)
	binary.LittleEndian.PutUint16(offsetsBuf[len(p.Groups)*4:], p.JointId)
	var dataStream bytes.Buffer
	for i := range p.Groups {
		groupOffset := PART_HEADER_SIZE + groupsOffset + uint32(dataStream.Len())
		binary.LittleEndian.PutUint32(offsetsBuf[i*4:], groupOffset)
		dataStream.Write(p.Groups[i].Marshal(off + groupOffset).Bytes())
	}

	var result bytes.Buffer
	result.Write(buf[:])
	result.Write(offsetsBuf)
	result.Write(dataStream.Bytes())
	return &result
}

func (v *Vector) Marshal() []byte {
	var buf [MESH_VECTOR_SIZE]byte
	binary.LittleEndian.PutUint16(buf[0:], v.Unk00)
	binary.LittleEndian.PutUint16(buf[2:], v.Unk02)
	for i, f := range v.Value {
		binary.LittleEndian.PutUint32(buf[4+i*4:], math.Float32bits(f))
	}
	return buf[:]
}

func (m *Mesh) MarshalBuffer() *bytes.Buffer {
	offsetsBuf := make([]byte, uint32(len(m.Parts)*4))
	partsOffset := uint32(len(offsetsBuf) + len(m.Vectors)*MESH_VECTOR_SIZE)
	var partsStream bytes.Buffer
	for i := range m.Parts {
		partOffset := MESH_HEADER_SIZE + partsOffset + uint32(partsStream.Len())
		binary.LittleEndian.PutUint32(offsetsBuf[i*4:], partOffset)
		partsStream.Write(m.Parts[i].Marshal(partOffset).Bytes())
	}

	var vectorsBuf bytes.Buffer
	for i := range m.Vectors {
		vectorsBuf.Write(m.Vectors[i].Marshal())
	}

	var buf [MESH_HEADER_SIZE]byte
	binary.LittleEndian.PutUint32(buf[0:], MESH_MAGIC)
	binary.LittleEndian.PutUint32(buf[4:], MESH_HEADER_SIZE+partsOffset+uint32(partsStream.Len()))
	binary.LittleEndian.PutUint32(buf[8:], uint32(len(m.Parts)))
	binary.LittleEndian.PutUint32(buf[0xc:], m.Unk0c)
	binary.LittleEndian.PutUint32(buf[0x10:], m.Unk10)
	binary.LittleEndian.PutUint32(buf[0x14:], m.Unk14)
	binary.LittleEndian.PutUint32(buf[0x18:], uint32(len(m.Vectors)))
	binary.LittleEndian.PutUint32(buf[0x20:], m.Flags0x20)
	binary.LittleEndian.PutUint32(buf[0x28:], m.Unk28)
	binary.LittleEndian.PutUint32(buf[0x2c:], m.Unk2c)
	binary.LittleEndian.PutUint32(buf[0x30:], m.Unk30)
	binary.LittleEndian.PutUint32(buf[0x34:], m.Unk34)
	copy(buf[0x38:0x50], utils.StringToBytesBuffer(m.NameOfRootJoint, 0x50-0x38, true))

	var result bytes.Buffer
	result.Write(buf[:])
	result.Write(offsetsBuf)
	result.Write(vectorsBuf.Bytes())
	result.Write(partsStream.Bytes())
	return &result
}
