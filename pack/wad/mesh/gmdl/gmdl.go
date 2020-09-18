package gmdl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/mogaika/god_of_war_browser/3rdparty/half"
)

const (
	GMDL_MAGIC = 0x474d444c
	MDL_MAGIC  = 0x4d4f444c

	HEADER_SIZE = 0x18
	MDL_SIZE    = 0xc
	STREAM_SIZE = 0x10
)

type Stream struct {
	Name   string
	Id     uint16
	Flags  uint16
	Values interface{}
}

type Object struct {
	BBox         [2][3]float32
	TextureIndex uint16
	UnkByte      uint8
	StreamStart  uint32
	StreamCount  uint32
	IndexStart   uint32
	IndexCount   uint32
	JointsMap    []uint32
}

type Model struct {
	Id    uint16
	Flags uint16

	Streams           map[string]Stream
	Indexes           []uint32
	UsedTexturesCount uint32
	Objects           []Object
}

type GMDL struct {
	Magic  uint32
	Unk4   uint32
	Unk8   uint32
	UnkC   uint32
	Models []Model
	Unk14  uint32
}

func (o *Object) fromBuf(bs *utils.BufStack) error {
	for i, vec := range o.BBox {
		for j := range vec {
			o.BBox[i][j] = bs.ReadBF()
		}
	}

	o.TextureIndex = bs.ReadBU16()
	o.UnkByte = bs.ReadByte()
	if o.UnkByte != 0 {
		return fmt.Errorf("o.UnkByte == %v", o.UnkByte)
	}
	o.StreamStart = bs.ReadBU32()
	o.StreamCount = bs.ReadBU32()
	o.IndexStart = bs.ReadBU32()
	o.IndexCount = bs.ReadBU32()
	o.JointsMap = make([]uint32, bs.ReadBU32())
	for i := range o.JointsMap {
		o.JointsMap[i] = bs.ReadBU32()
	}
	log.Printf("%#+v", o)
	bs.SetSize(bs.Pos())

	return nil
}

func (s *Stream) parseData(bs *utils.BufStack) error {
	switch s.Id {
	case 2: // POS0
		s.Values = make([][4]float32, bs.Size()/16)
	case 4: // BONI
		s.Values = make([][4]byte, bs.Size()/4)
	case 5: // TEX0 float16
		s.Values = make([][2]uint16, bs.Size()/4)
	case 6: // CLR0 float16
		s.Values = make([][4]uint16, bs.Size()/8)
	case 9: // NRM0
		s.Values = make([][4]uint8, bs.Size()/4)
	default:
		return fmt.Errorf("Unknown stream id %v (%v:%v)", bs, s.Name, s.Id)
	}

	var endian binary.ByteOrder = binary.BigEndian
	if config.GetPlayStationVersion() == config.PSVita {
		endian = binary.LittleEndian
	}

	if err := binary.Read(bytes.NewReader(bs.Raw()), endian, s.Values); err != nil {
		return fmt.Errorf("Error parsing stream data %v (%v): %v", bs, s.Name, err)
	}

	switch s.Id {
	case 5:
		uintArr := s.Values.([][2]uint16)
		floatArr := make([][2]float32, len(uintArr))
		for i, uintBlock := range uintArr {
			for j, v := range uintBlock {
				var floatVal float32
				// flp untextured meshes have +infinity, we can't convert it to json, so setting it to 0
				if v == 0x7c00 {
					floatVal = 0.0
				} else {
					floatVal = half.Float16(v).Float32()
				}

				floatArr[i][j] = floatVal
			}
		}
		s.Values = floatArr
	case 6:
		uintArr := s.Values.([][4]uint16)
		floatArr := make([][4]float32, len(uintArr))
		for i, uintBlock := range uintArr {
			for j, v := range uintBlock {
				floatArr[i][j] = half.Float16(v).Float32()
			}
		}
		s.Values = floatArr
	}
	return nil
}

