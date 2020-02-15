package mesh

import (
	"fmt"

	"github.com/mogaika/god_of_war_browser/fbx"
	"github.com/mogaika/god_of_war_browser/fbx/cache"
	"github.com/mogaika/god_of_war_browser/pack/wad"
)

type FbxExportObject struct {
	FbxGeometryId uint64
	FbxModelId    uint64

	Part       int
	Group      int
	Object     int
	MaterialId int
	InstanceId int
}

type FbxExportPart struct {
	FbxModel   *fbx.Model
	SkeletUsed bool
	Objects    []*FbxExportObject

	Part    int
	RawPart *Part
}

type FbxExporter struct {
	m     *Mesh
	Parts []*FbxExportPart
}

func uint8ColToF32(col uint16) float32 {
	return float32(col) / 255.0
}

func uint8AlpfaToF32(col uint16) float32 {
	return float32(col) / 128.0
}

func (fe *FbxExporter) exportObject(f *fbx.FBX, feo *FbxExportObject, fep *FbxExportPart) {
	o := &fe.m.Parts[feo.Part].Groups[feo.Group].Objects[feo.Object]
	feo.MaterialId = int(o.MaterialId)

	for _, jointMap := range o.JointMappers {
		if len(jointMap) == 0 {
			panic("wtf?")
		}
		if len(jointMap) != 1 || jointMap[0] != 0 {
			fep.SkeletUsed = true
		}
	}

	vertices := make([]float32, 0)
	indexes := make([]int, 0)
	uvindexes := make([]int, 0)
	rgba := make([]float32, 0)
	normals := make([]float32, 0)
	uv := make([]float32, 0)

	haveNorm := o.Packets[0][0].Norms.X != nil
	haveRgba := o.Packets[0][0].Blend.R != nil
	haveUV := o.Packets[0][0].Uvs.U != nil

	flip := false

	// first extract pos, color, norm
	for iPacket := range o.Packets[0] {
		packet := o.Packets[0][iPacket]

		for iVertex := range packet.Trias.X {
			if !packet.Trias.Skip[iVertex] {
				curIndex := len(vertices) / 3
				if flip {
					indexes = append(indexes, curIndex-2, curIndex-1, -(curIndex)-1)
					uvindexes = append(uvindexes, curIndex-2, curIndex-1, curIndex)
				} else {
					indexes = append(indexes, curIndex-1, curIndex-2, -(curIndex)-1)
					uvindexes = append(uvindexes, curIndex-1, curIndex-2, curIndex)
				}
				flip = !flip
			}

			vertices = append(vertices,
				packet.Trias.X[iVertex], packet.Trias.Y[iVertex], packet.Trias.Z[iVertex])
			if haveNorm {
				normals = append(normals,
					packet.Norms.X[iVertex], packet.Norms.Y[iVertex], packet.Norms.Z[iVertex])
			}
			if haveRgba {
				rgba = append(rgba,
					uint8ColToF32(packet.Blend.R[iVertex]),
					uint8ColToF32(packet.Blend.G[iVertex]),
					uint8ColToF32(packet.Blend.B[iVertex]),
					uint8AlpfaToF32(packet.Blend.A[iVertex]))
			}
			if haveUV {
				uv = append(uv,
					packet.Uvs.U[iVertex], -packet.Uvs.V[iVertex])
			}
		}
	}

	name := fmt.Sprintf("g%d_o%d_m%d_i%d", feo.Group, feo.Object, feo.MaterialId, feo.InstanceId)

	geometry := &fbx.Geometry{
		Id:                 f.GenerateId(),
		Name:               "Geometry::" + name,
		Element:            "Mesh",
		GeometryVersion:    124,
		Vertices:           vertices,
		PolygonVertexIndex: indexes,
		Layer:              &fbx.Layer{Version: 100},
	}
	feo.FbxGeometryId = geometry.Id

	if haveNorm {
		geometry.LayerElementNormal = &fbx.LayerElementShared{
			Version:                  102,
			MappingInformationType:   "ByVertice",
			ReferenceInformationType: "Direct",
			Normals:                  normals,
			NormalsW:                 make([]int, len(normals)/3),
		}
		geometry.Layer.LayerElement = append(geometry.Layer.LayerElement, fbx.LayerElement{
			Type:       "LayerElementNormal",
			TypedIndex: 0,
		})
	}

	if haveRgba {
		geometry.LayerElementColor = &fbx.LayerElementShared{
			Version:                  101,
			MappingInformationType:   "ByVertice",
			ReferenceInformationType: "Direct",
			Colors:                   rgba,
		}
		geometry.Layer.LayerElement = append(geometry.Layer.LayerElement, fbx.LayerElement{
			Type:       "LayerElementColor",
			TypedIndex: 0,
		})
	}

	if haveUV {
		geometry.LayerElementUV = &fbx.LayerElementShared{
			Version:                  101,
			Name:                     "UVMap",
			MappingInformationType:   "ByPolygonVertex",
			ReferenceInformationType: "IndexToDirect",
			UV:                       uv,
			UVIndex:                  uvindexes,
		}
		geometry.Layer.LayerElement = append(geometry.Layer.LayerElement, fbx.LayerElement{
			Type:       "LayerElementUV",
			TypedIndex: 0,
		})
	}

	geometry.LayerElementMaterial = &fbx.LayerElementShared{
		Version:                  101,
		MappingInformationType:   "AllSame",
		ReferenceInformationType: "IndexToDirect",
		Materials:                []int{0},
	}
	geometry.Layer.LayerElement = append(geometry.Layer.LayerElement, fbx.LayerElement{
		Type:       "LayerElementMaterial",
		TypedIndex: 0,
	})

	model := &fbx.Model{
		Id:      f.GenerateId(),
		Name:    "Model::" + name,
		Element: "Mesh",
		Version: 232,
		Shading: true,
		Culling: "CullingOff",
	}
	feo.FbxModelId = model.Id

	f.Objects.Model = append(f.Objects.Model, model)
	f.Objects.Geometry = append(f.Objects.Geometry, geometry)
	f.Connections.C = append(f.Connections.C, fbx.Connection{
		Type: "OO", Child: geometry.Id, Parent: model.Id,
	})

	fep.Objects = append(fep.Objects, feo)
}

