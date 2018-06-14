package toc

import (
	"io"
	"sort"

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

func sortFilesByEncounters(files map[string]*File) []*File {
	result := make([]*File, 0, len(files))
	for _, f := range files {
		result = append(result, f)
	}

	sort.Slice(result, func(i int, j int) bool {
		// if one of encounter of i file earlier then
		// all encounters of j file
		lesser := result[i]
		bigger := result[j]

		for _, le := range lesser.encounters {
			isAllEncountersLess := true
			for _, be := range bigger.encounters {
				if le.Pak > be.Pak || (le.Pak == be.Pak && le.Offset >= be.Offset) {
					isAllEncountersLess = false
				}
			}
			if isAllEncountersLess {
				return true
			}
		}
		return false
	})

	return result
}
