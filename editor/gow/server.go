package gow

import (
	"log"

	"github.com/google/uuid"
	"github.com/mogaika/god_of_war_browser/editor/core"
)

type ServerInstanceResource interface {
	core.Resource
	WadGroupEnd(p *core.Project, deps []uuid.UUID)
}

type RequireNoGroupMock struct{}

func (r RequireNoGroupMock) WadGroupEnd(p *core.Project, deps []uuid.UUID) {
	if len(deps) > 0 {
		log.Printf("Loaded grouped resource while required nogroup")
		// panic(deps)
	}
}

type DefaultGroupMock struct {
	Deps []core.Ref[core.Resource]
}

func (r *DefaultGroupMock) WadGroupEnd(p *core.Project, deps []uuid.UUID) {
	for _, dep := range deps {
		r.Deps = append(r.Deps, core.NewRef[core.Resource](dep))
	}
}

// ------------ defaults --------------

type DefaultServerInstanceResource struct {
	DefaultGroupMock

	Name string

	Data []byte
}

func (r *DefaultServerInstanceResource) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
}

type UnresolvedReference struct {
}

func (r *UnresolvedReference) RenderUI(p *core.Project) {}

type Tweak struct {
	Tag  uint16
	Data []byte
}

func (r *Tweak) RenderUI(p *core.Project) {}
