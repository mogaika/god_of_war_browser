package toc

import "github.com/mogaika/god_of_war_browser/vfs"
import "io"

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
	panic("Not implemented")
	return 0, nil
}

func (f *File) Copy(src io.Reader) error {
	panic("Not implemented")
	return nil
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	panic("Not implemented")
	return 0, nil
}
