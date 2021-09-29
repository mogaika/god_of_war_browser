package collision

import (
	"fmt"
	"io"
	"log"

	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/go-gl/mathgl/mgl32"
)

const RIBSHEET_HEADER_SIZE = 0x90

/*
	@10 => @2
	@2 =>
*/

type RibKDTreeNode struct {
	IsPolygon bool // If false, then plane

	// If IsPlane (IsPolygon == false)
	PlaneCoordinate    float32
	PlaneAxis          uint8 // 3 options, so probably x/y/z
	PlaneSubNodeHigher uint16

	// If IsContainsSome6Index == true
	PolygonIndex  uint32
	PolygonFlag   uint8
	PolygonsCount uint16
	PolygonUnk0x6 uint16
}

type RibContextZone struct {
	ints  []uint16
	Zones map[int][8][]uint16
}

func (rs2 *RibContextZone) Load(idx int) [8][]uint16 {
	if r, ok := rs2.Zones[idx]; ok {
		return r
	}

	var store [8][]uint16
	pos := idx

	if rs2.ints[pos] != 0 {
		store[0] = append(store[0], rs2.ints[pos])
		panic("heh")
	} else {
		pos++
		c1 := rs2.ints[pos]
		pos++
		for i := uint16(0); i < c1; i++ {
			key := rs2.ints[pos]
			pos++
			c2 := rs2.ints[pos]
			pos++
			for j := uint16(0); j < c2; j++ {
				store[key] = append(store[key], rs2.ints[pos])
				pos++
			}
		}
	}
	rs2.Zones[idx] = store
	return store
}

type RibMaterial struct {
	Name   string
	Values map[string]interface{}
}

type RibMaterialField struct {
	Name         string
	Type         uint32 // 0 = int, 1 = float, 2 = bit
	Unk0x24      uint32
	MaxValue     uint32
	DefaultValue uint32
	Unk0x30      uint32
	Unk0x34      uint32
	Unk0x38      uint16
	Unk0x3a      uint16
	OffsetBytes  uint32 // + 0x3C
	IsBitField1  uint32
	OffsetBits   uint32
	IsBitField2  uint32
}

type RibPolygon struct {
	IsQuad              bool
	QuadOrTriangleIndex uint16
}

type RibPolygonBase struct {
	Flags         uint16
	MaterialIndex uint16
}

type RibTriangle struct {
	RibPolygonBase
	Indexes [3]uint16
}

type RibQuad struct {
	RibPolygonBase
	Indexes [4]uint16
}

type RibZone struct {
	FloatArray  [6]float32
	ContextZone [8][]uint16
}

type ShapeRibSheet struct {
	FileSize  uint32        // + 0xC
	Unk0x10   uint32        // + 0x10 always = 0x1f
	Unk0x14   uint32        // + 0x14 always = 0x02140201
	LevelBBox [2]mgl32.Vec4 // + 0x18
	Unk0x38   uint16
	Unk0x3a   uint16
	// count of @1
	Unk0x3e uint16
	// count of @6
	Unk0x42 uint16 // usually = 0, only = 1 for DEST01
	Unk0x44 uint16
	// count of @7 @8 @9
	Unk0x4c uint16 // always = 0x40
	Unk0x4e uint16 // always = 0x40
	// count of @4 @5
	Unk0x54 uint16
	// count of @2
	Unk0x58 uint16
	// count of @3 @10

	Some1               []RibKDTreeNode
	Some3CxtNames       []string
	Some4Materials      []RibMaterial
	Some5FlagMaps       []RibMaterialField
	Some6               []RibPolygon
	Some7TrianglesIndex []RibTriangle
	Some8QuadsIndex     []RibQuad
	Some9Points         []mgl32.Vec3
	Some10              []RibZone
}

