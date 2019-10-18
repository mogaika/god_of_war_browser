package mesh

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	MESH_GOW1_HEADER_SIZE   = 0x50
	PART_GOW1_HEADER_SIZE   = 4
	GROUP_GOW1_HEADER_SIZE  = 0xC
	OBJECT_GOW1_HEADER_SIZE = 0x20
	MESH_GOW1_VECTOR_SIZE   = 0x14
)

func (o *Object) parseGow1(allb []byte, pos uint32, size uint32, exlog *utils.Logger) error {
	b := allb[pos:]
	o.Offset = pos
	o.RawDmaAndJointsData = b[OBJECT_GOW1_HEADER_SIZE:size]

	o.Type = binary.LittleEndian.Uint16(b[0:])
	if o.Type != 0xe && o.Type != 0x1d {
		// 0xe - dynamic
		// 0x1d - static (ignores 0 joint?)
		return fmt.Errorf("Unknown type %x", o.Type)
	}
	o.Unk02 = binary.LittleEndian.Uint16(b[2:])
	o.PacketsPerFilter = binary.LittleEndian.Uint32(b[4:])
	o.MaterialId = binary.LittleEndian.Uint16(b[8:])
	if jmLen := binary.LittleEndian.Uint16(b[0xa:]); jmLen != 0 {
		o.JointMapper = make([]uint32, jmLen)
	}
	o.Unk0c = binary.LittleEndian.Uint32(b[0xc:])
	o.Unk10 = binary.LittleEndian.Uint32(b[0x10:])
	o.UseInvertedMatrix = o.Unk10&0x40 != 0
	o.Unk14 = binary.LittleEndian.Uint32(b[0x14:])
	o.TextureLayersCount = b[0x18]
	o.Unk19 = b[0x19]
	o.NextFreeVUBufferId = binary.LittleEndian.Uint16(b[0x1a:])
	o.Unk1c = binary.LittleEndian.Uint16(b[0x1c:])
	o.SourceVerticesCount = binary.LittleEndian.Uint16(b[0x1e:])

	exlog.Printf("        | type: 0x%.4x  unk02: 0x%.4x packets per filter?: %d materialId: %d joints: %d",
		o.Type, o.Unk02, o.PacketsPerFilter, o.MaterialId, len(o.JointMapper))
	exlog.Printf("        | unk0c: 0x%.8x unk10: 0x%.8x unk14: 0x%.8x textureLayers: %d unk19: 0x%.2x next free vu buffer: 0x%.4x unk1c: 0x%.4x source vertices count: 0x%.4x ",
		o.Unk0c, o.Unk10, o.Unk14, o.TextureLayersCount, o.Unk19, o.NextFreeVUBufferId, o.Unk1c, o.SourceVerticesCount)
	exlog.Printf("      --===--\n%v\n", utils.SDump(o.RawDmaAndJointsData))

	dmaCalls := o.Unk0c * uint32(o.TextureLayersCount)
	o.Packets = make([][]Packet, dmaCalls)
	for iDmaChain := uint32(0); iDmaChain < dmaCalls; iDmaChain++ {
		packetOffset := o.Offset + OBJECT_GOW1_HEADER_SIZE + iDmaChain*o.PacketsPerFilter*0x10
		exlog.Printf("        - packets %d offset 0x%.8x pps 0x%.8x", iDmaChain, packetOffset, o.PacketsPerFilter)

		ds := NewMeshParserStream(allb, o, packetOffset, exlog)
		if err := ds.ParsePackets(); err != nil {
			return err
		}
		verts := 0
		trias := 0

		for ip := range ds.Packets {
			p := &ds.Packets[ip]
			verts += len(p.Trias.Skip)
			for it := range p.Trias.Skip {
				if !p.Trias.Skip[it] {
					trias += 1
				}
			}
		}
		exlog.Printf("         - - - - - - - - -  trias %d (0x%x)  verts %d (0x%x)",
			trias, trias, verts, verts)
		o.Packets[iDmaChain] = ds.Packets
	}
	//exlog.Printf("%v\n", utils.SDump(o.Packets[0]))
	if o.JointMapper != nil {
		// right after dma calls
		jointMapOffset := OBJECT_GOW1_HEADER_SIZE + dmaCalls*0x10*o.PacketsPerFilter
		for i := range o.JointMapper {
			jid := binary.LittleEndian.Uint32(b[jointMapOffset+uint32(i)*4:])
			o.JointMapper[i] = jid
		}
		exlog.Printf("              - joint map: %+#v", o.JointMapper)
	}

	return nil
}

func (g *Group) parseGow1(allb []byte, pos uint32, size uint32, exlog *utils.Logger) error {
	b := allb[pos:]
	g.Offset = pos

	g.Unk00 = binary.LittleEndian.Uint32(b[0:])
	g.Objects = make([]Object, binary.LittleEndian.Uint32(b[4:]))
	g.Unk08 = binary.LittleEndian.Uint32(b[8:])
	exlog.Printf("      | unk00: 0x%.8x unk08: 0x%.8x objects count: %v", g.Unk00, g.Unk08, len(g.Objects))

	for i := range g.Objects {
		objectOffset := binary.LittleEndian.Uint32(b[GROUP_GOW1_HEADER_SIZE+i*4:])
		exlog.Printf(" - - - object %d offset 0x%.8x", i, pos+objectOffset)

		objectEnd := size
		if i != len(g.Objects)-1 {
			objectEnd = binary.LittleEndian.Uint32(b[GROUP_GOW1_HEADER_SIZE+i*4+4:])
		}

		if err := g.Objects[i].parseGow1(allb, pos+objectOffset, objectEnd-objectOffset, exlog); err != nil {
			return fmt.Errorf("Error when parsing object %d: %v", i, err)
		}
	}

	return nil
}

