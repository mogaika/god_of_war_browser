package mdl

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mat "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	file_mesh "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
)

type Model struct {
	Unk4         uint32
	UnkFloats    [3]float32
	TextureCount uint32
	MeshCount    uint32
	JointsCount  uint32
	Someinfo     [10]uint32
}

const MODEL_MAGIC = 0x0002000f
const FILE_SIZE = 0x48

func NewFromData(rdr io.ReaderAt) (*Model, error) {
	var file [FILE_SIZE]byte
	_, err := rdr.ReadAt(file[:], 0)
	if err != nil {
		return nil, err
	}

	mdl := new(Model)

	mdl.Unk4 = binary.LittleEndian.Uint32(file[0x4:0x8])
	mdl.UnkFloats[0] = math.Float32frombits(binary.LittleEndian.Uint32(file[0x8:0xc]))
	mdl.UnkFloats[1] = math.Float32frombits(binary.LittleEndian.Uint32(file[0xc:0x10]))
	mdl.UnkFloats[2] = math.Float32frombits(binary.LittleEndian.Uint32(file[0x10:0x14]))
	mdl.TextureCount = binary.LittleEndian.Uint32(file[0x14:0x18])
	mdl.MeshCount = binary.LittleEndian.Uint32(file[0x18:0x1c])
	mdl.JointsCount = binary.LittleEndian.Uint32(file[0x1c:0x20])
	for i := range mdl.Someinfo {
		mdl.Someinfo[i] = binary.LittleEndian.Uint32(file[0x20+i*4:])
	}

	return mdl, nil
}

type Ajax struct {
	Raw       *Model
	Meshes    []*file_mesh.Mesh
	Materials []interface{}
	Other     []interface{}
}

func (mdl *Model) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	res := &Ajax{Raw: mdl}
	for _, i := range node.SubNodes {
		if nd := wad.Node(i); nd.Name[0] != ' ' {
			sn, err := wad.Get(i)
			if err != nil {
				return nil, fmt.Errorf("Error when extracting node %d->%s mdl info: %v", i, wad.Node(i).Name, err)
			} else {
				switch sn.(type) {
				case *file_mesh.Mesh:
					res.Meshes = append(res.Meshes, sn.(*file_mesh.Mesh))
				case *file_mat.Material:
					mat, err := sn.(*file_mat.Material).Marshal(wad, wad.Nodes[i])
					if err != nil {
						return nil, fmt.Errorf("Error when getting material info %d-'%s': %v", i, wad.Node(i).Name, err)
					}
					res.Materials = append(res.Materials, mat)
				default:
					res.Other = append(res.Other, "Unknown interface of "+reflect.TypeOf(sn).Name())
				}
			}
		}
	}
	return res, nil
}

func init() {
	wad.SetHandler(MODEL_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		return NewFromData(r)
	})
}
