package collision

import (
	"fmt"
	"io"

	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/go-gl/mathgl/mgl32"
)

const RIBSHEET_HEADER_SIZE = 0x90

type RibSome1 struct {
	IsContainsSome6Index bool

	Float0x4 float32

	IndexToSome6 uint32
	Unk0x4       uint16
	Unk0x6       uint16
}

type RibSome6 struct {
	IsQuad        bool
	IndexOfSome78 uint16
}

type RibSome78Base struct {
	Flags             uint16
	IndexOfSome4Debug uint16
}

type RibSome7Tria struct {
	RibSome78Base
	Indexes [3]uint16
}

type RibSome8Quad struct {
	RibSome78Base
	Indexes [4]uint16
}

type ShapeRibSheet struct {
	CountOfSomethingUnk38 uint16
	CountOfSomethingUnk3a uint16
	CountOfSome1          uint16 // count of @1 ?
	CountOfSomethingUnk3e uint16

	CountOfSomethingUnk40 uint16 // count of @6 ?
	CountOfSomethingUnk42 uint16
	CountOfSomethingUnk44 uint16
	CountOfSomethingUnk46 uint16
	CountOfSomethingUnk48 uint16
	CountOfSomethingUnk4a uint16 // count of @9
	CountOfSomethingUnk4c uint16
	CountOfSomethingUnk4e uint16

	CountOfSome4          uint16 // count of debug materials @4
	CountOfSomethingUnk52 uint16 // count of logical fields or flags @5
	CountOfSomethingUnk54 uint16
	CountOfSomethingUnk56 uint16
	CountOfSomethingUnk5a uint16
	CountOfSomethingUnk5c uint16

	OffsetToSome1  uint32 // offset to (element size 0x8) @1
	OffsetToSome2  uint32 // offset to @2
	OffsetToSome3  uint32 // offset to ctx names array of (element size 0x18) @3
	OffsetToSome4  uint32 // offset to debug materials (element size 0x40 but can vary?) @4
	OffsetToSome5  uint32 // offset to logical fields or flags (element size 0x4c) @5
	OffsetToSome6  uint32 // offset to (element size 0x2) @6
	OffsetToSome7  uint32 // offset to triangle indexes (element size 0xa) @7
	OffsetToSome8  uint32 // offset to quads indexes (element size 0xc) @8
	OffsetToSome9  uint32 // offset to floats array (element size 0xc)  @9
	OffsetToSome88 uint32
	OffsetToSome8c uint32

	Some1               []RibSome1
	Some3CxtNames       []string
	Some6               []RibSome6
	Some7TrianglesIndex []RibSome7Tria
	Some8QuadsIndex     []RibSome8Quad
	Some9Points         []mgl32.Vec3
}

