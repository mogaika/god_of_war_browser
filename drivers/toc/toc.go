package toc

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/vfs"
)

const TOC_FILE_NAME = "GODOFWAR.TOC"

type TableOfContent struct {
	dir                vfs.Directory
	files              map[string]*File
	paks               []vfs.File
	pa                 *PaksArray
	namingPolicy       *TocNamingPolicy
	packsArrayIndexing int // only for gow2
}

// interface vfs.Element
func (t *TableOfContent) Init(parent vfs.Directory) {}
func (t *TableOfContent) Name() string              { return "%TOC%" }
func (t *TableOfContent) IsDirectory() bool         { return true }

// interface vfs.Directory
func (t *TableOfContent) List() ([]string, error) {
	files := make([]string, 0, 256)
	for f := range t.files {
		files = append(files, f)
	}
	return files, nil
}

func (t *TableOfContent) GetElement(name string) (vfs.Element, error) {
	if f, ok := t.files[name]; !ok {
		return nil, fmt.Errorf("[toc] Cannot find file '%s' in toc", name)
	} else {
		return f, nil
	}
}
func (t *TableOfContent) Add(e vfs.Element) error  { panic("Not implemented") }
func (t *TableOfContent) Remove(name string) error { panic("Not impelemented") }

func (toc *TableOfContent) Unmarshal(b []byte) error {
	if config.GetGOWVersion() == config.GOWunknown {
		if b[2] != 0 {
			log.Println("[toc] Detected gow version: GOW1ps1")
			config.SetGOWVersion(config.GOW1ps2)
		} else {
			log.Println("[toc] Detected gow version: GOW2ps1")
			config.SetGOWVersion(config.GOW2ps2)
		}
	}
	switch config.GetGOWVersion() {
	case config.GOW1ps2:
		return toc.unmarshalGOW1(b)
	case config.GOW2ps2:
		return toc.unmarshalGOW2(b)
	default:
		return fmt.Errorf("[toc] Unknown GOW version: %d", config.GetGOWVersion())
	}
}

func (toc *TableOfContent) Marshal() []byte {
	switch config.GetGOWVersion() {
	case config.GOW1ps2:
		return toc.marshalGOW1()
	case config.GOW2ps2:
		return toc.marshalGOW2()
	default:
		log.Panicf("[toc] Unknown GOW version: %v", config.GetGOWVersion())
		return nil
	}
}

func (t *TableOfContent) findTocFile() (vfs.File, error) {
	for _, np := range defaultTocNamePair {
		f, err := vfs.DirectoryGetFile(t.dir, np.TocName)
		if err == nil {
			t.namingPolicy = &np
			return f, nil
		} else {
			log.Printf("[toc] Cannot open possible toc file '%s': %v", np.TocName, err)
		}
	}
	return nil, fmt.Errorf("[toc] Cannot find any of toc file")
}

func (t *TableOfContent) closePakStreams() error {
	errStrings := make([]string, 0, 8)
	if t.paks != nil {
		for _, p := range t.paks {
			if p != nil {
				if err := p.Close(); err != nil {
					errStrings = append(errStrings, fmt.Sprintf("Error closing '%s': %v", p.Name(), err))
				}
			}
		}
	}
	t.paks = nil
	if len(errStrings) == 0 {
		return nil
	} else {
		return fmt.Errorf("[toc] Cannot close streams: %v", strings.Join(errStrings, "; "))
	}
}

func (t *TableOfContent) getMaximumPossiblePak() PakIndex {
	max := PakIndex(0)
	for _, f := range t.files {
		for _, e := range f.encounters {
			if e.Pak > max {
				max = e.Pak
			}
		}
	}
	return max
}

func (t *TableOfContent) openPakStreams(readonly bool) error {
	if err := t.closePakStreams(); err != nil {
		return err
	}

	log.Printf("[toc] Opening streams (readonly: %v)", readonly)

	maxPaks := t.getMaximumPossiblePak()
	t.paks = make([]vfs.File, maxPaks+1)
	for i := range t.paks {
		name := t.namingPolicy.GetPakName(PakIndex(i))

		f, err := vfs.DirectoryGetFile(t.dir, name)
		if err != nil {
			log.Printf("[toc] [WARNING] Cannot get pak '%s': %v", name, err)
			break
		} else {
			if err := f.Open(readonly); err != nil {
				log.Printf("[toc] [WARNING] Cannot open pak '%s': %v", name, err)
				break
			}
			t.paks[i] = f
			log.Printf("[toc] Opened pak '%s'", name)
		}
	}
	t.pa = NewPaksArray(t.paks, t.packsArrayIndexing)
	return nil
}

func (t *TableOfContent) readTocFile() error {
	t.files = make(map[string]*File)

	if f, err := t.findTocFile(); err != nil {
		return err
	} else {
		if r, err := vfs.OpenFileAndGetReader(f, true); err != nil {
			return fmt.Errorf("[toc] Cannot open '%s': %v", f.Name(), err)
		} else {
			defer f.Close()
			if rawToc, err := ioutil.ReadAll(r); err != nil {
				return fmt.Errorf("[toc] ioutil.ReadAll(r): %v", err)
			} else {
				if err := t.Unmarshal(rawToc); err != nil {
					return fmt.Errorf("[toc]: Cannot unmarhsal: %v", err)
				}
			}
		}
	}
	return nil
}

func printFreeSpace(t *TableOfContent) {
	log.Printf("Free space: ")
	for _, fs := range constructFreeSpaceArray(t.files, t.paks) {
		log.Printf("[%d] 0x%.9x <=> 0x%.9x  0x%.7x  %dkB", fs.Pak, fs.Start, fs.End, fs.End-fs.Start, (fs.End-fs.Start)/1024)
		for _, f := range t.files {
			for _, e := range f.encounters {
				if e.Size != f.size {
					log.Printf("size  0x%.7x != 0x%.7x file '%v' offset %x:", e.Size, f.size, f.name, e)
				}
				if e.Pak == fs.Pak {
					if e.Offset+e.Size > fs.Start && e.Offset < fs.End {
						log.Printf("collision with file %s: 0x%.9x <=> 0x%.9x", f.name, e.Offset, e.Offset+e.Size)
					}
				}
			}
		}
	}
}

func NewTableOfContent(dir vfs.Directory) (*TableOfContent, error) {
	t := &TableOfContent{
		files: nil,
		dir:   dir,
	}

	if err := t.readTocFile(); err != nil {
		return nil, err
	}

	if err := t.openPakStreams(true); err != nil {
		return nil, err
	}

	// printFreeSpace(t)

	return t, nil
}
