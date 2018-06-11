package toc

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/drivers/pak"
)

const TOC_FILE_NAME = "GODOFWAR.TOC"

type File struct {
	name       string
	size       int64
	encounters []pak.Encounter
}

type TableOfContent struct {
	files map[string]*File
	pa    *pak.PaksArray
}

func NewTableOfContent(b []byte, pa *pak.PaksArray) *TableOfContent {
	return &TableOfContent{
		files: make(map[string]*File),
		pa:    pa}
}

func (toc *TableOfContent) Open(filename string) (*pak.PakEncounterReader, error) {
	if f, ok := toc.files[strings.ToLower(filename)]; ok {
		return toc.pa.NewReader(f.encounters[0]), nil
	} else {
		return nil, fmt.Errorf("[toc] Cannot find file '%s'", filename)
	}
}

func (toc *TableOfContent) Unmarshal(b []byte) error {
	if config.GodOfWarVersion == config.GOWunknown {
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
		return fmt.Errorf("Unknown GOW version: %d", config.GetGOWVersion())
	}
}

func (toc *TableOfContent) Marshal() []byte {
	switch config.GetGOWVersion() {
	case config.GOW1ps2:
		return toc.marshalGOW1()
	case config.GOW2ps2:
		return toc.marshalGOW2()
	default:
		log.Panicf("Unknown GOW version: %v", config.GetGOWVersion())
		return nil
	}
}

func (toc *TableOfContent) Update(map[string]*io.SectionReader) error {
	panic("Not implemented")
	return nil
}

// Remove files duplicates and free space to other files
func (toc *TableOfContent) Shrink() error {
	panic("Not implemented")
	return nil
}
