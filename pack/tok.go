package pack

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/utils"
)

type TokDriver struct {
	Files     map[string]*TokFile
	Streams   [TOK_PARTS_COUNT]*os.File
	Directory string
	Cache     *InstanceCache
}

func (p *TokDriver) GetFileNamesList() []string {
	return getFileNamesListFromTokMap(p.Files)
}

func getFileNamesListFromTokMap(files map[string]*TokFile) []string {
	result := make([]string, len(files))
	i := 0
	for name := range files {
		result[i] = name
		i++
	}
	return result
}

func (p *TokDriver) tokGetFileName() string {
	return filepath.Join(p.Directory, "GODOFWAR.TOC")
}

func (p *TokDriver) partGetFileName(packNumber int) string {
	return filepath.Join(p.Directory, fmt.Sprintf("PART%d.PAK", packNumber+1))
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
	for _, f := range p.Streams {
		if f != nil {
			f.Close()
		}
	}
}

func (p *TokDriver) getFile(fileName string) (*TokFile, error) {
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
						log.Println("WARNING: Cannot open pack stream '%s': %v", err)
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
	f, err := p.getFile(fileName)
	if err != nil {
		return err
	}

	if in.Size()/utils.SECTOR_SIZE > f.Size()/utils.SECTOR_SIZE {
		return fmt.Errorf("Size increase above sector boundary is not supported yet")
	}

	p.closeStreams()

	// update sizes in tok file, if changed
	if in.Size() != f.Size() {
		fTok, err := os.OpenFile(p.tokGetFileName(), os.O_RDWR, 0666)
		if err != nil {
			return fmt.Errorf("Cannot open '%s' for writing: %v", p.tokGetFileName(), err)
		}
		defer fTok.Close()

		var buf [TOK_ENCOUNTER_SIZE]byte
		for {
			if _, err := fTok.Read(buf[:]); err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			name, size, enc := unmarshalTokEntry(buf[:])
			if name == "" {
				break
			}
			if size != in.Size() {
				log.Printf("[pack] Warning! Tok entry '%s': incorrect file size, file may be unconsistent: %d != %d",
					name, size, in.Size())
			}
			if name == fileName {
				if _, err := fTok.Seek(-TOK_ENCOUNTER_SIZE, os.SEEK_CUR); err != nil {
					return err
				}
				if _, err := fTok.Write(marshalTokEntry(name, in.Size(), enc)); err != nil {
					return err
				}
			}
		}
	}

	var fileBuffer bytes.Buffer
	if _, err := io.Copy(&fileBuffer, in); err != nil {
		return err
	}

	for iPart := 0; iPart < TOK_PARTS_COUNT; iPart++ {
		fPart, err := os.OpenFile(p.partGetFileName(iPart), os.O_RDWR, 0666)
		defer fPart.Close()
		if err == nil {
			for _, enc := range f.Encounters {
				if enc.Pack == iPart {
					if _, err := fPart.WriteAt(fileBuffer.Bytes(), enc.Start); err != nil {
						return err
					}
				}
			}
		} else {
			log.Printf("[pack] Error opening '%s' for writing: %v", p.partGetFileName(iPart), err)
		}
		fPart.Close()
	}

	p.Cache = &InstanceCache{}

	return nil
}

func (p *TokDriver) parseTokFile() error {
	if tokStream, err := os.Open(p.tokGetFileName()); err == nil {
		defer tokStream.Close()
		log.Printf("[pack] Parsing tok '%s'", p.tokGetFileName())
		p.Files, err = tokPartsParseFiles(tokStream)
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
