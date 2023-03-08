package gow

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/core"
	"github.com/mogaika/god_of_war_browser/pack/wad/gfx"
)

type GFX struct {
	RequireNoGroupMock
	OG *gfx.GFX
}

func (r *GFX) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
	imgui.Separator()

	imgui.Selectable(fmt.Sprintf("psm: %q", r.OG.Psm))
	imgui.Selectable(fmt.Sprintf("W: %v", r.OG.Width))
	imgui.Selectable(fmt.Sprintf("H: %v", r.OG.Height))
	imgui.Selectable(fmt.Sprintf("realH: %v", r.OG.RealHeight))
	imgui.Selectable(fmt.Sprintf("encoding: 0x%x", r.OG.Encoding))
	imgui.Selectable(fmt.Sprintf("bpi: %v", r.OG.Bpi))
	imgui.Selectable(fmt.Sprintf("dataSize: %v", r.OG.DataSize))
	imgui.Selectable(fmt.Sprintf("datasCount: %v", len(r.OG.Data)))
}
