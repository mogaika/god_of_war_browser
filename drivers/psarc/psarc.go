package psarc

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/vfs"
)

type Psarc struct {
	f          vfs.File
	h          Header
	r          *io.SectionReader
	blockSizes []uint32
	entries    []Entry
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

	p.entries = make([]Entry, p.h.NumFiles)
	for i := range p.entries {
		p.entries[i].FromBuf(b[RAW_ENTRY_SIZE*i:])
	}

	return nil
}

func (p *Psarc) parseBlockSizes() error {
	blockStorageLen := 2
	if p.h.BlockSize > 0x10000 {
		blockStorageLen = 3
		if p.h.BlockSize > 0x1000000 {
			blockStorageLen = 4
		}
	}

	sizesStartOffset := RAW_HEADER_SIZE + RAW_ENTRY_SIZE*int64(p.h.NumFiles)
	sizesBufLen := int64(p.h.TotalTOCSize) - sizesStartOffset
	sizesCount := sizesBufLen / int64(blockStorageLen)

	b := make([]byte, sizesBufLen)
	if _, err := p.f.ReadAt(b, sizesStartOffset); err != nil {
		return fmt.Errorf("[psarc] BlockSizes ReadAt(..): %v", err)
	}

	p.blockSizes = make([]uint32, sizesCount)
	for i := range p.blockSizes {
		switch blockStorageLen {
		case 2:
			p.blockSizes[i] = uint32(binary.BigEndian.Uint16(b[i*2:]))
		case 3:
			p.blockSizes[i] = utils.Read24bitUint(binary.BigEndian, b[i*3:])
		case 4:
			p.blockSizes[i] = uint32(binary.BigEndian.Uint32(b[i*4:]))
		}
	}

	return nil
}

func (p *Psarc) parseManifest() error {
	p.entries[0].Name = "manifest"

	zr, err := zlib.NewReader(
		io.NewSectionReader(p.r, p.entries[0].StartOffset, p.entries[1].StartOffset-p.entries[0].StartOffset))
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
			p.entries[i].Name = strings.TrimSuffix(strings.TrimPrefix(name, "/"), "\n")
			p.entries[i].Name = strings.Replace(p.entries[i].Name, "/", "_", -1)
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
	if p.h.CompressionMethod[0] != 0x7a {
		return nil, fmt.Errorf("Only zlib compression supported (%#+v)", p.h.CompressionMethod)
	}
	if err := p.parseEntries(); err != nil {
		return nil, err
	}
	if err := p.parseBlockSizes(); err != nil {
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
	for i := range p.entries {
		result = append(result, p.entries[i].Name)
	}
	return result, nil
}

func (p *Psarc) GetElement(name string) (vfs.Element, error) {
	for i := range p.entries {
		if p.entries[i].Name == name {
			return &File{e: p.entries[i], p: p}, nil
		}
	}
	return nil, os.ErrNotExist
}
