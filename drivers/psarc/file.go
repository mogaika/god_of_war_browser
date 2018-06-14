package psarc

import (
	"io"

	"github.com/mogaika/god_of_war_browser/vfs"
)

type File struct {
	p *Psarc
	e Entry
}

// interface vfs.Element
func (f *File) Init(parent vfs.Directory) {}
func (f *File) Name() string              { return f.e.Name }
func (f *File) IsDirectory() bool         { return false }

// interface vfs.File
func (f *File) Size() int64                                    { return f.e.OriginalSize }
func (f *File) Open(readonly bool) error                       { return nil }
func (f *File) Close() error                                   { return nil }
func (f *File) Reader() (*io.SectionReader, error)             { panic("Not implemented") }
func (f *File) ReadAt(b []byte, off int64) (n int, err error)  { panic("Not implemented") }
func (f *File) Copy(src io.Reader) error                       { panic("read-only") }
func (f *File) WriteAt(b []byte, off int64) (n int, err error) { panic("read-only") }
