package flp

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
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
	if pos%4 != 0 {
		newPos := pos + 4 - pos%4
		if newPos&3 != 0 {
			panic(fmt.Sprintf("How it even possible? %x + 4 - %x = %x", pos, pos%4, newPos))
		}
		return newPos
	} else {
		return pos
	}
}

// Mesh instance linkage?
type Data1 struct {
	Type              uint16
	IdInThatTypeArray uint16
}

func (d1 *Data1) FromBuf(buf []byte) int {
	d1.Type = binary.LittleEndian.Uint16(buf[:])
	d1.IdInThatTypeArray = binary.LittleEndian.Uint16(buf[2:])
	return DATA1_ELEMENT_SIZE
}

type Data2 struct {
	MeshPartIndex               int16
	MeshPartObjectsDescriptions []Data2Subtype1 // Count equals to objects count in mesh part group
}

func (d2 *Data2) FromBuf(buf []byte) int {
	d2.MeshPartIndex = int16(binary.LittleEndian.Uint16(buf[:]))
	d2.MeshPartObjectsDescriptions = make([]Data2Subtype1, binary.LittleEndian.Uint16(buf[2:]))
	return DATA2_ELEMENT_SIZE
}

func (d2 *Data2) Parse(buf []byte, pos int) int {
	for j := range d2.MeshPartObjectsDescriptions {
		d2.MeshPartObjectsDescriptions[j] = Data2Subtype1{
			Color:             binary.LittleEndian.Uint32(buf[pos:]),
			TextureNameSecOff: binary.LittleEndian.Uint32(buf[pos+4:]),
		}
		pos += DATA2_SUBTYPE1_ELEMENT_SIZE
	}
	return pos
}

func (d2 *Data2) SetNameFromStringSector(stringsSector []byte) {
	for i := range d2.MeshPartObjectsDescriptions {
		d2.MeshPartObjectsDescriptions[i].SetNameFromStringSector(stringsSector)
	}
}

type Data2Subtype1 struct {
	// Texture Linkage
	Color             uint32
	TextureNameSecOff uint32
	TextureName       string
}

func (d2s1 *Data2Subtype1) SetNameFromStringSector(stringsSector []byte) {
	if d2s1.TextureNameSecOff != 0xffff {
		d2s1.TextureName = utils.BytesToString(stringsSector[d2s1.TextureNameSecOff:])
	}
}

type Data3 struct {
	CharsCount uint32
	// Flags
	// & 1 != 0 => CharNumberToSymbolIdMap contain 0x100 elements of symbol=>char map
	// & 1 == 0 => CharNumberToSymbolIdMap contain CharsCount elements of char=>symbol map
	Flags uint16

	Flag2Datas2             []Data2
	Flag4Datas2             []Data2
	SymbolWidths            []int16
	CharNumberToSymbolIdMap []int16 // Char to glyph map?
}

func (d3 *Data3) FromBuf(buf []byte) int {
	d3.CharsCount = binary.LittleEndian.Uint32(buf[:])
	d3.Flags = binary.LittleEndian.Uint16(buf[0xc:])
	return DATA3_ELEMENT_SIZE
}

