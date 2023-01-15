package model

import (
	"io"
	"path"
	"sort"

	"github.com/google/uuid"
	"github.com/inkyblackness/imgui-go/v4"
)

type Loader struct {
}

func (l *Loader) LoadMeta(T any) error {
	panic("not implemented")
}

func (l *Loader) LoadBinary(stream string) (io.ReadSeekCloser, error) {
	panic("not implemented")
}

type Saver struct{}

func (s *Saver) SaveMeta(T any) error {
	panic("not implemented")
}

func (s *Saver) SaveBinary(stream string) (io.WriteCloser, error) {
	panic("not implemented")
}

type Ref[T any] struct {
	UUID uuid.UUID
}

func (ref Ref[T]) IsEmpty() bool {
	return ref.UUID == uuid.UUID{}
}

func (ref Ref[T]) Get(p *Project) *T {
	if ref.IsEmpty() {
		return nil
	}

	meta, exists := p.resourcesByUUID[ref.UUID]
	if !exists {
		ref.UUID = uuid.UUID{}
		return nil
	}

	p.ensureLoaded(meta)

	if r := meta.loaded; r != nil {
		return r.(*T)
	} else {
		return nil
	}
}

func (p *Project) ensureLoaded(meta *ResourceMeta) {
	if meta.loaded != nil {
		return
	}
	panic("not implemented")
}

type ResourceMeta struct {
	uid    uuid.UUID
	path   string
	loaded any
}

type Project struct {
	path string

	resourcesByUUID map[uuid.UUID]*ResourceMeta
	loadedResources map[uuid.UUID]*ResourceMeta

	presenter ProjectPresenter
}

type Controller interface {
	Load(Loader) (any, error)
}

func NewProject(path string) (*Project, error) {
	return &Project{
		path:            path,
		resourcesByUUID: make(map[uuid.UUID]*ResourceMeta),
		loadedResources: make(map[uuid.UUID]*ResourceMeta),
	}, nil
}

func (p *Project) AddResource(path string, r any) uuid.UUID {
	p.presenter.Set()

	uid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	meta := &ResourceMeta{
		uid:    uid,
		path:   path,
		loaded: r,
	}

	p.resourcesByUUID[uid] = meta
	p.loadedResources[uid] = meta
	return uid
}

func OpenProject(path string) (*Project, error) {
	panic("not implemented")
}

func (p *Project) Save() error {
	panic("not implemented")
}

type projectViewDirectory struct {
	name      string
	resources []string
	dirs      []*projectViewDirectory
}

type ProjectPresenter struct {
	Dirty
	root projectViewDirectory
}

func (pp *ProjectPresenter) UpdateIfNeeded(p *Project) {
	if !pp.Dirty.NeedUpdate() {
		return
	}
	pathMapping := make(map[string]*projectViewDirectory)
	var getDirectory func(p string) *projectViewDirectory
	getDirectory = func(p string) *projectViewDirectory {
		if p == "" || p == "." {
			return &pp.root
		}

		if dir := pathMapping[p]; dir != nil {
			return dir
		}

		newDir := &projectViewDirectory{
			name: path.Base(p),
		}
		pathMapping[p] = newDir

		parent := getDirectory(path.Dir(p))
		parent.dirs = append(parent.dirs, newDir)

		return newDir
	}

	pp.root = projectViewDirectory{}
	for _, m := range p.resourcesByUUID {
		pDir, name := path.Split(m.path)
		pDir = path.Clean(pDir)
		dir := getDirectory(pDir)
		dir.resources = append(dir.resources, name)
	}

	var sortDir func(*projectViewDirectory)
	sortDir = func(dir *projectViewDirectory) {
		sort.Slice(dir.dirs, func(i, j int) bool {
			return dir.dirs[i].name < dir.dirs[j].name
		})
		sort.Strings(dir.resources)
		for _, subDir := range dir.dirs {
			sortDir(subDir)
		}
	}
	sortDir(&pp.root)
}

type ProjectEditorView struct {
	ProjectTreeView
}

func (v *ProjectEditorView) RenderUI(p *Project, framebufferSize [2]float32) {
	p.presenter.UpdateIfNeeded(p)

	imgui.SetNextWindowPos(imgui.Vec2{0, 0})
	imgui.SetNextWindowSizeConstraints(imgui.Vec2{64, framebufferSize[1]}, imgui.Vec2{1000, framebufferSize[1]})

	imgui.BeginV("Project view", nil, imgui.WindowFlagsNoCollapse)
	{
		v.ProjectTreeView.RenderUI(p, framebufferSize)
	}
	imgui.End()
}

type ProjectTreeView struct{}

func (v *ProjectTreeView) RenderUI(p *Project, framebufferSize [2]float32) {
	p.presenter.UpdateIfNeeded(p)

	var recursiveRender func(dir *projectViewDirectory)
	recursiveRender = func(dir *projectViewDirectory) {
		for _, subDir := range dir.dirs {
			if imgui.TreeNodeV(subDir.name, imgui.TreeNodeFlagsNone) {
				recursiveRender(subDir)
				imgui.TreePop()
			}
		}
		// imgui.Separator()
		for _, name := range dir.resources {
			imgui.Selectable(name)
		}
	}
	recursiveRender(&p.presenter.root)
}

// Inverse logic, so by default we always dirty
type Dirty struct {
	notDirty bool
}

func (d *Dirty) Set() {
	d.notDirty = false
}

func (d *Dirty) NeedUpdate() (isNeed bool) {
	isNeed = !d.notDirty
	if isNeed {
		d.notDirty = true
	}
	return
}
