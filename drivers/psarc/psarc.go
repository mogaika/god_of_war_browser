package psarc

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mogaika/god_of_war_browser/vfs"
)

type Psarc struct {
	f      vfs.File
	h      Header
	r      *io.SectionReader
	enties []Entry
}

func (p *Psarc) parseHeader() error {
	var b [RAW_HEADER_SIZE]byte
	if _, err := p.f.ReadAt(b[:], 0); err != nil {
		return fmt.Errorf("[psarc] Header ReadAt(..): %v", err)
	}
	p.h.FromBuf(b[:])
	return nil
}

func (p *Psarc) parseEntries() error {
	b := make([]byte, p.h.NumFiles*RAW_ENTRY_SIZE)
	if _, err := p.f.ReadAt(b, RAW_HEADER_SIZE); err != nil {
		return fmt.Errorf("[psarc] Entries ReadAt(..): %v", err)
	}

	p.enties = make([]Entry, p.h.NumFiles)
	for i := range p.enties {
		p.enties[i].FromBuf(b[RAW_ENTRY_SIZE*i:])
	}

	return nil
}

func (p *Psarc) parseManifest() error {
	p.enties[0].Name = "manifest"

	zr, err := zlib.NewReader(
		io.NewSectionReader(p.r, p.enties[0].StartOffset, p.enties[1].StartOffset-p.enties[0].StartOffset))
	if err != nil {
		return err
	}
	defer zr.Close()

	if rawManifest, err := ioutil.ReadAll(zr); err != nil {
		return err
	} else {
		b := bytes.NewBuffer(rawManifest)
		for i := 1; i < int(p.h.NumFiles); i++ {
			name, _ := b.ReadString('\n')
			p.enties[i].Name = strings.TrimSuffix(strings.TrimPrefix(name, "/"), "\n")
		}
	}

	return nil
}

func NewPsarcDriver(f vfs.File) (*Psarc, error) {
	p := &Psarc{f: f}
	if r, err := f.Reader(); err != nil {
		return nil, err
	} else {
		p.r = r
	}
	if err := p.parseHeader(); err != nil {
		return nil, err
	}
	if err := p.parseEntries(); err != nil {
		return nil, err
	}
	if err := p.parseManifest(); err != nil {
		return nil, err
	}
	return p, nil
}

// interface vfs.Element
func (p *Psarc) Init(parent vfs.Directory) {}
func (p *Psarc) Name() string              { return p.f.Name() }
func (p *Psarc) IsDirectory() bool         { return true }

// interface vfs.Directory
func (p *Psarc) Add(e vfs.Element) error  { panic("read-only") }
func (p *Psarc) Remove(name string) error { panic("read-only") }

func (p *Psarc) List() ([]string, error) {
	result := make([]string, 0, p.h.NumFiles)
	for i := range p.enties {
		result = append(result, p.enties[i].Name)
	}
	return result, nil
}

func (p *Psarc) GetElement(name string) (vfs.Element, error) {
	for i := range p.enties {
		if p.enties[i].Name == name {
			return &File{e: p.enties[i], p: p}, nil
		}
	}
	return nil, os.ErrNotExist
}
