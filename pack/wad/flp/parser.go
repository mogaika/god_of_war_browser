package flp

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/utils"
)

func (d1 *GlobalHandlerIndex) FromBuf(buf []byte) int {
	d1.TypeArrayId = binary.LittleEndian.Uint16(buf[:])
	d1.IdInThatTypeArray = binary.LittleEndian.Uint16(buf[2:])
	return DATA1_ELEMENT_SIZE
}

func (d2 *MeshPartReference) FromBuf(buf []byte) int {
	if config.GetGOWVersion() == config.GOW1 {
		d2.MeshPartIndex = int16(binary.LittleEndian.Uint16(buf[:]))
		d2.Materials = make([]MeshPartMaterialSlot, binary.LittleEndian.Uint16(buf[2:]))
	} else {
		d2.MeshPartIndex = int16(binary.LittleEndian.Uint16(buf[4:]))
		d2.Materials = make([]MeshPartMaterialSlot, binary.LittleEndian.Uint16(buf[6:]))
	}
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

func (d2s1 *MeshPartMaterialSlot) SetNameFromStringSector(stringsSector []byte) {
	if d2s1.TextureNameSecOff != 0xffff && d2s1.TextureNameSecOff != 0xffffffff {
		if config.GetGOWVersion() == config.GOW1 {
			d2s1.TextureName = utils.BytesToString(stringsSector[d2s1.TextureNameSecOff:])
		} else {
			d2s1.TextureName = fmt.Sprintf("!indexed %d", d2s1.TextureNameSecOff)
		}
	}
}

func (d3 *Font) FromBuf(buf []byte) int {
	if config.GetGOWVersion() == config.GOW1 {
		d3.CharsCount = binary.LittleEndian.Uint32(buf[:])
		d3.Unk04 = binary.LittleEndian.Uint16(buf[4:])
		d3.Size = int16(binary.LittleEndian.Uint16(buf[6:]))
		d3.Unk08 = binary.LittleEndian.Uint16(buf[8:])
		d3.Unk0a = binary.LittleEndian.Uint16(buf[0xa:])
		d3.Flags = binary.LittleEndian.Uint16(buf[0xc:])
		d3.Float020 = math.Float32frombits(binary.LittleEndian.Uint32(buf[0x20:]))
	} else {
		d3.CharsCount = binary.LittleEndian.Uint32(buf[0x10:])
		d3.Float020 = math.Float32frombits(binary.LittleEndian.Uint32(buf[0x14:]))
		d3.Unk04 = binary.LittleEndian.Uint16(buf[0x18:])
		d3.Size = int16(binary.LittleEndian.Uint16(buf[0x1a:]))
		d3.Unk08 = binary.LittleEndian.Uint16(buf[0x1c:])
		d3.Unk0a = binary.LittleEndian.Uint16(buf[0x1e:])
		d3.Flags = binary.LittleEndian.Uint16(buf[0x20:])
	}
	return DATA3_ELEMENT_SIZE
}

func (d3 *Font) Parse(buf []byte, pos int) int {
	if d3.Flags&(4|2) == (4 | 2) {
		panic("d3.Flags &(4|2) == (4|2)")
	}
	if d3.Flags&(2|4) != 0 {
		d3.MeshesRefs = make([]MeshPartReference, d3.CharsCount)
		for i := range d3.MeshesRefs {
			pos += d3.MeshesRefs[i].FromBuf(buf[pos:])
		}
		for i := range d3.MeshesRefs {
			pos = d3.MeshesRefs[i].Parse(buf, pos)
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

func (d4 *StaticLabel) FromBuf(buf []byte) int {
	if config.GetGOWVersion() == config.GOW1 {
		d4.Transformation.FromBuf(buf[0:])
		d4.tempRenderCommandBuffer = make([]byte, binary.LittleEndian.Uint32(buf[0x14:]))
		return DATA4_ELEMENT_SIZE
	} else {
		d4.Transformation.FromBuf(buf[4:])
		d4.tempRenderCommandBuffer = make([]byte, binary.LittleEndian.Uint32(buf[0x18:]))
		return DATA4_ELEMENT_SIZE_GOW2
	}
}

func (d4 *StaticLabel) Parse(f *FLP, buf []byte, pos int) int {
	res := posPad4(pos + copy(d4.tempRenderCommandBuffer, buf[pos:pos+len(d4.tempRenderCommandBuffer)]))
	d4.ParseRenderCommandList(d4.tempRenderCommandBuffer)
	d4.tempRenderCommandBuffer = nil
	return res
}

func (d5 *DynamicLabel) FromBuf(buf []byte) int {
	d5.valueNameSecOff = binary.LittleEndian.Uint16(buf[0:])
	d5.placeholderSecOff = binary.LittleEndian.Uint16(buf[2:])
	d5.FontHandler = GlobalHandler(binary.LittleEndian.Uint16(buf[4:]))
	d5.Width1 = binary.LittleEndian.Uint16(buf[6:])
	d5.BlendColor = binary.LittleEndian.Uint32(buf[8:])
	d5.StringLengthLimit = binary.LittleEndian.Uint16(buf[0xc:])
	d5.OffsetX1 = binary.LittleEndian.Uint16(buf[0xe:])
	d5.Width2 = binary.LittleEndian.Uint16(buf[0x12:])
	d5.OffsetX2 = binary.LittleEndian.Uint16(buf[0x16:])
	d5.Unk01a = binary.LittleEndian.Uint16(buf[0x1a:])
	d5.Unk01e = binary.LittleEndian.Uint16(buf[0x1e:])
	return DATA5_ELEMENT_SIZE
}

func (d5 *DynamicLabel) SetNameFromStringSector(stringsSector []byte) {
	if d5.valueNameSecOff != 0xffff {
		d5.ValueName = utils.BytesToString(stringsSector[d5.valueNameSecOff:])
	}
	if d5.placeholderSecOff != 0xffff {
		d5.Placeholder = utils.BytesToString(stringsSector[d5.placeholderSecOff:])
	}
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

func (d6 *Data6) SetNameFromStringSector(stringsSector []byte) {
	d6.Sub1.SetNameFromStringSector(stringsSector)
	for i := range d6.Sub2s {
		d6.Sub2s[i].SetNameFromStringSector(stringsSector)
	}
}

func (d6s1 *Data6Subtype1) FromBuf(buf []byte) int {
	if config.GetGOWVersion() == config.GOW1 {
		d6s1.TotalFramesCount = binary.LittleEndian.Uint16(buf[0:])
		d6s1.ElementsAnimation = make([]ElementAnimation, binary.LittleEndian.Uint16(buf[0x2:]))
		d6s1.FrameScriptLables = make([]FrameScriptLabel, binary.LittleEndian.Uint16(buf[0x4:]))
		d6s1.Width = binary.LittleEndian.Uint16(buf[0xa:])
	} else {
		d6s1.TotalFramesCount = binary.LittleEndian.Uint16(buf[8:])
		d6s1.ElementsAnimation = make([]ElementAnimation, binary.LittleEndian.Uint16(buf[0xa:]))
		d6s1.FrameScriptLables = make([]FrameScriptLabel, binary.LittleEndian.Uint16(buf[0xc:]))
		d6s1.Width = binary.LittleEndian.Uint16(buf[0xe:])
	}

	return DATA6_SUBTYPE1_ELEMENT_SIZE
}

func (d6s1 *Data6Subtype1) SetNameFromStringSector(stringsSector []byte) {
	for i := range d6s1.ElementsAnimation {
		d6s1.ElementsAnimation[i].SetNameFromStringSector(stringsSector)
	}
	for i := range d6s1.FrameScriptLables {
		d6s1.FrameScriptLables[i].SetNameFromStringSector(stringsSector)
	}
}

func (d6s1 *Data6Subtype1) Parse(buf []byte, pos int) int {
	pos = posPad4(pos)
	for i := range d6s1.ElementsAnimation {
		pos += d6s1.ElementsAnimation[i].FromBuf(buf[pos:])
	}
	for i := range d6s1.ElementsAnimation {
		pos = d6s1.ElementsAnimation[i].Parse(buf, pos)
	}

	pos = posPad4(pos)
	for i := range d6s1.FrameScriptLables {
		pos += d6s1.FrameScriptLables[i].FromBuf(buf[pos:])
	}
	for i := range d6s1.FrameScriptLables {
		pos = d6s1.FrameScriptLables[i].Parse(buf, pos)
	}
	return pos
}

func (d6s1s1 *ElementAnimation) FromBuf(buf []byte) int {
	if config.GetGOWVersion() == config.GOW1 {
		d6s1s1.FramesCount = binary.LittleEndian.Uint16(buf[0:])
		d6s1s1.KeyFrames = make([]KeyFrame, binary.LittleEndian.Uint16(buf[0x2:]))
	} else {
		d6s1s1.FramesCount = binary.LittleEndian.Uint16(buf[4:])
		d6s1s1.KeyFrames = make([]KeyFrame, binary.LittleEndian.Uint16(buf[0x6:]))
	}
	return DATA6_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE
}

func (d6s1s1 *ElementAnimation) Parse(buf []byte, pos int) int {
	pos = posPad4(pos)
	for i := range d6s1s1.KeyFrames {
		pos += d6s1s1.KeyFrames[i].FromBuf(buf[pos:])
	}
	return pos
}

func (d6s1s1 *ElementAnimation) SetNameFromStringSector(stringsSector []byte) {
	for i := range d6s1s1.KeyFrames {
		d6s1s1.KeyFrames[i].SetNameFromStringSector(stringsSector)
	}
}

func (d6s1s1s1 *KeyFrame) FromBuf(buf []byte) int {
	d6s1s1s1.WhenThisFrameEnds = binary.LittleEndian.Uint16(buf[0:])
	d6s1s1s1.ElementHandler = GlobalHandler(binary.LittleEndian.Uint16(buf[2:]))
	d6s1s1s1.TransformationId = binary.LittleEndian.Uint16(buf[4:])
	d6s1s1s1.ColorId = binary.LittleEndian.Uint16(buf[6:])
	d6s1s1s1.nameSecOff = uint16(binary.LittleEndian.Uint16(buf[8:]))
	return DATA6_SUBTYPE1_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE
}

func (d6s1s1s1 *KeyFrame) SetNameFromStringSector(stringsSector []byte) {
	if d6s1s1s1.nameSecOff != 0xffff {
		d6s1s1s1.Name = utils.BytesToString(stringsSector[d6s1s1s1.nameSecOff:])
	}
}

func (d6s1s2 *FrameScriptLabel) FromBuf(buf []byte) int {
	if config.GetGOWVersion() == config.GOW1 {
		d6s1s2.TriggerFrameNumber = binary.LittleEndian.Uint16(buf[:])
		d6s1s2.Subs = make([]Data6Subtype1Subtype2Subtype1, binary.LittleEndian.Uint16(buf[0x2:]))
	} else {
		d6s1s2.TriggerFrameNumber = binary.LittleEndian.Uint16(buf[4:])
		d6s1s2.Subs = make([]Data6Subtype1Subtype2Subtype1, binary.LittleEndian.Uint16(buf[0x6:]))
	}
	d6s1s2.labelNameSecOff = binary.LittleEndian.Uint16(buf[8:])
	return DATA6_SUBTYPE1_SUBTYPE2_ELEMENT_SIZE
}

func (d6s1s2 *FrameScriptLabel) Parse(buf []byte, pos int) int {
	pos = posPad4(pos)
	for i := range d6s1s2.Subs {
		pos += d6s1s2.Subs[i].FromBuf(buf[pos:])
	}
	for i := range d6s1s2.Subs {
		pos = d6s1s2.Subs[i].Parse(buf, pos)
	}
	return pos
}

func (d6s1s2 *FrameScriptLabel) SetNameFromStringSector(stringsSector []byte) {
	if d6s1s2.labelNameSecOff != 0xffff {
		d6s1s2.LabelName = utils.BytesToString(stringsSector[uint16(d6s1s2.labelNameSecOff):])
	}
	for i := range d6s1s2.Subs {
		d6s1s2.Subs[i].SetNameFromStringSector(stringsSector)
	}
}

func (d6s1s2s1 *Data6Subtype1Subtype2Subtype1) FromBuf(buf []byte) int {
	if config.GetGOWVersion() == config.GOW1 {
		d6s1s2s1.scriptDataLength = binary.LittleEndian.Uint32(buf[:])
	} else {
		d6s1s2s1.scriptDataLength = binary.LittleEndian.Uint32(buf[4:])
	}
	return DATA6_SUBTYPE1_SUBTYPE2_SUBTYPE1_ELEMENT_SIZE
}

func (d6s1s2s1 *Data6Subtype1Subtype2Subtype1) Parse(buf []byte, pos int) int {
	pos = posPad4(pos)
	d6s1s2s1.scriptData = buf[pos : pos+int(d6s1s2s1.scriptDataLength)]

	return pos + int(d6s1s2s1.scriptDataLength)
}

func (d6s1s2s1 *Data6Subtype1Subtype2Subtype1) SetNameFromStringSector(stringsSector []byte) {
	d6s1s2s1.Script = NewScriptFromData(d6s1s2s1.scriptData, stringsSector)
	d6s1s2s1.scriptData = nil
}

func (d6s2 *Data6Subtype2) FromBuf(buf []byte) int {
	if config.GetGOWVersion() == config.GOW1 {
		d6s2.scriptDataLength = binary.LittleEndian.Uint32(buf[:])
	} else {
		d6s2.scriptDataLength = binary.LittleEndian.Uint32(buf[4:])
	}
	d6s2.EventKeysMask = binary.LittleEndian.Uint32(buf[0x8:])
	d6s2.EventUnkMask = binary.LittleEndian.Uint16(buf[0xc:])
	return DATA6_SUBTYPE2_ELEMENT_SIZE
}

func (d6s2 *Data6Subtype2) Parse(buf []byte, pos int) int {
	pos = posPad4(pos)
	d6s2.scriptData = buf[pos : pos+int(d6s2.scriptDataLength)]
	return pos + int(d6s2.scriptDataLength)
}

func (d6s2 *Data6Subtype2) SetNameFromStringSector(stringsSector []byte) {
	d6s2.Script = NewScriptFromData(d6s2.scriptData, stringsSector)
	d6s2.scriptData = nil
}

func (d9 *Transformation) FromBuf(buf []byte) int {
	for i := range d9.Matrix {
		d9.Matrix[i] = float64(int32(binary.LittleEndian.Uint32(buf[i*4:]))) / 65536.0
	}
	d9.OffsetX = float64(int16(binary.LittleEndian.Uint16(buf[0x10:]))) / 16.0
	d9.OffsetY = float64(int16(binary.LittleEndian.Uint16(buf[0x12:]))) / 16.0
	return DATA9_ELEMENT_SIZE
}

func (d10 *BlendColor) FromBuf(buf []byte) int {
	for i := range d10.Color {
		d10.Color[i] = binary.LittleEndian.Uint16(buf[i*2:])
	}
	return DATA10_ELEMENT_SIZE
}

func (f *FLP) fromBuffer(buf []byte) error {
	var pos int
	f.Strings = make([]string, 0)

	if config.GetGOWVersion() == config.GOW1 {
		f.Unk04 = binary.LittleEndian.Uint32(buf[0x4:])
		f.Unk08 = binary.LittleEndian.Uint32(buf[0x8:])
		f.GlobalHandlersIndexes = make([]GlobalHandlerIndex, binary.LittleEndian.Uint32(buf[0xc:]))
		f.MeshPartReferences = make([]MeshPartReference, binary.LittleEndian.Uint32(buf[0x14:]))
		f.Fonts = make([]Font, binary.LittleEndian.Uint32(buf[0x1c:]))
		f.StaticLabels = make([]StaticLabel, binary.LittleEndian.Uint32(buf[0x24:]))
		f.DynamicLabels = make([]DynamicLabel, binary.LittleEndian.Uint32(buf[0x2c:]))
		f.Datas6 = make([]Data6, binary.LittleEndian.Uint32(buf[0x34:]))
		f.Datas7 = make([]Data6Subtype1, binary.LittleEndian.Uint32(buf[0x3c:]))
		f.Transformations = make([]Transformation, binary.LittleEndian.Uint16(buf[0x48:]))
		f.BlendColors = make([]BlendColor, binary.LittleEndian.Uint16(buf[0x50:]))
		pos = HEADER_SIZE
	} else {
		f.Unk04 = binary.LittleEndian.Uint32(buf[0x30:])
		f.Unk08 = binary.LittleEndian.Uint32(buf[0x34:])
		f.GlobalHandlersIndexes = make([]GlobalHandlerIndex, binary.LittleEndian.Uint32(buf[0x38:]))
		f.MeshPartReferences = make([]MeshPartReference, binary.LittleEndian.Uint32(buf[0x3c:]))
		f.Fonts = make([]Font, binary.LittleEndian.Uint32(buf[0x40:]))
		f.StaticLabels = make([]StaticLabel, binary.LittleEndian.Uint32(buf[0x44:]))
		f.DynamicLabels = make([]DynamicLabel, binary.LittleEndian.Uint32(buf[0x48:]))
		f.Datas6 = make([]Data6, binary.LittleEndian.Uint32(buf[0x4c:]))
		f.Datas7 = make([]Data6Subtype1, binary.LittleEndian.Uint32(buf[0x50:]))
		f.Transformations = make([]Transformation, binary.LittleEndian.Uint16(buf[0x54:]))
		f.BlendColors = make([]BlendColor, binary.LittleEndian.Uint16(buf[0x56:]))
		f.Strings = make([]string, 0)
		pos = HEADER_SIZE_GOW2
	}

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

	pos = posPad4(pos)
	for i := range f.Datas7 {
		pos += f.Datas7[i].FromBuf(buf[pos:])
	}
	for i := range f.Datas7 {
		pos = f.Datas7[i].Parse(buf, pos)
	}

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
		for j := range f.Fonts[i].MeshesRefs {
			f.Fonts[i].MeshesRefs[j].SetNameFromStringSector(stringsSector)
		}
	}

	for i := range f.DynamicLabels {
		f.DynamicLabels[i].SetNameFromStringSector(stringsSector)
	}

	for i := range f.Datas6 {
		f.Datas6[i].SetNameFromStringSector(stringsSector)
	}

	for i := range f.Datas7 {
		f.Datas7[i].SetNameFromStringSector(stringsSector)
	}
	f.Data8.SetNameFromStringSector(stringsSector)
}
