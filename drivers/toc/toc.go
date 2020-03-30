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
	dirty              bool
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
func (t *TableOfContent) Add(e vfs.Element) error { panic("Not implemented") }
func (t *TableOfContent) Remove(name string) error {
	if _, ok := t.files[name]; !ok {
		return fmt.Errorf("[toc] Cannot find file '%s' in toc", name)
	}
	t.dirty = true
	delete(t.files, name)
	if err := t.Sync(); err != nil {
		return fmt.Errorf("Sync error: %v", err)
	}
	return nil
}

func (toc *TableOfContent) Unmarshal(b []byte) error {
	if config.GetGOWVersion() == config.GOWunknown {
		if b[2] != 0 {
			log.Println("[toc] Detected gow version: GOW1")
			config.SetGOWVersion(config.GOW1)
		} else {
			log.Println("[toc] Detected gow version: GOW2")
			config.SetGOWVersion(config.GOW2)
		}
	}
	switch config.GetGOWVersion() {
	case config.GOW1:
		return toc.unmarshalGOW1(b)
	case config.GOW2:
		return toc.unmarshalGOW2(b)
	default:
		return fmt.Errorf("[toc] Unknown GOW version: %d", config.GetGOWVersion())
	}
}

func (toc *TableOfContent) Marshal() []byte {
	switch config.GetGOWVersion() {
	case config.GOW1:
		return toc.marshalGOW1()
	case config.GOW2:
		return toc.marshalGOW2()
	default:
		log.Panicf("[toc] Unknown GOW version: %v", config.GetGOWVersion())
		return nil
	}
}

func (t *TableOfContent) detectNamingPolicyTocOnly() (err error) {
	for _, policy := range defaultTocNamePair {
		t.namingPolicy = &policy
		if _, err = t.openTocFile(); err == nil {
			log.Printf("[toc] Detected toc file %q", policy.TocName)
			return nil
		} else {
			log.Printf("[toc] WARNING: Error opening possible toc file %q: %v", policy.TocName, err)
		}
	}
	return fmt.Errorf("[toc] Wasn't able to detect pak naming policy")
}

func (t *TableOfContent) detectNamingPolicyPakOnly() (err error) {
	for _, policy := range defaultTocNamePair {
		t.namingPolicy = &policy
		if err := t.openPakStreams(true); err == nil {
			log.Printf("[toc] Detected pack naming policy %+v", policy)
			return nil
		}
		log.Printf("[toc] WARNING: Error opening possible pak stream: %v", err)
	}
	return fmt.Errorf("[toc] Wasn't able to detect pak naming policy")
}

func (t *TableOfContent) openTocFile() (vfs.File, error) {
	return vfs.DirectoryGetFile(t.dir, t.namingPolicy.TocName)
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

func (t *TableOfContent) getMaximumPossiblePakIndex() PakIndex {
	max := PakIndex(0)

	if !t.namingPolicy.UseIndexing {
		return 0
	}

	switch t.packsArrayIndexing {
	case PACK_ADDR_INDEX:
		for _, f := range t.files {
			for _, e := range f.encounters {
				if e.Pak > max {
					max = e.Pak
				}
			}
		}
	case PACK_ADDR_ABSOLUTE:
		for i := PakIndex(0); ; i++ {
			if _, err := t.dir.GetElement(t.namingPolicy.GetPakName(i)); err != nil {
				break
			} else {
				max = i
			}
		}
	default:
		log.Panicf("Unknown pack array indexing: %v", t.packsArrayIndexing)
	}
	return max
}

func (t *TableOfContent) openPakStreams(readonly bool) error {
	if err := t.closePakStreams(); err != nil {
		return err
	}

	log.Printf("[toc] Opening streams (readonly: %v). Naming policy %+v", readonly, t.namingPolicy)

	maxPaks := t.getMaximumPossiblePakIndex()
	t.paks = make([]vfs.File, maxPaks+1)
	for i := range t.paks {
		name := t.namingPolicy.GetPakName(PakIndex(i))

		f, err := vfs.DirectoryGetFile(t.dir, name)
		if err != nil {
			log.Printf("[toc] [WARNING] Cannot get pak '%s': %v", name, err)
			if i == 0 {
				return fmt.Errorf("Wasn't able to get first pak %q: %v", name, err)
			}
			break
		} else {
			if err := f.Open(readonly); err != nil {
				log.Printf("[toc] [WARNING] Cannot open pak '%s': %v", name, err)
				if i == 0 {
					return fmt.Errorf("Wasn't able to open first pak %q: %v", name, err)
				}
			}
			t.paks[i] = f
			log.Printf("[toc] Opened pak '%s' (readonly: %v)", name, readonly)
		}
	}
	t.pa = NewPaksArray(t.paks, t.packsArrayIndexing)
	return nil
}

func (t *TableOfContent) readTocFile() error {
	t.files = make(map[string]*File)

	if f, err := t.openTocFile(); err != nil {
		return err
	} else {
		log.Printf("[toc] Used toc file %q", f.Name())
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
	log.Printf(" xXxXx Free space of files packed in toc/pak:")
	var totalFree int64

	for _, fs := range constructFreeSpaceArray(t.files, t.paks) {
		freeSpaceSize := fs.End - fs.Start
		log.Printf("[%d] 0x%.9x <=> 0x%.9x  0x%.7x  %dkB", fs.Pak, fs.Start, fs.End, freeSpaceSize, freeSpaceSize>>10)
		totalFree += freeSpaceSize

		// sanity check of files
		for _, f := range t.files {
			for _, e := range f.encounters {
				if e.Size != f.size {
					log.Printf("[WARNING] size  0x%.7x != 0x%.7x file '%v' offset %x:", e.Size, f.size, f.name, e)
				}
				if e.Pak == fs.Pak {
					if e.Offset+e.Size > fs.Start && e.Offset < fs.End {
						log.Printf("collision with file %s: 0x%.9x <=> 0x%.9x", f.name, e.Offset, e.Offset+e.Size)
					}
				}
			}
		}
	}
	if totalFree == 0 {
		log.Printf("       no free space found")
	}
	log.Printf(" xXxXx Total free space: 0x%x  %dkB  %dMb", totalFree, totalFree>>10, totalFree>>20)
}

func NewTableOfContent(dir vfs.Directory) (*TableOfContent, error) {
	t := &TableOfContent{
		files: nil,
		dir:   dir,
	}

	if err := t.detectNamingPolicyTocOnly(); err != nil {
		return nil, err
	}

	if err := t.readTocFile(); err != nil {
		return nil, err
	}

	if err := t.detectNamingPolicyPakOnly(); err != nil {
		return nil, err
	}

	if err := t.openPakStreams(true); err != nil {
		return nil, fmt.Errorf("Wasn't able to open streams: %v", err)
	}

	printFreeSpace(t)

	return t, nil
}
