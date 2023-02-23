package core

import (
	"io"
	"log"
	"path"
	"sort"

	"github.com/google/uuid"
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

	Presenter ProjectPresenter
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

func OpenProject(path string) (*Project, error) {
	panic("not implemented")
}

func (p *Project) Save() error {
	panic("not implemented")
}

type ProjectDirectory struct {
	Name      string
	Resources []string
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
		dir.Resources = append(dir.Resources, name)
		// log.Printf("adding %q at %q (%q)", name, dir.Name, pDir)
	}

	var sortDir func(*ProjectDirectory)
	sortDir = func(dir *ProjectDirectory) {
		sort.Slice(dir.Sub, func(i, j int) bool {
			return dir.Sub[i].Name < dir.Sub[j].Name
		})
		sort.Strings(dir.Resources)
		for _, subDir := range dir.Sub {
			sortDir(subDir)
		}
	}
	sortDir(&pp.Root)

	log.Printf("Presenter updated %+#v", pp.Root)
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
