package gow

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/core"
	"github.com/mogaika/god_of_war_browser/editor/rendercontext"
	"github.com/mogaika/god_of_war_browser/pack/wad/mat"
)

type MaterialLayer struct {
	og      mat.Layer
	color   [4]float32
	texture core.Ref[*Texture]
}

type Material struct {
	DefaultGroupMock

	Color  [4]float32
	Layers []*MaterialLayer
	OG     *mat.Material

	glInited bool
}

func (r *Material) useGL(p *core.Project) {
	rendercontext.Use(r)
	for _, layer := range r.Layers {
		if tex := layer.texture.Resolve(p); tex != nil {
			tex.useGL(p)
		}
	}

	if r.glInited {
		return
	}
	r.glInited = true
}

func (r *Material) ClearTempRenderData() {
	if !r.glInited {
		return
	}
	r.glInited = false
}

func (r *Material) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
	imgui.Separator()

	imgui.ColorEdit4V("##rgba_color", &r.Color,
		imgui.ColorEditFlagsFloat|imgui.ColorEditFlagsHDR|imgui.ColorEditFlagsNoLabel)

	imgui.Separator()
	for iLayer, layer := range r.Layers {
		if imgui.TreeNodeV(fmt.Sprintf("layer_%v", iLayer), imgui.TreeNodeFlagsDefaultOpen) {
			prefixWith := func(format string, v ...any) {
				w := imgui.CalcItemWidth()
				imgui.Textf(format, v...)
				imgui.SameLine()
				pos := imgui.CursorPos()
				imgui.SetCursorPos(imgui.Vec2{X: pos.X + w*0.5 + imgui.CurrentStyle().ItemInnerSpacing().X, Y: pos.Y})
				imgui.SetNextItemWidth(-1)
			}

			prefixWith("color")
			imgui.ColorEdit4V("##rgba_color", &layer.color,
				imgui.ColorEditFlagsFloat|imgui.ColorEditFlagsHDR|imgui.ColorEditFlagsNoLabel)

			prefixWith("f_unk")
			imgui.Textf("%v", layer.og.FloatUnk)

			if txr := layer.texture.Resolve(p); txr != nil {
				if txr.useGL(p) {
					for _, tid := range txr.glTextures {
						imgui.Image(imgui.TextureID(tid), txr.glSize.Times(2))
					}
				}
			}

			imgui.TreePop()
		}
	}
}

func (r *Material) RenderTooltip(p *core.Project) {
	for _, layer := range r.Layers {
		if txr := layer.texture.Resolve(p); txr != nil {
			txr.RenderTooltip(p)
		}
	}
}
