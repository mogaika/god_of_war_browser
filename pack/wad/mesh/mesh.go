package mesh

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/mesh/common"
	"github.com/mogaika/god_of_war_browser/pack/wad/mesh/gmdl"
	"github.com/mogaika/god_of_war_browser/utils"
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
		R, G, B, A []uint16 // actually uint8, only for marshaling (not base64-encode)
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

	// Types:
	//   0x1D - static mesh
	//   0x0E - dynamic or transparent
	//   everything else - lines
	Type                  uint16
	Unk02                 uint16 // always zero
	DmaTagsCountPerPacket uint32
	MaterialId            uint16
	JointMapElementsCount uint16

	// new dma program per each instance.
	// uses same buffers except rgba lighting, own jointmapper per instance
	InstancesCount uint32
	Flags          uint32 // if & 0x40 - then we get broken joints and diff between type 0x1D and others
	FlagsMask      uint32

	// new dma program per texture layer
	// uses same buffers except uv and rgba for  second layer
	TextureLayersCount    uint8
	TotalDmaProgramsCount uint8 // total dma programs count ?
	NextFreeVUBufferId    uint16
	Unk1c                 uint16 // source faces count
	SourceVerticesCount   uint16 // unique vertices count ?

	Packets             [][]Packet
	RawDmaAndJointsData []byte
	UseInvertedMatrix   bool
	JointMappers        [][]uint32
}

type Group struct {
	Offset uint32
	// Each part - lod group

	// value with minus, for last lod - very huge number
	HideDistance float32

	Objects []Object
	HasBbox uint32
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
	// last vector used in mdl

	Unk0c           uint32
	Unk10           uint32
	Unk14           uint32
	Flags0x20       uint32 // use last and first vectors for some rendering process
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

func (m *Mesh) AsCommonMesh() *common.Mesh {
	cMesh := &common.Mesh{
		Parts: make([]*common.Part, len(m.Parts)),
	}

	for iPart := range m.Parts {
		part := &m.Parts[iPart]
		cPart := &common.Part{
			LodGroups: make([]*common.LodGroup, len(part.Groups)),
		}
		cMesh.Parts[iPart] = cPart

		for iGroup := range part.Groups {
			group := &part.Groups[iGroup]
			cLodGroup := &common.LodGroup{
				Objects:      make([]*common.Object, len(group.Objects)),
				HideDistance: group.HideDistance,
			}
			cPart.LodGroups[iGroup] = cLodGroup

			for iObject := range group.Objects {
				object := &group.Objects[iObject]
				cObject := &common.Object{
					Vertices:       make([]common.Vertex, 0, 64),
					Indexes:        make([]int, 0, 128),
					JointMaps:      object.JointMappers,
					BlendColors:    make([][]common.RGBA, 0, 2),
					UVs:            make([][]common.UV, 0, 2),
					MaterialIndex:  int(object.MaterialId),
					PartIndex:      iPart,
					LodGroupIndex:  iGroup,
					ObjectIndex:    iObject,
					InstancesCount: int(object.InstancesCount),
					LayersCount:    int(object.TextureLayersCount),
				}
				cLodGroup.Objects[iObject] = cObject

				if object.Packets[0][0].Norms.X != nil {
					cObject.Normals = make([]common.Normal, 0, 64)
				}

				// fill vertices and normals from first packet
				for iInstance := 0; iInstance < int(object.InstancesCount); iInstance++ {
					for iLayer := 0; iLayer < int(object.TextureLayersCount); iLayer++ {
						iDmaPacket := iInstance*int(object.TextureLayersCount) + iLayer
						packets := object.Packets[iDmaPacket]

						var blendColors []common.RGBA
						if packets[0].Blend.R != nil {
							blendColors = make([]common.RGBA, 0, 64)
						}

						var uvs []common.UV
						if iInstance == 0 {
							// uv differs only for layers
							if packets[0].Uvs.U != nil {
								uvs = make([]common.UV, 0, 64)
							}
						}

						for iPacket := range packets {
							packet := &packets[iPacket]

							for iVertex := range packet.Trias.X {
								if iDmaPacket == 0 {
									// fill vertices, indexes and normals from first dma packet
									cObject.Vertices = append(cObject.Vertices,
										common.Vertex{
											Position: common.Position{
												packet.Trias.X[iVertex],
												packet.Trias.Y[iVertex],
												packet.Trias.Z[iVertex],
											},
											Weight: packet.Trias.Weight[iVertex],
											JointsIndexes: [2]uint16{
												packet.Joints[iVertex],
												packet.Joints2[iVertex]},
										})

									if !packet.Trias.Skip[iVertex] {
										curIndex := len(cObject.Vertices) - 1
										cObject.Indexes = append(cObject.Indexes,
											curIndex-2, curIndex-1, curIndex)
									}

									if cObject.Normals != nil {
										cObject.Normals = append(cObject.Normals,
											common.Normal{
												packet.Norms.X[iVertex],
												packet.Norms.Y[iVertex],
												packet.Norms.Z[iVertex],
											})
									}
								}

								if blendColors != nil {
									blendColors = append(blendColors,
										common.RGBA{
											R: uint8(packet.Blend.R[iVertex]),
											G: uint8(packet.Blend.G[iVertex]),
											B: uint8(packet.Blend.B[iVertex]),
											A: byte((float64(packet.Blend.A[iVertex]) / 128.0) * 255.0),
										})
								}
								if uvs != nil {
									uvs = append(uvs,
										common.UV{
											packet.Uvs.U[iVertex],
											packet.Uvs.V[iVertex],
										})
								}
							}
						}

						if blendColors != nil {
							cObject.BlendColors = append(cObject.BlendColors, blendColors)
						}
						if uvs != nil {
							cObject.UVs = append(cObject.UVs, uvs)
						}
					}
				}
			}
		}
	}

	return cMesh
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

		mesh, err := NewFromData(wrsrc.Tag.Data, &logger)
		if mesh.BaseBoneIndex != 0 {
			log.Printf("bbi: %d mesh: %s:%s j: %q",
				mesh.BaseBoneIndex, wrsrc.Wad.Name(), wrsrc.Tag.Name, mesh.NameOfRootJoint)
		}

		return mesh, err
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
		// log.Printf("\n%v", bs.StringTree())
		return g, err
	})
}
