package flp

import (
	"bytes"
	"encoding/binary"
	"math"
)

type StringPosReplacerReference struct {
	Position    int
	SizeInBytes int
}

type StringsIndexBuffer map[string][]StringPosReplacerReference

func NewStringsIndexBuffer() StringsIndexBuffer {
	return StringsIndexBuffer(make(map[string][]StringPosReplacerReference))
}

func (sb StringsIndexBuffer) Add(s string, pos int, size int) *StringPosReplacerReference {
	arr, ex := sb[s]
	if !ex {
		arr = make([]StringPosReplacerReference, 0)
	}
	arr = append(arr, StringPosReplacerReference{Position: pos, SizeInBytes: size})
	sb[s] = arr
	return &arr[len(arr)-1]
}

type FlpMarshaler struct {
	sbuffer StringsIndexBuffer
	buf     bytes.Buffer
}

func NewFlpMarshaler() *FlpMarshaler {
	return &FlpMarshaler{sbuffer: NewStringsIndexBuffer()}
}

func (fm *FlpMarshaler) w16(v uint16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	fm.buf.Write(buf[:])
}

func (fm *FlpMarshaler) w32(v uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	fm.buf.Write(buf[:])
}

func (fm *FlpMarshaler) pos() int {
	return fm.buf.Len()
}

func (fm *FlpMarshaler) skip(count int) {
	fm.buf.Write(make([]byte, count))
}

func (fm *FlpMarshaler) pad(align int) {
	fm.skip(fm.pos() % align)
}

func (fm *FlpMarshaler) pad4() {
	fm.pad(4)
}

func (fm *FlpMarshaler) addStringOffsetPlaceholder(s string, size int) {
	fm.sbuffer.Add(s, fm.pos(), size)
	fm.skip(size)
}

func (d1 *GlobalHandlerIndex) MarshalStruct(fm *FlpMarshaler) {
	fm.w16(d1.TypeArrayId)
	fm.w16(d1.IdInThatTypeArray)
}

func (d2 *MeshPartReference) MarshalStruct(fm *FlpMarshaler) {
	fm.w16(uint16(d2.MeshPartIndex))
	fm.w16(uint16(len(d2.Materials)))
	fm.skip(4) // placeholder for materials array memory pointer
}

func (d2 *MeshPartReference) MarshalData(fm *FlpMarshaler) {
	for j := range d2.Materials {
		fm.w32(d2.Materials[j].Color)
		fm.addStringOffsetPlaceholder(d2.Materials[j].TextureName, 4)
	}
}

func (d3 *Font) MarshalStruct(fm *FlpMarshaler) {
	fm.w32(d3.CharsCount)
	fm.w16(d3.Unk04)
	fm.w16(uint16(d3.Size))
	fm.w16(d3.Unk08)
	fm.w16(d3.Unk0a)
	fm.w16(d3.Flags)
	fm.skip(0x12) // placeholders for pointers and padding
	fm.w32(math.Float32bits(d3.Float020))
}

func (d3 *Font) MarshalData(fm *FlpMarshaler) {
	if d3.Flags&2 != 0 {
		for i := range d3.Flag2Datas2 {
			d3.Flag2Datas2[i].MarshalStruct(fm)
		}
		for i := range d3.Flag2Datas2 {
			d3.Flag2Datas2[i].MarshalData(fm)
		}
	}
	if d3.Flags&4 != 0 {
		for i := range d3.Flag4Datas2 {
			d3.Flag4Datas2[i].MarshalStruct(fm)
		}
		for i := range d3.Flag4Datas2 {
			d3.Flag4Datas2[i].MarshalData(fm)
		}
	}

	for _, symbolWidth := range d3.SymbolWidths {
		fm.w16(uint16(symbolWidth))
	}
	fm.pad4()

	for _, charNumberToSymbolId := range d3.CharNumberToSymbolIdMap {
		fm.w16(uint16(charNumberToSymbolId))
	}
	fm.pad4()
}

func (d4 *StaticLabel) MarshalStruct(fm *FlpMarshaler) {
	fm.skip(0x14) // unknown, probably unused
	fm.w32(uint32(len(d4.RenderCommandsList)))
	fm.skip(0xc) // pointer placeholder and unknown stuff
}

