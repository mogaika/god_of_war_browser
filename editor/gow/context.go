package gow

import "github.com/mogaika/god_of_war_browser/editor/core"

type Context struct {
	RequireNoGroupMock
	Objects []core.Ref[*GameObject]
}

func (r *Context) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
}
