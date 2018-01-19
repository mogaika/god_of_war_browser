package mesh

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Packet struct {
	Uvs struct {
		U, V []float32
	}
	Trias struct {
		X, Y, Z []float32
		Skip    []bool
	}
	Norms struct {
		X, Y, Z []float32
	}
	Blend struct {
		R, G, B, A []uint16 // actually uint8, only for marshaling
	}
	Joints                 []uint16
	Joints2                []uint16
	Offset                 uint32
	VertexMeta             []byte
	Boundaries             [4]float32 // center pose (xyz) and radius (w)
	HasTransparentBlending bool
}

type Object struct {
	Offset uint32

	Type                uint16
	Unk02               uint16
	PacketsPerFilter    uint32
	MaterialId          uint16
	JointMapper         []uint32
	Unk0c               uint32
	Unk10               uint32 // if & 0x40 - then we get broken joints and diff between type 0x1D and others
	Unk14               uint32
	TextureLayersCount  uint8
	Unk19               uint8
	NextFreeVUBufferId  uint16
	Unk1c               uint16
	SourceVerticesCount uint16

	Packets             [][]Packet
	RawDmaAndJointsData []byte
}

type Group struct {
	Offset uint32

	Unk00   uint32
	Objects []Object
	Unk08   uint32
}

type Part struct {
	Offset uint32

	Unk00   uint16
	Groups  []Group
	JointId uint16 // parent joint
}

type Vector struct {
	Unk00 uint16
	Unk02 uint16
	Value [4]float32
}

type Mesh struct {
	Parts   []Part
	Vectors []Vector

	Unk0c           uint32
	Unk10           uint32
	Unk14           uint32
	Flags0x20       uint32
	NameOfRootJoint string
	Unk28           uint32
	Unk2c           uint32
	Unk30           uint32
	Unk34           uint32
}

const (
	MESH_MAGIC         = 0x0001000f
	MESH_HEADER_SIZE   = 0x50
	PART_HEADER_SIZE   = 4
	GROUP_HEADER_SIZE  = 0xC
	OBJECT_HEADER_SIZE = 0x20
	MESH_VECTOR_SIZE   = 0x14
)

func (o *Object) Parse(allb []byte, pos uint32, size uint32, exlog *Logger) error {
	b := allb[pos:]
	o.Offset = pos
	o.RawDmaAndJointsData = b[OBJECT_HEADER_SIZE:size]

	o.Type = binary.LittleEndian.Uint16(b[0:])
	o.Unk02 = binary.LittleEndian.Uint16(b[2:])
	o.PacketsPerFilter = binary.LittleEndian.Uint32(b[4:])
	o.MaterialId = binary.LittleEndian.Uint16(b[8:])
	if jmLen := binary.LittleEndian.Uint16(b[0xa:]); jmLen != 0 {
		o.JointMapper = make([]uint32, binary.LittleEndian.Uint16(b[0xa:]))
	}
	o.Unk0c = binary.LittleEndian.Uint32(b[0xc:])
	o.Unk10 = binary.LittleEndian.Uint32(b[0x10:])
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
		packetOffset := o.Offset + OBJECT_HEADER_SIZE + iDmaChain*o.PacketsPerFilter*0x10
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
	exlog.Printf("%v\n", utils.SDump(o.Packets[0]))
	if o.JointMapper != nil {
		// right after dma calls
		jointMapOffset := OBJECT_HEADER_SIZE + dmaCalls*0x10*o.PacketsPerFilter
		for i := range o.JointMapper {
			o.JointMapper[i] = binary.LittleEndian.Uint32(b[jointMapOffset+uint32(i)*4:])
		}
		exlog.Printf("              - joint map: %+#v", o.JointMapper)
	}

	return nil
}

