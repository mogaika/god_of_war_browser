package collision

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"

	"github.com/go-gl/mathgl/mgl32"
)

const BALLHULL_HEADER_SIZE = 0x68

const (
	BALLHULL_SECTION_INFONAME  = 0 // 0x3c
	BALLHULL_SECTION_POSVECTOR = 4 // 0x4c
)

type ShapeBallHull struct {
	Vector        mgl32.Vec4
	Offsets       [10]uint32
	FileSize      uint32
	Some4cVectors []mgl32.Vec4
}

func (c *ShapeBallHull) GetSectionSize(section int) uint32 {
	if section == 9 {
		return c.FileSize - c.Offsets[section]
	} else {
		return c.Offsets[section+1] - c.Offsets[section]
	}
}

func NewBallHull(f io.ReaderAt, wrtw io.Writer) (*ShapeBallHull, error) {
	buf := make([]byte, BALLHULL_HEADER_SIZE)
	if _, err := f.ReadAt(buf, 0); err != nil {
		return nil, err
	}

	bh := &ShapeBallHull{
		FileSize: binary.LittleEndian.Uint32(buf[0x10:0x14]),
	}

	if err := binary.Read(bytes.NewReader(buf[0x1c:0x2c]), binary.LittleEndian, &bh.Vector); err != nil {
		panic(err)
	}

	for i := range bh.Offsets {
		bh.Offsets[i] = binary.LittleEndian.Uint32(buf[0x3c+i*4:])
	}

	bh.Some4cVectors = make([]mgl32.Vec4, bh.GetSectionSize(BALLHULL_SECTION_POSVECTOR)/0x10)
	for i := range bh.Some4cVectors {
		if err := binary.Read(
			io.NewSectionReader(f, int64(bh.Offsets[BALLHULL_SECTION_POSVECTOR]+uint32(i)*0x10), 0x10),
			binary.LittleEndian, &bh.Some4cVectors[i]); err != nil {
			panic(err)
		}
	}

	var medusaObj = mgl32.Mat4{0.0022545615, 0, 0.6209665, 0, 0, 0.6209706, 0, 0, -0.6209665, 0, 0.0022545615, 0, -467.78577, 427.11475, 198.80093, 1}
	log.Println(mgl32.TransformCoordinate(bh.Vector.Vec3(), medusaObj))

	var medusaInst = mgl32.Mat4{0.003626258810982108, 0, -0.9999935030937195, 0, 0, 1, 0, 0, 0.9999935030937195, 0, 0.003626258810982108, 0, -158.70399475097656, -343.90899658203125, -377.2359924316406, 1}
	log.Println(medusaInst.Mul4(medusaObj))
	log.Println(mgl32.TransformCoordinate(bh.Vector.Vec3(), medusaInst.Mul4(medusaObj)))

	return bh, nil
}
