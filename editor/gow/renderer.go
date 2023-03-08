package gow

import (
	"math"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/core"
	"github.com/mogaika/god_of_war_browser/editor/r3d"
)

type GLRenderer struct {
	frame glFrame
}

func (r *GLRenderer) ClearTempRenderData() {}

type glMesh struct {
	glVAO uint32
	glVBO uint32
	glEBO uint32

	indexesCount          int32
	materialId            uint16
	jointMaps             [][]uint32
	instancesCount        int
	useInvertedBoneMatrix bool
}

type glMeshVertex struct {
	pos     mgl32.Vec3
	normal  mgl32.Vec3
	uv      mgl32.Vec2
	rgba    [4]uint8
	jointId [4]uint8
	weight  [4]float32
}

type glRenderCall struct {
	mesh       *glMesh
	meshBundle *Meshes
	material   *Material
	model      *Model
	object     *Object
	transform  mgl32.Mat4

	// TODO:
	instanceId uint8
	// textureLayerId uint8
}

type hashViewElement struct {
	/*
		steps:
			1 - reflections render
			2 - stensil buffers to mask mirrors (do not draw depth + color)
			3 - sky render (use setnsil buffer to mask mirrors)
				TODO: write black full screen quad before sky to clear color but not touching reflections
			4 - world render: depth clear after sky, disable stensil

		hash:
			- 1 bit  0 - mirror, 1 - not mirror
			- 3 bits 0 - , 1 - (do not render shadow), 2 - , 3 - , 4 -
			- 3 bits 0 - , 1 - , 2 - , 3 - , 4 -
			- 1 bit  0 - , 1 -
			- 3 bits 0 - , 1 - , 2 - ,
	*/

	hash         uint64
	renderCallId int
}

type glFrame struct {
	queue    []glRenderCall
	hashView []hashViewElement
}

func (r *GLRenderer) AddCall(call glRenderCall) {
	r.frame.queue = append(r.frame.queue, call)
}

var rotationStartTime = time.Now()

func (r *GLRenderer) Render(p *core.Project, fbSize imgui.Vec2) {
	r.RenderCustom(p, fbSize, mgl32.Vec3{}, 200)
}

func (r *GLRenderer) RenderCustom(p *core.Project, fbSize imgui.Vec2, target mgl32.Vec3, distance float32) {
	dp := r3d.GetDefaultProgram()
	matricesBuffer := make([]mgl32.Mat4, 200)

	camera := r3d.NewOrbitController(
		target, distance,
		20, float32(math.Mod(time.Since(rotationStartTime).Seconds()*0.1, 1.0))*360,
	)

	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.DepthMask(true)

	projection := mgl32.Perspective(mgl32.DegToRad(50), fbSize.X/fbSize.Y, 1.0, 10000.0)

	view := camera.GetViewMatrix()
	projectViewMat := projection.Mul4(view)

	gl.UseProgram(dp.Program.Id)
	gl.UniformMatrix4fv(dp.UProjectView, 1, false, &projectViewMat[0])
	gl.Uniform4f(dp.UColor, 1.0, 1.0, 1.0, 1.0)
	gl.ActiveTexture(gl.TEXTURE0)

	for _, call := range r.frame.queue {
		gl.UniformMatrix4fv(dp.UModel, 1, false, &call.transform[0])

		hasTexture := false
		if mat := call.material; mat != nil && len(mat.Layers) != 0 {
			if txr := mat.Layers[0].texture.Resolve(p); txr != nil {
				gl.BindTexture(gl.TEXTURE_2D, txr.glTextures[0])
				hasTexture = true
			}
		}

		if !hasTexture {
			gl.BindTexture(gl.TEXTURE_2D, 0)
			gl.Uniform1i(dp.UUseTexture, 0)
		} else {
			gl.Uniform1i(dp.UUseTexture, 1)
		}

		if call.object != nil {
			o := call.object

			jointMap := call.mesh.jointMaps[call.instanceId]
			for iJoint, mappedJoint := range jointMap {
				if mappedJoint < call.meshBundle.OG.SkeletJoints {
					// matricesBuffer[iJoint] = o.OG.Joints[mappedJoint].BindToJointMat

					if call.mesh.useInvertedBoneMatrix {
						matricesBuffer[iJoint] = o.glRenderJoints[mappedJoint]
					} else {
						if iJoint < len(o.OG.Joints) { // check for ui
							matricesBuffer[iJoint] = o.OG.Joints[mappedJoint].ObjectToJoint
						}
					}
				} else {
					blendJoints := &call.meshBundle.OG.BlendJoints[mappedJoint-call.meshBundle.OG.SkeletJoints]
					var blendResult mgl32.Mat4
					for i, jId := range blendJoints.JointIds {
						if call.mesh.useInvertedBoneMatrix {
							blendResult.Add(o.glRenderJoints[jId].Mul(blendJoints.Weights[i]))
						} else {
							blendResult.Add(o.OG.Joints[jId].ObjectToJoint.Mul(blendJoints.Weights[i]))
						}
						// blendResult.Add(o.OG.Joints[jId].BindToJointMat.Mul(blendJoints.Weights[i]))
					}
					matricesBuffer[iJoint] = blendResult
				}
			}

			gl.Uniform1i(dp.UUseBones, 1)
			dp.SetBoneMatrices(matricesBuffer[:len(jointMap)])
		} else {
			gl.Uniform1i(dp.UUseBones, 0)
		}

		gl.BindVertexArray(call.mesh.glVAO)
		gl.DrawElements(gl.TRIANGLES, call.mesh.indexesCount, gl.UNSIGNED_INT, unsafe.Pointer(nil))
	}

	r.frame.queue = r.frame.queue[:0]
}