func NewRibSheet(bs *utils.BufStack, wrtw io.Writer) (*ShapeRibSheet, error) {
	headerbs := bs.SubBuf("header", 0).SetSize(RIBSHEET_HEADER_SIZE)

	countOfSome1 := headerbs.LU16(0x3c)
	countOfSome6 := headerbs.LU16(0x40)
	countOfSome7 := headerbs.LU16(0x46)
	countOfSome8 := headerbs.LU16(0x48)
	countOfSome9 := headerbs.LU16(0x4a)
	countOfSome4 := headerbs.LU16(0x50)
	countOfSome5 := headerbs.LU16(0x52)
	countOfSome2 := headerbs.LU16(0x56)
	countOfSome3 := headerbs.LU16(0x5a)
	countOfSome10 := headerbs.LU16(0x5c)

	_ = countOfSome2

	rib := &ShapeRibSheet{
		FileSize: headerbs.LU32(0xc),
		Unk0x10:  headerbs.LU32(0x10),
		Unk0x14:  headerbs.LU32(0x14),

		Unk0x38: headerbs.LU16(0x38),
		Unk0x3a: headerbs.LU16(0x3a),
		Unk0x3e: headerbs.LU16(0x3e),

		Unk0x42: headerbs.LU16(0x42),
		Unk0x44: headerbs.LU16(0x44),

		Unk0x4c: headerbs.LU16(0x4c),
		Unk0x4e: headerbs.LU16(0x4e),

		Unk0x54: headerbs.LU16(0x54),
		Unk0x58: headerbs.LU16(0x58),
	}

	var some2 RibContextZone

	offsetToSome1 := headerbs.LU32(0x64)
	offsetToSome2 := headerbs.LU32(0x68)
	offsetToSome3 := headerbs.LU32(0x6c)
	offsetToSome4 := headerbs.LU32(0x70)
	offsetToSome5 := headerbs.LU32(0x74)
	offsetToSome6 := headerbs.LU32(0x78)
	offsetToSome7 := headerbs.LU32(0x7c)
	offsetToSome8 := headerbs.LU32(0x80)
	offsetToSome9 := headerbs.LU32(0x84)
	offsetToSome10 := headerbs.LU32(0x8c)

	if rib.Unk0x10 != 0x1f {
		panic("Unk0x10")
	}
	if rib.Unk0x14 != 0x02140201 {
		panic("Unk0x14")
	}
	if rib.Unk0x4c != 0x40 {
		panic("Unk0x4c")
	}
	if rib.Unk0x4e != 0x40 {
		panic("Unk0x4e")
	}
	if rib.Unk0x58 != 0 {
		panic("Unk0x58")
	}

	// utils.LogDump(rib)

	for i := range rib.LevelBBox {
		for j := range rib.LevelBBox[i] {
			rib.LevelBBox[i][j] = headerbs.LF(0x18 + i*16 + j*4)
		}
	}

	{
		some1Bs := bs.SubBuf("some1Buffer",
			int(offsetToSome1)).SetSize(int(offsetToSome2 - offsetToSome1))

		rib.Some1 = make([]RibKDTreeNode, countOfSome1)

		for i := range rib.Some1 {
			off := i * 8

			flag := some1Bs.LU32(off + 0)

			rib.Some1[i].IsPolygon = flag&1 != 0

			rib.Some1[i].PlaneAxis = uint8(flag>>1) & 3
			rib.Some1[i].PlaneSubNodeHigher = some1Bs.LU16(off + 2)
			rib.Some1[i].PlaneCoordinate = some1Bs.LF(off + 4)

			rib.Some1[i].PolygonIndex = (flag >> 1) & 0xFF_FFFF
			rib.Some1[i].PolygonFlag = uint8((flag >> 1) >> 24)
			rib.Some1[i].PolygonsCount = some1Bs.LU16(off + 4)
			rib.Some1[i].PolygonUnk0x6 = some1Bs.LU16(off + 6)
		}
	}

	{
		some2Bs := bs.SubBuf("some2Buffer",
			int(offsetToSome2)).SetSize(int(offsetToSome3 - offsetToSome2))

		ints := make([]uint16, some2Bs.Size()/2)

		for i := range ints {
			ints[i] = some2Bs.ReadLU16()
		}
		some2.ints = ints
		some2.Zones = make(map[int][8][]uint16)
	}

	{
		some3Bs := bs.SubBuf("some3CxtNamesBuffer",
			int(offsetToSome3)).SetSize(int(offsetToSome4 - offsetToSome3))

		rib.Some3CxtNames = make([]string, countOfSome3)

		for i := range rib.Some3CxtNames {
			rib.Some3CxtNames[i] = some3Bs.ReadStringBuffer(0x18)
		}
	}

	{
		some5Bs := bs.SubBuf("some5Buffer",
			int(offsetToSome5)).SetSize(int(offsetToSome6 - offsetToSome5))

		rib.Some5FlagMaps = make([]RibMaterialField, countOfSome5)

		for i := range rib.Some5FlagMaps {
			s5bs := some5Bs.SubBuf("some5", i*0x4c).SetSize(0x4c)
			rib.Some5FlagMaps[i].Name = s5bs.ReadStringBuffer(0x20)
			rib.Some5FlagMaps[i].Type = s5bs.ReadLU32()
			rib.Some5FlagMaps[i].Unk0x24 = s5bs.ReadLU32()
			rib.Some5FlagMaps[i].MaxValue = s5bs.ReadLU32()
			rib.Some5FlagMaps[i].DefaultValue = s5bs.ReadLU32()
			rib.Some5FlagMaps[i].Unk0x30 = s5bs.ReadLU32()
			rib.Some5FlagMaps[i].Unk0x34 = s5bs.ReadLU32()
			rib.Some5FlagMaps[i].Unk0x38 = s5bs.ReadLU16()
			rib.Some5FlagMaps[i].Unk0x3a = s5bs.ReadLU16()
			rib.Some5FlagMaps[i].OffsetBytes = s5bs.ReadLU32()
			rib.Some5FlagMaps[i].IsBitField1 = s5bs.ReadLU32()
			rib.Some5FlagMaps[i].OffsetBits = s5bs.ReadLU32()
			rib.Some5FlagMaps[i].IsBitField2 = s5bs.ReadLU32()
			s5bs.SetName(rib.Some5FlagMaps[i].Name)
			s5bs.VerifySize(0x4c)
		}
	}

	{
		some4Bs := bs.SubBuf("some4Buffer",
			int(offsetToSome4)).SetSize(int(offsetToSome5 - offsetToSome4))
		elemSize := int(offsetToSome5-offsetToSome4) / int(countOfSome4)

		rib.Some4Materials = make([]RibMaterial, countOfSome4)
		for i := range rib.Some4Materials {
			s4bs := some4Bs.SubBuf("some4", i*elemSize).SetSize(elemSize)
			s4bsValue := s4bs.SubBuf("fields", 0x18).SetSize(elemSize - 0x18)

			rib.Some4Materials[i].Name = s4bs.ReadStringBuffer(0x18)
			rib.Some4Materials[i].Values = make(map[string]interface{})
			for _, field := range rib.Some5FlagMaps {
				var value interface{}

				switch field.Type {
				case 0:
					value = s4bsValue.LU32(int(field.OffsetBytes))
				case 1:
					value = float32(s4bsValue.LU32(int(field.OffsetBytes)))
				case 2:
					value = ((s4bsValue.LU32(int(field.OffsetBytes)) >> (field.OffsetBits - 0x8*field.OffsetBytes)) & 1) != 0
				}

				rib.Some4Materials[i].Values[field.Name] = value
			}
		}
	}

	{
		some6Bs := bs.SubBuf("some6Buffer",
			int(offsetToSome6)).SetSize(int(offsetToSome7 - offsetToSome6))

		rib.Some6 = make([]RibPolygon, countOfSome6)

		for i := range rib.Some6 {
			off := i * 2

			data := some6Bs.LU16(off)
			rib.Some6[i].IsQuad = data&1 != 0
			rib.Some6[i].QuadOrTriangleIndex = data >> 1
		}
	}

	{
		triaIndexBs := bs.SubBuf("triaIndexBuffer",
			int(offsetToSome7)).SetSize(int(offsetToSome8 - offsetToSome7))

		rib.Some7TrianglesIndex = make([]RibTriangle, countOfSome7)
		for i := range rib.Some7TrianglesIndex {
			off := i * 10
			rib.Some7TrianglesIndex[i].Flags = triaIndexBs.LU16(off + 0)
			rib.Some7TrianglesIndex[i].MaterialIndex = triaIndexBs.LU16(off + 2)
			rib.Some7TrianglesIndex[i].Indexes[0] = triaIndexBs.LU16(off + 4)
			rib.Some7TrianglesIndex[i].Indexes[1] = triaIndexBs.LU16(off + 6)
			rib.Some7TrianglesIndex[i].Indexes[2] = triaIndexBs.LU16(off + 8)
		}
	}

	{
		quadIndexBs := bs.SubBuf("quadIndexBuffer",
			int(offsetToSome8)).SetSize(int(offsetToSome9 - offsetToSome8))

		rib.Some8QuadsIndex = make([]RibQuad, countOfSome8)
		for i := range rib.Some8QuadsIndex {
			off := i * 12
			rib.Some8QuadsIndex[i].Flags = quadIndexBs.LU16(off + 0)
			rib.Some8QuadsIndex[i].MaterialIndex = quadIndexBs.LU16(off + 2)
			rib.Some8QuadsIndex[i].Indexes[0] = quadIndexBs.LU16(off + 4)
			rib.Some8QuadsIndex[i].Indexes[1] = quadIndexBs.LU16(off + 6)
			rib.Some8QuadsIndex[i].Indexes[2] = quadIndexBs.LU16(off + 8)
			rib.Some8QuadsIndex[i].Indexes[3] = quadIndexBs.LU16(off + 10)
		}
	}

	{
		floatsBs := bs.SubBuf("floatsBuffer",
			int(offsetToSome9)).SetSize(int(offsetToSome10 - offsetToSome9))

		rib.Some9Points = make([]mgl32.Vec3, countOfSome9)
		for i := range rib.Some9Points {
			off := i * 12
			rib.Some9Points[i][0] = floatsBs.LF(off + 0)
			rib.Some9Points[i][1] = floatsBs.LF(off + 4)
			rib.Some9Points[i][2] = floatsBs.LF(off + 8)
		}
	}

	{
		ctxRef := bs.SubBuf("ctxRef",
			int(offsetToSome10)).SetSize(int(countOfSome10) * 0x1c)

		rib.Some10 = make([]RibZone, countOfSome10)
		for i := range rib.Some10 {
			off := i * 0x1c
			for j := range rib.Some10[i].FloatArray {
				rib.Some10[i].FloatArray[j] = ctxRef.LF(off + j*4)
			}
			ctxZoneRef := ctxRef.LU32(off + 0x18)
			rib.Some10[i].ContextZone = some2.Load(int(ctxZoneRef))
			/*
				log.Printf("[%.2d](%.3d) %v :: %v",
					i,
					rib.Some10[i].ContextZone,
					rib.Some2.Load(int(rib.Some10[i].ContextZone)),
					rib.Some10[i].FloatArray)
			*/
		}
	}

	// rib.PrintKDTree()
	/*
		log.Printf("rib1count %d", len(rib.Some1))
		log.Printf("rib2count %d", len(rib.Some2.Zones))
		log.Printf("rib3count %d", len(rib.Some3CxtNames))
		log.Printf("rib4count %d", len(rib.Some4Materials))
		log.Printf("rib5count %d", len(rib.Some5FlagMaps))
		log.Printf("rib6count %d", len(rib.Some6))
		log.Printf("rib7count %d", len(rib.Some7TrianglesIndex))
		log.Printf("rib8count %d", len(rib.Some8QuadsIndex))
		log.Printf("rib9count %d", len(rib.Some9Points))
		log.Printf("rib10count %d", len(rib.Some10))
	*/
	// utils.LogDump(rib.Some1)
	// utils.LogDump(rib.Some3CxtNames)
	// utils.LogDump(rib.Some5FlagMaps)
	// utils.LogDump(rib.Some10)
	// utils.LogDump(rib.Some5FlagMaps)
	// utils.LogDump(rib.Some4Materials)

	/*
		for _, mat := range rib.Some4Materials {
			log.Printf(" Material %q:", mat.Name)
			for _, val := range mat.Values {
				log.Printf("  - %v: %v", val.Name, val.Value)
			}
		}
	*/

	/*
		maxin3 := uint16(0)
		for _, zone := range rib.Some2.Zones {
			for _, sz := range zone {
				for _, v := range sz {
					if v > maxin3 {
						maxin3 = v
					}
				}
			}
		}
		log.Printf("len2 %d len3 %d  maxin3 %d",
			len(rib.Some2.Zones), len(rib.Some3CxtNames), maxin3)
		log.Printf("count s10 %d, contexts: %v", rib.CountOfSome10, rib.Some3CxtNames)
	*/

	rib.WriteDebugObject(wrtw)

	return rib, nil
}