func NewRibSheet(bs *utils.BufStack, wrtw io.Writer) (*ShapeRibSheet, error) {
	headerbs := bs.SubBuf("header", 0).SetSize(RIBSHEET_HEADER_SIZE)

	rib := &ShapeRibSheet{
		CountOfSomethingUnk38: headerbs.LU16(0x38),
		CountOfSomethingUnk3a: headerbs.LU16(0x3a),
		CountOfSome1:          headerbs.LU16(0x3c),
		CountOfSomethingUnk3e: headerbs.LU16(0x3e),

		CountOfSomethingUnk40: headerbs.LU16(0x40),
		CountOfSomethingUnk42: headerbs.LU16(0x42),
		CountOfSomethingUnk44: headerbs.LU16(0x44),
		CountOfSomethingUnk46: headerbs.LU16(0x46),
		CountOfSomethingUnk48: headerbs.LU16(0x48),
		CountOfSomethingUnk4a: headerbs.LU16(0x4a),
		CountOfSomethingUnk4c: headerbs.LU16(0x4c),
		CountOfSomethingUnk4e: headerbs.LU16(0x4e),

		CountOfSome4:          headerbs.LU16(0x50),
		CountOfSomethingUnk52: headerbs.LU16(0x52),
		CountOfSomethingUnk54: headerbs.LU16(0x54),
		CountOfSomethingUnk56: headerbs.LU16(0x56),
		CountOfSomethingUnk5a: headerbs.LU16(0x5a),
		CountOfSomethingUnk5c: headerbs.LU16(0x5c),

		OffsetToSome1:  headerbs.LU32(0x64),
		OffsetToSome2:  headerbs.LU32(0x68),
		OffsetToSome3:  headerbs.LU32(0x6c),
		OffsetToSome4:  headerbs.LU32(0x70),
		OffsetToSome5:  headerbs.LU32(0x74),
		OffsetToSome6:  headerbs.LU32(0x78),
		OffsetToSome7:  headerbs.LU32(0x7c),
		OffsetToSome8:  headerbs.LU32(0x80),
		OffsetToSome9:  headerbs.LU32(0x84),
		OffsetToSome88: headerbs.LU32(0x88),
		OffsetToSome8c: headerbs.LU32(0x8c),
	}

	utils.LogDump(rib)

	//sizeOfSome4DebugMaterial := (rib.OffsetToSome5 - rib.OffsetToSome4) / uint32(rib.CountOfSome4)

	{
		some1Bs := bs.SubBuf("some1Buffer",
			int(rib.OffsetToSome1)).SetSize(int(rib.OffsetToSome2 - rib.OffsetToSome1))

		rib.Some1 = make([]RibSome1, some1Bs.Size()/0x8)

		for i := range rib.Some1 {
			off := i * 8

			flag := some1Bs.LU32(off + 0)

			rib.Some1[i].IsContainsSome6Index = flag&1 != 0

			rib.Some1[i].IndexToSome6 = flag >> 1

			rib.Some1[i].Float0x4 = some1Bs.LF(off + 4)
			rib.Some1[i].Unk0x4 = some1Bs.LU16(off + 4)
			rib.Some1[i].Unk0x6 = some1Bs.LU16(off + 6)
		}
	}
	utils.LogDump(rib.Some1)

	{
		some3Bs := bs.SubBuf("some3CxtNamesBuffer",
			int(rib.OffsetToSome3)).SetSize(int(rib.OffsetToSome4 - rib.OffsetToSome3))

		rib.Some3CxtNames = make([]string, some3Bs.Size()/0x18)

		for i := range rib.Some3CxtNames {
			rib.Some3CxtNames[i] = some3Bs.ReadStringBuffer(0x18)
		}
	}
	utils.LogDump(rib.Some3CxtNames)

	{
		some6Bs := bs.SubBuf("some6Buffer",
			int(rib.OffsetToSome6)).SetSize(int(rib.OffsetToSome7 - rib.OffsetToSome6))

		rib.Some6 = make([]RibSome6, some6Bs.Size()/0x2)

		for i := range rib.Some6 {
			off := i * 2

			data := some6Bs.LU16(off)
			rib.Some6[i].IsQuad = data&1 != 0
			rib.Some6[i].IndexOfSome78 = data >> 1
		}
	}

	{
		triaIndexBs := bs.SubBuf("triaIndexBuffer",
			int(rib.OffsetToSome7)).SetSize(int(rib.OffsetToSome8 - rib.OffsetToSome7))

		rib.Some7TrianglesIndex = make([]RibSome7Tria, triaIndexBs.Size()/0xa)
		for i := range rib.Some7TrianglesIndex {
			off := i * 10
			rib.Some7TrianglesIndex[i].Flags = triaIndexBs.LU16(off + 0)
			rib.Some7TrianglesIndex[i].IndexOfSome4Debug = triaIndexBs.LU16(off + 2)
			rib.Some7TrianglesIndex[i].Indexes[0] = triaIndexBs.LU16(off + 4)
			rib.Some7TrianglesIndex[i].Indexes[1] = triaIndexBs.LU16(off + 6)
			rib.Some7TrianglesIndex[i].Indexes[2] = triaIndexBs.LU16(off + 8)
		}
	}

	{
		quadIndexBs := bs.SubBuf("quadIndexBuffer",
			int(rib.OffsetToSome8)).SetSize(int(rib.OffsetToSome9 - rib.OffsetToSome8))

		rib.Some8QuadsIndex = make([]RibSome8Quad, quadIndexBs.Size()/0xc)
		for i := range rib.Some8QuadsIndex {
			off := i * 12
			rib.Some8QuadsIndex[i].Flags = quadIndexBs.LU16(off + 0)
			rib.Some8QuadsIndex[i].IndexOfSome4Debug = quadIndexBs.LU16(off + 2)
			rib.Some8QuadsIndex[i].Indexes[0] = quadIndexBs.LU16(off + 4)
			rib.Some8QuadsIndex[i].Indexes[1] = quadIndexBs.LU16(off + 6)
			rib.Some8QuadsIndex[i].Indexes[2] = quadIndexBs.LU16(off + 8)
			rib.Some8QuadsIndex[i].Indexes[3] = quadIndexBs.LU16(off + 10)
		}
	}

	{
		floatsBs := bs.SubBuf("floatsBuffer",
			int(rib.OffsetToSome9)).SetSize(int(rib.OffsetToSome88 - rib.OffsetToSome9))

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

		for some4Idx := uint16(0); some4Idx < rib.CountOfSome4; some4Idx++ {
			fmt.Fprintf(wrtw, "o material%d\n", some4Idx)

			//fmt.Fprintf(wrtw, "o triangles\n")
			for _, triIndex := range rib.Some7TrianglesIndex {
				if triIndex.IndexOfSome4Debug == some4Idx {
					fmt.Fprintf(wrtw, "f %d %d %d\n",
						triIndex.Indexes[0]+1, triIndex.Indexes[1]+1, triIndex.Indexes[2]+1)
				}
			}
			//fmt.Fprintf(wrtw, "o quads\n")
			for _, quadIndex := range rib.Some8QuadsIndex {
				if quadIndex.IndexOfSome4Debug == some4Idx {
					fmt.Fprintf(wrtw, "f %d %d %d\n",
						quadIndex.Indexes[0]+1, quadIndex.Indexes[1]+1, quadIndex.Indexes[2]+1)
					fmt.Fprintf(wrtw, "f %d %d %d\n",
						quadIndex.Indexes[3]+1, quadIndex.Indexes[0]+1, quadIndex.Indexes[2]+1)
				}
			}
		}
	}

	return rib, nil
}
