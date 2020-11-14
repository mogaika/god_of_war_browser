package mesh

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	MESH_GOW2_HEADER_SIZE   = 0x18
	PART_GOW2_HEADER_SIZE   = 4
	GROUP_GOW2_HEADER_SIZE  = 0x8
	OBJECT_GOW2_HEADER_SIZE = 0x20
	MESH_GOW2_VECTOR_SIZE   = 0x14
)

func (o *Object) parseGow2(allb []byte, pos uint32, size uint32, exlog *utils.Logger) error {
	b := allb[pos:]
	o.Offset = pos
	o.RawDmaAndJointsData = b[OBJECT_GOW2_HEADER_SIZE:size]

	o.Type = binary.LittleEndian.Uint16(b[0:])
	o.Unk02 = binary.LittleEndian.Uint16(b[2:])
	o.DmaTagsCountPerPacket = binary.LittleEndian.Uint32(b[4:])
	o.MaterialId = binary.LittleEndian.Uint16(b[8:])
	o.JointMappers = make([][]uint32, 1)
	if jmLen := binary.LittleEndian.Uint16(b[0xa:]); jmLen != 0 {
		o.JointMappers[0] = make([]uint32, binary.LittleEndian.Uint16(b[0xa:]))
	}
	o.InstancesCount = binary.LittleEndian.Uint32(b[0xc:])
	o.Flags = binary.LittleEndian.Uint32(b[0x10:])
	o.FlagsMask = binary.LittleEndian.Uint32(b[0x14:])
	o.TextureLayersCount = b[0x18]
	o.TotalDmaProgramsCount = b[0x19]
	o.NextFreeVUBufferId = binary.LittleEndian.Uint16(b[0x1a:])
	o.Unk1c = binary.LittleEndian.Uint16(b[0x1c:])
	o.SourceVerticesCount = binary.LittleEndian.Uint16(b[0x1e:])

	exlog.Printf("        | type: 0x%.4x  unk02: 0x%.4x packets per filter?: %d materialId: %d joints: %d",
		o.Type, o.Unk02, o.DmaTagsCountPerPacket, o.MaterialId, len(o.JointMappers[0]))
	exlog.Printf("        | unk0c: 0x%.8x unk10: 0x%.8x unk14: 0x%.8x textureLayers: %d unk19: 0x%.2x next free vu buffer: 0x%.4x unk1c: 0x%.4x source vertices count: 0x%.4x ",
		o.InstancesCount, o.Flags, o.FlagsMask, o.TextureLayersCount, o.TotalDmaProgramsCount, o.NextFreeVUBufferId, o.Unk1c, o.SourceVerticesCount)
	exlog.Printf("      --===--\n%v\n", utils.SDump(o.RawDmaAndJointsData))

	dmaCalls := o.InstancesCount * uint32(o.TextureLayersCount)
	o.Packets = make([][]Packet, dmaCalls)
	for iDmaChain := uint32(0); iDmaChain < dmaCalls; iDmaChain++ {
		packetOffset := o.Offset + OBJECT_GOW1_HEADER_SIZE + iDmaChain*o.DmaTagsCountPerPacket*0x10
		exlog.Printf("        - packets %d offset 0x%.8x pps 0x%.8x", iDmaChain, packetOffset, o.DmaTagsCountPerPacket)

		ds := NewMeshParserStream(allb, o, packetOffset, nil, exlog)
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
	if len(o.Packets) == 0 || o.Packets[0] == nil {
		log.Printf(" object have %d packets", len(o.Packets))
	} else {
		exlog.Printf("%v\n", utils.SDump(o.Packets[0]))
	}
	if o.JointMappers[0] != nil {
		// right after dma calls
		jointMapOffset := OBJECT_GOW1_HEADER_SIZE + dmaCalls*0x10*o.DmaTagsCountPerPacket
		for i := range o.JointMappers[0] {
			o.JointMappers[0][i] = binary.LittleEndian.Uint32(b[jointMapOffset+uint32(i)*4:])
		}
		exlog.Printf("              - joint map: %+#v", o.JointMappers[0])
	}

	return nil
}

func (g *Group) parseGow2(allb []byte, pos uint32, size uint32, exlog *utils.Logger) error {
	b := allb[pos:]
	g.Offset = pos

	g.HideDistance = math.Float32frombits(binary.LittleEndian.Uint32(b[0:]))
	g.Objects = make([]Object, binary.LittleEndian.Uint16(b[4:]))
	g.HasBbox = binary.LittleEndian.Uint32(b[8:])
	exlog.Printf("      | unk00: 0x%.8x unk08: 0x%.8x objects count: %v", g.HideDistance, g.HasBbox, len(g.Objects))

	for i := range g.Objects {
		objectOffset := binary.LittleEndian.Uint32(b[GROUP_GOW2_HEADER_SIZE+i*4:])
		exlog.Printf(" - - - object %d offset 0x%.8x", i, pos+objectOffset)

		objectEnd := size
		if i != len(g.Objects)-1 {
			objectEnd = binary.LittleEndian.Uint32(b[GROUP_GOW2_HEADER_SIZE+i*4+4:])
		}

		if err := g.Objects[i].parseGow2(allb, pos+objectOffset, objectEnd-objectOffset, exlog); err != nil {
			return fmt.Errorf("Error when parsing object %d: %v", i, err)
		}
	}

	return nil
}

func (p *Part) parseGow2(allb []byte, pos uint32, size uint32, exlog *utils.Logger) error {
	b := allb[pos:]
	p.Offset = pos
	p.Unk00 = binary.LittleEndian.Uint16(b[:])
	p.Groups = make([]Group, binary.LittleEndian.Uint16(b[2:]))
	p.JointId = binary.LittleEndian.Uint16(b[len(p.Groups)*4+PART_GOW2_HEADER_SIZE:])
	exlog.Printf("    | unk00: 0x%.4x jointid: %d groups count: %v", p.Unk00, p.JointId, len(p.Groups))

	for i := range p.Groups {
		groupOffset := binary.LittleEndian.Uint32(b[PART_GOW2_HEADER_SIZE+i*4:])
		exlog.Printf(" - - group %d offset 0x%.8x", i, pos+groupOffset)

		groupEnd := size
		if i != len(p.Groups)-1 {
			groupEnd = binary.LittleEndian.Uint32(b[PART_GOW2_HEADER_SIZE+i*4+4:])
		}

		if err := p.Groups[i].parseGow2(allb, pos+groupOffset, groupEnd-groupOffset, exlog); err != nil {
			return fmt.Errorf("Error when parsing group %d: %v", i, err)
		}
	}
	return nil
}

func (m *Mesh) parseGow2(b []byte, exlog *utils.Logger) error {
	if binary.LittleEndian.Uint32(b) != MESH_MAGIC {
		return fmt.Errorf("Invalid mesh magic")
	}

	mdlCommentStart := binary.LittleEndian.Uint32(b[4:])
	b = b[:mdlCommentStart]

	m.Parts = make([]Part, binary.LittleEndian.Uint16(b[8:]))

	for i := range m.Parts {
		partOffset := binary.LittleEndian.Uint32(b[MESH_GOW2_HEADER_SIZE+i*4:])
		exlog.Printf(" - part %d offset 0x%.8x", i, partOffset)

		partEnd := mdlCommentStart
		if i != len(m.Parts)-1 {
			partEnd = binary.LittleEndian.Uint32(b[MESH_GOW2_HEADER_SIZE+i*4+4:])
		}

		if err := m.Parts[i].parseGow2(b, partOffset, partEnd-partOffset, exlog); err != nil {
			return fmt.Errorf("Error when parsing part %d: %v", i, err)
		}
	}

	return nil
}