func (rib *ShapeRibSheet) BuildMeshForKDTree() (vertices [][3]float32, triangles [][3]int) {
	type Point = [3]float32
	type Bbox = [2]Point // lower / upper
	type Rect = [4]Point

	rects := make([]Rect, 0)

	var fillRects func(id uint16, bbox Bbox)
	fillRects = func(id uint16, bbox Bbox) {
		n := &rib.Some1[id]
		if n.IsPolygon {
			return
		}

		switch n.PlaneAxis {
		case 0:
			rects = append(rects, Rect{
				Point{n.PlaneCoordinate, bbox[0][1], bbox[0][2]},
				Point{n.PlaneCoordinate, bbox[0][1], bbox[1][2]},
				Point{n.PlaneCoordinate, bbox[1][1], bbox[0][2]},
				Point{n.PlaneCoordinate, bbox[1][1], bbox[1][2]},
			})
		case 1:
			rects = append(rects, Rect{
				Point{bbox[0][0], n.PlaneCoordinate, bbox[0][2]},
				Point{bbox[0][0], n.PlaneCoordinate, bbox[1][2]},
				Point{bbox[1][0], n.PlaneCoordinate, bbox[0][2]},
				Point{bbox[1][0], n.PlaneCoordinate, bbox[1][2]},
			})
		case 2:
			rects = append(rects, Rect{
				Point{bbox[0][0], bbox[0][1], n.PlaneCoordinate},
				Point{bbox[0][0], bbox[1][1], n.PlaneCoordinate},
				Point{bbox[1][0], bbox[0][1], n.PlaneCoordinate},
				Point{bbox[1][0], bbox[1][1], n.PlaneCoordinate},
			})
		}

		bboxLo := bbox
		bboxLo[0][n.PlaneAxis] = n.PlaneCoordinate
		fillRects(id+1, bboxLo)

		bboxHi := bbox
		bboxHi[1][n.PlaneAxis] = n.PlaneCoordinate
		fillRects(n.PlaneSubNodeHigher, bboxHi)
	}

	fillRects(0, Bbox{
		Point{rib.LevelBBox[0][0], rib.LevelBBox[0][1], rib.LevelBBox[0][2]},
		Point{rib.LevelBBox[1][0], rib.LevelBBox[1][1], rib.LevelBBox[1][2]},
	})

	vertices = make([]Point, 0)
	triangles = make([][3]int, 0)
	vertHash := make(map[Point]int)

	unhash := func(p Point) int {
		if i, ok := vertHash[p]; ok {
			return i
		} else {
			i := len(vertices)
			vertices = append(vertices, p)
			vertHash[p] = i
			return i
		}
	}

	for _, rect := range rects {
		triangles = append(triangles,
			[3]int{unhash(rect[0]), unhash(rect[1]), unhash(rect[2])},
			[3]int{unhash(rect[2]), unhash(rect[1]), unhash(rect[3])},
		)
	}

	return
}

