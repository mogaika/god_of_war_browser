package pack

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mogaika/udf"
)

// Actually structs for second layer starts from 0x0fdf98000
// Because boot-space for iso/udf fs is 0x8000 we get 0x0fdf90000 as layer start
// But 0x0fdf92000 is end of part1 on my iso
// What i'm doin wrong?
var IsoSecondLayerStart int64 = 0x0fdf90000

type IsoDriver struct {
	Files       map[string]*TokFile
	IsoFile     *os.File
	IsoLayers   [2]*udf.Udf
	PackStreams [TOK_PARTS_COUNT]*io.SectionReader
	IsoPath     string
	Cache       *InstanceCache
}

func (p *IsoDriver) openIsoFile(name string) *udf.File {
	name = strings.ToLower(name)
	for _, layer := range p.IsoLayers {
		if layer != nil {
			for _, f := range layer.ReadDir(nil) {
				if strings.ToLower(f.Name()) == name {
					return &f
				}
			}
		}
	}
	return nil
}

func (p *IsoDriver) prepareStreams() error {
	if p.IsoFile == nil || p.IsoLayers[0] == nil {
		var err error
		if p.IsoFile, err = os.Open(p.IsoPath); err != nil {
			return err
		}

		p.IsoLayers[0] = udf.NewUdfFromReader(p.IsoFile)

		isoinfo, err := p.IsoFile.Stat()
		if err == nil {
			if isoinfo.Size() > IsoSecondLayerStart {
				r := io.NewSectionReader(p.IsoFile, IsoSecondLayerStart, isoinfo.Size()-IsoSecondLayerStart)
				log.Printf("[pack] Detected second layer of iso file (size:%d)", r.Size())
				p.IsoLayers[1] = udf.NewUdfFromReader(r)
			}
		}

		for i := range p.PackStreams {
			if f := p.openIsoFile(fmt.Sprintf("PART%d.PAK", i+1)); f != nil {
				p.PackStreams[i] = f.NewReader()
			} else {
				p.PackStreams[i] = nil
			}
		}
	}
	return nil
}

func (p *IsoDriver) parseFilesFromTok() error {
	var err error
	tok := p.openIsoFile("GODOFWAR.TOC").NewReader()
	p.Files, err = tokPartsParseFiles(tok)
	return err
}

func (p *IsoDriver) GetFileNamesList() []string {
	return getFileNamesListFromTokMap(p.Files)
}

func (p *IsoDriver) GetFile(fileName string) (PackFile, error) {
	return p.Files[fileName], nil
}

func (p *IsoDriver) GetFileReader(fileName string) (PackFile, *io.SectionReader, error) {
	if err := p.prepareStreams(); err != nil {
		return nil, nil, err
	}
	if f, exists := p.Files[fileName]; exists {
		for packNumber := range p.PackStreams {
			if p.PackStreams[packNumber] != nil {
				for _, enc := range f.Encounters {
					if enc.Pack == packNumber {
						return f, io.NewSectionReader(p.PackStreams[packNumber], enc.Start, f.Size()), nil
					}
				}
			}
		}
		return f, nil, fmt.Errorf("Cannot open stream for '%s'", fileName)
	} else {
		return nil, nil, fmt.Errorf("File '%s' not found", fileName)
	}
}

func (p *IsoDriver) GetInstance(fileName string) (interface{}, error) {
	return defaultGetInstanceCachedHandler(p, p.Cache, fileName)
}

func (p *IsoDriver) UpdateFile(fileName string, in *io.SectionReader) error {
	return fmt.Errorf("Not supported yet")
}

func NewPackFromIso(isoPath string) (*IsoDriver, error) {
	p := &IsoDriver{
		IsoPath: isoPath,
		Cache:   &InstanceCache{},
	}
	if err := p.prepareStreams(); err != nil {
		return nil, err
	}
	return p, p.parseFilesFromTok()
}
