package gow

import (
	"github.com/google/uuid"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/core"
)

type Archive struct {
	Contexts []core.Ref[*GameContext]
	Other    []core.Ref[core.Resource]
}

func (r *Archive) WadGroupEnd(p *core.Project, deps []uuid.UUID) {
	for _, dep := range deps {
		if ref := core.NewRef[*GameContext](dep); ref.Resolve(p) != nil {
			r.Contexts = append(r.Contexts, ref)
		} else {
			r.Other = append(r.Other, core.NewRef[core.Resource](dep))
		}
	}
}

func (r *Archive) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
}

func (r *Archive) Render3D(p *core.Project, fbSize imgui.Vec2) {
	var ren GLRenderer
	var bbox helperBBox

	for _, rGameCtx := range r.Contexts {
		if gameCtx := rGameCtx.Resolve(p); gameCtx != nil {
			gameCtx.Render3DAdd(p, &ren, &bbox)
		}
	}

	ren.RenderCustom(p, fbSize, bbox.Center(), bbox.Size()*0.5)
}