func (rib *ShapeRibSheet) PrintKDTree() {
	deepSum := 0
	deepCount := 0
	polyCountSum := 0
	polyCountCount := 0
	polyCountMax := uint16(0)
	var printNode func(id uint16, tab string, deep int)
	printNode = func(id uint16, tab string, deep int) {
		deep++
		n := &rib.Some1[id]
		if n.IsPolygon {
			log.Printf("%s - [%.4d] polyg %.4d (%d,%d)", tab, id, n.PolygonIndex, n.PolygonsCount, n.PolygonUnk0x6)
			if n.PolygonUnk0x6 != 0 {
				//panic("")
			}
			deepSum += deep
			deepCount++
			if n.PolygonsCount > polyCountMax {
				polyCountMax = n.PolygonsCount
			}
			polyCountSum += int(n.PolygonsCount)
			polyCountCount++
		} else {
			log.Printf("%s = [%.4d] plane %d %f", tab, id, n.PlaneAxis, n.PlaneCoordinate)
			tab += "  "
			printNode(id+1, tab, deep)
			printNode(n.PlaneSubNodeHigher, tab, deep)
		}
	}

	printNode(0, " ", 0)
	log.Printf("Avg deep of kdtree %d/%d = %f",
		deepSum, deepCount, float32(deepSum)/float32(deepCount))
	log.Printf("Max poly per node %d. Avg poly per node %d/%d = %f",
		polyCountMax, polyCountSum, polyCountCount, float32(polyCountSum)/float32(polyCountCount))
}

