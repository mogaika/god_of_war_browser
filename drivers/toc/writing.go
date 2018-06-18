package toc

import (
	"bytes"
	"fmt"
	"log"

	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/vfs"
)

func (toc *TableOfContent) Sync() error {
	var result error
	for _, f := range toc.paks {
		if s, ok := f.(vfs.Syncer); ok {
			if err := s.Sync(); err != nil && result == nil {
				result = err
			}
		}
	}
	return result
}

func (toc *TableOfContent) UpdateFile(name string, b []byte) error {
	f, ok := toc.files[name]
	if !ok {
		return fmt.Errorf("[toc] Cannot find file with name: '%s'", name)
	}

	if err := toc.openPakStreams(false); err != nil {
		return fmt.Errorf("[toc] UpdateFile=>openPakStreams: %v", err)
	}
	defer func() {
		toc.Sync()
		toc.readTocFile()
		toc.openPakStreams(true)
	}()

	newSize := int64(len(b))
	if utils.GetRequiredSectorsCount(f.size) <= utils.GetRequiredSectorsCount(newSize) {
		// we can reuse space
		f.size = newSize
		for i := range f.encounters {
			f.encounters[i].Size = newSize

			if _, err := toc.pa.NewReaderWriter(f.encounters[i]).WriteAt(b, 0); err != nil {
				return fmt.Errorf("[toc] size <= oldsize, UpdateFile=>WriteAt: %v", err)
			}
		}

		if err := toc.updateToc(); err != nil {
			return fmt.Errorf("[toc] size <= oldsize, UpdateFile=>updateToc: %v", err)
		}
		return nil
	}

	f.encounters = make([]Encounter, 0)

	fs := toc.findFreeSpaceForFile(newSize)
	if fs == nil {
		log.Printf("[toc] There is no free space in paks, trying to shrink data and find place for file")
		if err := toc.Shrink(); err != nil {
			return fmt.Errorf("[toc] Cannot shrink files: %v", err)
		}

		fs = toc.findFreeSpaceForFile(newSize)
	}

	if fs != nil {
		e := Encounter{
			Offset: fs.Start,
			Size:   f.size,
			Pak:    fs.Pak}
		f.encounters = append(f.encounters, e)

		if _, err := toc.pa.NewReaderWriter(e).WriteAt(b, 0); err != nil {
			return fmt.Errorf("[toc] size > oldsize, UpdateFile=>WriteAt: %v", err)
		}
		if err := toc.updateToc(); err != nil {
			return fmt.Errorf("[toc] size > oldsize, UpdateFile=>updateToc: %v", err)
		}
		return nil
	}
	return fmt.Errorf("[toc] There is no free space available in packs. WORKAROUND: Manually increase size of paks files and try again.")
}

func (toc *TableOfContent) findFreeSpaceForFile(size int64) *FreeSpace {
	freeSpaces := constructFreeSpaceArray(toc.files, toc.paks)
	for iFreeSpace := range freeSpaces {
		fs := &freeSpaces[iFreeSpace]
		if fs.End-fs.Start >= size {
			result := freeSpaces[iFreeSpace]
			return &result
		}
	}
	return nil
}

func (t *TableOfContent) Shrink() error {
	panic("Not implemented")
}

func (t *TableOfContent) updateToc() error {
	tocFile, err := t.findTocFile()
	if err != nil {
		return fmt.Errorf("[toc] Cannot get dir element: %v", err)
	}

	return vfs.OpenFileAndCopy(tocFile, bytes.NewReader(t.Marshal()))
}
