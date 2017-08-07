package flp

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

const (
	FLP_MAGIC   = 0x21
	HEADER_SIZE = 0x60

	DATA1_ELEMENT_SIZE                            = 0x4
	DATA2_ELEMENT_SIZE                            = 0x8
	DATA2_SUBTYPE1_ELEMENT_SIZE                   = 0x8
	DATA3_ELEMENT_SIZE                            = 0x24
	DATA4_ELEMENT_SIZE                            = 0x24
	DATA5_ELEMENT_SIZE                            = 0x20
	DATA6_ELEMENT_SIZE                            = 0xc
	DATA6_SUBTYPE1_ELEMENT_SIZE                   = 0x18
	DATA6_SUBTYPE2_ELEMENT_SIZE                   = 0x10
	DATA6_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE          = 0x8
	DATA6_SUBTYPE1_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE = 0xa
	DATA6_SUBTYPE1_SUBTYPE2_ELEMENT_SIZE          = 0xc
	DATA6_SUBTYPE1_SUBTYPE2_SUBTYPE1_ELEMENT_SIZE = 0x8
)

func posPad4(pos int) int {
	return pos + (pos % 4)
}

type Data1 struct {
	Off_0 uint16
	Off_2 uint16
}

func (d1 *Data1) FromBuf(buf []byte) int {
	d1.Off_0 = binary.LittleEndian.Uint16(buf[:])
	d1.Off_2 = binary.LittleEndian.Uint16(buf[2:])
	return DATA1_ELEMENT_SIZE
}

type Data2 struct {
	Off_0 uint16
	Sub1s []Data2Subtype1
}

func (d2 *Data2) FromBuf(buf []byte) int {
	d2.Off_0 = binary.LittleEndian.Uint16(buf[:])
	d2.Sub1s = make([]Data2Subtype1, binary.LittleEndian.Uint16(buf[2:]))
	return DATA2_ELEMENT_SIZE
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

	Flag2Datas2          []Data2
	Flag4Datas2          []Data2
	SomeInfoForDatas2    []int16
	AnotherInfoForDatas2 []int16
}

func (d3 *Data3) FromBuf(buf []byte) int {
	d3.Off_0 = binary.LittleEndian.Uint32(buf[:])
	d3.Flags = binary.LittleEndian.Uint16(buf[0xc:])
	return DATA3_ELEMENT_SIZE
}

func (d3 *Data3) Parse(buf []byte, pos int) int {
	if d3.Flags&2 != 0 {
		d3.Flag2Datas2 = make([]Data2, d3.Off_0)
		for i := range d3.Flag2Datas2 {
			pos += d3.Flag2Datas2[i].FromBuf(buf[pos:])
		}
		for i := range d3.Flag2Datas2 {
			pos = d3.Flag2Datas2[i].Parse(buf, pos)
		}
	}
	if d3.Flags&4 != 0 {
		d3.Flag4Datas2 = make([]Data2, d3.Off_0)
		for i := range d3.Flag4Datas2 {
			pos += d3.Flag4Datas2[i].FromBuf(buf[pos:])
		}
		for i := range d3.Flag4Datas2 {
			pos = d3.Flag4Datas2[i].Parse(buf, pos)
		}
	}

	d3.SomeInfoForDatas2 = make([]int16, d3.Off_0)
	for i := range d3.SomeInfoForDatas2 {
		d3.SomeInfoForDatas2[i] = int16(binary.LittleEndian.Uint16(buf[pos+i*2:]))
		pos += 2
	}
	pos = posPad4(pos)

	if d3.Flags&1 != 0 {
		d3.AnotherInfoForDatas2 = make([]int16, 0x100)
	} else {
		d3.AnotherInfoForDatas2 = make([]int16, d3.Off_0)
	}
	for i := range d3.AnotherInfoForDatas2 {
		d3.AnotherInfoForDatas2[i] = int16(binary.LittleEndian.Uint16(buf[pos+i*2:]))
		pos += 2
	}

	return posPad4(pos)
}

