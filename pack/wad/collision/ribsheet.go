package collision

import (
	"fmt"
	"io"

	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/go-gl/mathgl/mgl32"
)

const RIBSHEET_HEADER_SIZE = 0x90

type ShapeRibSheet struct {
	CountOfSomethingUnk38 uint16
	CountOfSomethingUnk3a uint16
	CountOfSomethingUnk3c uint16
	CountOfSomethingUnk3e uint16

	CountOfSomethingUnk40 uint16
	CountOfSomethingUnk42 uint16
	CountOfSomethingUnk44 uint16
	CountOfSomethingUnk46 uint16
	CountOfSomethingUnk48 uint16
	CountOfSomethingUnk4a uint16 // count of floats @9
	CountOfSomethingUnk4c uint16
	CountOfSomethingUnk4e uint16

	CountOfSomethingUnk50 uint16 // count of debug materials @4
	CountOfSomethingUnk52 uint16 // count of logical fields or flags @5
	CountOfSomethingUnk54 uint16 //
	CountOfSomethingUnk56 uint16
	CountOfSomethingUnk5a uint16
	CountOfSomethingUnk5c uint16

	OffsetToSome64 uint32
	OffsetToSome68 uint32
	OffsetToSome6c uint32 // offset to ctx names array of (element size 0x18) @3
	OffsetToSome70 uint32 // offset to debug materials (element size 0x40) @4
	OffsetToSome74 uint32 // offset to logical fields or flags (element size 0x4c) @5
	OffsetToSome78 uint32
	OffsetToSome7c uint32 // offset to triangle indexes (element size 0xa) @7
	OffsetToSome80 uint32 // offset to quads indexes (element size 0xc) @8
	OffsetToSome84 uint32 // offset to floats array (element size 0xc)  @9
	OffsetToSome88 uint32
	OffsetToSome8c uint32

	Some7TrianglesIndex [][3]uint16
	Some8QuadsIndex     [][4]uint16
	Some9Points         []mgl32.Vec3
}

func NewRibSheet(bs *utils.BufStack, wrtw io.Writer) (*ShapeRibSheet, error) {
	headerbs := bs.SubBuf("header", 0).SetSize(RIBSHEET_HEADER_SIZE)

	rib := &ShapeRibSheet{
		CountOfSomethingUnk38: headerbs.LU16(0x38),
		CountOfSomethingUnk3a: headerbs.LU16(0x3a),
		CountOfSomethingUnk3c: headerbs.LU16(0x3c),
		CountOfSomethingUnk3e: headerbs.LU16(0x3e),

		CountOfSomethingUnk40: headerbs.LU16(0x40),
		CountOfSomethingUnk42: headerbs.LU16(0x42),
		CountOfSomethingUnk44: headerbs.LU16(0x44),
		CountOfSomethingUnk46: headerbs.LU16(0x46),
		CountOfSomethingUnk48: headerbs.LU16(0x48),
		CountOfSomethingUnk4a: headerbs.LU16(0x4a),
		CountOfSomethingUnk4c: headerbs.LU16(0x4c),
		CountOfSomethingUnk4e: headerbs.LU16(0x4e),

		CountOfSomethingUnk50: headerbs.LU16(0x50),
		CountOfSomethingUnk52: headerbs.LU16(0x52),
		CountOfSomethingUnk54: headerbs.LU16(0x54),
		CountOfSomethingUnk56: headerbs.LU16(0x56),
		CountOfSomethingUnk5a: headerbs.LU16(0x5a),
		CountOfSomethingUnk5c: headerbs.LU16(0x5c),

		OffsetToSome64: headerbs.LU32(0x64),
		OffsetToSome68: headerbs.LU32(0x68),
		OffsetToSome6c: headerbs.LU32(0x6c),
		OffsetToSome70: headerbs.LU32(0x70),
		OffsetToSome74: headerbs.LU32(0x74),
		OffsetToSome78: headerbs.LU32(0x78),
		OffsetToSome7c: headerbs.LU32(0x7c),
		OffsetToSome80: headerbs.LU32(0x80),
		OffsetToSome84: headerbs.LU32(0x84),
		OffsetToSome88: headerbs.LU32(0x88),
		OffsetToSome8c: headerbs.LU32(0x8c),
	}

	{
		triaIndexBs := bs.SubBuf("triaIndexBuffer", int(rib.OffsetToSome7c)).SetSize(int(rib.OffsetToSome80 - rib.OffsetToSome7c))

		rib.Some7TrianglesIndex = make([][3]uint16, triaIndexBs.Size()/0xa)
		for i := range rib.Some7TrianglesIndex {
			off := i * 10
			rib.Some7TrianglesIndex[i][0] = triaIndexBs.LU16(off + 4)
			rib.Some7TrianglesIndex[i][1] = triaIndexBs.LU16(off + 6)
			rib.Some7TrianglesIndex[i][2] = triaIndexBs.LU16(off + 8)
		}
	}

	{
		quadIndexBs := bs.SubBuf("quadIndexBuffer", int(rib.OffsetToSome80)).SetSize(int(rib.OffsetToSome84 - rib.OffsetToSome80))

		rib.Some8QuadsIndex = make([][4]uint16, quadIndexBs.Size()/0xc)
		for i := range rib.Some8QuadsIndex {
			off := i * 12
			rib.Some8QuadsIndex[i][0] = quadIndexBs.LU16(off + 4)
			rib.Some8QuadsIndex[i][1] = quadIndexBs.LU16(off + 6)
			rib.Some8QuadsIndex[i][2] = quadIndexBs.LU16(off + 8)
			rib.Some8QuadsIndex[i][3] = quadIndexBs.LU16(off + 10)
		}
	}

	{
		floatsBs := bs.SubBuf("floatsBuffer", int(rib.OffsetToSome84)).SetSize(int(rib.OffsetToSome88 - rib.OffsetToSome84))

		rib.Some9Points = make([]mgl32.Vec3, floatsBs.Size()/0xc)
		for i := range rib.Some9Points {
			off := i * 12
			rib.Some9Points[i][0] = floatsBs.LF(off + 0)
			rib.Some9Points[i][1] = floatsBs.LF(off + 4)
			rib.Some9Points[i][2] = floatsBs.LF(off + 8)
		}
		// log.Printf("%d points loaded", len(rib.Some9Points))
	}

	// log.Printf("TREEEEEEEEEE\n%s", bs.StringTree())

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
