package toc

import (
	"encoding/binary"

	"github.com/mogaika/god_of_war_browser/utils"
)

const GOW1_ENTRY_SIZE = 24

type RawTocEntryGOW1 struct {
	Name   string
	Size   int64
	Offset int64
	Pak    PakIndex
}

func (rte *RawTocEntryGOW1) Unmarshal(buffer []byte) {
	rte.Name = utils.BytesToString(buffer[0:12])
	rte.Size = int64(binary.LittleEndian.Uint32(buffer[16:20]))
	rte.Pak = PakIndex(binary.LittleEndian.Uint32(buffer[12:16]))
	rte.Offset = int64(binary.LittleEndian.Uint32(buffer[20:24])) * utils.SECTOR_SIZE
}

func (rte *RawTocEntryGOW1) Marshal() []byte {
	buf := make([]byte, GOW1_ENTRY_SIZE)
	copy(buf[:12], utils.StringToBytesBuffer(rte.Name, 12, false))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(rte.Pak))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(rte.Size))
	binary.LittleEndian.PutUint32(buf[20:24], uint32(utils.GetRequiredSectorsCount(rte.Offset)))
	return buf
}

func (toc *TableOfContent) unmarshalGOW1(b []byte) error {
	toc.files = make(map[string]*File)
	var raw RawTocEntryGOW1
	for len(b) >= GOW1_ENTRY_SIZE {
		raw.Unmarshal(b)
		e := Encounter{
			Offset: raw.Offset,
			Size:   raw.Size,
			Pak:    raw.Pak,
		}
		if f, ok := toc.files[raw.Name]; ok {
			f.encounters = append(f.encounters, e)
		} else {
			newF := &File{
				name:       raw.Name,
				size:       raw.Size,
				encounters: make([]Encounter, 1, 8),
				toc:        toc,
			}
			newF.encounters[0] = e
			toc.files[raw.Name] = newF
		}

		b = b[GOW1_ENTRY_SIZE:]
	}

	return nil
}

func (toc *TableOfContent) marshalGOW1() []byte {
	panic("Not implemented")
	return nil
}