type Data4 struct {
	Payload []byte
}

func (d4 *Data4) FromBuf(buf []byte) int {
	d4.Payload = make([]byte, binary.LittleEndian.Uint32(buf[0x14:]))
	return DATA4_ELEMENT_SIZE
}

func (d4 *Data4) Parse(buf []byte, pos int) int {
	return posPad4(pos + copy(d4.Payload, buf[pos:pos+len(d4.Payload)]))
}

type Data5 struct{}

func (d5 *Data5) FromBuf(buf []byte) int {
	return DATA5_ELEMENT_SIZE
}

type Data6 struct {
	Sub1  Data6Subtype1
	Sub2s []Data6Subtype2
}

func (d6 *Data6) FromBuf(buf []byte) int {
	d6.Sub2s = make([]Data6Subtype2, binary.LittleEndian.Uint16(buf[0x8:]))
	return DATA6_ELEMENT_SIZE
}

func (d6 *Data6) Parse(buf []byte, pos int) int {
	log.Printf("d6 parsing pos: %#x", pos)
	pos += d6.Sub1.FromBuf(buf)
	pos = d6.Sub1.Parse(buf, pos)
	for i := range d6.Sub2s {
		pos += d6.Sub2s[i].FromBuf(buf)
	}
	for i := range d6.Sub2s {
		pos = d6.Sub2s[i].Parse(buf, pos)
	}
	return pos
}

type Data6Subtype1 struct {
	Sub1s []Data6Subtype1Subtype1
	Sub2s []Data6Subtype1Subtype2
}

func (d6s1 *Data6Subtype1) FromBuf(buf []byte) int {
	d6s1.Sub1s = make([]Data6Subtype1Subtype1, binary.LittleEndian.Uint16(buf[0x2:]))
	d6s1.Sub2s = make([]Data6Subtype1Subtype2, binary.LittleEndian.Uint16(buf[0x4:]))
	return DATA6_SUBTYPE1_ELEMENT_SIZE
}

func (d6s1 *Data6Subtype1) Parse(buf []byte, pos int) int {
	for i := range d6s1.Sub1s {
		pos += d6s1.Sub1s[i].FromBuf(buf)
	}
	pos = posPad4(pos)
	for i := range d6s1.Sub1s {
		pos = d6s1.Sub1s[i].Parse(buf, pos)
	}
	pos = posPad4(pos)
	for i := range d6s1.Sub2s {
		pos += d6s1.Sub2s[i].FromBuf(buf)
	}
	pos = posPad4(pos)
	log.Printf("d6sub2 parsing pos: %#x", pos)
	for i := range d6s1.Sub2s {
		pos = d6s1.Sub2s[i].Parse(buf, pos)
	}
	pos = posPad4(pos)
	return pos
}

type Data6Subtype1Subtype1 struct {
	Subs []Data6Subtype1Subtype1Subtype1
}

func (d6s1s1 *Data6Subtype1Subtype1) FromBuf(buf []byte) int {
	d6s1s1.Subs = make([]Data6Subtype1Subtype1Subtype1, binary.LittleEndian.Uint16(buf[0x2:]))
	return DATA6_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE
}

func (d6s1s1 *Data6Subtype1Subtype1) Parse(buf []byte, pos int) int {
	for i := range d6s1s1.Subs {
		pos += d6s1s1.Subs[i].FromBuf(buf)
	}
	return posPad4(pos)
}

type Data6Subtype1Subtype1Subtype1 struct{}

func (d6s1s1s1 *Data6Subtype1Subtype1Subtype1) FromBuf(buf []byte) int {
	return DATA6_SUBTYPE1_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE
}

type Data6Subtype1Subtype2 struct {
	Subs []Data6Subtype1Subtype2Subtype1
}

