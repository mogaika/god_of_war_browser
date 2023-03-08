package core

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

type RefI interface {
	Uid() uuid.UUID
	Exists() bool
	ResolveAny(p *Project) any
}

type Ref[T Resource] struct {
	uuid uuid.UUID
}

func NewRef[T Resource](target uuid.UUID) Ref[T] {
	return Ref[T]{uuid: target}
}

func (ref Ref[T]) Uid() uuid.UUID {
	return ref.uuid
}

func (ref Ref[T]) Exists() bool {
	return ref.uuid != uuid.UUID{}
}

func (ref Ref[T]) ResolveAny(p *Project) any {
	if !ref.Exists() {
		return nil
	}

	meta, exists := p.resourcesByUUID[ref.uuid]
	if !exists {
		return nil
	}

	p.ensureLoaded(meta)
	return meta.loaded
}

func (ref Ref[T]) Resolve(p *Project) T {
	if casted, ok := ref.ResolveAny(p).(T); ok {
		return casted
	} else {
		var nilT T
		return nilT
	}
}

func (ref Ref[T]) Meta(p *Project) *ResourceMeta {
	return p.resourcesByUUID[ref.uuid]
}

func (p *Project) ensureLoaded(meta *ResourceMeta) {
	if meta.loaded != nil {
		return
	}
	panic("not implemented")
}

type Resource interface {
	RenderUI(*Project)
}

type ResourceWithTooltip interface {
	Resource
	RenderTooltip(*Project)
}

type ResourceWith3D interface {
	Resource
	Render3D(p *Project, fbSize imgui.Vec2)
}

type ResourceMeta struct {
	uid    uuid.UUID
	path   string
	loaded Resource
}

func (rm *ResourceMeta) Uid() uuid.UUID { return rm.uid }

func (rm *ResourceMeta) Path() string { return rm.path }

type Project struct {
	path string

	resourcesByUUID map[uuid.UUID]*ResourceMeta
	loadedResources map[uuid.UUID]*ResourceMeta

	Presenter ProjectPresenter
}

func NewProject(path string) (*Project, error) {
	return &Project{
		path:            path,
		resourcesByUUID: make(map[uuid.UUID]*ResourceMeta),
		loadedResources: make(map[uuid.UUID]*ResourceMeta),
	}, nil
}

func (p *Project) AddResource(path string, r Resource) uuid.UUID {
	p.Presenter.Set()

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

func (p *Project) GetResourceByPath(path string) (uuid.UUID, bool) {
	for _, meta := range p.resourcesByUUID {
		if meta.path == path {
			return meta.uid, true
		}
	}
	return uuid.Nil, false
}

func OpenProject(path string) (*Project, error) {
	panic("not implemented")
}

func (p *Project) Save() error {
	panic("not implemented")
}

type ProjectDirectoryResource struct {
	Name string
	UID  uuid.UUID
}

type ProjectDirectory struct {
	Name      string
	Resources []ProjectDirectoryResource
	Sub       []*ProjectDirectory
}

type ProjectPresenter struct {
	Dirty
	Root ProjectDirectory
}

func (pp *ProjectPresenter) UpdateIfNeeded(p *Project) {
	if !pp.Dirty.NeedUpdate() {
		return
	}

	pathMapping := make(map[string]*ProjectDirectory)
	var getDirectory func(p string) *ProjectDirectory
	getDirectory = func(p string) *ProjectDirectory {
		if p == "" || p == "." || p == "/" {
			return &pp.Root
		}

		if dir := pathMapping[p]; dir != nil {
			return dir
		}

		newDir := &ProjectDirectory{
			Name: path.Base(p),
		}
		pathMapping[p] = newDir

		parent := getDirectory(path.Dir(p))
		parent.Sub = append(parent.Sub, newDir)

		return newDir
	}

	pp.Root = ProjectDirectory{}
	for _, m := range p.resourcesByUUID {
		pDir, name := path.Split(m.path)
		pDir = path.Clean(pDir)
		dir := getDirectory(pDir)
		dir.Resources = append(dir.Resources, ProjectDirectoryResource{
			Name: name,
			UID:  m.uid,
		})
		// log.Printf("adding %q at %q (%q)", name, dir.Name, pDir)
	}

	var sortDir func(*ProjectDirectory)
	sortDir = func(dir *ProjectDirectory) {
		sort.Slice(dir.Sub, func(i, j int) bool {
			return dir.Sub[i].Name < dir.Sub[j].Name
		})
		sort.Slice(dir.Resources, func(i, j int) bool {
			return dir.Resources[i].Name < dir.Resources[j].Name
		})
		for _, subDir := range dir.Sub {
			sortDir(subDir)
		}
	}
	sortDir(&pp.Root)

	// log.Printf("Presenter updated %+#v", pp.Root)
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
