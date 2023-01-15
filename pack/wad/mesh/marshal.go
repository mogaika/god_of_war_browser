package mesh

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/mogaika/god_of_war_browser/utils"
)

func (o *Object) Marshal() *bytes.Buffer {
	var buf [OBJECT_GOW1_HEADER_SIZE]byte

	binary.LittleEndian.PutUint16(buf[0:], o.Type)
	binary.LittleEndian.PutUint16(buf[2:], o.Unk02)
	binary.LittleEndian.PutUint32(buf[4:], o.DmaTagsCountPerPacket)
	binary.LittleEndian.PutUint16(buf[8:], o.MaterialId)
	binary.LittleEndian.PutUint16(buf[0xa:], o.JointMapElementsCount)
	binary.LittleEndian.PutUint32(buf[0xc:], o.InstancesCount)
	binary.LittleEndian.PutUint32(buf[0x10:], o.Flags)
	binary.LittleEndian.PutUint32(buf[0x14:], o.FlagsMask)
	buf[0x18] = o.TextureLayersCount
	buf[0x19] = o.TotalDmaProgramsCount
	binary.LittleEndian.PutUint16(buf[0x1a:], o.NextFreeVUBufferId)
	binary.LittleEndian.PutUint16(buf[0x1c:], o.Unk1c)
	binary.LittleEndian.PutUint16(buf[0x1e:], o.SourceVerticesCount)

	dmaAndJointsData := make([]byte, len(o.RawDmaAndJointsData))
	copy(dmaAndJointsData, o.RawDmaAndJointsData)
	jointMapsOffset := o.InstancesCount * uint32(o.TextureLayersCount) * 0x10 * o.DmaTagsCountPerPacket
	for iJm, jm := range o.JointMappers {
		jointMapOffset := jointMapsOffset + uint32(iJm)*uint32(o.JointMapElementsCount)*4
		for i := range jm {
			binary.LittleEndian.PutUint32(dmaAndJointsData[jointMapOffset+uint32(i)*4:], jm[i])
		}
	}

	var result bytes.Buffer
	result.Write(buf[:])
	result.Write(dmaAndJointsData)
	return &result
}

func (g *Group) Marshal(off uint32) *bytes.Buffer {
	var buf [GROUP_GOW1_HEADER_SIZE]byte
	binary.LittleEndian.PutUint32(buf[0:], math.Float32bits(g.HideDistance))
	binary.LittleEndian.PutUint32(buf[4:], uint32(len(g.Objects)))
	binary.LittleEndian.PutUint32(buf[8:], g.HasBbox)

	objectsOffset := uint32(len(g.Objects) * 4)

	needToPad := (objectsOffset + off + GROUP_GOW1_HEADER_SIZE) % 0x10
	if needToPad != 0 {
		objectsOffset += 0x10 - needToPad
	}

	offsetsBuf := make([]byte, objectsOffset)
	var dataStream bytes.Buffer
	for i := range g.Objects {
		binary.LittleEndian.PutUint32(offsetsBuf[i*4:], GROUP_GOW1_HEADER_SIZE+objectsOffset+uint32(dataStream.Len()))
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
	var buf [PART_GOW1_HEADER_SIZE]byte
	binary.LittleEndian.PutUint16(buf[0:], p.Unk00)
	binary.LittleEndian.PutUint16(buf[2:], uint16(len(p.Groups)))

	groupsOffset := uint32(len(p.Groups)*4) + 4
	offsetsBuf := make([]byte, groupsOffset)
	binary.LittleEndian.PutUint16(offsetsBuf[len(p.Groups)*4:], p.JointId)
	var dataStream bytes.Buffer
	for i := range p.Groups {
		groupOffset := PART_GOW1_HEADER_SIZE + groupsOffset + uint32(dataStream.Len())
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
	var buf [MESH_GOW1_VECTOR_SIZE]byte
	binary.LittleEndian.PutUint16(buf[0:], v.Unk00)
	binary.LittleEndian.PutUint16(buf[2:], v.Unk02)
	for i, f := range v.Value {
		binary.LittleEndian.PutUint32(buf[4+i*4:], math.Float32bits(f))
	}
	return buf[:]
}

func (m *Mesh) MarshalBuffer() *bytes.Buffer {
	var vectorsBuf bytes.Buffer
	for i := range m.Vectors {
		vectorsBuf.Write(m.Vectors[i].Marshal())
	}

	var blendJointsBuf bytes.Buffer
	for i := range m.BlendJoints {
		bJoint := &m.BlendJoints[i]
		binary.Write(&blendJointsBuf, binary.LittleEndian, uint32(len(bJoint.JointIds)))
		for i := range bJoint.JointIds {
			binary.Write(&blendJointsBuf, binary.LittleEndian, uint32(bJoint.JointIds[i]))
			binary.Write(&blendJointsBuf, binary.LittleEndian, float32(bJoint.Weights[i]))
		}
	}

	partOffsetsBuf := make([]byte, uint32(len(m.Parts)*4))
	partsOffset := MESH_GOW1_HEADER_SIZE + uint32(len(partOffsetsBuf)+vectorsBuf.Len()+blendJointsBuf.Len())

	var partsStream bytes.Buffer
	for i := range m.Parts {
		partOffset := partsOffset + uint32(partsStream.Len())
		binary.LittleEndian.PutUint32(partOffsetsBuf[i*4:], partOffset)
		partsStream.Write(m.Parts[i].Marshal(partOffset).Bytes())
	}

	var buf [MESH_GOW1_HEADER_SIZE]byte
	binary.LittleEndian.PutUint32(buf[0:], MESH_MAGIC)
	binary.LittleEndian.PutUint32(buf[4:], partsOffset+uint32(partsStream.Len()))
	binary.LittleEndian.PutUint32(buf[8:], uint32(len(m.Parts)))
	binary.LittleEndian.PutUint32(buf[0xc:], m.Unk0c)
	binary.LittleEndian.PutUint32(buf[0x10:], m.Unk10)
	binary.LittleEndian.PutUint32(buf[0x14:], m.Unk14)
	binary.LittleEndian.PutUint32(buf[0x18:], uint32(len(m.Vectors)))
	binary.LittleEndian.PutUint32(buf[0x20:], m.Flags0x20)
	binary.LittleEndian.PutUint32(buf[0x28:], m.SkeletJoints)
	binary.LittleEndian.PutUint32(buf[0x2c:], uint32(len(m.BlendJoints)))
	binary.LittleEndian.PutUint32(buf[0x30:], m.Unk30)
	binary.LittleEndian.PutUint32(buf[0x34:], m.BaseBoneIndex)
	copy(buf[0x38:0x50], utils.StringToBytesBuffer(m.NameOfRootJoint, 0x50-0x38, true))

	var result bytes.Buffer
	result.Write(buf[:])
	result.Write(partOffsetsBuf)
	result.Write(vectorsBuf.Bytes())
	result.Write(blendJointsBuf.Bytes())
	result.Write(partsStream.Bytes())
	return &result
}
