package toc

import (
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
		if err := toc.readTocFile(); err != nil {
			log.Printf("[toc] Cannot parse toc file after updating toc file '%s': %v", name, err)
		}
	}()

	newSize := int64(len(b))
	f.encounters = make([]Encounter, 0)

	fs := toc.findFreeSpaceForFile(newSize)
	if fs == nil && false {
		log.Printf("[toc] There is no free space in paks, trying to remove file replicas (dups)")
		if err := toc.RemoveReplicas(); err != nil {
			return fmt.Errorf("[toc] Cannot remove replicas: %v", err)
		}
		fs = toc.findFreeSpaceForFile(newSize)
	}
	if fs == nil || true {
		log.Printf("[toc] There is no free space in paks, trying to shrink data and find place for file")
		if err := toc.Shrink(); err != nil {
			return fmt.Errorf("[toc] Cannot shrink files: %v", err)
		}
		fs = toc.findFreeSpaceForFile(newSize)
	}

	if fs == nil {
		return fmt.Errorf("[toc] There is no free space available in packs. WORKAROUND: Manually increase size of paks files and try again.")
	}

	f.size = newSize
	e := Encounter{
		Offset: fs.Start,
		Size:   newSize,
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

func (toc *TableOfContent) findFreeSpaceForFile(size int64) *FreeSpace {
	freeSpaces := constructFreeSpaceArray(toc.files, toc.paks)
	log.Printf("Free space table")
	for iFreeSpace := range freeSpaces {
		fs := &freeSpaces[iFreeSpace]
		log.Printf("  - free space %+#v", fs)
		if fs.End-fs.Start >= size {
			result := freeSpaces[iFreeSpace]
			return &result
		}
	}
	return nil
}

func (t *TableOfContent) RemoveReplicas() error {
	for _, f := range t.files {
		if len(f.encounters) > 1 {
			f.encounters = f.encounters[:1]
		}
	}
	return t.updateToc()
}

func (t *TableOfContent) Shrink() error {
	sortedFiles := sortFilesByEncounters(t.files)
	paksUsage := paksAsFreeSpaces(t.paks)
	alreadyProcessedFiles := make(map[string]*File)
	for _, f := range sortedFiles {
		if _, already := alreadyProcessedFiles[f.name]; !already {
			alreadyProcessedFiles[f.name] = f
			if len(f.encounters) != 0 {
				oldsencs := f.encounters
				oldE := oldsencs[0]

				f.encounters = make([]Encounter, 1, 1)
				newE := &f.encounters[0]
				newE.Size = oldE.Size

				for iPakUsage, pu := range paksUsage {
					if pu.End-pu.Start >= newE.Size {
						newE.Offset = pu.Start
						newE.Pak = pu.Pak
						paksUsage[iPakUsage].Start += utils.GetRequiredSectorsCount(oldE.Size) * utils.SECTOR_SIZE
						break
					}
				}

				log.Printf("Moving '%s' from %v to %v", f.name, oldE, newE)
				if err := t.pa.Move(oldE, *newE); err != nil {
					return fmt.Errorf("[toc] SHRINK ERROR! PROBABLY YOU LOSE YOUR DATA !: %v", err)
				}
			}
		}
	}
	return t.updateToc()
}

func (t *TableOfContent) updateToc() error {
	tocFile, err := t.findTocFile()
	if err != nil {
		return fmt.Errorf("[toc] updateToc: Cannot get dir element: %v", err)
	}
	if err := tocFile.Open(false); err != nil {
		return fmt.Errorf("[toc] updateToc: Cannot open file: %v", err)
	}
	defer tocFile.Close()
	if _, err := tocFile.WriteAt(t.Marshal(), 0); err != nil {
		return fmt.Errorf("[toc] updateToc: Error writing toc: %v", err)
	}
	return nil
}