func (fe *FbxExporter) exportPart(f *fbx.FBX, fep *FbxExportPart) {
	part := &fe.m.Parts[fep.Part]
	fep.Objects = make([]*FbxExportObject, 0)
	fep.RawPart = part

	for iGroup := range part.Groups {
		group := &part.Groups[iGroup]
		for iObject := range group.Objects {
			object := &group.Objects[iObject]

			for iInstance := uint32(0); iInstance < object.InstancesCount; iInstance++ {
				feo := &FbxExportObject{
					Part:       fep.Part,
					Group:      iGroup,
					Object:     iObject,
					InstanceId: int(iInstance),
				}
				fe.exportObject(f, feo, fep)
			}
		}
	}

	name := fmt.Sprintf("part%d", fep.Part)
	model := &fbx.Model{
		Id:      f.GenerateId(),
		Name:    "Model::" + name,
		Element: "Null",
		Version: 232,
		Shading: true,
		Culling: "CullingOff",
	}

	fep.FbxModel = model
	f.Objects.Model = append(f.Objects.Model, model)

	for _, object := range fep.Objects {
		f.Connections.C = append(f.Connections.C, fbx.Connection{
			Type: "OO", Parent: model.Id, Child: object.FbxModelId,
		})
	}

	fe.Parts = append(fe.Parts, fep)
}

func (m *Mesh) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbx.FBX, cache *cache.Cache) *FbxExporter {
	fe := &FbxExporter{
		m:     m,
		Parts: make([]*FbxExportPart, 0),
	}
	defer cache.Add(wrsrc.Tag.Id, fe)

	if f.Objects.Geometry == nil {
		f.Objects.Geometry = make([]*fbx.Geometry, 0)
	}

	for iPart := range m.Parts {
		fe.exportPart(f, &FbxExportPart{
			Part: iPart,
		})
	}
	fe.m = nil // free memory

	return fe
}

func (m *Mesh) ExportFbxDefault(wrsrc *wad.WadNodeRsrc) *fbx.FBX {
	f := fbx.NewFbx()
	fe := m.ExportFbx(wrsrc, f, cache.NewCache())

	for _, part := range fe.Parts {
		f.Connections.C = append(f.Connections.C, fbx.Connection{
			Type: "OO", Parent: 0, Child: part.FbxModel.Id,
		})
	}

	f.CountDefinitions()

	return f
}
