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
	DATA9_ELEMENT_SIZE                            = 0x14
	DATA10_ELEMENT_SIZE                           = 0x8
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
type GlobalHandlerIndex struct {
	TypeArrayId       uint16
	IdInThatTypeArray uint16
}

func (d1 *GlobalHandlerIndex) FromBuf(buf []byte) int {
	d1.TypeArrayId = binary.LittleEndian.Uint16(buf[:])
	d1.IdInThatTypeArray = binary.LittleEndian.Uint16(buf[2:])
	return DATA1_ELEMENT_SIZE
}

type MeshPartReference struct {
	MeshPartIndex int16
	Materials     []MeshPartMaterialSlot // Count equals to objects count in mesh part group
}

func (d2 *MeshPartReference) FromBuf(buf []byte) int {
	d2.MeshPartIndex = int16(binary.LittleEndian.Uint16(buf[:]))
	d2.Materials = make([]MeshPartMaterialSlot, binary.LittleEndian.Uint16(buf[2:]))
	return DATA2_ELEMENT_SIZE
}

func (d2 *MeshPartReference) Parse(buf []byte, pos int) int {
	for j := range d2.Materials {
		d2.Materials[j] = MeshPartMaterialSlot{
			Color:             binary.LittleEndian.Uint32(buf[pos:]),
			TextureNameSecOff: binary.LittleEndian.Uint32(buf[pos+4:]),
		}
		pos += DATA2_SUBTYPE1_ELEMENT_SIZE
	}
	return pos
}

func (d2 *MeshPartReference) SetNameFromStringSector(stringsSector []byte) {
	for i := range d2.Materials {
		d2.Materials[i].SetNameFromStringSector(stringsSector)
	}
}

type MeshPartMaterialSlot struct {
	// Texture Linkage
	Color             uint32
	TextureNameSecOff uint32
	TextureName       string
}

func (d2s1 *MeshPartMaterialSlot) SetNameFromStringSector(stringsSector []byte) {
	if d2s1.TextureNameSecOff != 0xffff {
		d2s1.TextureName = utils.BytesToString(stringsSector[d2s1.TextureNameSecOff:])
	}
}

type Font struct {
	CharsCount uint32
	// Flags
	// & 1 != 0 => CharNumberToSymbolIdMap contain 0x100 elements of symbol=>char map
	// & 1 == 0 => CharNumberToSymbolIdMap contain CharsCount elements of char=>symbol map
	Flags uint16

	Flag2Datas2             []MeshPartReference
	Flag4Datas2             []MeshPartReference
	SymbolWidths            []int16
	CharNumberToSymbolIdMap []int16 // Char to glyph map?
}

func (d3 *Font) FromBuf(buf []byte) int {
	d3.CharsCount = binary.LittleEndian.Uint32(buf[:])
	d3.Flags = binary.LittleEndian.Uint16(buf[0xc:])
	return DATA3_ELEMENT_SIZE
}

func (d3 *Font) Parse(buf []byte, pos int) int {
	if d3.Flags&2 != 0 {
		d3.Flag2Datas2 = make([]MeshPartReference, d3.CharsCount)
		for i := range d3.Flag2Datas2 {
			pos += d3.Flag2Datas2[i].FromBuf(buf[pos:])
		}
		for i := range d3.Flag2Datas2 {
			pos = d3.Flag2Datas2[i].Parse(buf, pos)
		}
	}
	if d3.Flags&4 != 0 {
		d3.Flag4Datas2 = make([]MeshPartReference, d3.CharsCount)
		for i := range d3.Flag4Datas2 {
			pos += d3.Flag4Datas2[i].FromBuf(buf[pos:])
		}
		for i := range d3.Flag4Datas2 {
			pos = d3.Flag4Datas2[i].Parse(buf, pos)
		}
	}

	d3.SymbolWidths = make([]int16, d3.CharsCount)
	for i := range d3.SymbolWidths {
		d3.SymbolWidths[i] = int16(binary.LittleEndian.Uint16(buf[pos:]))
		pos += 2
	}
	pos = posPad4(pos)

	if d3.Flags&1 != 0 {
		d3.CharNumberToSymbolIdMap = make([]int16, 0x100)
	} else {
		d3.CharNumberToSymbolIdMap = make([]int16, d3.CharsCount)
	}
	for i := range d3.CharNumberToSymbolIdMap {
		d3.CharNumberToSymbolIdMap[i] = int16(binary.LittleEndian.Uint16(buf[pos:]))
		pos += 2
	}

	return posPad4(pos)
}

