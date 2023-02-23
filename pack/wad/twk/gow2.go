package twk

import (
	"encoding/binary"
	"encoding/hex"
	"sort"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	TWK_GOW2_Data_Tag    = 12
	TWK_GOW2_Exports_Tag = 13
	TWK_GOW2_Imports_Tag = 14
	TWK_GOW2_Hashes_Tag  = 15
	TWK_GOW2_Types_Tag   = 16
)

type GOW2TWK struct {
	Entries []*GOW2TWKEntry   `yaml:"entries"`
	Hashes  map[string]string `yaml:"hashes"`
}

type GOW2TWKEntry struct {
	Name string `yaml:"name"`
	Data []any  `yaml:"data"`
}

type GOW2TWKLabel struct {
	Label    string `yaml:"label"`
	TypeMark uint32 `yaml:"type_mark"`
}

type GOW2TWKImport struct {
	Import string `yaml:"import"`
}

type GOW2TWKData struct {
	Hex string `yaml:"hex"`
	Str string `yaml:"str"`
}

func NewGOW2TWK(data []byte, rsrc *wad.WadNodeRsrc) *GOW2TWK {
	t13raw, _, err := rsrc.Wad.GetInstanceFromTag(rsrc.Tag.Id + 1)
	if err != nil {
		panic(err)
	}

	t14raw, _, err := rsrc.Wad.GetInstanceFromTag(rsrc.Tag.Id + 2)
	if err != nil {
		panic(err)
	}

	t15raw, _, err := rsrc.Wad.GetInstanceFromTag(rsrc.Tag.Id + 3)
	if err != nil {
		panic(err)
	}

	t16raw, _, err := rsrc.Wad.GetInstanceFromTag(rsrc.Tag.Id + 4)
	if err != nil {
		panic(err)
	}

	result := &GOW2TWK{}
	exports := t13raw.(GOW2TWK13)
	imports := t14raw.(GOW2TWK14)
	hashes := t15raw.(GOW2TWK15)
	labels := t16raw.(GOW2TWK16)

	type slicer struct {
		what   any
		offset uint32
		t      int
	}

	const (
		// in order we want to them to be parsed
		t_export = iota
		t_label
		t_import
	)

	spliters := make([]slicer, 0)
	for _, e := range exports {
		spliters = append(spliters, slicer{e, e.Offset, t_export})
	}
	for _, i := range imports {
		spliters = append(spliters, slicer{i, i.Offset, t_import})
	}
	for _, l := range labels {
		spliters = append(spliters, slicer{l, l.Offset, t_label})
	}

	sort.Slice(spliters, func(i, j int) bool {
		if spliters[i].offset == spliters[j].offset {
			return spliters[i].t < spliters[j].t
		} else {
			return spliters[i].offset < spliters[j].offset
		}
	})

	utils.LogDump(spliters)

	var currentEntry *GOW2TWKEntry
	var lastOffset uint32
	flushDataTill := func(newOffset uint32) {
		if newOffset != lastOffset {
			remainingData := data[lastOffset:newOffset]
			str := make([]byte, len(remainingData))
			for i, b := range remainingData {
				if b >= 0x20 && b < 0x80 {
					str[i] = b
				} else {
					str[i] = '.'
				}
			}
			currentEntry.Data = append(currentEntry.Data,
				GOW2TWKData{
					Hex: hex.EncodeToString(remainingData),
					Str: utils.BytesToString(str),
				},
			)
		}
		lastOffset = newOffset
	}

	for _, split := range spliters {
		switch split.t {
		case t_export:
			if currentEntry != nil {
				flushDataTill(split.offset)
				result.Entries = append(result.Entries, currentEntry)
			}
			currentEntry = &GOW2TWKEntry{
				Name: split.what.(GOW2TWKRAWReference).Name,
			}
		case t_import:
			flushDataTill(split.offset)
			currentEntry.Data = append(currentEntry.Data,
				GOW2TWKImport{Import: split.what.(GOW2TWKRAWReference).Name},
			)
			lastOffset += 4 // pointer
		case t_label:
			flushDataTill(split.offset)
			currentEntry.Data = append(currentEntry.Data,
				GOW2TWKLabel{
					Label:    split.what.(GOW2TWK16Entry).Name,
					TypeMark: split.what.(GOW2TWK16Entry).Type,
				},
			)
		}
	}
	if currentEntry != nil {
		flushDataTill(uint32(len(data)))
		result.Entries = append(result.Entries, currentEntry)
	}

	result.Hashes = make(map[string]string)
	for h, v := range hashes {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], h)
		result.Hashes[hex.EncodeToString(b[:])] = v
	}

	utils.LogDump(result)

	_, _, _, _ = exports, imports, hashes, labels

	return result
}

