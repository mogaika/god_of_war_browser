package collision

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/mogaika/god_of_war_browser/utils"
)

type ribSheetMarshaler struct {
}

func (rm *ribSheetMarshaler) alignBuf(buf *bytes.Buffer, align int) {
	if off := buf.Len() % align; off != 0 {
		buf.Write(make([]byte, align-off))
	}
}

func (rm *ribSheetMarshaler) insertAlignedSection(buf *bytes.Buffer, section []byte, align int) (offset uint32) {
	rm.alignBuf(buf, align)
	offset = uint32(buf.Len())
	buf.Write(section)
	return
}

func (rm *ribSheetMarshaler) getSection1(rib *ShapeRibSheet) []byte {
	section1 := make([]byte, len(rib.Some1)*8)
	for i, s1 := range rib.Some1 {
		buf := section1[i*8 : (i+1)*8]

		if s1.IsPolygon {
			flag := uint32(1)
			flag |= s1.PolygonIndex << 1
			flag |= (uint32(s1.PolygonFlag) << 24) << 1
			binary.LittleEndian.PutUint32(buf[0:], flag)
			binary.LittleEndian.PutUint16(buf[4:], s1.PolygonsCount)
			binary.LittleEndian.PutUint16(buf[6:], s1.PolygonUnk0x6)
		} else {
			binary.LittleEndian.PutUint32(buf[0:], uint32(s1.PlaneAxis)<<1)
			binary.LittleEndian.PutUint16(buf[2:], s1.PlaneSubNodeHigher)
			binary.LittleEndian.PutUint32(buf[4:], math.Float32bits(s1.PlaneCoordinate))
		}
	}
	return section1
}

func (rm *ribSheetMarshaler) getSection3(rib *ShapeRibSheet) []byte {
	section3 := make([]byte, len(rib.Some3CxtNames)*0x18)
	for i, cxtName := range rib.Some3CxtNames {
		copy(section3[i*0x18:], utils.StringToBytesBuffer(cxtName, 0x18, true))
	}
	return section3
}

func (rm *ribSheetMarshaler) findSection4ElementSize(rib *ShapeRibSheet) int {
	maxOffset := 0
	for _, fm := range rib.Some5FlagMaps {
		if int(fm.OffsetBytes) > maxOffset {
			maxOffset = int(fm.OffsetBytes)
		}
	}
	return maxOffset + 4 + 0x18
}
func (rm *ribSheetMarshaler) getSection4(rib *ShapeRibSheet) []byte {
	elementSize := rm.findSection4ElementSize(rib)
	section4 := make([]byte, len(rib.Some4Materials)*elementSize)
	for i, material := range rib.Some4Materials {
		buf := section4[i*elementSize : (i+1)*elementSize]

		copy(buf[0:0x18], utils.StringToBytesBuffer(material.Name, 0x18, true))

		valsBuf := buf[0x18:]

		for _, f := range rib.Some5FlagMaps {
			val := material.Values[f.Name]

			switch f.Type {
			case 0:
				binary.LittleEndian.PutUint32(valsBuf[f.OffsetBytes:], val.(uint32))
			case 1:
				binary.LittleEndian.PutUint32(valsBuf[f.OffsetBytes:], math.Float32bits(val.(float32)))
			case 2:
				oldBits := binary.LittleEndian.Uint32(valsBuf[f.OffsetBytes:])
				bit := uint32(0)
				if val.(bool) {
					bit = 1
				}

				mask := ^(uint32(1) << (f.OffsetBits - 0x8*f.OffsetBytes))
				value := bit << (f.OffsetBits - 0x8*f.OffsetBytes)
				newBits := (oldBits & mask) | value
				binary.LittleEndian.PutUint32(valsBuf[f.OffsetBytes:], newBits)
			default:
				panic("Invalid value type")
			}
		}
	}
	return section4
}

func (rm *ribSheetMarshaler) getSection5(rib *ShapeRibSheet) []byte {
	section5 := make([]byte, len(rib.Some5FlagMaps)*0x4c)
	for i, s5 := range rib.Some5FlagMaps {
		buf := section5[i*0x4c : (i+1)*0x4c]

		copy(buf[:0x20], utils.StringToBytesBuffer(s5.Name, 0x20, true))
		binary.LittleEndian.PutUint32(buf[0x20:], s5.Type)
		binary.LittleEndian.PutUint32(buf[0x24:], s5.Unk0x24)
		binary.LittleEndian.PutUint32(buf[0x28:], s5.MaxValue)
		binary.LittleEndian.PutUint32(buf[0x2c:], s5.DefaultValue)
		binary.LittleEndian.PutUint32(buf[0x30:], s5.Unk0x30)
		binary.LittleEndian.PutUint32(buf[0x34:], s5.Unk0x34)
		binary.LittleEndian.PutUint16(buf[0x38:], s5.Unk0x38)
		binary.LittleEndian.PutUint16(buf[0x3a:], s5.Unk0x3a)
		binary.LittleEndian.PutUint32(buf[0x3c:], s5.OffsetBytes)
		binary.LittleEndian.PutUint32(buf[0x40:], s5.IsBitField1)
		binary.LittleEndian.PutUint32(buf[0x44:], s5.OffsetBits)
		binary.LittleEndian.PutUint32(buf[0x48:], s5.IsBitField2)
	}
	return section5
}

