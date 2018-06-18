package toc

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/mogaika/god_of_war_browser/vfs"
)

type File struct {
	name       string
	size       int64
	encounters []Encounter
	toc        *TableOfContent
}

// interface vfs.Element
func (f *File) Init(parent vfs.Directory) {}
func (f *File) Name() string              { return f.name }
func (f *File) IsDirectory() bool         { return false }

// interface vfs.File
func (f *File) Size() int64              { return f.size }
func (f *File) Open(readonly bool) error { return nil }
func (f *File) Close() error             { return nil }
func (f *File) Reader() (*io.SectionReader, error) {
	return io.NewSectionReader(f.toc.pa.NewReaderWriter(f.encounters[0]), 0, f.size), nil
}

func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	return f.toc.pa.NewReaderWriter(f.encounters[0]).ReadAt(b, off)
}

func (f *File) Copy(src io.Reader) error {
	if b, err := ioutil.ReadAll(src); err != nil {
		return fmt.Errorf("[toc] File Copy(..) ioutil.ReadAll: %v", err)
	} else {
		return f.toc.UpdateFile(f.name, b)
	}
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	panic("Not implemented")
}

func (f *File) Sync() error {
	return f.toc.Sync()
}
