package tocdriver

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/toc"
	"github.com/mogaika/god_of_war_browser/utils"
)

type TocDriver struct {
	Files      map[string]*toc.File
	Streams    []*os.File
	PacksCount int
	Directory  string
	Cache      *pack.InstanceCache
}

func (p *TocDriver) GetFileNamesList() []string {
	return GetFileNamesListFromTocMap(p.Files)
}

func GetFileNamesListFromTocMap(files map[string]*toc.File) []string {
	result := make([]string, len(files))
	i := 0
	for name := range files {
		result[i] = name
		i++
	}
	return result
}

func (p *TocDriver) tocGetFileName() string {
	return filepath.Join(p.Directory, toc.GetTocFileName())
}

func (p *TocDriver) partGetFileName(packNumber int) string {
	return filepath.Join(p.Directory, toc.GenPartFileName(packNumber))
}

func (p *TocDriver) prepareStream(packNumber int) error {
	if p.Streams[packNumber] == nil {
		if f, err := os.Open(p.partGetFileName(packNumber)); err != nil {
			return err
		} else {
			p.Streams[packNumber] = f
		}
	}
	return nil
}

func (p *TocDriver) closeStreams() {
	for i, f := range p.Streams {
		if f != nil {
			f.Close()
		}
		p.Streams[i] = nil
	}
}

func (p *TocDriver) getFile(fileName string) (*toc.File, error) {
	if f, exists := p.Files[fileName]; exists {
		return f, nil
	} else {
		return nil, fmt.Errorf("Cannot find '%s' file in pack", fileName)
	}
}

func (p *TocDriver) GetFile(fileName string) (pack.PackFile, error) {
	return p.getFile(fileName)
}

func (p *TocDriver) GetFileReader(fileName string) (pack.PackFile, *io.SectionReader, error) {
	if f, err := p.getFile(fileName); err == nil {
		for packNumber := range p.Streams {
			for _, enc := range f.Encounters {
				if enc.Pack == packNumber {
					if err := p.prepareStream(packNumber); err != nil {
						log.Printf("WARNING: Cannot open pack stream %d: %v", packNumber, err)
					}
					return f, io.NewSectionReader(p.Streams[packNumber], enc.Start, f.Size()), nil
				}
			}
		}
		return f, nil, fmt.Errorf("Cannot open stream for '%s'", fileName)
	} else {
		return nil, nil, err
	}
}

func (p *TocDriver) GetInstance(fileName string) (interface{}, error) {
	return pack.GetInstanceCachedHandler(p, p.Cache, fileName)
}

func (p *TocDriver) UpdateFile(fileName string, in *io.SectionReader) error {
	defer p.parseTocFile()

	f, err := p.getFile(fileName)
	if err != nil {
		return err
	}
	p.closeStreams()

	fParts := make([]*os.File, len(p.Streams))
	partWriters := make([]utils.ReaderWriterAt, len(p.Streams))
	defer func() {
		for _, part := range fParts {
			if part != nil {
				part.Close()
			}
		}
	}()
	for iPart := range fParts {
		if part, err := os.OpenFile(p.partGetFileName(iPart), os.O_RDWR, 0666); err == nil {
			fParts[iPart] = part
			partWriters[iPart] = utils.NewReaderWriterAtFromFile(part)
		} else {
			return fmt.Errorf("Cannot open '%s' for writing: %v", p.partGetFileName(iPart), err)
		}
	}

	fToc, err := os.OpenFile(p.tocGetFileName(), os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("Cannot open tocfile '%s' for writing: %v", p.tocGetFileName(), err)
	}
	defer fToc.Close()

	ftocoriginal, _ := ioutil.ReadAll(fToc)
	fToc.Seek(0, os.SEEK_SET)

	err = toc.UpdateFile(bytes.NewReader(ftocoriginal), fToc, partWriters, f, in)

	p.Cache = &pack.InstanceCache{}

	return err
}

func (p *TocDriver) parseTocFile() error {
	if tocStream, err := os.Open(p.tocGetFileName()); err == nil {
		defer tocStream.Close()
		log.Printf("[pack] Parsing toc '%s'", p.tocGetFileName())
		var tocEntries []toc.Entry
		p.Files, tocEntries, err = toc.ParseFiles(tocStream)
		p.PacksCount = toc.GetPacksCount(tocEntries)
		p.Streams = make([]*os.File, p.PacksCount)
		return err
	} else {
		return err
	}
}

func NewPackFromToc(gamePath string) (*TocDriver, error) {
	p := &TocDriver{
		Directory: gamePath,
		Cache:     &pack.InstanceCache{},
	}

	return p, p.parseTocFile()
}
