package gow

import (
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/core"
	"github.com/mogaika/god_of_war_browser/editor/r3d"
	"github.com/mogaika/god_of_war_browser/editor/rendercontext"
	"github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	"github.com/mogaika/god_of_war_browser/pack/wad/mesh"
)

type Meshes struct {
	RequireNoGroupMock
	OG *mesh.Mesh

	glInited bool
	glMeshes []*glMesh
}

func (r *Meshes) useGL(p *core.Project) {
	rendercontext.Use(r)
	if r.glInited {
		return
	}
	r.glInited = true

	cm := r.OG.AsCommonMesh()
	for iPart, part := range cm.Parts {
		for iLod, lod := range part.LodGroups {
			for iObject, object := range lod.Objects {
				m := &glMesh{
					indexesCount:          int32(len(object.Indexes)),
					jointMaps:             object.JointMaps,
					materialId:            uint16(object.MaterialIndex),
					instancesCount:        object.InstancesCount,
					useInvertedBoneMatrix: r.OG.Parts[iPart].Groups[iLod].Objects[iObject].UseInvertedMatrix,
				}

				vertices := make([]glMeshVertex, len(object.Vertices))
				for i, pos := range object.Vertices {
					vertices[i].pos = pos.Position
					vertices[i].weight[0] = pos.JointWeights[0]
					vertices[i].weight[1] = pos.JointWeights[1]
					vertices[i].jointId[0] = uint8(pos.JointsIndexes[0])
					vertices[i].jointId[1] = uint8(pos.JointsIndexes[1])
				}

				if len(object.UVs) != 0 {
					for i, uv := range object.UVs[0] {
						vertices[i].uv = uv
					}
				}
				if len(object.BlendColors) != 0 {
					for i, rgba := range object.BlendColors[0] {
						vertices[i].rgba = [4]uint8{rgba.R, rgba.G, rgba.B, rgba.A}
					}
				} else {
					for i := range vertices {
						vertices[i].rgba = [4]uint8{0xff, 0xff, 0xff, 0xff}
					}
				}
				if len(object.Normals) != 0 {
					for i, normal := range object.Normals {
						vertices[i].normal = normal
					}
				}

				var vertex glMeshVertex
				stride := int(unsafe.Sizeof(vertex))
				dp := r3d.GetDefaultProgram()

				gl.GenVertexArrays(1, &m.glVAO)
				gl.BindVertexArray(m.glVAO)

				gl.GenBuffers(1, &m.glVBO)
				gl.BindBuffer(gl.ARRAY_BUFFER, m.glVBO)
				gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*stride, gl.Ptr(vertices), gl.STATIC_DRAW)

				gl.VertexAttribPointerWithOffset(uint32(dp.APosition), 3, gl.FLOAT, false, int32(stride), unsafe.Offsetof(vertex.pos))
				gl.EnableVertexAttribArray(uint32(dp.APosition))

				gl.VertexAttribPointerWithOffset(uint32(dp.AColor), 4, gl.UNSIGNED_BYTE, true, int32(stride), unsafe.Offsetof(vertex.rgba))
				gl.EnableVertexAttribArray(uint32(dp.AColor))
				/*
					gl.VertexAttribPointerWithOffset(uint32(dp.ANormal), 3, gl.FLOAT, false, int32(stride), unsafe.Offsetof(vertex.normal))
					gl.EnableVertexAttribArray(uint32(dp.ANormal))
				*/
				gl.VertexAttribPointerWithOffset(uint32(dp.AUV), 2, gl.FLOAT, false, int32(stride), unsafe.Offsetof(vertex.uv))
				gl.EnableVertexAttribArray(uint32(dp.AUV))

				gl.VertexAttribPointerWithOffset(uint32(dp.ABoneWeights), 2, gl.FLOAT, false, int32(stride), unsafe.Offsetof(vertex.weight))
				gl.EnableVertexAttribArray(uint32(dp.ABoneWeights))

				gl.VertexAttribIPointerWithOffset(uint32(dp.ABoneIndices), 2, gl.UNSIGNED_BYTE, int32(stride), unsafe.Offsetof(vertex.jointId))
				gl.EnableVertexAttribArray(uint32(dp.ABoneIndices))

				gl.GenBuffers(1, &m.glEBO)
				gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.glEBO)
				gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(object.Indexes), gl.Ptr(object.Indexes), gl.STATIC_DRAW)

				runtime.KeepAlive(vertices)

				r.glMeshes = append(r.glMeshes, m)
			}
		}
	}

	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
}

func (r *Meshes) ClearTempRenderData() {
	if !r.glInited {
		return
	}
	r.glInited = false

	for _, mesh := range r.glMeshes {
		gl.DeleteVertexArrays(1, &mesh.glVAO)
		gl.DeleteBuffers(1, &mesh.glVBO)
		gl.DeleteBuffers(1, &mesh.glEBO)
	}
	r.glMeshes = nil
}

func (r *Meshes) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
}

func (r *Meshes) Render3D(p *core.Project, fbSize imgui.Vec2) {
	r.useGL(p)

	var ren GLRenderer

	for _, mesh := range r.glMeshes {
		gl.BindVertexArray(mesh.glVAO)

		ren.AddCall(glRenderCall{
			meshBundle: r,
			mesh:       mesh,
			transform:  mgl32.Ident4(),
		})
	}

	ren.Render(p, fbSize)
}

type Model struct {
	OG *mdl.Model

	Materials []core.Ref[*Material]
	Meshes    []core.Ref[*Meshes]
	Other     []core.Ref[ServerInstanceResource]
}

func (r *Model) WadGroupEnd(p *core.Project, deps []uuid.UUID) {
	for _, dep := range deps {
		if ref := core.NewRef[*Material](dep); ref.Resolve(p) != nil {
			r.Materials = append(r.Materials, ref)
		} else if ref := core.NewRef[*Meshes](dep); ref.Resolve(p) != nil {
			r.Meshes = append(r.Meshes, ref)
		} else {
			r.Other = append(r.Other, core.NewRef[ServerInstanceResource](dep))
		}
	}
}

func (r *Model) useGL(p *core.Project) {
	rendercontext.Use(r)
	for _, mesh := range r.Meshes {
		mesh.Resolve(p).useGL(p)
	}
	for _, mat := range r.Materials {
		mat.Resolve(p).useGL(p)
	}
}

func (r *Model) ClearTempRenderData() {}

func (r *Model) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
}

func (r *Model) Render3DAdd(p *core.Project, ren *GLRenderer, object *Object, matrix mgl32.Mat4) {
	r.useGL(p)
	for _, meshBundleRef := range r.Meshes {
		meshBundle := meshBundleRef.Resolve(p)

		for _, mesh := range meshBundle.glMeshes {
			gl.BindVertexArray(mesh.glVAO)

			var material *Material
			if mesh.materialId < uint16(len(r.Materials)) {
				material = r.Materials[mesh.materialId].Resolve(p)
			}

			for instanceId := 0; instanceId < mesh.instancesCount; instanceId += 1 {
				ren.AddCall(glRenderCall{
					meshBundle: meshBundle,
					mesh:       mesh,
					material:   material,
					transform:  matrix,
					model:      r,
					object:     object,
					instanceId: uint8(instanceId),
				})
			}
		}
	}
}

func (r *Model) Render3D(p *core.Project, fbSize imgui.Vec2) {
	var ren GLRenderer

	r.Render3DAdd(p, &ren, nil, mgl32.Ident4())

	ren.Render(p, fbSize)
}
