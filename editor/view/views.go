package view

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/core"
)

type ProjectEditorView struct {
	ProjectTreeView
}

func (v *ProjectEditorView) RenderUI(p *core.Project, framebufferSize [2]float32) {
	p.Presenter.UpdateIfNeeded(p)

	imgui.SetNextWindowPos(imgui.Vec2{0, 0})
	imgui.SetNextWindowSizeConstraints(imgui.Vec2{64, framebufferSize[1]}, imgui.Vec2{1000, framebufferSize[1]})

	imgui.BeginV("Project view", nil, imgui.WindowFlagsNoCollapse)
	{
		v.ProjectTreeView.RenderUI(p, framebufferSize)
	}
	imgui.End()
}

type ProjectTreeView struct{}

func (v *ProjectTreeView) RenderUI(p *core.Project, framebufferSize [2]float32) {
	p.Presenter.UpdateIfNeeded(p)

	var recursiveRender func(dir *core.ProjectDirectory)
	recursiveRender = func(dir *core.ProjectDirectory) {
		for _, subDir := range dir.Sub {
			if imgui.TreeNodeV(subDir.Name, imgui.TreeNodeFlagsNone) {
				recursiveRender(subDir)
				imgui.TreePop()
			}
		}
		// imgui.Separator()
		for _, name := range dir.Resources {
			imgui.Selectable(name)
		}
	}
	recursiveRender(&p.Presenter.Root)
}

type ResourceView struct {
	resource *core.ResourceMeta
}