func (g *Group) Parse(allb []byte, pos uint32, size uint32, exlog *Logger) error {
	b := allb[pos:]
	g.Offset = pos

	g.Unk00 = binary.LittleEndian.Uint32(b[0:])
	g.Objects = make([]Object, binary.LittleEndian.Uint32(b[4:]))
	g.Unk08 = binary.LittleEndian.Uint32(b[8:])
	exlog.Printf("      | unk00: 0x%.8x unk08: 0x%.8x objects count: %v", g.Unk00, g.Unk08, len(g.Objects))

	for i := range g.Objects {
		objectOffset := binary.LittleEndian.Uint32(b[GROUP_HEADER_SIZE+i*4:])
		exlog.Printf(" - - - object %d offset 0x%.8x", i, pos+objectOffset)

		objectEnd := size
		if i != len(g.Objects)-1 {
			objectEnd = binary.LittleEndian.Uint32(b[GROUP_HEADER_SIZE+i*4+4:])
		}

		if err := g.Objects[i].Parse(allb, pos+objectOffset, objectEnd-objectOffset, exlog); err != nil {
			return fmt.Errorf("Error when parsing object %d: %v", i, err)
		}
	}

	return nil
}

func (p *Part) Parse(allb []byte, pos uint32, size uint32, exlog *Logger) error {
	b := allb[pos:]
	p.Offset = pos
	p.Unk00 = binary.LittleEndian.Uint16(b[:])
	p.Groups = make([]Group, binary.LittleEndian.Uint16(b[2:]))
	p.JointId = binary.LittleEndian.Uint16(b[len(p.Groups)*4+PART_HEADER_SIZE:])
	exlog.Printf("    | unk00: 0x%.4x jointid: %d groups count: %v", p.Unk00, p.JointId, len(p.Groups))

	for i := range p.Groups {
		groupOffset := binary.LittleEndian.Uint32(b[PART_HEADER_SIZE+i*4:])
		exlog.Printf(" - - group %d offset 0x%.8x", i, pos+groupOffset)

		groupEnd := size
		if i != len(p.Groups)-1 {
			groupEnd = binary.LittleEndian.Uint32(b[PART_HEADER_SIZE+i*4+4:])
		}

		if err := p.Groups[i].Parse(allb, pos+groupOffset, groupEnd-groupOffset, exlog); err != nil {
			return fmt.Errorf("Error when parsing group %d: %v", i, err)
		}
	}
	return nil
}

func (v *Vector) Parse(allb []byte, pos uint32, exlog *Logger) {
	b := allb[pos:]
	v.Unk00 = binary.LittleEndian.Uint16(b[0:])
	v.Unk02 = binary.LittleEndian.Uint16(b[2:])

	for i := range v.Value {
		v.Value[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[4+i*4:]))
	}

	exlog.Printf(" <- unk00: 0x%.4x unk02: 0x%.4x : %v", v.Unk00, v.Unk02, v.Value)
}

func (m *Mesh) Parse(b []byte, exlog *Logger) error {
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

	vectorsStart := len(m.Parts)*4 + MESH_HEADER_SIZE
	exlog.Printf(" - strange vectors starting at 0x%.8x count %d", vectorsStart, len(m.Vectors))
	for i := range m.Vectors {
		m.Vectors[i].Parse(b, uint32(vectorsStart+i*MESH_VECTOR_SIZE), exlog)
	}

	for i := range m.Parts {
		partOffset := binary.LittleEndian.Uint32(b[MESH_HEADER_SIZE+i*4:])
		exlog.Printf(" - part %d offset 0x%.8x", i, partOffset)

		partEnd := mdlCommentStart
		if i != len(m.Parts)-1 {
			partEnd = binary.LittleEndian.Uint32(b[MESH_HEADER_SIZE+i*4+4:])
		}

		if err := m.Parts[i].Parse(b, partOffset, partEnd-partOffset, exlog); err != nil {
			return fmt.Errorf("Error when parsing part %d: %v", i, err)
		}
	}

	return nil
}

func NewFromData(b []byte, exlog *Logger) (*Mesh, error) {
	m := &Mesh{}
	if err := m.Parse(b, exlog); err != nil {
		return nil, err
	} else {
		return m, nil
	}
}

func (m *Mesh) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return m, nil
}

func init() {
	wad.SetHandler(MESH_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		/*
			fpath := filepath.Join("logs", wrsrc.Wad.Name(), fmt.Sprintf("%.4d-%s.mesh.log", wrsrc.Tag.Id, wrsrc.Tag.Name))
			os.MkdirAll(filepath.Dir(fpath), 0777)
			f, _ := os.Create(fpath)
			defer f.Close()
			logger := Logger{f}
			//logger := Logger{os.Stdout}
		*/

		return NewFromData(wrsrc.Tag.Data, nil)
	})
}