type StaticLabel struct {
	RenderCommandsList []byte `json:"-"`
}

func (d4 *StaticLabel) FromBuf(buf []byte) int {
	d4.RenderCommandsList = make([]byte, binary.LittleEndian.Uint32(buf[0x14:]))
	utils.LogDump("D4", buf[:DATA4_ELEMENT_SIZE])
	return DATA4_ELEMENT_SIZE
}

func (d4 *StaticLabel) Parse(f *FLP, buf []byte, pos int) int {
	res := posPad4(pos + copy(d4.RenderCommandsList, buf[pos:pos+len(d4.RenderCommandsList)]))
	//log.Printf("<<<<<<<<<<<<<< + Parsing data4 at 0x%.6x + >>>>>>>>>>>>>>", pos)
	//utils.LogDump("D44444 PAYLOAD", d4.Payload)
	for i := 0; i < len(d4.RenderCommandsList); {
		cmd := d4.RenderCommandsList[i]
		i += 1
		if cmd&0x80 != 0 {
			if cmd&8 != 0 {
				log.Printf("  Set resource id %d as font with height %d", binary.LittleEndian.Uint16(d4.RenderCommandsList[i:]), binary.LittleEndian.Uint16(d4.RenderCommandsList[i+2:]))
				i += 4
			}
			if cmd&4 != 0 {
				log.Printf("  Set blend color %v", d4.RenderCommandsList[i:i+4])
				i += 4
			}
			if cmd&2 != 0 {
				log.Printf("  Set f20reg x offset %d", binary.LittleEndian.Uint16(d4.RenderCommandsList[i:]))
				i += 2
			}
			if cmd&1 != 0 {
				log.Printf("  Set f22reg y offset %d", binary.LittleEndian.Uint16(d4.RenderCommandsList[i:]))
				i += 2
			}
		} else {
			for j := byte(0); j < cmd; j++ {
				glyph := int16(binary.LittleEndian.Uint16(d4.RenderCommandsList[i:]))
				width := uint32(binary.LittleEndian.Uint16(d4.RenderCommandsList[i+2:]))
				i += 4
				log.Printf("  # Print glyph \"%v\"  with width %v", glyph, width)
			}
		}
	}
	return res
}

type DynamicLabel struct {
	ValueNameSecOff    uint16
	ValueName          string
	PlaceholderSecOff2 uint16
	Placeholder        string
	Payload            []byte `json:"-"`
}

func (d5 *DynamicLabel) FromBuf(buf []byte) int {
	d5.Payload = buf[:DATA5_ELEMENT_SIZE]
	d5.ValueNameSecOff = binary.LittleEndian.Uint16((buf[0:]))
	d5.PlaceholderSecOff2 = binary.LittleEndian.Uint16((buf[2:]))
	return DATA5_ELEMENT_SIZE
}

func (d5 *DynamicLabel) SetNameFromStringSector(stringsSector []byte) {
	if d5.ValueNameSecOff != 0xffff {
		d5.ValueName = utils.BytesToString(stringsSector[d5.ValueNameSecOff:])
	}
	if d5.PlaceholderSecOff2 != 0xffff {
		d5.Placeholder = utils.BytesToString(stringsSector[d5.PlaceholderSecOff2:])
	}
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
	WhenThisFrameEnds uint16 // in frameNumberUnits
	IndexInData1Array uint16
	TransformationId  uint16
	Unkint3           uint16
	NameSecOff        int16
	Name              string
}

func (d6s1s1s1 *Data6Subtype1Subtype1Subtype1) FromBuf(buf []byte) int {
	d6s1s1s1.WhenThisFrameEnds = binary.LittleEndian.Uint16(buf[0:])
	d6s1s1s1.IndexInData1Array = binary.LittleEndian.Uint16(buf[2:])
	d6s1s1s1.TransformationId = binary.LittleEndian.Uint16(buf[4:])
	d6s1s1s1.TransformationId = binary.LittleEndian.Uint16(buf[6:])
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
	LabelNameSecOff int16
	LabelName       string
	Subs            []Data6Subtype1Subtype2Subtype1
}

func (d6s1s2 *Data6Subtype1Subtype2) FromBuf(buf []byte) int {
	d6s1s2.Subs = make([]Data6Subtype1Subtype2Subtype1, binary.LittleEndian.Uint16(buf[0x2:]))
	d6s1s2.LabelNameSecOff = int16(binary.LittleEndian.Uint16(buf[8:]))
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
	if d6s1s2.LabelNameSecOff != -1 {
		d6s1s2.LabelName = utils.BytesToString(stringsSector[d6s1s2.LabelNameSecOff:])
	}
}

