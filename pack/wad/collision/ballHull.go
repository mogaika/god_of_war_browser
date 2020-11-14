package collision

import (
	"io"

	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/go-gl/mathgl/mgl32"
)

const BALLHULL_HEADER_SIZE = 0x68

const (
	BALLHULL_SECTION_INFONAME  = 0 // 0x3c
	BALLHULL_SECTION_POSVECTOR = 4 // 0x4c
	BALLHULL_SECTION_5CVECTOR  = 8 // 0x5c
)

type ShapeBallHull struct {
	Vector        mgl32.Vec4 // bbox or bsphere
	Offsets       [10]uint32
	FileSize      uint32
	Some4cVectors []mgl32.Vec4
	Some5cVectors []mgl32.Vec4
}

func (c *ShapeBallHull) GetSectionSize(section int) uint32 {
	if section == 9 {
		return c.FileSize - c.Offsets[section]
	} else {
		return c.Offsets[section+1] - c.Offsets[section]
	}
}

func NewBallHull(bs *utils.BufStack, wrtw io.Writer) (*ShapeBallHull, error) {
	bsHeader := bs.SubBuf("ballhull_header", 0).SetSize(BALLHULL_HEADER_SIZE)

	bh := &ShapeBallHull{
		FileSize: bsHeader.LU32(0x10),
	}

	for i := range bh.Vector {
		bh.Vector[i] = bsHeader.LF(0x1c + 4*i)
	}

	for i := range bh.Offsets {
		bh.Offsets[i] = bsHeader.LU32(0x3c + i*4)
	}

	bh.Some4cVectors = make([]mgl32.Vec4, bh.GetSectionSize(BALLHULL_SECTION_POSVECTOR)/0x10)
	bsSome4cVectors := bs.SubBuf("some4cVectors", int(bh.Offsets[BALLHULL_SECTION_POSVECTOR])).SetSize(0x10 * len(bh.Some4cVectors))

	for iVec := range bh.Some4cVectors {
		vec := &bh.Some4cVectors[iVec]
		for j := range vec {
			vec[j] = bsSome4cVectors.ReadLF()
		}
	}

	bh.Some5cVectors = make([]mgl32.Vec4, bh.GetSectionSize(BALLHULL_SECTION_5CVECTOR)/0x10)
	bsSome5cVectors := bs.SubBuf("some5cVectors", int(bh.Offsets[BALLHULL_SECTION_5CVECTOR])).SetSize(0x10 * len(bh.Some5cVectors))

	for iVec := range bh.Some5cVectors {
		vec := &bh.Some5cVectors[iVec]
		for j := range vec {
			vec[j] = bsSome5cVectors.ReadLF()
		}
	}

	return bh, nil
}
