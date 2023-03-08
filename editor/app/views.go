package app

import (
	"go/ast"
	"go/parser"
	"log"

	"github.com/google/uuid"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/core"
	"github.com/mogaika/god_of_war_browser/editor/r3d"
)

type ProjectEditorView struct {
	selectedResource uuid.UUID

	v3d r3d.View3D

	filter  string
	expr    ast.Expr
	exprErr error
}

func (v *ProjectEditorView) RenderUI(p *core.Project, vStart, vEnd imgui.Vec2) {
	p.Presenter.UpdateIfNeeded(p)

	imgui.SetNextWindowPos(vStart)
	imgui.SetNextWindowSizeConstraints(imgui.Vec2{X: 64, Y: vEnd.Y}, imgui.Vec2{X: vEnd.X - vStart.X, Y: vEnd.Y})

	imgui.BeginV("Project view", nil, imgui.WindowFlagsNoCollapse)
	{
		pushed := v.exprErr != nil
		if pushed {
			imgui.PushStyleColor(imgui.StyleColorFrameBg, imgui.Vec4{X: 0.2, Y: 0, Z: 0, W: 1})
		}
		if imgui.InputTextV("filter", &v.filter, imgui.InputTextFlagsNone, nil) {
			log.Printf("%q", v.filter)

			if v.filter != "" {
				v.expr, v.exprErr = parser.ParseExpr(v.filter)
				log.Printf("Err: %v", v.exprErr)
			} else {
				v.expr, v.exprErr = nil, nil
			}
		}
		if pushed {
			imgui.PopStyleColor()
		}

		if v.expr != nil {
			core.ReflectView(p, v.expr)
		}

		var recursiveRender func(dir *core.ProjectDirectory)
		recursiveRender = func(dir *core.ProjectDirectory) {
			for _, subDir := range dir.Sub {
				if imgui.TreeNodeV(subDir.Name, imgui.TreeNodeFlagsNone) {
					recursiveRender(subDir)
					imgui.TreePop()
				}
			}
			// imgui.Separator()
			for _, r := range dir.Resources {
				imgui.PushID(r.UID.String())
				if imgui.SelectableV(r.Name, r.UID == v.selectedResource, 0, imgui.Vec2{}) {
					v.selectedResource = r.UID
				}
				if imgui.IsItemHovered() {
					ref := core.NewRef[core.ResourceWithTooltip](r.UID)
					imgui.BeginTooltip()
					if resource := ref.Resolve(p); resource != nil {
						resource.RenderTooltip(p)
					} else {
						imgui.Textf("%q", ref.Meta(p).Path())
						imgui.Textf("%q", ref.Uid().String())
					}
					imgui.EndTooltip()
				}
				imgui.PopID()
			}
		}
		recursiveRender(&p.Presenter.Root)
	}
	vStart.X += imgui.WindowWidth()
	imgui.End()

	imgui.SetNextWindowPos(vStart)
	imgui.SetNextWindowSizeConstraints(imgui.Vec2{X: 64, Y: vEnd.Y}, imgui.Vec2{X: vEnd.X - vStart.X, Y: vEnd.Y})

	imgui.BeginV("Resource view", nil, imgui.WindowFlagsNoCollapse)
	if v.selectedResource != uuid.Nil {
		ref := core.NewRef[core.Resource](v.selectedResource)

		path := ref.Meta(p).Path()
		imgui.InputTextV("path", &path, imgui.InputTextFlagsReadOnly, nil)

		ref.Resolve(p).RenderUI(p)
	}
	vStart.X += imgui.WindowWidth()
	imgui.End()

	imgui.SetNextWindowPos(vStart)
	imgui.SetNextWindowSize(imgui.Vec2{X: vEnd.X - vStart.X, Y: vEnd.Y})
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{})
	imgui.BeginV("Scene view", nil,
		imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoScrollWithMouse|imgui.WindowFlagsNoResize)
	// if v.selectedResource != uuid.Nil {
	{
		size := imgui.WindowSize()
		size = imgui.WindowContentRegionMax().Minus(imgui.WindowContentRegionMin())
		v.v3d.BeforeRender3D(int32(size.X), int32(size.Y))

		if v.selectedResource != uuid.Nil {
			ref := core.NewRef[core.ResourceWith3D](v.selectedResource)
			if resource := ref.Resolve(p); resource != nil {
				resource.Render3D(p, size)
			}
		}

		texture := v.v3d.AfterRender3D()
		imgui.ImageV(texture, size,
			imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0},
			imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1},
			imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
	}
	vStart.X += imgui.WindowWidth()
	imgui.End()
	imgui.PopStyleVar()
}