func (p *Part) parseGow1(allb []byte, pos uint32, size uint32, exlog *utils.Logger) error {
	b := allb[pos:]
	p.Offset = pos
	p.Unk00 = binary.LittleEndian.Uint16(b[:])
	p.Groups = make([]Group, binary.LittleEndian.Uint16(b[2:]))
	p.JointId = binary.LittleEndian.Uint16(b[len(p.Groups)*4+PART_GOW1_HEADER_SIZE:])
	exlog.Printf("    | unk00: 0x%.4x jointid: %d groups count: %v", p.Unk00, p.JointId, len(p.Groups))

	for i := range p.Groups {
		groupOffset := binary.LittleEndian.Uint32(b[PART_GOW1_HEADER_SIZE+i*4:])
		exlog.Printf(" - - group %d offset 0x%.8x", i, pos+groupOffset)

		groupEnd := size
		if i != len(p.Groups)-1 {
			groupEnd = binary.LittleEndian.Uint32(b[PART_GOW1_HEADER_SIZE+i*4+4:])
		}

		if err := p.Groups[i].parseGow1(allb, pos+groupOffset, groupEnd-groupOffset, exlog); err != nil {
			return fmt.Errorf("Error when parsing group %d: %v", i, err)
		}
	}
	return nil
}

func (v *Vector) parseGow1(allb []byte, pos uint32, exlog *utils.Logger) {
	b := allb[pos:]
	v.Unk00 = binary.LittleEndian.Uint16(b[0:])
	v.Unk02 = binary.LittleEndian.Uint16(b[2:])

	for i := range v.Value {
		v.Value[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[4+i*4:]))
	}

	exlog.Printf(" <- unk00: 0x%.4x unk02: 0x%.4x : %v", v.Unk00, v.Unk02, v.Value)
}

func (m *Mesh) parseGow1(b []byte, exlog *utils.Logger) error {
	if binary.LittleEndian.Uint32(b) != MESH_MAGIC {
		return fmt.Errorf("Invalid mesh magic")
	}

	// remove Comments, game don't use them anyway
	mdlCommentStart := binary.LittleEndian.Uint32(b[4:])
	b = b[:mdlCommentStart]

	m.Parts = make([]Part, binary.LittleEndian.Uint32(b[8:]))
	m.Unk0c = binary.LittleEndian.Uint32(b[0xc:])
	m.Unk10 = binary.LittleEndian.Uint32(b[0x10:])
	m.Unk14 = binary.LittleEndian.Uint32(b[0x14:])
	m.Vectors = make([]Vector, binary.LittleEndian.Uint32(b[0x18:]))
	m.Flags0x20 = binary.LittleEndian.Uint32(b[0x20:])
	m.Unk28 = binary.LittleEndian.Uint32(b[0x28:])
	m.Unk2c = binary.LittleEndian.Uint32(b[0x2c:])
	m.Unk30 = binary.LittleEndian.Uint32(b[0x30:])
	m.Unk34 = binary.LittleEndian.Uint32(b[0x34:])
	m.NameOfRootJoint = utils.BytesToString(b[0x38:0x50])

	exlog.Printf("unk0c 0x%.8x  unk10 0x%.8x  unk14 0x%.8x", m.Unk0c, m.Unk10, m.Unk14)
	exlog.Printf("root joint '%s' flags 0x%.8x", m.NameOfRootJoint, m.Flags0x20)
	exlog.Printf("unk28 0x%.8x  unk2c 0x%.8x  unk30 0x%.8x  unk34 0x%.8x", m.Unk28, m.Unk2c, m.Unk30, m.Unk34)

	vectorsStart := len(m.Parts)*4 + MESH_GOW1_HEADER_SIZE
	exlog.Printf(" - strange vectors starting at 0x%.8x count %d", vectorsStart, len(m.Vectors))
	for i := range m.Vectors {
		m.Vectors[i].parseGow1(b, uint32(vectorsStart+i*MESH_GOW1_VECTOR_SIZE), exlog)
	}

	for i := range m.Parts {
		partOffset := binary.LittleEndian.Uint32(b[MESH_GOW1_HEADER_SIZE+i*4:])
		exlog.Printf(" - part %d offset 0x%.8x", i, partOffset)

		partEnd := mdlCommentStart
		if i != len(m.Parts)-1 {
			partEnd = binary.LittleEndian.Uint32(b[MESH_GOW1_HEADER_SIZE+i*4+4:])
		}

		if err := m.Parts[i].parseGow1(b, partOffset, partEnd-partOffset, exlog); err != nil {
			return fmt.Errorf("Error when parsing part %d: %v", i, err)
		}
	}

	return nil
}