func (t *GOW2TWK) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return t, nil
}

type GOW2TWKRAWReference struct {
	Offset uint32
	Name   string
}

type GOW2TWK13 []GOW2TWKRAWReference

func (m GOW2TWK13) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return m, nil
}

type GOW2TWK14 []GOW2TWKRAWReference

func (m GOW2TWK14) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return m, nil
}

type GOW2TWK15 map[uint32]string

func (m GOW2TWK15) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return m, nil
}

type GOW2TWK16 []GOW2TWK16Entry

type GOW2TWK16Entry struct {
	GOW2TWKRAWReference
	Type uint32
}

func (m GOW2TWK16) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return m, nil
}

func NewGOWTWKStringMap(bs *utils.BufStack) []GOW2TWKRAWReference {
	count := bs.ReadLU32()

	result := make([]GOW2TWKRAWReference, count)
	for i := uint32(0); i < count; i++ {
		t12Off := bs.ReadLU32()
		stringOff := bs.ReadLU32()

		result[i] = GOW2TWKRAWReference{
			Offset: t12Off,
			Name:   utils.BytesToString(bs.Raw()[stringOff:]),
		}
	}

	return result
}

func NewGOWTWK15(bs *utils.BufStack) GOW2TWK15 {
	count := bs.ReadLU32()

	result := make(map[uint32]string, count)
	for i := uint32(0); i < count; i++ {
		t12Off := bs.ReadLU32()
		stringOff := bs.ReadLU32()
		result[t12Off] = utils.BytesToString(bs.Raw()[stringOff:])
	}

	return result
}

func NewGOWTWK16(bs *utils.BufStack) GOW2TWK16 {
	count := bs.ReadLU32()

	result := make([]GOW2TWK16Entry, count)
	for i := uint32(0); i < count; i++ {
		t12Off := bs.ReadLU32()
		stringOff := bs.ReadLU32()
		unk08 := bs.ReadLU32()

		result[i] = GOW2TWK16Entry{
			GOW2TWKRAWReference: GOW2TWKRAWReference{
				Offset: t12Off,
				Name:   utils.BytesToString(bs.Raw()[stringOff:]),
			},
			Type: unk08,
		}
	}

	return result
}

func init() {
	wad.SetTagHandler(config.GOW2, TWK_GOW2_Data_Tag, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewGOW2TWK(wrsrc.Tag.Data, wrsrc), nil
	})
	wad.SetTagHandler(config.GOW2, TWK_GOW2_Exports_Tag, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return GOW2TWK13(NewGOWTWKStringMap(utils.NewBufStack("twk_exports", wrsrc.Tag.Data))), nil
	})
	wad.SetTagHandler(config.GOW2, TWK_GOW2_Imports_Tag, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return GOW2TWK14(NewGOWTWKStringMap(utils.NewBufStack("twk_imports", wrsrc.Tag.Data))), nil
	})
	wad.SetTagHandler(config.GOW2, TWK_GOW2_Hashes_Tag, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return GOW2TWK15(NewGOWTWK15(utils.NewBufStack("twk_hashes", wrsrc.Tag.Data))), nil
	})
	wad.SetTagHandler(config.GOW2, TWK_GOW2_Types_Tag, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewGOWTWK16(utils.NewBufStack("twk_types", wrsrc.Tag.Data)), nil
	})
}
