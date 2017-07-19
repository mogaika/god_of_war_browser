package pack

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/mogaika/god_of_war_browser/tok"
	"github.com/mogaika/udf"
)

// Actually structs for second layer starts from 0x0fdf98000
// Because boot-space for iso/udf fs is 0x8000 we get 0x0fdf90000 as layer start
// But 0x0fdf92000 is end of part1 on my iso
// What i'm doin wrong?
var IsoSecondLayerStart int64 = 0x0fdf90000

type IsoDriver struct {
	Files       map[string]*tok.File
	IsoFile     *os.File
	IsoLayers   [2]*udf.Udf
	PackStreams [tok.PARTS_COUNT]*io.SectionReader
	IsoPath     string
	Cache       *InstanceCache
}

func (p *IsoDriver) openIsoFile(name string) (*udf.File, int) {
	name = strings.ToLower(name)
	for iLayer, layer := range p.IsoLayers {
		if layer != nil {
			for _, f := range layer.ReadDir(nil) {
				if strings.ToLower(f.Name()) == name {
					return &f, iLayer
				}
			}
		}
	}
	return nil, -1
}

func (p *IsoDriver) prepareStreams() error {
	if p.IsoFile == nil || p.IsoLayers[0] == nil {
		var err error
		if p.IsoFile, err = os.OpenFile(p.IsoPath, os.O_RDWR|os.O_SYNC, 0666); err != nil {
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
			if f, _ := p.openIsoFile(tok.GenPartFileName(i)); f != nil {
				p.PackStreams[i] = f.NewReader()
			} else {
				p.PackStreams[i] = nil
			}
		}
	}
	return nil
}

func (p *IsoDriver) parseFilesFromTok() error {
	if err := p.prepareStreams(); err != nil {
		return err
	}
	var err error
	tokIso, _ := p.openIsoFile(tok.FILE_NAME)
	p.Files, err = tok.ParseFiles(tokIso.NewReader())
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

func (p *IsoDriver) closeStreams() {
	for i := range p.IsoLayers {
		p.IsoLayers[i] = nil
	}
	log.Println("Close: ", p.IsoFile.Close())
	p.IsoFile = nil
}

type IsoFileWriterAt struct {
	f   *os.File
	off int64
}

func (ifw *IsoFileWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = ifw.f.WriteAt(p, ifw.off+off)
	log.Println("Writing at ", off, "=", ifw.off+off, " size:", len(p), " err:", err)
	return n, err
}

func (p *IsoDriver) openIsoFileWriterAt(file string) *IsoFileWriterAt {
	fstr, layer := p.openIsoFile(file)
	filestart := udf.SECTOR_SIZE * (int64(fstr.FileEntry().AllocationDescriptors[0].Location) + int64(fstr.Udf.PartitionStart()))
	if layer == 1 {
		filestart += IsoSecondLayerStart
	}
	log.Println("filestart ", file, filestart)
	return &IsoFileWriterAt{
		f:   p.IsoFile,
		off: filestart,
	}
}

func (p *IsoDriver) UpdateFile(fileName string, in *io.SectionReader) error {
	tokUdf, _ := p.openIsoFile(tok.FILE_NAME)
	tokOriginal, err := ioutil.ReadAll(tokUdf.NewReader())
	if err != nil {
		panic(err)
	}

	f := p.Files[fileName]

	defer func() {
		p.parseFilesFromTok()
		p.Cache = &InstanceCache{}
	}()

	var tokbuf bytes.Buffer

	var packStreams [tok.PARTS_COUNT]io.WriterAt
	for i := range packStreams {
		packStreams[i] = p.openIsoFileWriterAt(tok.GenPartFileName(i))
	}

	if err := tok.UpdateFile(bytes.NewBuffer(tokOriginal), &tokbuf, packStreams, f, in); err == nil {
		_, err = p.openIsoFileWriterAt(tok.FILE_NAME).WriteAt(tokbuf.Bytes(), 0)
	} else {
		panic(err)
	}
	return err
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
