package mesh

import (
	"fmt"

	"github.com/mogaika/god_of_war_browser/utils/gltfutils"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

type GLTFObjectExported struct {
	GLTFMesh      *gltf.Mesh
	GLTFMeshIndex uint32
	MaterialId    int
}

type GLTFMeshExported struct {
	Objects []*GLTFObjectExported
}

func (m *Mesh) ExportGLTF(wrsrc *wad.WadNodeRsrc, gltfCacher *gltfutils.GLTFCacher) (*GLTFMeshExported, error) {
	doc := gltfCacher.Doc
	tfme := &GLTFMeshExported{
		Objects: make([]*GLTFObjectExported, 0),
	}
	defer gltfCacher.AddCache(wrsrc.Tag.Id, tfme)

	cm := m.AsCommonMesh()

	for iPart, part := range cm.Parts {
		for iLodGroup, lodGroup := range part.LodGroups {
			for iObject, object := range lodGroup.Objects {
				verticesCount := len(object.Vertices)

				var positionAccessor, weightsAccessor, normalsAccessor, indicesAccessor uint32
				var uvAccessors, colorAccessors []uint32

				weights := make([][4]float32, verticesCount)
				{
					positions := make([][3]float32, verticesCount)
					for iVertex := range object.Vertices {
						positions[iVertex] = object.Vertices[iVertex].Position

						weights[iVertex] = [4]float32{
							object.Vertices[iVertex].Weight,
							1.0 - object.Vertices[iVertex].Weight,
							0, 0}
					}
					positionAccessor = modeler.WritePosition(doc, positions)
					weightsAccessor = modeler.WriteWeights(doc, weights)
				}

				{
					indices := make([]uint32, len(object.Indexes))
					for i, index := range object.Indexes {
						indices[i] = uint32(index)
					}
					indicesAccessor = modeler.WriteIndices(doc, indices)
				}

				{
					var normals [][3]float32
					if object.Normals != nil {
						normals = make([][3]float32, verticesCount)
						for iVertex, normal := range object.Normals {
							if normal.Len() > 0.5 {
								normal = normal.Normalize()
							}
							normals[iVertex] = normal
						}
						normalsAccessor = modeler.WriteNormal(doc, normals)
					}
				}

				{
					if object.UVs != nil {
						uvAccessors = make([]uint32, len(object.UVs))
						for iLayer := range object.UVs {
							uvs := make([][2]float32, verticesCount)
							for iVertex, uv := range object.UVs[iLayer] {
								uvs[iVertex] = [2]float32{uv[0], uv[1]}
							}
							uvAccessors[iLayer] = modeler.WriteTextureCoord(doc, uvs)
						}
					}
				}

				{
					if object.BlendColors != nil {
						colorAccessors = make([]uint32, len(object.BlendColors))
						for iLayer := range object.BlendColors {
							colors := make([][4]uint8, verticesCount)
							for iVertex, color := range object.BlendColors[iLayer] {
								colors[iVertex] = [4]uint8{color.R, color.G, color.B, color.A}
							}
							colorAccessors[iLayer] = modeler.WriteColor(doc, colors)
						}
					}
				}

				for iInstance := 0; iInstance < object.InstancesCount; iInstance++ {
					attributes := make(map[string]uint32)

					attributes["POSITION"] = positionAccessor
					attributes["WEIGHTS_0"] = weightsAccessor
					if object.Normals != nil {
						attributes["NORMAL"] = normalsAccessor
					}
					for iLayer := range uvAccessors {
						attributes[fmt.Sprintf("TEXCOORD_%d", iLayer)] = uvAccessors[iLayer]
					}
					if colorAccessors != nil {
						for iLayer := 0; iLayer < object.LayersCount; iLayer++ {
							attributes[fmt.Sprintf("COLOR_%d", iLayer)] = colorAccessors[iLayer]
						}
					}

					{
						instanceJointsMap := object.JointMaps[iInstance]
						joints := make([][4]uint16, verticesCount)
						for iVertex := range object.Vertices {
							jointIndexes := object.Vertices[iVertex].JointsIndexes
							joints[iVertex] = [4]uint16{
								uint16(instanceJointsMap[jointIndexes[0]]),
								uint16(instanceJointsMap[jointIndexes[1]]),
								0, 0}

							for i, weight := range weights[iVertex] {
								if weight == 0 {
									joints[iVertex][i] = 0
								}
							}
						}

						attributes["JOINTS_0"] = modeler.WriteJoints(doc, joints)
					}

					gltfMesh := &gltf.Mesh{
						Name: fmt.Sprintf("p%d_lod%d_o%d_i%d", iPart, iLodGroup, iObject, iInstance),
						Primitives: []*gltf.Primitive{
							&gltf.Primitive{
								Indices:    &indicesAccessor,
								Attributes: attributes,
							},
						},
					}

					doc.Meshes = append(doc.Meshes, gltfMesh)
					tfme.Objects = append(tfme.Objects,
						&GLTFObjectExported{
							GLTFMesh:      gltfMesh,
							GLTFMeshIndex: uint32(len(doc.Meshes) - 1),
							MaterialId:    object.MaterialIndex,
						},
					)
				}
			}
		}
	}

	return tfme, nil
}

func (m *Mesh) ExportGLTFDefault(wrsrc *wad.WadNodeRsrc) (*gltf.Document, error) {
	gltfCacher := gltfutils.NewCacher()
	doc := gltfCacher.Doc

	tfme, err := m.ExportGLTF(wrsrc, gltfCacher)
	if err != nil {
		return nil, err
	}

	doc.Materials = append(doc.Materials, &gltf.Material{
		Name:        "default",
		DoubleSided: true,
	})

	for _, object := range tfme.Objects {
		for _, primitive := range object.GLTFMesh.Primitives {
			primitive.Material = gltf.Index(0)
		}
		object.GLTFMesh.Name = wrsrc.Name() + object.GLTFMesh.Name
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(len(doc.Nodes)))
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: object.GLTFMesh.Name,
			Mesh: gltf.Index(object.GLTFMeshIndex),
		})
	}

	return doc, nil
}
