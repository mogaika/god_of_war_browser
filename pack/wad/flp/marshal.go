package flp

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/mogaika/god_of_war_browser/utils"
)

type StringPosReplacerReference struct {
	Position         int
	SizeInBytes      int
	AllowEmptyString bool
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

func (fm *FlpMarshaler) compileStringAndReturnFile() *bytes.Buffer {
	buf := fm.buf.Bytes()

	var stringSection bytes.Buffer
	stringSection.WriteByte(0) // empty string at start

	for k, v := range map[string][]StringPosReplacerReference(fm.sbuffer) {
		var offbuf [4]byte
		off := stringSection.Len()

		if k == "" {
			off = -1
		} else {
			stringSection.Write(utils.StringToBytes(k, true))
		}

		binary.LittleEndian.PutUint32(offbuf[:], uint32(off))
		for _, e := range v {
			//log.Printf("String ref at %x:%d = %x to %s", e.Position, e.SizeInBytes, off, k)
			if k == "" && e.AllowEmptyString {
				copy(buf[e.Position:e.Position+e.SizeInBytes], make([]byte, e.SizeInBytes))
			} else {
				copy(buf[e.Position:e.Position+e.SizeInBytes], offbuf[:e.SizeInBytes])
			}
		}
	}

	// update flp field with string section size
	binary.LittleEndian.PutUint32(buf[0x58:], uint32(stringSection.Len()))

	fm.buf.Write(stringSection.Bytes())
	return &fm.buf
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
	if fm.pos()%align != 0 {
		fm.skip(align - fm.pos()%align)
	}
}

func (fm *FlpMarshaler) pad4() {
	fm.pad(4)
}

func (fm *FlpMarshaler) addScriptStringRefs(script *Script) {
	if script.staticStringRef != nil {
		pos := fm.pos()
		for str := range script.staticStringRef {
			for _, offset := range script.staticStringRef[str] {
				fm.sbuffer.Add(str, pos+int(offset), 2).AllowEmptyString = true
			}
		}
	}
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
		fm.addStringOffsetPlaceholder(d2.Materials[j].TextureName, 2)
		fm.skip(2)
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
	if d3.Flags&(4|2) == (4 | 2) {
		panic("d3.Flags &(4|2) == (4|2")
	}
	if d3.Flags&2 != 0 || d3.Flags&4 != 0 {
		for i := range d3.MeshesRefs {
			d3.MeshesRefs[i].MarshalStruct(fm)
		}
		for i := range d3.MeshesRefs {
			d3.MeshesRefs[i].MarshalData(fm)
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
	d4.Transformation.MarshalStruct(fm)
	d4.tempRenderCommandBuffer = d4.MarshalRenderCommandList()
	fm.w32(uint32(len(d4.tempRenderCommandBuffer)))
	fm.skip(0xc) // pointer placeholder and unknown stuff
}

func (d4 *StaticLabel) MarshalData(fm *FlpMarshaler) {
	fm.buf.Write(d4.tempRenderCommandBuffer)
	d4.tempRenderCommandBuffer = nil
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
	d6.Sub1.MarshalStruct(fm)
	d6.Sub1.MarshalData(fm)

	for i := range d6.Sub2s {
		d6.Sub2s[i].MarshalStruct(fm)
	}
	for i := range d6.Sub2s {
		d6.Sub2s[i].MarshalData(fm)
	}
}

func (d6s1 *Data6Subtype1) MarshalStruct(fm *FlpMarshaler) {
	fm.w16(d6s1.TotalFramesCount)
	fm.w16(uint16(len(d6s1.ElementsAnimation)))
	fm.w16(uint16(len(d6s1.FrameScriptLables)))
	fm.skip(4)
	fm.w16(d6s1.Width)
	fm.skip(0xc) // placeholders for pointers
}

func (d6s1 *Data6Subtype1) MarshalData(fm *FlpMarshaler) {
	fm.pad4()
	for i := range d6s1.ElementsAnimation {
		d6s1.ElementsAnimation[i].MarshalStruct(fm)
	}
	for i := range d6s1.ElementsAnimation {
		d6s1.ElementsAnimation[i].MarshalData(fm)
	}
	fm.pad4()
	for i := range d6s1.FrameScriptLables {
		d6s1.FrameScriptLables[i].MarshalStruct(fm)
	}
	for i := range d6s1.FrameScriptLables {
		d6s1.FrameScriptLables[i].MarshalData(fm)
	}
}

func (d6s1s1 *ElementAnimation) MarshalStruct(fm *FlpMarshaler) {
	fm.w16(d6s1s1.FramesCount)
	fm.w16(uint16(len(d6s1s1.KeyFrames)))
	fm.skip(4) // placeholder for pointer
}

func (d6s1s1 *ElementAnimation) MarshalData(fm *FlpMarshaler) {
	fm.pad4()
	for i := range d6s1s1.KeyFrames {
		d6s1s1.KeyFrames[i].MarshalStruct(fm)
	}
}

func (d6s1s1s1 *KeyFrame) MarshalStruct(fm *FlpMarshaler) {
	fm.w16(d6s1s1s1.WhenThisFrameEnds)
	fm.w16(uint16(d6s1s1s1.ElementHandler))
	fm.w16(d6s1s1s1.TransformationId)
	fm.w16(d6s1s1s1.ColorId)
	fm.addStringOffsetPlaceholder(d6s1s1s1.Name, 2)
}

func (d6s1s2 *FrameScriptLabel) MarshalStruct(fm *FlpMarshaler) {
	fm.w16(d6s1s2.TriggerFrameNumber)
	fm.w16(uint16(len(d6s1s2.Subs)))
	fm.skip(4) // array pointer placeholder
	fm.addStringOffsetPlaceholder(d6s1s2.LabelName, 2)
	fm.skip(2)
}

func (d6s1s2 *FrameScriptLabel) MarshalData(fm *FlpMarshaler) {
	fm.pad4()
	for i := range d6s1s2.Subs {
		d6s1s2.Subs[i].MarshalStruct(fm)
	}
	for i := range d6s1s2.Subs {
		d6s1s2.Subs[i].MarshalData(fm)
	}
}

func (d6s1s2s1 *Data6Subtype1Subtype2Subtype1) MarshalStruct(fm *FlpMarshaler) {
	fm.w32(uint32(len(d6s1s2s1.Script.Marshal())))
	fm.skip(4) // placeholder for pointer to array
}

func (d6s1s2s1 *Data6Subtype1Subtype2Subtype1) MarshalData(fm *FlpMarshaler) {
	fm.pad4()
	fm.addScriptStringRefs(d6s1s2s1.Script)
	fm.buf.Write(d6s1s2s1.Script.marshaled)
}

func (d6s2 *Data6Subtype2) MarshalStruct(fm *FlpMarshaler) {
	fm.w32(uint32(len(d6s2.Script.Marshal())))
	fm.skip(4) // placeholder for pointer to script payload
	fm.w32(d6s2.EventKeysMask)
	fm.w16(d6s2.EventUnkMask)
	fm.skip(2)
}

func (d6s2 *Data6Subtype2) MarshalData(fm *FlpMarshaler) {
	fm.pad4()
	fm.addScriptStringRefs(d6s2.Script)
	fm.buf.Write(d6s2.Script.marshaled)
}

func (d9 *Transformation) MarshalStruct(fm *FlpMarshaler) {
	for _, v := range d9.Matrix {
		fm.w32(uint32(int32(v * 65536.0)))
	}
	fm.w16(uint16(int16(d9.OffsetX * 16.0)))
	fm.w16(uint16(int16(d9.OffsetY * 16.0)))

}

func (d10 *BlendColor) MarshalStruct(fm *FlpMarshaler) {
	for _, v := range d10.Color {
		fm.w16(v)
	}
}

func (f *FLP) marshalBufferHeader(fm *FlpMarshaler) {
	fm.w32(FLP_MAGIC)
	fm.w32(f.Unk04)
	fm.w32(f.Unk08)
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
	writeArraydescr(0) // fill this field later (string data offset)
	if fm.pos() != HEADER_SIZE {
		panic("Wrong header generated")
	}
}

func (f *FLP) marshalBufferWithHeader() *bytes.Buffer {
	if config.GetPlayStationVersion() == config.PS3 {
		log.Panicf("Unsupported playstation version")
	}

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

	for i := range f.DynamicLabels {
		f.DynamicLabels[i].MarshalStruct(fm)
	}

	for i := range f.Datas6 {
		f.Datas6[i].MarshalStruct(fm)
	}
	for i := range f.Datas6 {
		f.Datas6[i].MarshalData(fm)
	}
	fm.pad4()

	for i := range f.Datas7 {
		f.Datas7[i].MarshalStruct(fm)
	}
	for i := range f.Datas7 {
		f.Datas7[i].MarshalData(fm)
	}
	fm.pad4()

	f.Data8.MarshalStruct(fm)
	f.Data8.MarshalData(fm)
	fm.pad4()

	for i := range f.Transformations {
		f.Transformations[i].MarshalStruct(fm)
	}

	for i := range f.BlendColors {
		f.BlendColors[i].MarshalStruct(fm)
	}

	return fm.compileStringAndReturnFile()
}
