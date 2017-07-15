package pack

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
)

type PackDriver interface {
	GetFileNamesList() []string
	GetFile(fileName string) (PackFile, error)
	GetFileReader(fileName string) (PackFile, *io.SectionReader, error)
	GetInstance(fileName string) (interface{}, error)
	UpdateFile(fileName string, in *io.SectionReader) error
}

type PackFile interface {
	Name() string
	Size() int64
}

type FileLoader func(pf PackFile, r *io.SectionReader) (interface{}, error)

var gHandlers map[string]FileLoader = make(map[string]FileLoader, 0)

func SetHandler(format string, ldr FileLoader) {
	gHandlers[strings.ToUpper(format)] = ldr
}

func CallHandler(pf PackFile, r *io.SectionReader) (interface{}, error) {
	ext := strings.ToUpper(filepath.Ext(pf.Name()))

	if h, found := gHandlers[ext]; found {
		return h(pf, r)
	} else {
		return nil, fmt.Errorf("Cannot find handler for '%s' extension", ext)
	}
}

type InstanceCacheEntry struct {
	Name     string
	Instance interface{}
}

type InstanceCache struct {
	Cache [8]*InstanceCacheEntry
	Pos   int
}

func (ic *InstanceCache) getEntry(name string) *InstanceCacheEntry {
	for _, ce := range ic.Cache {
		if ce != nil && ce.Name == name {
			return ce
		}
	}
	return nil
}

func (ic *InstanceCache) Get(name string) interface{} {
	if e := ic.getEntry(name); e != nil {
		return e.Instance
	}
	return nil
}

func (ic *InstanceCache) Add(name string, val interface{}) {
	if e := ic.getEntry(name); e != nil {
		e.Instance = val
	} else {
		ic.Cache[ic.Pos] = &InstanceCacheEntry{
			Name:     name,
			Instance: val,
		}
		ic.Pos = (ic.Pos + 1) % len(ic.Cache)
		if ic.Pos == 0 {
			runtime.GC()
		}
	}
}
