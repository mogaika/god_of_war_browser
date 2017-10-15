package pack

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/tok"
	"github.com/mogaika/god_of_war_browser/utils"
)

type TokDriver struct {
	Files     map[string]*tok.File
	Streams   [tok.PARTS_COUNT]*os.File
	Directory string
	Cache     *InstanceCache
}

func (p *TokDriver) GetFileNamesList() []string {
	return getFileNamesListFromTokMap(p.Files)
}

func getFileNamesListFromTokMap(files map[string]*tok.File) []string {
	result := make([]string, len(files))
	i := 0
	for name := range files {
		result[i] = name
		i++
	}
	return result
}

func (p *TokDriver) tokGetFileName() string {
	return filepath.Join(p.Directory, tok.FILE_NAME)
}

func (p *TokDriver) partGetFileName(packNumber int) string {
	return filepath.Join(p.Directory, tok.GenPartFileName(packNumber))
}

func (p *TokDriver) prepareStream(packNumber int) error {
	if p.Streams[packNumber] == nil {
		if f, err := os.Open(p.partGetFileName(packNumber)); err != nil {
			return err
		} else {
			p.Streams[packNumber] = f
		}
	}
	return nil
}

func (p *TokDriver) closeStreams() {
	for i, f := range p.Streams {
		if f != nil {
			f.Close()
		}
		p.Streams[i] = nil
	}
}

func (p *TokDriver) getFile(fileName string) (*tok.File, error) {
	if f, exists := p.Files[fileName]; exists {
		return f, nil
	} else {
		return nil, fmt.Errorf("Cannot find '%s' file in pack", fileName)
	}
}

func (p *TokDriver) GetFile(fileName string) (PackFile, error) {
	return p.getFile(fileName)
}

func (p *TokDriver) GetFileReader(fileName string) (PackFile, *io.SectionReader, error) {
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

func (p *TokDriver) GetInstance(fileName string) (interface{}, error) {
	return defaultGetInstanceCachedHandler(p, p.Cache, fileName)
}

func (p *TokDriver) UpdateFile(fileName string, in *io.SectionReader) error {
	defer p.parseTokFile()

	f, err := p.getFile(fileName)
	if err != nil {
		return err
	}
	p.closeStreams()

	var fParts [tok.PARTS_COUNT]*os.File
	var partWriters [tok.PARTS_COUNT]utils.ReaderWriterAt
	defer func() {
		for _, part := range fParts {
			if part != nil {
				part.Close()
			}
		}
	}()
	for iPart := range fParts {
		if part, err := os.OpenFile(p.partGetFileName(iPart), os.O_RDWR, 0666); err == nil {
			if finfo, err := part.Stat(); err == nil {
				fParts[iPart] = part
				partWriters[iPart] = utils.NewReaderWriterAt(part, part, finfo.Size())
			} else {
				return fmt.Errorf("Cannot stat '%s' for writing: %v", p.partGetFileName(iPart), err)
			}
		} else {
			return fmt.Errorf("Cannot open '%s' for writing: %v", p.partGetFileName(iPart), err)
		}
	}

	fTok, err := os.OpenFile(p.tokGetFileName(), os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("Cannot open tokfile '%s' for writing: %v", p.tokGetFileName(), err)
	}
	defer fTok.Close()

	ftokoriginal, _ := ioutil.ReadAll(fTok)
	fTok.Seek(0, os.SEEK_SET)

	err = tok.UpdateFile(bytes.NewReader(ftokoriginal), fTok, partWriters, f, in)

	p.Cache = &InstanceCache{}

	return err
}

func (p *TokDriver) parseTokFile() error {
	if tokStream, err := os.Open(p.tokGetFileName()); err == nil {
		defer tokStream.Close()
		log.Printf("[pack] Parsing tok '%s'", p.tokGetFileName())
		p.Files, _, err = tok.ParseFiles(tokStream)
		return err
	} else {
		return err
	}
}

func NewPackFromTok(gamePath string) (*TokDriver, error) {
	p := &TokDriver{
		Directory: gamePath,
		Cache:     &InstanceCache{},
	}

	return p, p.parseTokFile()
}