func (rib *ShapeRibSheet) WriteDebugObject(wrtw io.Writer) {

	{
		//fmt.Fprintf(wrtw, "o ribsheet\n")
		for _, point := range rib.Some9Points {
			fmt.Fprintf(wrtw, "v %f %f %f\n", point[0], point[1], point[2])
		}
		pointsBaseIndex := len(rib.Some9Points)
		/*
			{
				var renderNode func(id uint16)
				renderNode = func(id uint16) {
					n := &rib.Some1[id]
					if n.IsPolygon {
						fmt.Fprintf(wrtw, "o node_%d\n", id)

						for i := uint32(0); i < uint32(n.PolygonsCount); i++ {
							polygonId := n.PolygonIndex + i
							idx := uint32(rib.Some6[polygonId].IndexOfSome7Or8)

							if rib.Some6[polygonId].IsQuad {
								quadIndex := rib.Some8QuadsIndex[idx]
								fmt.Fprintf(wrtw, "f %d %d %d\n",
									quadIndex.Indexes[0]+1, quadIndex.Indexes[1]+1, quadIndex.Indexes[2]+1)
								fmt.Fprintf(wrtw, "f %d %d %d\n",
									quadIndex.Indexes[3]+1, quadIndex.Indexes[0]+1, quadIndex.Indexes[2]+1)
							} else {
								triIndex := rib.Some7TrianglesIndex[idx]
								fmt.Fprintf(wrtw, "f %d %d %d\n",
									triIndex.Indexes[0]+1, triIndex.Indexes[1]+1, triIndex.Indexes[2]+1)
							}
						}
					} else {
						renderNode(id + 1)
						renderNode(n.PlaneSubNodeHigher)
					}
				}
				renderNode(0)
			}
		*/

		for some4Idx := uint16(0); some4Idx < uint16(len(rib.Some4Materials)); some4Idx++ {
			fmt.Fprintf(wrtw, "o material%d\n", some4Idx)

			//fmt.Fprintf(wrtw, "o triangles\n")
			for _, triIndex := range rib.Some7TrianglesIndex {
				if triIndex.MaterialIndex == some4Idx {
					fmt.Fprintf(wrtw, "f %d %d %d\n",
						triIndex.Indexes[0]+1, triIndex.Indexes[1]+1, triIndex.Indexes[2]+1)
				}
			}
			//fmt.Fprintf(wrtw, "o quads\n")
			for _, quadIndex := range rib.Some8QuadsIndex {
				if quadIndex.MaterialIndex == some4Idx {
					fmt.Fprintf(wrtw, "f %d %d %d\n",
						quadIndex.Indexes[0]+1, quadIndex.Indexes[1]+1, quadIndex.Indexes[2]+1)
					fmt.Fprintf(wrtw, "f %d %d %d\n",
						quadIndex.Indexes[3]+1, quadIndex.Indexes[0]+1, quadIndex.Indexes[2]+1)
				}
			}
		}

		for is10, s10 := range rib.Some10 {
			idxStart := pointsBaseIndex

			fmt.Fprintf(wrtw, "o s10_%d\n", is10)

			fa := s10.FloatArray
			points := [8]mgl32.Vec3{
				{fa[3], fa[1], fa[2]},
				{fa[3], fa[1], fa[5]},
				{fa[0], fa[1], fa[5]},
				{fa[0], fa[1], fa[2]},

				{fa[3], fa[4], fa[2]},
				{fa[3], fa[4], fa[5]},
				{fa[0], fa[4], fa[5]},
				{fa[0], fa[4], fa[2]},
			}

			for _, point := range points {
				fmt.Fprintf(wrtw, "v %f %f %f\n", point[0], point[1], point[2])
				pointsBaseIndex++
			}

			trias := [12][3]int{
				{2, 3, 4}, {5, 8, 7},
				{5, 6, 2}, {2, 6, 7},
				{7, 8, 4}, {5, 1, 4},
				{1, 2, 4}, {6, 5, 7},
				{1, 5, 2}, {3, 2, 7},
				{3, 7, 4}, {8, 5, 4},
			}

			for _, tria := range trias {
				fmt.Fprintf(wrtw, "f %d %d %d\n",
					tria[0]+idxStart, tria[1]+idxStart, tria[2]+idxStart)
			}

		}

		/*
			vertices, triangles := rib.BuildMeshForKDTree()

			off := len(rib.Some9Points)

			fmt.Fprintf(wrtw, "o ribsheet\n")
			for _, point := range vertices {
				fmt.Fprintf(wrtw, "v %f %f %f\n", point[0], point[1], point[2])
			}

			for _, triangle := range triangles {
				fmt.Fprintf(wrtw, "f %d %d %d\n", triangle[0]+1+off, triangle[1]+1+off, triangle[2]+1+off)
			}
		*/
	}
}
