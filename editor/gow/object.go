package gow

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/core"
	"github.com/mogaika/god_of_war_browser/editor/r3d"
	"github.com/mogaika/god_of_war_browser/editor/rendercontext"
	"github.com/mogaika/god_of_war_browser/pack/wad/inst"
	"github.com/mogaika/god_of_war_browser/pack/wad/obj"
)

type Object struct {
	OG *obj.Object

	Model core.Ref[*Model]
	Other []core.Ref[ServerInstanceResource]

	glInited       bool
	glRenderJoints []mgl32.Mat4
}

func (r *Object) WadGroupEnd(p *core.Project, deps []uuid.UUID) {
	for _, dep := range deps {
		if ref := core.NewRef[*Model](dep); ref.Resolve(p) != nil {
			if r.Model.Uid() != uuid.Nil {
				panic("multiple models")
			}
			r.Model = ref
		} else {
			r.Other = append(r.Other, core.NewRef[ServerInstanceResource](dep))
		}
	}
}

func (r *Object) useGL(p *core.Project) {
	rendercontext.Use(r)
	if r.glInited {
		return
	}
	r.glInited = true

	r.glRenderJoints = make([]mgl32.Mat4, len(r.OG.Joints))
	for iJoint := range r.OG.Joints {
		joint := &r.OG.Joints[iJoint]
		r.glRenderJoints[iJoint] = joint.ObjectToJoint.Mul4(joint.BindToJointMat)
	}

	if mdl := r.Model.Resolve(p); mdl != nil {
		mdl.useGL(p)
	}
}

func (r *Object) ClearTempRenderData() {
	if !r.glInited {
		return
	}
	r.glInited = false
}

func (r *Object) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
}

func (r *Object) Render3DAdd(p *core.Project, ren *GLRenderer, matrix mgl32.Mat4) {
	r.useGL(p)

	if mdl := r.Model.Resolve(p); mdl != nil {
		mdl.Render3DAdd(p, ren, r, matrix)
	}
}

func (r *Object) Render3D(p *core.Project, fbSize imgui.Vec2) {
	var ren GLRenderer

	r.Render3DAdd(p, &ren, mgl32.Ident4())

	ren.Render(p, fbSize)
}

type GameObject struct {
	RequireNoGroupMock
	OG *inst.Instance

	Object core.Ref[*Object]
	Other  []core.Ref[ServerInstanceResource]
}

func (r *GameObject) WadGroupEnd(p *core.Project, deps []uuid.UUID) {
	for _, dep := range deps {
		r.Other = append(r.Other, core.NewRef[ServerInstanceResource](dep))
	}
}

func (r *GameObject) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
}

func (r *GameObject) GetTransform() mgl32.Mat4 {
	tv := r.OG.Position1

	// glm.Mat4FromRotationTranslationScale()

	transform := mgl32.Ident4()
	transform = transform.Mul4(mgl32.Translate3D(tv.X(), tv.Y(), tv.Z()))

	rv := r.OG.Rotation
	quat := mgl32.AnglesToQuat(rv[0], rv[1], rv[2], mgl32.XYZ)
	transform = transform.Mul4(quat.Mat4())

	/*
		transform = transform.Mul4(mgl32.HomogRotate3DX(r.OG.Rotation[0]))
		transform = transform.Mul4(mgl32.HomogRotate3DY(r.OG.Rotation[1]))
		transform = transform.Mul4(mgl32.HomogRotate3DZ(r.OG.Rotation[2]))
	*/

	transform = transform.Mul4(mgl32.Scale3D(r.OG.Rotation[3], r.OG.Rotation[3], r.OG.Rotation[3]))

	return transform
}

func (r *GameObject) Render3DAdd(p *core.Project, ren *GLRenderer) {
	if o := r.Object.Resolve(p); o != nil {
		o.Render3DAdd(p, ren, r.GetTransform())
	}
}

func (r *GameObject) Render3D(p *core.Project, fbSize imgui.Vec2) {
	var ren GLRenderer

	r.Render3DAdd(p, &ren)

	ren.RenderCustom(p, fbSize, r.OG.Position2.Vec3(), 200)
}

func (r *GameObject) RenderTooltip(p *core.Project) {
	const PREVIEW_SIZE = 256
	var v3d r3d.View3D
	size := imgui.Vec2{X: PREVIEW_SIZE, Y: PREVIEW_SIZE}

	v3d.BeforeRender3D(int32(PREVIEW_SIZE), int32(size.Y))

	r.Render3D(p, size)

	texture := v3d.AfterRender3D()
	imgui.ImageV(texture, size,
		imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0},
		imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1},
		imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
}

type GameContext struct {
	Data []byte
	Name string

	Objects []core.Ref[*GameObject]
	Other   []core.Ref[ServerInstanceResource]
}

func (r *GameContext) WadGroupEnd(p *core.Project, deps []uuid.UUID) {
	for _, dep := range deps {
		if ref := core.NewRef[*GameObject](dep); ref.Resolve(p) != nil {
			r.Objects = append(r.Objects, ref)
		} else {
			r.Other = append(r.Other, core.NewRef[ServerInstanceResource](dep))
		}
	}
}

func (r *GameContext) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
}

func (r *GameContext) Render3DAdd(p *core.Project, ren *GLRenderer, bbox *helperBBox) {
	for _, object := range r.Objects {
		if o := object.Resolve(p); o != nil {
			if bbox != nil {
				if oo := o.Object.Resolve(p); oo != nil {
					if mdl := oo.Model.Resolve(p); mdl != nil {
						bbox.ExpandToPoint(o.OG.Position2.Vec3())
					}
				}
			}

			o.Render3DAdd(p, ren)
		}
	}
}

func (r *GameContext) Render3D(p *core.Project, fbSize imgui.Vec2) {
	var ren GLRenderer
	var bbox helperBBox

	r.Render3DAdd(p, &ren, &bbox)

	ren.RenderCustom(p, fbSize, bbox.Center(), bbox.Size()*0.5)
}
