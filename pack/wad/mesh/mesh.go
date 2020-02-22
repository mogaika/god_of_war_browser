package mesh

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/mesh/gmdl"
)

type Packet struct {
	Uvs struct {
		U, V []float32
	}
	Trias struct {
		X, Y, Z []float32
		Skip    []bool
		Weight  []float32
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

	Type                  uint16
	Unk02                 uint16
	DmaTagsCountPerPacket uint32
	MaterialId            uint16
	JointMapElementsCount uint16
	// new dma program per each instance.
	// uses same buffers except rgba lighting
	// have own jointmappers
	InstancesCount      uint32
	Flags               uint32 // if & 0x40 - then we get broken joints and diff between type 0x1D and others
	FlagsMask           uint32
	TextureLayersCount  uint8
	Unk19               uint8 // total dma programs count ?
	NextFreeVUBufferId  uint16
	Unk1c               uint16 // source faces count
	SourceVerticesCount uint16 // unique vertices count ?

	Packets             [][]Packet
	RawDmaAndJointsData []byte
	UseInvertedMatrix   bool
	JointMappers        [][]uint32
}

type Group struct {
	Offset uint32

	HideDistance float32
	Objects      []Object
	HasBbox      uint32
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
	BaseBoneIndex   uint32
}

const (
	MESH_MAGIC = 0x0001000f
	GMDL_MAGIC = 0x0003000f
)

func NewFromData(b []byte, exlog *utils.Logger) (*Mesh, error) {
	m := &Mesh{}
	switch config.GetGOWVersion() {
	case config.GOW1:
		if err := m.parseGow1(b, exlog); err != nil {
			return nil, err
		} else {
			return m, nil
		}
	case config.GOW2:
		if err := m.parseGow2(b, exlog); err != nil {
			return nil, err
		} else {
			return m, nil
		}
	default:
		panic("err")
	}
}

func (m *Mesh) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return m, nil
}

func init() {
	wad.SetHandler(config.GOW1, MESH_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {

		fpath := filepath.Join("logs", wrsrc.Wad.Name(), fmt.Sprintf("%.4d-%s.mesh.log", wrsrc.Tag.Id, wrsrc.Tag.Name))
		os.MkdirAll(filepath.Dir(fpath), 0777)
		f, _ := os.Create(fpath)
		defer f.Close()
		logger := utils.Logger{f}

		//logger := utils.Logger{ioutil.Discard}

		return NewFromData(wrsrc.Tag.Data, &logger)
	})
	wad.SetHandler(config.GOW2, MESH_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		fpath := filepath.Join("logs_gow2", wrsrc.Wad.Name(), fmt.Sprintf("%.4d-%s.mesh.log", wrsrc.Tag.Id, wrsrc.Tag.Name))
		os.MkdirAll(filepath.Dir(fpath), 0777)
		f, _ := os.Create(fpath)
		defer f.Close()
		logger := utils.Logger{f}
		//logger := Logger{io.MultiWriter(os.Stdout, f)}

		return NewFromData(wrsrc.Tag.Data, &logger)
	})
	wad.SetHandler(config.GOW1, GMDL_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		bs := utils.NewBufStack("resource", wrsrc.Tag.Data[:]).SetSize(int(wrsrc.Size()))
		g, err := gmdl.NewGMDL(bs.SubBuf("gmdl", 4).Expand().SetName(wrsrc.Name()))
		log.Printf("\n%v", bs.StringTree())
		return g, err
	})
}
