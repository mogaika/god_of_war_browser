package collision

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
)

const RIBSHEET_HEADER_SIZE = 0x90

type ShapeRibSheet struct {
	CountOfSomethingUnk50 uint16
	CountOfSomethingUnk52 uint16
	CountOfSomethingUnk54 uint16
	CountOfSomethingUnk5a uint16
	CountOfSomethingUnk5c uint16

	OffsetToSome1  uint32
	OffsetToSome2  uint32
	OffsetToSome3  uint32 // offset_to ctxNames   array of char[0x18]
	OffsetToSome4  uint32
	OffsetToSome5  uint32
	OffsetToSome6  uint32
	OffsetToSome7  uint32 // ?? triangle indexes
	OffsetToSome8  uint32 // ?? quads indexes ??
	OffsetToSome9  uint32 // ?? floats array
	OffsetToSome10 uint32

	Some7TrianglesIndex [][3]uint16
	Some8QuadsIndex     [][4]uint16
	Some9Points         []mgl32.Vec3
}

func NewRibSheet(f io.ReaderAt, wrtw io.Writer) (*ShapeRibSheet, error) {
	buf := make([]byte, RIBSHEET_HEADER_SIZE)
	if _, err := f.ReadAt(buf, 0); err != nil {
		return nil, err
	}

	rib := &ShapeRibSheet{
		CountOfSomethingUnk50: binary.LittleEndian.Uint16(buf[0x50:0x52]),
		CountOfSomethingUnk52: binary.LittleEndian.Uint16(buf[0x52:0x54]),
		CountOfSomethingUnk54: binary.LittleEndian.Uint16(buf[0x54:0x56]),
		CountOfSomethingUnk5a: binary.LittleEndian.Uint16(buf[0x5a:0x5c]),
		CountOfSomethingUnk5c: binary.LittleEndian.Uint16(buf[0x5c:0x5e]),

		OffsetToSome1:  binary.LittleEndian.Uint32(buf[0x64:0x68]),
		OffsetToSome2:  binary.LittleEndian.Uint32(buf[0x68:0x6c]),
		OffsetToSome3:  binary.LittleEndian.Uint32(buf[0x6c:0x70]),
		OffsetToSome4:  binary.LittleEndian.Uint32(buf[0x70:0x74]),
		OffsetToSome5:  binary.LittleEndian.Uint32(buf[0x74:0x78]),
		OffsetToSome6:  binary.LittleEndian.Uint32(buf[0x78:0x7c]),
		OffsetToSome7:  binary.LittleEndian.Uint32(buf[0x7c:0x80]),
		OffsetToSome8:  binary.LittleEndian.Uint32(buf[0x80:0x84]),
		OffsetToSome9:  binary.LittleEndian.Uint32(buf[0x84:0x8c]),
		OffsetToSome10: binary.LittleEndian.Uint32(buf[0x8c:0x90]),
	}

	{
		indexBuf := make([]byte, rib.OffsetToSome8-rib.OffsetToSome7)
		if _, err := f.ReadAt(indexBuf, int64(rib.OffsetToSome7)); err != nil {
			return nil, err
		}

		rib.Some7TrianglesIndex = make([][3]uint16, len(indexBuf)/10)
		for i := range rib.Some7TrianglesIndex {
			triaBuf := indexBuf[i*10 : (i+1)*10]
			rib.Some7TrianglesIndex[i][0] = binary.LittleEndian.Uint16(triaBuf[4:6])
			rib.Some7TrianglesIndex[i][1] = binary.LittleEndian.Uint16(triaBuf[6:8])
			rib.Some7TrianglesIndex[i][2] = binary.LittleEndian.Uint16(triaBuf[8:10])
		}
	}

	{
		quadIndexBuf := make([]byte, rib.OffsetToSome9-rib.OffsetToSome8)
		if _, err := f.ReadAt(quadIndexBuf, int64(rib.OffsetToSome8)); err != nil {
			return nil, err
		}

		rib.Some8QuadsIndex = make([][4]uint16, len(quadIndexBuf)/12)
		for i := range rib.Some8QuadsIndex {
			quadBuf := quadIndexBuf[i*12 : (i+1)*12]
			rib.Some8QuadsIndex[i][0] = binary.LittleEndian.Uint16(quadBuf[4:6])
			rib.Some8QuadsIndex[i][1] = binary.LittleEndian.Uint16(quadBuf[6:8])
			rib.Some8QuadsIndex[i][2] = binary.LittleEndian.Uint16(quadBuf[8:10])
			rib.Some8QuadsIndex[i][3] = binary.LittleEndian.Uint16(quadBuf[10:12])
		}
	}

	{
		floatsBuf := make([]byte, rib.OffsetToSome10-rib.OffsetToSome9)
		if _, err := f.ReadAt(floatsBuf, int64(rib.OffsetToSome9)); err != nil {
			return nil, err
		}

		rib.Some9Points = make([]mgl32.Vec3, len(floatsBuf)/0xc)
		for i := range rib.Some9Points {
			if err := binary.Read(bytes.NewReader(floatsBuf[i*0xc:(i+1)*0xc]), binary.LittleEndian, &rib.Some9Points[i]); err != nil {
				return nil, err
			}
		}
		//log.Printf("%d points loaded", len(rib.Some9Points))
	}

	{
		fmt.Fprintf(wrtw, "o ribsheet\n")
		for _, point := range rib.Some9Points {
			fmt.Fprintf(wrtw, "v %f %f %f\n", point[0], point[1], point[2])
		}
		fmt.Fprintf(wrtw, "o triangles\n")
		for _, triIndex := range rib.Some7TrianglesIndex {
			fmt.Fprintf(wrtw, "f %d %d %d\n", triIndex[0]+1, triIndex[1]+1, triIndex[2]+1)
		}
		fmt.Fprintf(wrtw, "o quads\n")
		for _, quadIndex := range rib.Some8QuadsIndex {
			fmt.Fprintf(wrtw, "f %d %d %d\n", quadIndex[0]+1, quadIndex[1]+1, quadIndex[2]+1)
			fmt.Fprintf(wrtw, "f %d %d %d\n", quadIndex[3]+1, quadIndex[0]+1, quadIndex[2]+1)
		}
	}

	return rib, nil
}
