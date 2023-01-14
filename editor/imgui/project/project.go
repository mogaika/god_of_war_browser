package project

import (
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/inkyblackness/imgui-go/v4"
)

type Project struct {
	path             string
	resources        map[uuid.UUID]*Resource
	selectedResource *Resource

	needUpdateByKind bool
	byKind           []kindedList
}

type kindedList struct {
	kind      Kind
	resources []*Resource
}

func NewProject(path string) *Project {
	return &Project{
		path:      path,
		resources: make(map[uuid.UUID]*Resource),
	}
}

func (s *Project) newUniqueId() uuid.UUID {
	for {
		newId, err := uuid.NewRandom()
		if err != nil {
			panic(err)
		}

		if _, exists := s.resources[newId]; !exists {
			return newId
		}
	}
}

func (s *Project) AddResource(name string, data IResource) *Resource {
	id := s.newUniqueId()
	r := &Resource{id: id, name: name, kind: data.Kind(), data: data}

	if true {
		// sanity check
		for _, r := range s.resources {
			if r.data == data {
				panic(r.GetName())
			}
		}
	}

	s.resources[id] = r
	s.needUpdateByKind = true
	return r
}

func (s *Project) updateByKind() {
	s.byKind = make([]kindedList, 0)
	kindMap := make(map[Kind][]*Resource)
	for _, r := range s.resources {
		kindMap[r.GetKind()] = append(kindMap[r.GetKind()], r)
	}
	for kind, resources := range kindMap {
		sort.Slice(resources, func(i, j int) bool {
			cmp := strings.Compare(resources[i].name, resources[j].name)
			if cmp != 0 {
				return cmp < 0
			}
			return strings.Compare(resources[i].id.String(), resources[i].id.String()) < 0
		})
		s.byKind = append(s.byKind, kindedList{
			kind:      kind,
			resources: resources,
		})
	}
	sort.Slice(s.byKind, func(i, j int) bool {
		return strings.Compare(string(s.byKind[i].kind), string(s.byKind[j].kind)) < 0
	})
}

func (s *Project) Get(id uuid.UUID) *Resource { return s.resources[id] }

func (s *Project) OpenResource(r *Resource) { s.selectedResource = r }

func (s *Project) RenderUI(framebufferSize [2]float32) {
	if s.needUpdateByKind {
		s.updateByKind()
		s.needUpdateByKind = false
	}

	imgui.SetNextWindowPos(imgui.Vec2{0, 0})
	imgui.SetNextWindowSizeConstraints(imgui.Vec2{64, framebufferSize[1]}, imgui.Vec2{1000, framebufferSize[1]})

	imgui.BeginV("Project view", nil, imgui.WindowFlagsNoCollapse)
	{
		for i, km := range s.byKind {
			imgui.PushIDInt(i)

			treeNodeFlags := imgui.TreeNodeFlagsNone // imgui.TreeNodeFlagsDefaultOpen
			if s.selectedResource != nil && s.selectedResource.kind == km.kind {
				treeNodeFlags |= imgui.TreeNodeFlagsSelected
			}
			if imgui.TreeNodeV(string(km.kind), treeNodeFlags) {
				for i, r := range km.resources {
					imgui.PushIDInt(i)

					if imgui.SelectableV(r.GetName(), r == s.selectedResource, imgui.SelectableFlagsNone, imgui.Vec2{}) {
						s.OpenResource(r)
					}
					imgui.SameLineV(340, 0)
					imgui.Selectable(r.GetID().String())

					imgui.PopID()
				}
				imgui.TreePop()
			}
			imgui.PopID()
		}
	}
	lastSize := imgui.WindowSize()
	imgui.End()

	if s.selectedResource != nil {
		imgui.SetNextWindowPos(imgui.Vec2{lastSize.X, 0})
		imgui.SetNextWindowSizeConstraints(imgui.Vec2{64, framebufferSize[1]}, imgui.Vec2{1000, framebufferSize[1]})
		imgui.BeginV("Resource view", nil, imgui.WindowFlagsNoCollapse)

		/*
			if imgui.TreeNodeV("Resource", imgui.TreeNodeFlagsDefaultOpen) {
				imgui.InputTextV("Name", &s.selectedResource.name, imgui.InputTextFlagsReadOnly, nil)
				imgui.Separator()
				imgui.TreePop()
			}
		*/
		imgui.InputTextV("Name", &s.selectedResource.name, imgui.InputTextFlagsReadOnly, nil)
		imgui.Separator()

		s.selectedResource.data.RenderUI()

		imgui.End()
	}
}

func UIReference(p *Project, label string, r **Resource) {
	imgui.PushID(label)

	imgui.Text(label)
	imgui.SameLine()
	if *r == nil {
		imgui.Button("none")
	} else {
		if imgui.Button((*r).GetName()) {
			p.OpenResource(*r)
		}
	}

	imgui.PopID()
}
