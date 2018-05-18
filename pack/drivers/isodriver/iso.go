package isodriver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/pack/drivers/tocdriver"
	"github.com/mogaika/god_of_war_browser/toc"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/udf"
)

type IsoDriver struct {
	Files            map[string]*toc.File
	PaksCount        int
	IsoFile          *os.File
	IsoLayers        [2]*udf.Udf
	PackStreams      []*io.SectionReader
	IsoPath          string
	Cache            *pack.InstanceCache
	SecondLayerStart int64
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

func (p *IsoDriver) prepareIsoStreams() error {
	if p.IsoFile == nil || p.IsoLayers[0] == nil {
		var err error
		if p.IsoFile, err = os.OpenFile(p.IsoPath, os.O_RDWR|os.O_SYNC, 0666); err != nil {
			return err
		}

		p.IsoLayers[0] = udf.NewUdfFromReader(p.IsoFile)

		isoinfo, err := p.IsoFile.Stat()

		// Hack that allow detect second layer volume
		if err == nil {
			var volSizeBuf [4]byte
			// primary volume description sector + offset of volume space size
			if _, err := p.IsoFile.ReadAt(volSizeBuf[:], 0x10*2048+80); err != nil {
				log.Printf("[pack] Error when detecting second layer: Read vol size buf error: %v", err)
			} else {
				// minus 16 boot sectors, because they do not replicated over layers (volumes)
				volumeSize := int64(binary.LittleEndian.Uint32(volSizeBuf[:])-16) * utils.SECTOR_SIZE
				if volumeSize+32*utils.SECTOR_SIZE < isoinfo.Size() {
					p.IsoLayers[1] = udf.NewUdfFromReader(io.NewSectionReader(p.IsoFile, volumeSize, isoinfo.Size()-volumeSize))
					log.Printf("[pack] Detected second layer of disk. Start: %x (%x)", volumeSize+16*utils.SECTOR_SIZE, volumeSize)
					p.SecondLayerStart = volumeSize

					/*
						TODO: remove and and function to dump packs in folder
						f, err := os.Create("PART2.PAK")
						defer f.Close()
						if err != nil {
							panic(err)
						}
						io.Copy(f, io.NewSectionReader(p.IsoFile, volumeSize, isoinfo.Size()-volumeSize))
					*/
				}
			}
		} else {
			log.Printf("[pack] Cannot stat iso file: %v", err)
		}
	}
	return nil
}

func (p *IsoDriver) preparePackStreams() error {
	if p.PaksCount == 0 {
		panic("p.PaksCount == 0, probably you have some problems with reading toc file or pack files naming")
	}
	if p.PackStreams != nil {
		return nil
	}
	log.Println("[pack] Preparing pack streams")
	p.PackStreams = make([]*io.SectionReader, p.PaksCount)
	for i := range p.PackStreams {
		packName := toc.GenPartFileName(i)
		if f, _ := p.openIsoFile(packName); f != nil {
			p.PackStreams[i] = f.NewReader()
		} else {
			log.Println("[pack] WARNING: Cannot open pack stream '%s'", packName)
			p.PackStreams[i] = nil
		}
	}
	return nil
}

func (p *IsoDriver) parseFilesFromTok() error {
	var err error
	var tocEntries []toc.Entry

	if err := p.prepareIsoStreams(); err != nil {
		return err
	}

	tocIso, _ := p.openIsoFile(toc.GetTocFileName())

	if p.Files, tocEntries, err = toc.ParseFiles(tocIso.NewReader()); err != nil {
		return err
	}

	p.PaksCount = toc.GetPacksCount(tocEntries)
	return p.preparePackStreams()
}

func (p *IsoDriver) GetFileNamesList() []string {
	return tocdriver.GetFileNamesListFromTocMap(p.Files)
}

func (p *IsoDriver) GetFile(fileName string) (pack.PackFile, error) {
	return p.Files[fileName], nil
}

func (p *IsoDriver) GetFileReader(fileName string) (pack.PackFile, *io.SectionReader, error) {
	if err := p.prepareIsoStreams(); err != nil {
		return nil, nil, err
	}
	if err := p.preparePackStreams(); err != nil {
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
	return pack.GetInstanceCachedHandler(p, p.Cache, fileName)
}

func (p *IsoDriver) closeStreams() {
	for i := range p.IsoLayers {
		p.IsoLayers[i] = nil
	}
	log.Println("[pack] Closing pack streams")
	p.PackStreams = nil
	log.Println("[pack] Close: ", p.IsoFile.Close())
	p.IsoFile = nil
}

type IsoFileReaderWriterAt struct {
	f    *os.File
	off  int64
	isof *udf.File
}

func (ifw *IsoFileReaderWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = ifw.f.WriteAt(p, ifw.off+off)
	return n, err
}

func (ifw *IsoFileReaderWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	return ifw.f.ReadAt(p, ifw.off+off)
}

func (ifw *IsoFileReaderWriterAt) Size() int64 {
	return ifw.isof.Size()
}

func (p *IsoDriver) openIsoFileReaderWriterAt(file string) *IsoFileReaderWriterAt {
	fstr, layer := p.openIsoFile(file)
	filestart := fstr.GetFileOffset()
	if layer == 1 {
		filestart += p.SecondLayerStart
	}
	log.Println("filestart ", file, filestart)
	return &IsoFileReaderWriterAt{
		f:    p.IsoFile,
		off:  filestart,
		isof: fstr,
	}
}

func (p *IsoDriver) UpdateFile(fileName string, in *io.SectionReader) error {
	tocUdf, _ := p.openIsoFile(toc.GetTocFileName())
	tocOriginal, err := ioutil.ReadAll(tocUdf.NewReader())
	if err != nil {
		panic(err)
	}

	f := p.Files[fileName]

	defer func() {
		p.parseFilesFromTok()
		p.Cache = &pack.InstanceCache{}
	}()

	var tocbuf bytes.Buffer

	packStreams := make([]utils.ReaderWriterAt, len(p.PackStreams))
	for i := range packStreams {
		packStreams[i] = p.openIsoFileReaderWriterAt(toc.GenPartFileName(i))
	}

	if err := toc.UpdateFile(bytes.NewBuffer(tocOriginal), &tocbuf, packStreams, f, in); err == nil {
		_, err = p.openIsoFileReaderWriterAt(toc.GetTocFileName()).WriteAt(tocbuf.Bytes(), 0)
	} else {
		panic(err)
	}
	return err
}

func NewPackFromIso(isoPath string) (*IsoDriver, error) {
	p := &IsoDriver{
		IsoPath: isoPath,
		Cache:   &pack.InstanceCache{},
	}
	return p, p.parseFilesFromTok()
}
