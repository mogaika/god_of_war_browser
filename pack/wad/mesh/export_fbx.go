package mesh

import (
	"fmt"
	"path/filepath"

	"github.com/mogaika/fbx/builders/bfbx73"

	"github.com/mogaika/fbx"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils/fbxbuilder"
)

type FbxExportObject struct {
	FbxGeometryId int64
	FbxGeometry   *fbx.Node
	FbxModelId    int64
	FbxModel      *fbx.Node

	Part       int
	Group      int
	Object     int
	MaterialId int
	InstanceId int

	VerticeToJoint     [][2]uint16
	VerticeJointWeight []float32
	AffectedByJoints   map[uint16]struct{}
}

type FbxExportPart struct {
	//FbxModel   *fbx.Node
	//FbxModelId int64
	Objects []*FbxExportObject

	Part    int
	RawPart *Part
}

type FbxExporter struct {
	m     *Mesh
	Parts []*FbxExportPart
}

func uint8ColToF64(col uint16) float64 {
	return float64(col) / 255.0
}

func uint8AlpfaToF64(col uint16) float64 {
	return float64(col) / 128.0
}

func (fe *FbxExporter) exportObject(f *fbxbuilder.FBXBuilder, feo *FbxExportObject, fep *FbxExportPart) {
	o := &fe.m.Parts[feo.Part].Groups[feo.Group].Objects[feo.Object]
	feo.MaterialId = int(o.MaterialId)

	for _, jointMap := range o.JointMappers {
		if len(jointMap) == 0 {
			panic("wtf?")
		}
	}

	vertices := make([]float64, 0)
	indexes := make([]int32, 0)
	uvindexes := make([]int32, 0)
	rgba := make([]float64, 0)
	normals := make([]float64, 0)
	uv := make([]float64, 0)

	haveNorm := o.Packets[0][0].Norms.X != nil
	haveRgba := o.Packets[0][0].Blend.R != nil
	haveUV := o.Packets[0][0].Uvs.U != nil

	flip := false

	verticeToJoint := make([][2]uint16, 0)
	verticeJointWeight := make([]float32, 0)

	// first extract pos, color, norm
	for iPacket := range o.Packets[0] {
		packet := o.Packets[0][iPacket]

		for iVertex := range packet.Trias.X {
			if !packet.Trias.Skip[iVertex] {
				curIndex := int32(len(vertices) / 3)
				if flip {
					indexes = append(indexes, curIndex-2, curIndex-1, -(curIndex)-1)
					uvindexes = append(uvindexes, curIndex-2, curIndex-1, curIndex)
				} else {
					indexes = append(indexes, curIndex-1, curIndex-2, -(curIndex)-1)
					uvindexes = append(uvindexes, curIndex-1, curIndex-2, curIndex)
				}
				flip = !flip
			}

			joint1 := uint16(o.JointMappers[0][packet.Joints[0][iVertex]])
			joint2 := uint16(o.JointMappers[0][packet.Joints[1][iVertex]])
			weight := packet.Trias.Weight[iVertex]

			vertices = append(vertices,
				float64(packet.Trias.X[iVertex]),
				float64(packet.Trias.Y[iVertex]),
				float64(packet.Trias.Z[iVertex]))

			verticeToJoint = append(verticeToJoint, [2]uint16{joint1, joint2})
			verticeJointWeight = append(verticeJointWeight, weight)

			if weight < 1.0 {
				feo.AffectedByJoints[joint1] = struct{}{}
			}
			if weight > 0.0 {
				feo.AffectedByJoints[joint2] = struct{}{}
			}

			if haveNorm {
				normals = append(normals,
					float64(packet.Norms.X[iVertex]),
					float64(packet.Norms.Y[iVertex]),
					float64(packet.Norms.Z[iVertex]))
			}
			if haveRgba {
				rgba = append(rgba,
					uint8ColToF64(packet.Blend.R[iVertex]),
					uint8ColToF64(packet.Blend.G[iVertex]),
					uint8ColToF64(packet.Blend.B[iVertex]),
					uint8AlpfaToF64(packet.Blend.A[iVertex]))
			}
			if haveUV {
				uv = append(uv,
					float64(packet.Uvs.U[iVertex]), float64(-packet.Uvs.V[iVertex]))
			}
		}
	}

	name := fmt.Sprintf("p%d_g%d_o%d_m%d_i%d", feo.Part, feo.Group, feo.Object, feo.MaterialId, feo.InstanceId)

	feo.FbxGeometryId = f.GenerateId()

	geometryLayer := bfbx73.Layer(0).AddNodes(
		bfbx73.Version(100),
	)

	//geometry := bfbx73.Geometry(feo.FbxGeometryId, name+"\x00\x01Geometry", "Mesh").AddNodes(
	geometry := bfbx73.Geometry(feo.FbxGeometryId, "\x00\x01Geometry", "Mesh").AddNodes(
		bfbx73.Properties70().AddNodes(
			bfbx73.P("Color", "ColorRGB", "Color", "", float64(1), float64(1), float64(1)),
		),
		bfbx73.GeometryVersion(124),
		bfbx73.Vertices(vertices),
		bfbx73.PolygonVertexIndex(indexes),
		geometryLayer,
	)

	if haveNorm {
		geometry.AddNode(
			bfbx73.LayerElementNormal(0).AddNodes(
				bfbx73.Version(101),
				bfbx73.Name(""),
				bfbx73.MappingInformationType("ByVertice"),
				bfbx73.ReferenceInformationType("Direct"),
				bfbx73.Normals(normals),
			),
		)
		geometryLayer.AddNode(
			bfbx73.LayerElement().AddNodes(
				bfbx73.Type("LayerElementNormal"),
				bfbx73.TypedIndex(0),
			),
		)
	}

	if haveRgba {
		geometry.AddNode(
			bfbx73.LayerElementColor(0).AddNodes(
				bfbx73.Version(101),
				bfbx73.Name(""),
				bfbx73.MappingInformationType("ByVertice"),
				bfbx73.ReferenceInformationType("Direct"),
				bfbx73.Colors(rgba),
			),
		)
		geometryLayer.AddNode(
			bfbx73.LayerElement().AddNodes(
				bfbx73.Type("LayerElementColor"),
				bfbx73.TypedIndex(0),
			),
		)
	}

	if haveUV {
		geometry.AddNode(
			bfbx73.LayerElementUV(0).AddNodes(
				bfbx73.Version(101),
				bfbx73.Name(""),
				bfbx73.MappingInformationType("ByPolygonVertex"),
				bfbx73.ReferenceInformationType("IndexToDirect"),
				bfbx73.UV(uv),
				bfbx73.UVIndex(uvindexes),
			),
		)
		geometryLayer.AddNode(
			bfbx73.LayerElement().AddNodes(
				bfbx73.Type("LayerElementUV"),
				bfbx73.TypedIndex(0),
			),
		)
	}

	geometry.AddNode(
		bfbx73.LayerElementMaterial(0).AddNodes(
			bfbx73.Version(101),
			bfbx73.Name(""),
			bfbx73.MappingInformationType("AllSame"),
			bfbx73.ReferenceInformationType("IndexToDirect"),
			bfbx73.Materials([]int32{0}),
		),
	)
	geometryLayer.AddNode(
		bfbx73.LayerElement().AddNodes(
			bfbx73.Type("LayerElementMaterial"),
			bfbx73.TypedIndex(0),
		),
	)

	feo.VerticeToJoint = verticeToJoint
	feo.VerticeJointWeight = verticeJointWeight
	feo.FbxGeometry = geometry
	feo.FbxModelId = f.GenerateId()
	feo.FbxModel = bfbx73.Model(feo.FbxModelId, name+"\x00\x01Model", "Mesh").AddNodes(
		bfbx73.Version(232),
		bfbx73.Properties70().AddNodes(
			bfbx73.P("InheritType", "enum", "", "", int32(1)),
			bfbx73.P("DefaultAttributeIndex", "int", "Integer", "", int32(0)),
			bfbx73.P("Lcl Translation", "Lcl Translation", "", "A", float64(0), float64(0), float64(0)),
			bfbx73.P("Lcl Rotation", "Lcl Rotation", "", "A", float64(0), float64(0), float64(0)),
			bfbx73.P("Lcl Scaling", "Lcl Scaling", "", "A", float64(1), float64(1), float64(1)),
		),
		bfbx73.Shading(true),
		bfbx73.Culling("CullingOff"),
	)

	f.AddObjects(feo.FbxModel, geometry)
	f.AddConnections(bfbx73.C("OO", feo.FbxGeometryId, feo.FbxModelId))

	fep.Objects = append(fep.Objects, feo)
}