func (d4 *StaticLabel) MarshalData(fm *FlpMarshaler) {
	fm.buf.Write(d4.RenderCommandsList)
	fm.pad4()
}

func (d5 *DynamicLabel) MarshalStruct(fm *FlpMarshaler) {
	fm.addStringOffsetPlaceholder(d5.ValueName, 2)
	fm.addStringOffsetPlaceholder(d5.Placeholder, 2)
	fm.w16(uint16(d5.FontHandler))
	fm.w16(d5.Width1)
	fm.w32(d5.BlendColor)
	fm.w16(d5.StringLengthLimit)
	fm.w16(d5.OffsetX1)
	fm.skip(2)
	fm.w16(d5.Width2)
	fm.skip(2)
	fm.w16(d5.OffsetX2)
	fm.skip(2)
	fm.w16(d5.Unk01a)
	fm.skip(2)
	fm.w16(d5.Unk01e)
}

func (d6 *Data6) MarshalStruct(fm *FlpMarshaler) {
	fm.skip(8) // pointer placeholders
	fm.w16(uint16(len(d6.Sub2s)))
	fm.skip(2)
}

func (d6 *Data6) MarshalData(fm *FlpMarshaler) {
	fm.pad4()
	//d6.Sub1.MarshalStruct(fm)
	//d6.Sub1.MarshalData(fm)

	//	for i := range d6.Sub2s {
	//		d6.Sub2s[i].MarshalStruct(fm)
	//	}
	//	for i := range d6.Sub2s {
	//		d6.Sub2s[i].MarshalData(fm)
	//	}
}

/*
func (d6s1 *Data6Subtype1) FromBuf(buf []byte) int {

	d6s1.Sub1s = make([]Data6Subtype1Subtype1, binary.LittleEndian.Uint16(buf[0x2:]))
	d6s1.Sub2s = make([]Data6Subtype1Subtype2, binary.LittleEndian.Uint16(buf[0x4:]))
	return DATA6_SUBTYPE1_ELEMENT_SIZE
}

func (d6s1 *Data6Subtype1) Parse(buf []byte, pos int) int {
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
}*/

func (f *FLP) marshalBufferHeader(fm *FlpMarshaler) {
	fm.w32(FLP_MAGIC)
	fm.skip(8)
	writeArraydescr := func(count int) {
		fm.w32(uint32(count))
		fm.w32(0x61ace01d) // placeholder for in-memory pointer
	}
	writeArraydescr(len(f.GlobalHandlersIndexes))
	writeArraydescr(len(f.MeshPartReferences))
	writeArraydescr(len(f.Fonts))
	writeArraydescr(len(f.StaticLabels))
	writeArraydescr(len(f.DynamicLabels))
	writeArraydescr(len(f.Datas6))
	writeArraydescr(len(f.Datas7))
	fm.skip(4) // placeholder for data8 single instance pointer
	writeArraydescr(len(f.Transformations))
	writeArraydescr(len(f.BlendColors))
	writeArraydescr(0x88005553) // fill this field later (string data offset)
	if fm.pos() != HEADER_SIZE {
		panic("Wrong header generated")
	}
}

func (f *FLP) marshalBuffer() *bytes.Buffer {
	fm := NewFlpMarshaler()

	f.marshalBufferHeader(fm)

	for i := range f.GlobalHandlersIndexes {
		f.GlobalHandlersIndexes[i].MarshalStruct(fm)
	}

	for i := range f.MeshPartReferences {
		f.MeshPartReferences[i].MarshalStruct(fm)
	}
	for i := range f.MeshPartReferences {
		f.MeshPartReferences[i].MarshalData(fm)
	}

	for i := range f.Fonts {
		f.Fonts[i].MarshalStruct(fm)
	}
	for i := range f.Fonts {
		f.Fonts[i].MarshalData(fm)
	}

	for i := range f.StaticLabels {
		f.StaticLabels[i].MarshalStruct(fm)
	}
	for i := range f.StaticLabels {
		f.StaticLabels[i].MarshalData(fm)
	}
	/*
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
	*/
	return &fm.buf
}
