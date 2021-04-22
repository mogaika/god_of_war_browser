package toc

import (
	"encoding/binary"
	"sort"

	"github.com/mogaika/god_of_war_browser/utils"
)

const GOW2_ENTRY_SIZE = 36
const GOW2_DVDDL_SPLITLINE = 10000000

type RawTocEntryGOW2 struct {
	Name         string
	Size         int64
	EntriesCount uint32
	EntriesStart uint32
}

func (rte *RawTocEntryGOW2) Unmarshal(buffer []byte) {
	rte.Name = utils.BytesToString(buffer[0:24])
	rte.Size = int64(binary.LittleEndian.Uint32(buffer[24:28]))
	rte.EntriesCount = binary.LittleEndian.Uint32(buffer[28:32])
	rte.EntriesStart = binary.LittleEndian.Uint32(buffer[32:36])
}

func (rte *RawTocEntryGOW2) Marshal() []byte {
	buf := make([]byte, GOW2_ENTRY_SIZE)
	copy(buf[:24], utils.StringToBytesBuffer(rte.Name, 24, false))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(rte.Size))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(rte.EntriesCount))
	binary.LittleEndian.PutUint32(buf[32:36], uint32(rte.EntriesStart))
	return buf
}

func (toc *TableOfContent) unmarshalGOW2(b []byte) error {
	countOfFiles := binary.LittleEndian.Uint32(b)
	entriesBuffer := b[4 : countOfFiles*GOW2_ENTRY_SIZE]
	offsetsArray := b[4+countOfFiles*GOW2_ENTRY_SIZE:]
	files := make(map[string]*File)
	var raw RawTocEntryGOW2
	for iFile := uint32(0); iFile < countOfFiles; iFile++ {
		raw.Unmarshal(entriesBuffer[iFile*GOW2_ENTRY_SIZE:])

		file := &File{
			name:       raw.Name,
			size:       raw.Size,
			encounters: make([]Encounter, raw.EntriesCount),
			toc:        toc}

		for i := range file.encounters {
			file.encounters[i] = Encounter{
				Offset: int64(binary.LittleEndian.Uint32(offsetsArray[(raw.EntriesStart+uint32(i))*4:])),
				Size:   file.size}
		}
		files[file.name] = file
	}

	// need to convert offset value to pak id+offset or just offset
	isHaveSectorsBiggerThenSpliline := false
	for _, f := range files {
		for i := range f.encounters {
			if f.encounters[i].Offset >= GOW2_DVDDL_SPLITLINE {
				isHaveSectorsBiggerThenSpliline = true
			}
		}
	}
	if isHaveSectorsBiggerThenSpliline {
		// official dvd-9-dl disks
		toc.packsArrayIndexing = PACK_ADDR_INDEX
	} else {
		// custom dvd-5 rip
		toc.packsArrayIndexing = PACK_ADDR_ABSOLUTE
	}

	for _, f := range files {
		for i := range f.encounters {
			off := f.encounters[i].Offset
			f.encounters[i].Pak = PakIndex(off / GOW2_DVDDL_SPLITLINE)
			f.encounters[i].Offset = (off % GOW2_DVDDL_SPLITLINE) * utils.SECTOR_SIZE
		}
	}

	toc.files = files
	return nil
}

func (toc *TableOfContent) marshalGOW2() []byte {
	rawFiles := make([]RawTocEntryGOW2, 0, len(toc.files))

	totalEncounters := 0

	for _, file := range toc.files {
		rawFiles = append(rawFiles, RawTocEntryGOW2{
			Name:         file.name,
			Size:         file.size,
			EntriesCount: uint32(len(file.encounters)),
		})
		totalEncounters += len(file.encounters)
	}

	sort.Slice(rawFiles, func(i int, j int) bool {
		return rawFiles[i].Name < rawFiles[j].Name
	})

	buffer := make([]byte, 4+len(rawFiles)*GOW2_ENTRY_SIZE+totalEncounters*4)
	binary.LittleEndian.PutUint32(buffer, uint32(len(rawFiles)))

	entriesBuf := buffer[4 : 4+len(rawFiles)*GOW2_ENTRY_SIZE]
	encountersBuf := buffer[4+len(rawFiles)*GOW2_ENTRY_SIZE:]

	currentEncounter := 0
	for iEntry, entry := range rawFiles {
		entry.EntriesStart = uint32(currentEncounter)

		file := toc.files[entry.Name]
		for _, encounter := range file.encounters {
			off := encounter.Offset / utils.SECTOR_SIZE

			if toc.packsArrayIndexing == PACK_ADDR_INDEX {
				off += int64(encounter.Pak) * GOW2_DVDDL_SPLITLINE
			}

			binary.LittleEndian.PutUint32(encountersBuf[currentEncounter*4:], uint32(off))
			currentEncounter++
		}
		copy(entriesBuf[iEntry*GOW2_ENTRY_SIZE:], entry.Marshal())
	}

	return buffer
}