func (rm *ribSheetMarshaler) getSection6(rib *ShapeRibSheet) []byte {
	section6 := make([]byte, len(rib.Some6)*2)
	for i, s6 := range rib.Some6 {
		val := s6.QuadOrTriangleIndex << 1
		if s6.IsQuad {
			val |= 1
		}
		binary.LittleEndian.PutUint16(section6[i*2:], val)
	}
	return section6
}

func (rm *ribSheetMarshaler) getSection7(rib *ShapeRibSheet) []byte {
	section7 := make([]byte, len(rib.Some7TrianglesIndex)*10)
	for i, s7 := range rib.Some7TrianglesIndex {
		buf := section7[i*10 : (i+1)*10]

		binary.LittleEndian.PutUint16(buf[0:], s7.Flags)
		binary.LittleEndian.PutUint16(buf[2:], s7.MaterialIndex)
		binary.LittleEndian.PutUint16(buf[4:], s7.Indexes[0])
		binary.LittleEndian.PutUint16(buf[6:], s7.Indexes[1])
		binary.LittleEndian.PutUint16(buf[8:], s7.Indexes[2])
	}
	return section7
}

func (rm *ribSheetMarshaler) getSection8(rib *ShapeRibSheet) []byte {
	section8 := make([]byte, len(rib.Some8QuadsIndex)*12)
	for i, s8 := range rib.Some8QuadsIndex {
		buf := section8[i*12 : (i+1)*12]

		binary.LittleEndian.PutUint16(buf[0:], s8.Flags)
		binary.LittleEndian.PutUint16(buf[2:], s8.MaterialIndex)
		binary.LittleEndian.PutUint16(buf[4:], s8.Indexes[0])
		binary.LittleEndian.PutUint16(buf[6:], s8.Indexes[1])
		binary.LittleEndian.PutUint16(buf[8:], s8.Indexes[2])
		binary.LittleEndian.PutUint16(buf[10:], s8.Indexes[3])
	}
	return section8
}

func (rm *ribSheetMarshaler) getSection9(rib *ShapeRibSheet) []byte {
	section9 := make([]byte, len(rib.Some9Points)*12)
	for i, s9 := range rib.Some9Points {
		buf := section9[i*12 : (i+1)*12]

		binary.LittleEndian.PutUint32(buf[0:], math.Float32bits(s9[0]))
		binary.LittleEndian.PutUint32(buf[4:], math.Float32bits(s9[1]))
		binary.LittleEndian.PutUint32(buf[8:], math.Float32bits(s9[2]))
	}
	return section9
}

func (rm *ribSheetMarshaler) getSection2and10(rib *ShapeRibSheet) (section2 []byte, count2 int, section10 []byte) {
	s2ints := make([]uint16, 0)

	fillContextZone := func(cz [8][]uint16) (idx uint32) {
		idx = uint32(len(s2ints))

		s2ints = append(s2ints, 0)
		s2ints = append(s2ints, uint16(len(cz)))
		for key, zoneIds := range cz {
			s2ints = append(s2ints, uint16(key))
			s2ints = append(s2ints, uint16(len(zoneIds)))
			for _, zoneId := range zoneIds {
				s2ints = append(s2ints, uint16(zoneId))
			}
		}
		count2++
		return
	}

	section10 = make([]byte, len(rib.Some10)*0x1c)

	for i, s10 := range rib.Some10 {
		buf := section10[i*0x1c : (i+1)*0x1c]
		for j, v := range s10.FloatArray {
			binary.LittleEndian.PutUint32(buf[j*4:], math.Float32bits(v))
		}
		binary.LittleEndian.PutUint32(buf[0x18:], fillContextZone(s10.ContextZone))
	}

	section2 = make([]byte, len(s2ints)*2)
	for i, v := range s2ints {
		binary.LittleEndian.PutUint16(section2[i*2:], v)
	}

	return section2, count2, section10
}

