package mdl

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"reflect"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mat "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	file_mesh "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	file_gmdl "github.com/mogaika/god_of_war_browser/pack/wad/mesh/gmdl"
	file_scr "github.com/mogaika/god_of_war_browser/pack/wad/scr"
)

type Model struct {
	Unk4         uint32     // Always zero
	UnkFloats    [3]float32 // Used in LOD related meshes. Maybe range for rendering
	TextureCount uint32
	MeshCount    uint32
	JointsCount  uint32

	/*
		[0] & 0x8 - is animated or gui
		[0] & 0x10 - simple animated
		[0] & 0x40 - is breakable? (and some enemy)
		[0] & 0x80 - joint weight double blending (hero or complicated models)
	*/
	Flags  uint32
	Unk24  uint32 // Unused
	Ints28 [4]int32
	Unk38  uint32 // Unused
	Unk3c  uint32
	Unk40  uint32 // count of some animated texture layers to separate anims???
	Unk44  uint32 // Probably unused (different types)
}

const MODEL_MAGIC = 0x0002000f

func NewFromData(buf []byte) (*Model, error) {
	mdl := new(Model)

	mdl.Unk4 = binary.LittleEndian.Uint32(buf[0x4:0x8])
	mdl.UnkFloats[0] = math.Float32frombits(binary.LittleEndian.Uint32(buf[0x8:0xc]))
	mdl.UnkFloats[1] = math.Float32frombits(binary.LittleEndian.Uint32(buf[0xc:0x10]))
	mdl.UnkFloats[2] = math.Float32frombits(binary.LittleEndian.Uint32(buf[0x10:0x14]))
	mdl.TextureCount = binary.LittleEndian.Uint32(buf[0x14:0x18])
	mdl.MeshCount = binary.LittleEndian.Uint32(buf[0x18:0x1c])
	mdl.JointsCount = binary.LittleEndian.Uint32(buf[0x1c:0x20])

	mdl.Flags = binary.LittleEndian.Uint32(buf[0x20:])
	mdl.Unk24 = binary.LittleEndian.Uint32(buf[0x24:])
	for i := range mdl.Ints28 {
		mdl.Ints28[i] = int32(binary.LittleEndian.Uint32(buf[0x28+i*4:]))
	}
	mdl.Unk38 = binary.LittleEndian.Uint32(buf[0x38:])
	mdl.Unk3c = binary.LittleEndian.Uint32(buf[0x3c:])
	mdl.Unk40 = binary.LittleEndian.Uint32(buf[0x40:])
	mdl.Unk44 = binary.LittleEndian.Uint32(buf[0x44:])

	return mdl, nil
}

type Ajax struct {
	Raw       *Model
	Meshes    []*file_mesh.Mesh
	Materials []interface{}
	Scripts   []interface{}
	Other     []interface{}
	GMDL      interface{}
}

func (mdl *Model) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	res := &Ajax{Raw: mdl}
	for _, i := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(i)
		name := n.Tag.Name
		sn, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id)
		if err != nil {
			// TODO: improve
			if config.GetGOWVersion() == config.GOW1 {
				return nil, fmt.Errorf("Error when extracting node %d->%s mdl info: %v", i, name, err)
			}
		} else {
			switch sn.(type) {
			case *file_mesh.Mesh:
				res.Meshes = append(res.Meshes, sn.(*file_mesh.Mesh))
			case *file_mat.Material:
				mat, err := sn.(*file_mat.Material).Marshal(wrsrc.Wad.GetNodeResourceByNodeId(n.Id))
				if err != nil {
					return nil, fmt.Errorf("Error when getting material info %d-'%s': %v", i, name, err)
				}
				res.Materials = append(res.Materials, mat)
			case *file_scr.ScriptParams:
				scr, err := sn.(*file_scr.ScriptParams).Marshal(wrsrc.Wad.GetNodeResourceByNodeId(n.Id))
				if err != nil {
					return nil, fmt.Errorf("Error when getting script info %d-'%s': %v", i, name, err)
				}
				res.Scripts = append(res.Scripts, scr)
			case *file_gmdl.GMDL:
				gmdl, err := sn.(*file_gmdl.GMDL).Marshal(wrsrc.Wad.GetNodeResourceByNodeId(n.Id))
				if err != nil {
					return nil, fmt.Errorf("Error when getting gmdl %d-'%s': %v", i, name, err)
				}
				res.GMDL = gmdl
			default:
				log.Printf("MDL: Added unknown interface of %s", reflect.TypeOf(sn).Name())
			}
		}
	}
	return res, nil
}

func init() {
	wad.SetServerHandler(config.GOW1, MODEL_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		mdl, err := NewFromData(wrsrc.Tag.Data)
		if err == nil {
			/*
				fprefx := fmt.Sprintf("%.4d-%s", wrsrc.Tag.Id, wrsrc.Tag.Name)

				mdlpath := filepath.Join("mdl", wrsrc.Wad.Name(), fprefx+".obj")
				os.MkdirAll(filepath.Dir(mdlpath), 0777)
				fMdl, _ := os.Create(mdlpath)
				defer fMdl.Close()

				mtlPath := filepath.Join("mdl", wrsrc.Wad.Name(), fprefx+".mtl")
				fMtl, _ := os.Create(mtlPath)
				defer fMtl.Close()

				textures, err := mdl.ExportObj(wrsrc, nil, mtlPath, fMdl, fMtl)
				if err == nil {
					for tname, t := range textures {
						pngPath := filepath.Join("mdl", wrsrc.Wad.Name(), tname+".png")
						f, err := os.Create(pngPath)
						if err == nil {
							defer f.Close()
							f.Write(t)
						}
					}
				}
			*/
		}
		return mdl, err
	})

	wad.SetServerHandler(config.GOW2, MODEL_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