func (d6s1s2 *Data6Subtype1Subtype2) FromBuf(buf []byte) int {
	d6s1s2.Subs = make([]Data6Subtype1Subtype2Subtype1, binary.LittleEndian.Uint16(buf[0x2:]))
	return DATA6_SUBTYPE1_SUBTYPE2_ELEMENT_SIZE
}

func (d6s1s2 *Data6Subtype1Subtype2) Parse(buf []byte, pos int) int {
	for i := range d6s1s2.Subs {
		pos += d6s1s2.Subs[i].FromBuf(buf)
	}
	return pos
}

type Data6Subtype1Subtype2Subtype1 struct {
	Payload []byte
}

func (d6s1s2s1 *Data6Subtype1Subtype2Subtype1) FromBuf(buf []byte) int {
	d6s1s2s1.Payload = make([]byte, binary.LittleEndian.Uint32(buf[:]))
	return DATA6_SUBTYPE1_SUBTYPE2_SUBTYPE1_ELEMENT_SIZE
}
func (d6s1s2s1 *Data6Subtype1Subtype2Subtype1) Parse(buf []byte, pos int) int {
	return posPad4(pos + copy(d6s1s2s1.Payload, buf[pos:pos+len(d6s1s2s1.Payload)]))
}

type Data6Subtype2 struct {
	Payload []byte
}

func (d6s2 *Data6Subtype2) FromBuf(buf []byte) int {
	d6s2.Payload = make([]byte, binary.LittleEndian.Uint32(buf[:]))
	return DATA6_SUBTYPE2_ELEMENT_SIZE
}

func (d6s2 *Data6Subtype2) Parse(buf []byte, pos int) int {
	return posPad4(pos + copy(d6s2.Payload, buf[pos:pos+len(d6s2.Payload)]))
}

type FLP struct {
	Datas1 []Data1
	Datas2 []Data2
	Datas3 []Data3
	Datas4 []Data4
	Datas5 []Data5
	Datas6 []Data6
}

func (f *FLP) fromBuffer(buf []byte) error {
	f.Datas1 = make([]Data1, binary.LittleEndian.Uint32(buf[0xc:]))
	f.Datas2 = make([]Data2, binary.LittleEndian.Uint32(buf[0x14:]))
	f.Datas3 = make([]Data3, binary.LittleEndian.Uint32(buf[0x1c:]))
	f.Datas4 = make([]Data4, binary.LittleEndian.Uint32(buf[0x24:]))
	f.Datas5 = make([]Data5, binary.LittleEndian.Uint32(buf[0x2c:]))
	f.Datas6 = make([]Data6, binary.LittleEndian.Uint32(buf[0x34:]))

	pos := HEADER_SIZE
	for i := range f.Datas1 {
		pos += f.Datas1[i].FromBuf(buf[pos:])
	}

	for i := range f.Datas2 {
		pos += f.Datas2[i].FromBuf(buf[pos:])
	}
	for i := range f.Datas2 {
		pos = f.Datas2[i].Parse(buf, pos)
	}

	for i := range f.Datas3 {
		pos += f.Datas3[i].FromBuf(buf[pos:])
	}
	for i := range f.Datas3 {
		pos = f.Datas3[i].Parse(buf, pos)
	}

	for i := range f.Datas4 {
		pos += f.Datas4[i].FromBuf(buf[pos:])
	}
	for i := range f.Datas4 {
		pos = f.Datas4[i].Parse(buf, pos)
	}

	for i := range f.Datas5 {
		pos += f.Datas5[i].FromBuf(buf[pos:])
	}
	log.Printf("after fdata5: %#x", pos)

	for i := range f.Datas6 {
		pos += f.Datas6[i].FromBuf(buf[pos:])
	}
	for i := range f.Datas6 {
		pos = f.Datas6[i].Parse(buf, pos)
		if i == 3 {
			return nil
		}
	}
	log.Printf("after fdata6: %#x", pos)

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