func (rib *ShapeRibSheet) Marshal() []byte {
	rm := ribSheetMarshaler{}

	var buf bytes.Buffer
	buf.Write(make([]byte, RIBSHEET_HEADER_SIZE))

	section2, count2, section10 := rm.getSection2and10(rib)

	offset1 := rm.insertAlignedSection(&buf, rm.getSection1(rib), 0x8)
	offset2 := rm.insertAlignedSection(&buf, section2, 0x8)
	offset3 := rm.insertAlignedSection(&buf, rm.getSection3(rib), 0x8)
	offset4 := rm.insertAlignedSection(&buf, rm.getSection4(rib), 0x8)
	offset5 := rm.insertAlignedSection(&buf, rm.getSection5(rib), 0x8)
	offset6 := rm.insertAlignedSection(&buf, rm.getSection6(rib), 0x8)
	offset7 := rm.insertAlignedSection(&buf, rm.getSection7(rib), 0x8)
	offset8 := rm.insertAlignedSection(&buf, rm.getSection8(rib), 0x8)
	offset9 := rm.insertAlignedSection(&buf, rm.getSection9(rib), 0x8)
	offset10 := rm.insertAlignedSection(&buf, section10, 0x8)

	raw := buf.Bytes()
	header := raw[:RIBSHEET_HEADER_SIZE]

	binary.LittleEndian.PutUint32(header[0x0:], COLLISION_MAGIC)
	copy(raw[4:], utils.StringToBytesBuffer("SheetHdr", 8, false))
	binary.LittleEndian.PutUint32(header[0xc:], uint32(len(raw)))
	binary.LittleEndian.PutUint32(header[0x10:], rib.Unk0x10)
	binary.LittleEndian.PutUint32(header[0x14:], rib.Unk0x14)
	for i := range rib.LevelBBox {
		for j := range rib.LevelBBox[i] {
			binary.LittleEndian.PutUint32(header[0x18+i*0x10+j*4:], math.Float32bits(rib.LevelBBox[i][j]))
		}
	}

	binary.LittleEndian.PutUint16(header[0x38:], rib.Unk0x38)
	binary.LittleEndian.PutUint16(header[0x3a:], rib.Unk0x3a)
	binary.LittleEndian.PutUint16(header[0x3c:], uint16(len(rib.Some1)))
	binary.LittleEndian.PutUint16(header[0x3e:], rib.Unk0x3e)
	binary.LittleEndian.PutUint16(header[0x40:], uint16(len(rib.Some6)))
	binary.LittleEndian.PutUint16(header[0x42:], rib.Unk0x42)
	binary.LittleEndian.PutUint16(header[0x44:], rib.Unk0x44)
	binary.LittleEndian.PutUint16(header[0x46:], uint16(len(rib.Some7TrianglesIndex)))
	binary.LittleEndian.PutUint16(header[0x48:], uint16(len(rib.Some8QuadsIndex)))
	binary.LittleEndian.PutUint16(header[0x4a:], uint16(len(rib.Some9Points)))
	binary.LittleEndian.PutUint16(header[0x4c:], rib.Unk0x4c)
	binary.LittleEndian.PutUint16(header[0x4e:], rib.Unk0x4e)
	binary.LittleEndian.PutUint16(header[0x50:], uint16(len(rib.Some4Materials)))
	binary.LittleEndian.PutUint16(header[0x52:], uint16(len(rib.Some5FlagMaps)))
	binary.LittleEndian.PutUint16(header[0x54:], rib.Unk0x54)
	binary.LittleEndian.PutUint16(header[0x56:], uint16(count2))
	binary.LittleEndian.PutUint16(header[0x5a:], uint16(len(rib.Some3CxtNames)))
	binary.LittleEndian.PutUint16(header[0x5c:], uint16(len(rib.Some10)))
	binary.LittleEndian.PutUint32(header[0x60:], 0xFFFF_ffff)
	binary.LittleEndian.PutUint32(header[0x64:], offset1)
	binary.LittleEndian.PutUint32(header[0x68:], offset2)
	binary.LittleEndian.PutUint32(header[0x6c:], offset3)
	binary.LittleEndian.PutUint32(header[0x70:], offset4)
	binary.LittleEndian.PutUint32(header[0x74:], offset5)
	binary.LittleEndian.PutUint32(header[0x78:], offset6)
	binary.LittleEndian.PutUint32(header[0x7c:], offset7)
	binary.LittleEndian.PutUint32(header[0x80:], offset8)
	binary.LittleEndian.PutUint32(header[0x84:], offset9)
	binary.LittleEndian.PutUint32(header[0x88:], offset10) // not used, so copy of offset10
	binary.LittleEndian.PutUint32(header[0x8c:], offset10)

	return raw
}