func (s *Stream) fromBuf(bs *utils.BufStack) error {
	switch config.GetPlayStationVersion() {
	case config.PS3:
		s.Name = bs.ReadStringBuffer(4)
		s.Id = bs.ReadBU16()
		s.Flags = bs.ReadBU16()
	case config.PSVita:
		s.Name = utils.ReverseString(bs.ReadStringBuffer(4))
		s.Id = bs.ReadLU16()
		s.Flags = bs.ReadLU16()
	default:
		panic("unknown ps version")
	}

	bs.SetName(s.Name)

	if s.Flags != 2 {
		return fmt.Errorf("s.Flags == %v", s.Flags)
	}

	var dataOffset, dataSize uint32

	switch config.GetPlayStationVersion() {
	case config.PS3:
		dataOffset = bs.ReadBU32()
		dataSize = bs.ReadBU32()
	case config.PSVita:
		dataOffset = bs.ReadLU32()
		dataSize = bs.ReadLU32()
	default:
		panic("unknown ps version")
	}

	bs.SetSize(bs.Pos()).VerifySize(STREAM_SIZE)
	bsData := bs.Parent().Parent().SubBuf("buffer", int(dataOffset)).SetName(s.Name).SetSize(int(dataSize))
	return s.parseData(bsData)
}

func (m *Model) fromBuf(bs *utils.BufStack) error {
	bsHeader := bs.SubBuf("mdlHeader", 0)

	if magic := bsHeader.ReadBU32(); magic != MDL_MAGIC {
		return fmt.Errorf("Magic mismatch %v != %v", magic, MDL_MAGIC)
	}
	m.Id = bsHeader.ReadBU16()
	m.Flags = bsHeader.ReadBU16()

	switch config.GetPlayStationVersion() {
	case config.PS3:
		if m.Flags != 8 {
			return fmt.Errorf("m.Flags == %v", m.Flags)
		}
	case config.PSVita:
		if m.Flags != 9 {
			return fmt.Errorf("m.Flags == %v", m.Flags)
		}
	default:
		panic("unknown ps version")
	}

	indexesOffset := int(bsHeader.ReadBU32())
	bsHeader.SetSize(bsHeader.Pos()).VerifySize(MDL_SIZE)

	bsData := bsHeader.SubBufFollowing("data").Expand()

	bsStreamsCount := bsData.SubBuf("streamsCount", 0).SetSize(4)

	var streamsCount int
	switch config.GetPlayStationVersion() {
	case config.PS3:
		streamsCount = int(bsStreamsCount.ReadBU32())
	case config.PSVita:
		streamsCount = int(bsStreamsCount.ReadLU32())
	}

	bsStreams := bsData.SubBuf("streams", 4).SetSize(streamsCount * STREAM_SIZE)

	m.Streams = make(map[string]Stream)
	for i := 0; i < streamsCount; i++ {
		bsStream := bsStreams.SubBuf("stream", STREAM_SIZE*i)
		var stream Stream
		if err := stream.fromBuf(bsStream); err != nil {
			return fmt.Errorf("Error parsing buf %v: %v", bsStream, err)
		}
		m.Streams[stream.Name] = stream
	}

	bsIndexes := bsData.SubBuf("indexes", indexesOffset)

	switch config.GetPlayStationVersion() {
	case config.PS3:
		m.Indexes = make([]uint32, bsIndexes.ReadBU32())
		for i := range m.Indexes {
			m.Indexes[i] = bsIndexes.ReadBU32()
		}
		bsIndexes.SetSize(bsIndexes.Pos()).VerifySize(len(m.Indexes)*4 + 4)
	case config.PSVita:
		log.Printf("Unknown floats: %v %v", bsIndexes.ReadLF(), bsIndexes.ReadLF())
		log.Printf("Unknown int: %v", bsIndexes.ReadBU32())
		m.Indexes = make([]uint32, bsIndexes.ReadBU32())
		log.Printf("Indexes: %v (%.8x)", len(m.Indexes), len(m.Indexes))
		for i := range m.Indexes {
			m.Indexes[i] = uint32(bsIndexes.ReadBU16())
		}
		bsIndexes.SetSize(bsIndexes.Pos()).VerifySize(len(m.Indexes)*2 + 4 + 12)
	}

	bsObjects := bsIndexes.SubBufFollowing("objects")
	bsObjects.Expand()

	bsObjectsMeta := bsObjects.SubBuf("meta", 0)
	m.UsedTexturesCount = bsObjectsMeta.ReadBU32()
	objectsCount := bsObjectsMeta.ReadBU32()
	bsObjectsMeta.SetSize(bsObjectsMeta.Pos()).VerifySize(8)

	m.Objects = make([]Object, objectsCount)
	var currentObject *utils.BufStack
	if len(m.Objects) != 0 {
		currentObject = bsObjects.SubBuf("object", 8)
		for i := range m.Objects {
			m.Objects[i].fromBuf(currentObject)
			if i != len(m.Objects)-1 {
				currentObject = currentObject.SubBufFollowing("object")
			}
		}
	}

	return nil
}

