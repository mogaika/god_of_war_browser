package toc

import (
	"bytes"
	"fmt"
	"log"

	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/vfs"
)

func (toc *TableOfContent) UpdateFile(name string, b []byte) error {
	f, ok := toc.files[name]
	if !ok {
		return fmt.Errorf("[toc] Cannot find file with name: '%s'", name)
	}

	if err := toc.openPakStreams(false); err != nil {
		return fmt.Errorf("[toc] UpdateFile=>openPakStreams: %v", err)
	}
	// we do not need them rw all time
	defer toc.openPakStreams(true)

	// refresh toc struct
	defer toc.readTocFile()

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

	if fs := toc.findFreeSpaceForFile(newSize); fs != nil {
		if err := toc.replicaFileInFreeSpace(f, *fs, b); err != nil {
			return fmt.Errorf("[toc] cannot create replica of file: %v", err)
		}
		return nil
	}

	log.Printf("[toc] There is no free space in paks, trying to shrink data and find place for file")

	if err := toc.Shrink(); err != nil {
		return fmt.Errorf("[toc] Cannot shrink files: %v", err)
	}

	if fs := toc.findFreeSpaceForFile(newSize); fs != nil {
		if err := toc.replicaFileInFreeSpace(f, *fs, b); err != nil {
			return fmt.Errorf("[toc] cannot create replica of file after shrinking: %v", err)
		}
		return nil
	}

	return fmt.Errorf("[toc] There is no free space available in packs. WORKAROUND: Manually increase size of paks files and try again.")
}

func (toc *TableOfContent) replicaFileInFreeSpace(f *File, fs FreeSpace, b []byte) error {
	e := Encounter{
		Offset: fs.Start,
		Size:   f.size,
		Pak:    fs.Pak}
	f.encounters = append(f.encounters, e)

	if _, err := toc.pa.NewReaderWriter(e).WriteAt(b, 0); err != nil {
		return fmt.Errorf("[toc] replicaFileInFreeSpace=>WriteAt: %v", err)
	}
	if err := toc.updateToc(); err != nil {
		return fmt.Errorf("[toc] replicaFileInFreeSpace=>updateToc: %v", err)
	}
	return nil
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