func (d3 *Data3) Parse(buf []byte, pos int) int {
	if d3.Flags&2 != 0 {
		d3.Flag2Datas2 = make([]Data2, d3.CharsCount)
		for i := range d3.Flag2Datas2 {
			pos += d3.Flag2Datas2[i].FromBuf(buf[pos:])
		}
		for i := range d3.Flag2Datas2 {
			pos = d3.Flag2Datas2[i].Parse(buf, pos)
		}
	}
	if d3.Flags&4 != 0 {
		d3.Flag4Datas2 = make([]Data2, d3.CharsCount)
		for i := range d3.Flag4Datas2 {
			pos += d3.Flag4Datas2[i].FromBuf(buf[pos:])
		}
		for i := range d3.Flag4Datas2 {
			pos = d3.Flag4Datas2[i].Parse(buf, pos)
		}
	}

	d3.SymbolWidths = make([]int16, d3.CharsCount)
	for i := range d3.SymbolWidths {
		d3.SymbolWidths[i] = int16(binary.LittleEndian.Uint16(buf[pos+i*2:]))
		pos += 2
	}
	pos = posPad4(pos)

	if d3.Flags&1 != 0 {
		d3.CharNumberToSymbolIdMap = make([]int16, 0x100)
	} else {
		d3.CharNumberToSymbolIdMap = make([]int16, d3.CharsCount)
	}
	for i := range d3.CharNumberToSymbolIdMap {
		d3.CharNumberToSymbolIdMap[i] = int16(binary.LittleEndian.Uint16(buf[pos+i*2:]))
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

type Data5 struct {
	Payload []byte
}

func (d5 *Data5) FromBuf(buf []byte) int {
	d5.Payload = buf[:DATA5_ELEMENT_SIZE]
	utils.LogDump("D5 PAYLOAD", d5.Payload)
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
	pos = posPad4(pos)
	pos += d6.Sub1.FromBuf(buf[pos:])
	pos = d6.Sub1.Parse(buf, pos)

	for i := range d6.Sub2s {
		pos += d6.Sub2s[i].FromBuf(buf[pos:])
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

func (d6s1 *Data6Subtype1) SetNameFromStringSector(stringsSector []byte) {
	for i := range d6s1.Sub1s {
		d6s1.Sub1s[i].SetNameFromStringSector(stringsSector)
	}
	for i := range d6s1.Sub2s {
		d6s1.Sub2s[i].SetNameFromStringSector(stringsSector)
	}
}

func (d6s1 *Data6Subtype1) Parse(buf []byte, pos int) int {
	//log.Printf("d6sub1 parsing pos: %#x {%d,%d} < b34c,b3bc,b608,e694,f214,f5e8,f720,f84c,fa20,fd0c,12398,123ac,12448,124f3,1278f",
	//	pos, len(d6s1.Sub1s), len(d6s1.Sub2s))

	pos = posPad4(pos)
	for i := range d6s1.Sub1s {
		pos += d6s1.Sub1s[i].FromBuf(buf[pos:])
	}
	for i := range d6s1.Sub1s {
		pos = d6s1.Sub1s[i].Parse(buf, pos)
	}

	pos = posPad4(pos)
	for i := range d6s1.Sub2s {
		pos += d6s1.Sub2s[i].FromBuf(buf[pos:])
	}
	for i := range d6s1.Sub2s {
		//log.Printf("d6sub1sub2 %d parsing pos: %#x << [0x123cc,0x124b0,0x124d6]", i, pos)
		pos = d6s1.Sub2s[i].Parse(buf, pos)
	}
	return pos
}

type Data6Subtype1Subtype1 struct {
	// FRAMES PROBABLY (way of animation)
	Subs []Data6Subtype1Subtype1Subtype1
}

func (d6s1s1 *Data6Subtype1Subtype1) FromBuf(buf []byte) int {
	d6s1s1.Subs = make([]Data6Subtype1Subtype1Subtype1, binary.LittleEndian.Uint16(buf[0x2:]))
	return DATA6_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE
}

func (d6s1s1 *Data6Subtype1Subtype1) Parse(buf []byte, pos int) int {
	//log.Printf("d6sub1sub1 parsing pos: %#x {%d} < b354,b3c4,b4e4,b610,e69c,f21c,f5f0,f728,f854,fa28,fd14,123a0,123b4,12458", pos, len(d6s1s1.Subs))
	pos = posPad4(pos)
	for i := range d6s1s1.Subs {
		pos += d6s1s1.Subs[i].FromBuf(buf[pos:])
	}
	return pos
}

func (d6s1s1 *Data6Subtype1Subtype1) SetNameFromStringSector(stringsSector []byte) {
	for i := range d6s1s1.Subs {
		d6s1s1.Subs[i].SetNameFromStringSector(stringsSector)
	}
}

// Special symbol (mesh + text?)
type Data6Subtype1Subtype1Subtype1 struct {
	IdexInData1Array uint16
	NameSecOff       int16
	Name             string
	Data             []byte
}

func (d6s1s1s1 *Data6Subtype1Subtype1Subtype1) FromBuf(buf []byte) int {
	d6s1s1s1.Data = buf[:DATA6_SUBTYPE1_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE]
	d6s1s1s1.IdexInData1Array = binary.LittleEndian.Uint16(buf[0:])
	d6s1s1s1.NameSecOff = int16(binary.LittleEndian.Uint16(buf[8:]))
	return DATA6_SUBTYPE1_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE
}

func (d6s1s1s1 *Data6Subtype1Subtype1Subtype1) SetNameFromStringSector(stringsSector []byte) {
	if d6s1s1s1.NameSecOff != -1 {
		d6s1s1s1.Name = utils.BytesToString(stringsSector[d6s1s1s1.NameSecOff:])
	}
}

type Data6Subtype1Subtype2 struct {
	// Frame
	FrameNameSecOff int16
	FrameName       string
	Subs            []Data6Subtype1Subtype2Subtype1
}

func (d6s1s2 *Data6Subtype1Subtype2) FromBuf(buf []byte) int {
	d6s1s2.Subs = make([]Data6Subtype1Subtype2Subtype1, binary.LittleEndian.Uint16(buf[0x2:]))
	d6s1s2.FrameNameSecOff = int16(binary.LittleEndian.Uint16(buf[8:]))
	return DATA6_SUBTYPE1_SUBTYPE2_ELEMENT_SIZE
}

func (d6s1s2 *Data6Subtype1Subtype2) Parse(buf []byte, pos int) int {
	pos = posPad4(pos)
	for i := range d6s1s2.Subs {
		pos += d6s1s2.Subs[i].FromBuf(buf[pos:])
	}
	for i := range d6s1s2.Subs {
		pos = d6s1s2.Subs[i].Parse(buf, pos)
	}
	return pos
}

func (d6s1s2 *Data6Subtype1Subtype2) SetNameFromStringSector(stringsSector []byte) {
	if d6s1s2.FrameNameSecOff != -1 {
		d6s1s2.FrameName = utils.BytesToString(stringsSector[d6s1s2.FrameNameSecOff:])
	}
}

type Data6Subtype1Subtype2Subtype1 struct {
	Script  *Script
	Payload []byte
}

func (d6s1s2s1 *Data6Subtype1Subtype2Subtype1) FromBuf(buf []byte) int {
	d6s1s2s1.Payload = make([]byte, binary.LittleEndian.Uint32(buf[:]))
	return DATA6_SUBTYPE1_SUBTYPE2_SUBTYPE1_ELEMENT_SIZE
}
func (d6s1s2s1 *Data6Subtype1Subtype2Subtype1) Parse(buf []byte, pos int) int {
	pos = posPad4(pos)
	// utils.Dump("d6s1s2s1 _ PAYLOAD", buf[pos:pos+len(d6s1s2s1.Payload)])
	d6s1s2s1.Script = NewScriptFromData(buf[pos : pos+len(d6s1s2s1.Payload)])
	return pos + copy(d6s1s2s1.Payload, buf[pos:pos+len(d6s1s2s1.Payload)])
}

type Data6Subtype2 struct {
	Payload []byte
	Script  *Script
}

func (d6s2 *Data6Subtype2) FromBuf(buf []byte) int {
	d6s2.Payload = make([]byte, binary.LittleEndian.Uint32(buf[:]))
	return DATA6_SUBTYPE2_ELEMENT_SIZE
}

func (d6s2 *Data6Subtype2) Parse(buf []byte, pos int) int {
	pos = posPad4(pos)
	//utils.LogDump("d6s2 payload", buf[pos:pos+len(d6s2.Payload)])
	d6s2.Script = NewScriptFromData(buf[pos : pos+len(d6s2.Payload)])
	return pos + copy(d6s2.Payload, buf[pos:pos+len(d6s2.Payload)])
}

type FLP struct {
	Datas1  []Data1 // Global id table
	Datas2  []Data2 // Textures
	Datas3  []Data3 // Font declaration
	Datas4  []Data4
	Datas5  []Data5
	Datas6  []Data6
	Datas7  []Data6Subtype1
	Data8   Data6Subtype1 // Root logic node
	Strings []string
}

func (f *FLP) fromBuffer(buf []byte) error {
	f.Datas1 = make([]Data1, binary.LittleEndian.Uint32(buf[0xc:]))
	f.Datas2 = make([]Data2, binary.LittleEndian.Uint32(buf[0x14:]))
	f.Datas3 = make([]Data3, binary.LittleEndian.Uint32(buf[0x1c:]))
	f.Datas4 = make([]Data4, binary.LittleEndian.Uint32(buf[0x24:]))
	f.Datas5 = make([]Data5, binary.LittleEndian.Uint32(buf[0x2c:]))
	f.Datas6 = make([]Data6, binary.LittleEndian.Uint32(buf[0x34:]))
	f.Datas7 = make([]Data6Subtype1, binary.LittleEndian.Uint32(buf[0x3c:]))
	f.Strings = make([]string, 0)

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
		log.Println("Data5 pos ", pos)
		pos += f.Datas5[i].FromBuf(buf[pos:])
	}

	for i := range f.Datas6 {
		pos += f.Datas6[i].FromBuf(buf[pos:])
	}
	for i := range f.Datas6 {
		pos = f.Datas6[i].Parse(buf, pos)
	}
	log.Printf("after fdata6: %#x == 0xffdf", pos)

	pos = posPad4(pos)
	for i := range f.Datas7 {
		pos += f.Datas7[i].FromBuf(buf[pos:])
	}
	log.Printf("fdata7count: %#x == 0x17d  | after fdata7buf: %#x == 0x12398", len(f.Datas7), pos)
	for i := range f.Datas7 {
		pos = f.Datas7[i].Parse(buf, pos)
	}
	log.Printf("after fdata7: %#x == 0x3e570", pos)

	pos = posPad4(pos)
	pos += f.Data8.FromBuf(buf[pos:])
	pos = f.Data8.Parse(buf, pos)

	pos = posPad4(pos)

	// more sections that we ignore
	pos += int(binary.LittleEndian.Uint16(buf[0x48:])) * 0x14
	pos += int(binary.LittleEndian.Uint16(buf[0x50:])) * 8

	stringsSectorStart := pos
	log.Printf("string sec start: %#x == 0x72cf8  {size? or strings count: %#x}", pos, binary.LittleEndian.Uint16(buf[0x58:]))
	for {
		if pos >= len(buf)-1 {
			break
		}
		s := utils.BytesToString(buf[pos:])
		pos += utils.BytesStringLength(buf[pos:])
		pos += 1
		f.Strings = append(f.Strings, s)
	}

	f.SetNameFromStringSector(buf[stringsSectorStart:])

	return nil
}

func (f *FLP) SetNameFromStringSector(stringsSector []byte) {
	for i := range f.Datas2 {
		f.Datas2[i].SetNameFromStringSector(stringsSector)
	}
	for i := range f.Datas3 {
		for j := range f.Datas3[i].Flag4Datas2 {
			f.Datas3[i].Flag4Datas2[j].SetNameFromStringSector(stringsSector)
		}
		for j := range f.Datas3[i].Flag2Datas2 {
			f.Datas3[i].Flag2Datas2[j].SetNameFromStringSector(stringsSector)
		}
	}

	for i := range f.Datas6 {
		f.Datas6[i].Sub1.SetNameFromStringSector(stringsSector)
	}

	for i := range f.Datas7 {
		f.Datas7[i].SetNameFromStringSector(stringsSector)
	}
	f.Data8.SetNameFromStringSector(stringsSector)
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
