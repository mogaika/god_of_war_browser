package flp

import (
	"encoding/binary"
	"fmt"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

const (
	FLP_MAGIC   = 0x21
	HEADER_SIZE = 0x60

	DATA1_ELEMENT_SIZE          = 0x4
	DATA2_ELEMENT_SIZE          = 0x8
	DATA2_SUBTYPE1_ELEMENT_SIZE = 0x8
	DATA3_ELEMENT_SIZE          = 0x24
	DATA4_ELEMENT_SIZE          = 0x24
)

type Data1 struct {
	Off_0 uint16
	Off_2 uint16
}

type Data2 struct {
	Off_0 uint16
	Sub1s []Data2Subtype1
}

func (d2 *Data2) Parse(buf []byte, pos int) int {
	for j := range d2.Sub1s {
		d2.Sub1s[j] = Data2Subtype1{
			Color: binary.LittleEndian.Uint32(buf[pos:]),
			Int1:  int16(binary.LittleEndian.Uint16(buf[pos+4:])),
			Int2:  int16(binary.LittleEndian.Uint16(buf[pos+6:])),
		}
		pos += DATA2_SUBTYPE1_ELEMENT_SIZE
	}
	return pos
}

type Data2Subtype1 struct {
	Color      uint32 //maybe
	Int1, Int2 int16
}

type Data3 struct {
	Off_0 uint32
	Flags uint16
}

type FLP struct {
	Datas1 []Data1
	Datas2 []Data2
	Datas3 []Data3
}

func (f *FLP) fromBuffer(buf []byte) error {
	data1Start := HEADER_SIZE
	f.Datas1 = make([]Data1, binary.LittleEndian.Uint32(buf[0xc:]))
	for i := range f.Datas1 {
		f.Datas1[i] = Data1{
			Off_0: binary.LittleEndian.Uint16(buf[data1Start+i*DATA1_ELEMENT_SIZE:]),
			Off_2: binary.LittleEndian.Uint16(buf[data1Start+i*DATA1_ELEMENT_SIZE+2:]),
		}
	}

	data2Start := data1Start + len(f.Datas1)*DATA1_ELEMENT_SIZE
	f.Datas2 = make([]Data2, binary.LittleEndian.Uint32(buf[0x14:]))
	for i := range f.Datas2 {
		f.Datas2[i] = Data2{
			Off_0: binary.LittleEndian.Uint16(buf[data2Start+i*DATA2_ELEMENT_SIZE:]),
			Sub1s: make([]Data2Subtype1, binary.LittleEndian.Uint16(buf[data2Start+i*DATA2_ELEMENT_SIZE+2:])),
		}
	}

	data2s1Start := data2Start + len(f.Datas2)*DATA2_ELEMENT_SIZE
	pos := data2s1Start
	for i := range f.Datas2 {
		pos = f.Datas2[i].Parse(buf, pos)
	}

	data3Start := pos
	f.Datas3 = make([]Data3, binary.LittleEndian.Uint32(buf[0x1c:]))
	for i := range f.Datas3 {
		f.Datas3[i] = Data3{
			Off_0: binary.LittleEndian.Uint32(buf[data3Start+i*DATA3_ELEMENT_SIZE:]),
			Flags: binary.LittleEndian.Uint16(buf[data3Start+i*DATA3_ELEMENT_SIZE+0xc:]),
		}
	}

	return nil
}

func NewFromData(buf []byte) (*FLP, error) {
	f := &FLP{}
	if err := f.fromBuffer(buf); err != nil {
		return nil, fmt.Errorf("Error when reading flp header: %v", err)
	}
	return f, nil
}

func (f *FLP) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return f, nil
}

func init() {
	wad.SetHandler(FLP_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