func (fe *FbxExporter) exportPart(f *fbxbuilder.FBXBuilder, fep *FbxExportPart) {
	part := &fe.m.Parts[fep.Part]
	fep.Objects = make([]*FbxExportObject, 0)
	fep.RawPart = part

	for iGroup := range part.Groups {
		group := &part.Groups[iGroup]
		for iObject := range group.Objects {
			object := &group.Objects[iObject]

			for iInstance := uint32(0); iInstance < object.InstancesCount; iInstance++ {
				feo := &FbxExportObject{
					Part:             fep.Part,
					Group:            iGroup,
					Object:           iObject,
					InstanceId:       int(iInstance),
					AffectedByJoints: make(map[uint16]struct{}),
				}
				fe.exportObject(f, feo, fep)
			}
		}
	}

	/*
		name := fmt.Sprintf("part%d", fep.Part)

		fep.FbxModelId = f.GenerateId()
		fep.FbxModel = bfbx73.Model(fep.FbxModelId, name+"\x00\x01Model", "Null").AddNodes(
			bfbx73.Version(232),
			bfbx73.Properties70().AddNodes(
				bfbx73.P("InheritType", "enum", "", "", int32(1)),
				bfbx73.P("DefaultAttributeIndex", "int", "Integer", "", int32(0)),
				bfbx73.P("Lcl Translation", "Lcl Translation", "", "A", float64(0), float64(0), float64(0)),
			),
			bfbx73.Shading(true),
			bfbx73.Culling("CullingOff"),
		)

		nodeAttribute := bfbx73.NodeAttribute(
			f.GenerateId(), name+"\x00\x01NodeAttribute", "Null").AddNodes(
			bfbx73.TypeFlags("Null"),
		)
	*/
	//f.AddObjects(fep.FbxModel, nodeAttribute)
	//f.AddConnections(bfbx73.C("OO", nodeAttribute.Properties[0].(int64), fep.FbxModelId))
	//for _, object := range fep.Objects {
	//f.AddConnections(bfbx73.C("OO", object.FbxModelId, fep.FbxModelId))
	//}

	fe.Parts = append(fe.Parts, fep)
}

func (m *Mesh) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbxbuilder.FBXBuilder) *FbxExporter {
	fe := &FbxExporter{
		m:     m,
		Parts: make([]*FbxExportPart, 0),
	}
	defer f.AddCache(wrsrc.Tag.Id, fe)

	for iPart := range m.Parts {
		fe.exportPart(f, &FbxExportPart{
			Part: iPart,
		})
	}
	fe.m = nil // free memory

	return fe
}

func (m *Mesh) ExportFbxDefault(wrsrc *wad.WadNodeRsrc) *fbxbuilder.FBXBuilder {
	f := fbxbuilder.NewFBXBuilder(filepath.Join(wrsrc.Wad.Name(), wrsrc.Name()))
	fe := m.ExportFbx(wrsrc, f)

	for _, part := range fe.Parts {
		//f.AddConnections(bfbx73.C("OO", part.FbxModelId, 0))
		for _, object := range part.Objects {
			f.AddConnections(bfbx73.C("OO", object.FbxModelId, 0))
		}
	}

	return f
}
