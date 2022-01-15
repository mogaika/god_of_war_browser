package collision

import (
	"fmt"
	"io"

	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/go-gl/mathgl/mgl32"
)

const BALLHULL_HEADER_SIZE = 0x68

const (
	BALLHULL_SECTION_MATERIAL           = 0  // 0x3c []material
	BALLHULL_SECTION_BALLS_JOINTS       = 1  // 0x40 []byte
	BALLHULL_SECTION_BALLS_SCRIPTMARK   = 2  // 0x44 []byte
	BALLHULL_SECTION_BALLS_MAPMATERIAL  = 3  // 0x48 []byte
	BALLHULL_SECTION_BALLS_COORDS       = 4  // 0x4c []vec4
	BALLHULL_SECTION_MESHES_PLANESCOUNT = 5  // 0x50 []byte
	BALLHULL_SECTION_MESHES_JOINTS      = 6  // 0x54 []byte
	BALLHULL_SECTION_MESHES_SCRIPTMARK  = 7  // 0x58 []byte
	BALLHULL_SECTION_MESHES_BBOXES      = 8  // 0x5c []vec4
	BALLHULL_SECTION_MESHES_MAPMATERIAL = 9  // 0x60 []byte
	BALLHULL_SECTION_MESHES_PLANES      = 10 // 0x64 []vec4
)

type BallHullBall struct {
	Coord      mgl32.Vec4
	Joint      byte
	ScriptMark byte
	Material   byte
}

type BallHullMesh struct {
	BBox       mgl32.Vec4
	Planes     []mgl32.Vec4 // Hesse normal form of plane representation
	Materials  []int8       // int8 just so json marshal not performs base64
	Joint      byte
	ScriptMark byte
}

type ShapeBallHull struct {
	Type         uint32     // +0x04 0 - object, 1 - camera, 2 - sensor, 3 - soundem, (4 - ribsheet, in-engine const only)
	FileSize     uint32     // +0x10
	BallsCount   uint32     // +0x14
	MeshesCount  uint32     // +0x18
	BSphere      mgl32.Vec4 // +0x1c bbox or bsphere
	BSphereJoint uint32     // +0x2c
	// Unk0x30        uint32     // +0x30 either 0, either 0x1f. looks like not used in game
	MaterialsCount uint32     // +0x34
	MaterialSize   uint32     // +0x38
	Offsets        [11]uint32 // +0x3c

	Balls  []*BallHullBall
	Meshes []*BallHullMesh

	DbgMesh *ShapeDbgHdr // for export only
}

func (c *ShapeBallHull) GetSectionSize(section int) uint32 {
	if section == 9 {
		return c.FileSize - c.Offsets[section]
	} else {
		return c.Offsets[section+1] - c.Offsets[section]
	}
}

func readVec4(bs *utils.BufStack, vec *mgl32.Vec4) {
	for i := range vec {
		vec[i] = bs.ReadLF()
	}
}

func NewBallHull(bs *utils.BufStack, wrtw io.Writer) (*ShapeBallHull, error) {
	bsHeader := bs.SubBuf("ballhull_header", 0).SetSize(BALLHULL_HEADER_SIZE)

	bh := &ShapeBallHull{
		Type:           bsHeader.LU32(0x4),
		FileSize:       bsHeader.LU32(0x10),
		BallsCount:     bsHeader.LU32(0x14),
		MeshesCount:    bsHeader.LU32(0x18),
		BSphereJoint:   bsHeader.LU32(0x2c),
		MaterialsCount: bsHeader.LU32(0x34),
		MaterialSize:   bsHeader.LU32(0x38),
	}

	utils.ReadBytes(&bh.BSphere, bsHeader.Raw()[0x1c:0x2c])
	utils.ReadBytes(&bh.Offsets, bsHeader.Raw()[0x3c:0x68])

	bss := make([]*utils.BufStack, 11)
	for i := range bss {
		bss[i] = bs.SubBuf(fmt.Sprintf("data%d", i), int(bh.Offsets[i]))
		if i < len(bss)-1 {
			bss[i].SetSize(int(bh.Offsets[i+1] - bh.Offsets[i]))
		} else {
			bss[i].Expand()
		}
	}

	bh.Balls = make([]*BallHullBall, bh.BallsCount)
	for i := range bh.Balls {
		b := &BallHullBall{}

		readVec4(bss[BALLHULL_SECTION_BALLS_COORDS], &b.Coord)
		b.Material = bss[BALLHULL_SECTION_BALLS_MAPMATERIAL].ReadByte()
		b.Joint = bss[BALLHULL_SECTION_BALLS_JOINTS].ReadByte()
		b.ScriptMark = bss[BALLHULL_SECTION_BALLS_SCRIPTMARK].ReadByte()

		bh.Balls[i] = b
	}

	bh.Meshes = make([]*BallHullMesh, bh.MeshesCount)
	for i := range bh.Meshes {
		m := &BallHullMesh{}
		readVec4(bss[BALLHULL_SECTION_MESHES_BBOXES], &m.BBox)

		planesCount := bss[BALLHULL_SECTION_MESHES_PLANESCOUNT].ReadByte()
		m.Planes = make([]mgl32.Vec4, planesCount)
		m.Materials = make([]int8, planesCount)
		m.Joint = bss[BALLHULL_SECTION_MESHES_JOINTS].ReadByte()
		m.ScriptMark = bss[BALLHULL_SECTION_MESHES_SCRIPTMARK].ReadByte()

		for vi := range m.Planes {
			m.Materials[vi] = int8(bss[BALLHULL_SECTION_MESHES_MAPMATERIAL].ReadByte())
			readVec4(bss[BALLHULL_SECTION_MESHES_PLANES], &m.Planes[vi])
		}

		bh.Meshes[i] = m
	}

	return bh, nil
}
