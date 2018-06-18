package vfs

import (
	"io"
)

// must contain only metadata (filename) as long as possible
// (before List/Open/GetElement/Remove/Add calls)
type Element interface {
	Init(parent Directory)
	Name() string
	IsDirectory() bool
}

type File interface {
	Element
	Size() int64
	Open(readonly bool) error
	Close() error
	Reader() (*io.SectionReader, error)
	ReadAt(b []byte, off int64) (n int, err error)
	Copy(src io.Reader) error
	WriteAt(b []byte, off int64) (n int, err error)
}

type Directory interface {
	Element
	List() ([]string, error)
	GetElement(name string) (Element, error)
	Add(e Element) error
	Remove(name string) error
}

type Syncer interface {
	Sync() error
}