func (g *GMDL) fromBuf(bs *utils.BufStack) error {
	bsHeader := bs.SubBuf("gmdlHeader", 0)

	g.Magic = bsHeader.ReadBU32()
	g.Unk4 = bsHeader.ReadBU32()
	g.Unk8 = bsHeader.ReadBU32()
	g.UnkC = bsHeader.ReadBU32()
	log.Printf("%#+v", g)
	g.Models = make([]Model, bsHeader.ReadBU32())
	g.Unk14 = bsHeader.ReadBU32()

	bsHeader.SetSize(bsHeader.Pos()).VerifySize(HEADER_SIZE)

	if g.Magic != GMDL_MAGIC {
		return fmt.Errorf("Magic mismatch %v != %v", g.Magic, GMDL_MAGIC)
	}
	if g.Unk4 != 1 {
		return fmt.Errorf("Unk4 == %v", g.Unk4)
	}
	if g.Unk8 != 1 {
		// TODO: Check
		// wad_olymp01.wad_ps3 => sunFlare == 2
		// wad_athn01a.wad_ps3 => trapDoorFlap1 == 2
		//return fmt.Errorf("Unk8 == %v", g.Unk8)
	}
	if g.UnkC != 1 {
		// TODO: Check
		// wad_athn01a.wad_ps3 => trapDoorFlap1 == 2
		//return fmt.Errorf("UnkC == %v", g.UnkC)
	}
	if g.Unk14 != 1 {
		// TODO: Check
		// wad_athn02a.wad_ps3 => ~MD_lodSandBag08 == 2
		// return fmt.Errorf("Unk14 == %v", g.Unk14)
	}

	bsModelOffsets := bsHeader.SubBufFollowing("modelOffsets").SetSize(len(g.Models) * 4)
	modelOffsets := make([]int, len(g.Models))
	for i := range g.Models {
		modelOffsets[i] = int(bsModelOffsets.ReadBU32())
	}

	for i := range g.Models {
		modelBs := bs.SubBuf("model", modelOffsets[i]).SetName(fmt.Sprintf("mdl%d", i))
		if i == len(g.Models)-1 {
			modelBs.Expand()
		} else {
			modelBs.SetSize(modelOffsets[i+1] - modelOffsets[i])
		}

		if err := g.Models[i].fromBuf(modelBs); err != nil {
			return fmt.Errorf("Error parsing model %v: %v", modelBs, err)
		}
	}
	return nil
}

func NewGMDL(bs *utils.BufStack) (*GMDL, error) {
	gmdl := &GMDL{}
	if err := gmdl.fromBuf(bs); err != nil {
		return nil, err
	} else {
		return gmdl, nil
	}
}

func (g *GMDL) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return g, nil
}