type Data6Subtype1Subtype2Subtype1 struct {
	Script  *Script
	Payload []byte `json:"-"`
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

type Transformation struct {
	Ints  [4]int32 // used as floats, and divided by 65536.0
	Half1 uint16   // used as float also
	Half2 uint16   // used as float too
}

func (d9 *Transformation) FromBuf(buf []byte) int {
	for i := range d9.Ints {
		d9.Ints[i] = int32(binary.LittleEndian.Uint32(buf[i*4:]))
	}
	d9.Half1 = binary.LittleEndian.Uint16(buf[0x10:])
	d9.Half2 = binary.LittleEndian.Uint16(buf[0x12:])
	return DATA9_ELEMENT_SIZE
}

type BlendColor struct {
	// in range [0, 256]. used 16 bits to better multiply
	Color [4]uint16 // rgba
}

func (d10 *BlendColor) FromBuf(buf []byte) int {
	for i := range d10.Color {
		d10.Color[i] = binary.LittleEndian.Uint16(buf[i*2:])
	}
	return DATA10_ELEMENT_SIZE
}

type FLP struct {
	GlobalHandlersIndexes []GlobalHandlerIndex
	MeshPartReferences    []MeshPartReference
	Fonts                 []Font
	StaticLabels          []StaticLabel
	DynamicLabels         []DynamicLabel
	Datas6                []Data6
	Datas7                []Data6Subtype1
	Data8                 Data6Subtype1 // Root logic node
	Transformations       []Transformation
	BlendColors           []BlendColor
	Strings               []string
}

func (f *FLP) fromBuffer(buf []byte) error {
	f.GlobalHandlersIndexes = make([]GlobalHandlerIndex, binary.LittleEndian.Uint32(buf[0xc:]))
	f.MeshPartReferences = make([]MeshPartReference, binary.LittleEndian.Uint32(buf[0x14:]))
	f.Fonts = make([]Font, binary.LittleEndian.Uint32(buf[0x1c:]))
	f.StaticLabels = make([]StaticLabel, binary.LittleEndian.Uint32(buf[0x24:]))
	f.DynamicLabels = make([]DynamicLabel, binary.LittleEndian.Uint32(buf[0x2c:]))
	f.Datas6 = make([]Data6, binary.LittleEndian.Uint32(buf[0x34:]))
	f.Datas7 = make([]Data6Subtype1, binary.LittleEndian.Uint32(buf[0x3c:]))
	f.Transformations = make([]Transformation, binary.LittleEndian.Uint16(buf[0x48:]))
	f.BlendColors = make([]BlendColor, binary.LittleEndian.Uint16(buf[0x50:]))
	f.Strings = make([]string, 0)

	pos := HEADER_SIZE
	for i := range f.GlobalHandlersIndexes {
		pos += f.GlobalHandlersIndexes[i].FromBuf(buf[pos:])
	}

	for i := range f.MeshPartReferences {
		pos += f.MeshPartReferences[i].FromBuf(buf[pos:])
	}
	for i := range f.MeshPartReferences {
		pos = f.MeshPartReferences[i].Parse(buf, pos)
	}

	for i := range f.Fonts {
		pos += f.Fonts[i].FromBuf(buf[pos:])
	}
	for i := range f.Fonts {
		pos = f.Fonts[i].Parse(buf, pos)
	}

	for i := range f.StaticLabels {
		pos += f.StaticLabels[i].FromBuf(buf[pos:])
	}
	for i := range f.StaticLabels {
		pos = f.StaticLabels[i].Parse(f, buf, pos)
	}

	for i := range f.DynamicLabels {
		pos += f.DynamicLabels[i].FromBuf(buf[pos:])
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
	for i := range f.Transformations {
		pos += f.Transformations[i].FromBuf(buf[pos:])
	}
	for i := range f.BlendColors {
		pos += f.BlendColors[i].FromBuf(buf[pos:])
	}

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
	for i := range f.MeshPartReferences {
		f.MeshPartReferences[i].SetNameFromStringSector(stringsSector)
	}
	for i := range f.Fonts {
		for j := range f.Fonts[i].Flag4Datas2 {
			f.Fonts[i].Flag4Datas2[j].SetNameFromStringSector(stringsSector)
		}
		for j := range f.Fonts[i].Flag2Datas2 {
			f.Fonts[i].Flag2Datas2[j].SetNameFromStringSector(stringsSector)
		}
	}

	for i := range f.DynamicLabels {
		f.DynamicLabels[i].SetNameFromStringSector(stringsSector)
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
