package collision

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/mogaika/god_of_war_browser/utils"
)

// Convex Polyhedron
type DbgMesh struct {
	Vertices []mgl32.Vec4
	Indices  []uint16 // 2 indexes per edge
}

type ShapeDbgHdr struct {
	Meshes []DbgMesh
}

func NewDbgHdr(bs *utils.BufStack) (*ShapeDbgHdr, error) {
	headerBs := bs.SubBuf("header", 0).SetSize(0x20)

	totalSize := int(headerBs.LU32(0xc))
	vCountOffset := int(headerBs.LU32(0x10))
	iCountOffset := int(headerBs.LU32(0x14))
	vDataOffset := int(headerBs.LU32(0x18))
	iDataOffset := int(headerBs.LU32(0x1c))

	vCountBs := bs.SubBuf("vcount", vCountOffset).SetSize(iCountOffset - vCountOffset)
	iCountBs := bs.SubBuf("icount", iCountOffset).SetSize(vDataOffset - iCountOffset)
	vDataBs := bs.SubBuf("vdata", vDataOffset).SetSize(iDataOffset - vDataOffset)
	iDataBs := bs.SubBuf("idata", iDataOffset).SetSize(totalSize - iDataOffset)

	dm := &ShapeDbgHdr{
		Meshes: make([]DbgMesh, 0),
	}

	vReader := bytes.NewReader(vDataBs.Raw())
	iReader := bytes.NewReader(iDataBs.Raw())
	for i := 0; i*2 < vCountBs.Size(); i++ {
		vCount := vCountBs.LU16(i * 2)
		iCount := iCountBs.LU16(i * 2)
		if vCount == 0 || iCount == 0 {
			break
		}

		vertices := make([]mgl32.Vec4, vCount)
		indices := make([]uint16, iCount)

		if err := binary.Read(vReader, binary.LittleEndian, vertices); err != nil {
			return nil, errors.Wrapf(err, "Failed ot read vertices")
		}
		if err := binary.Read(iReader, binary.LittleEndian, indices); err != nil {
			return nil, errors.Wrapf(err, "Failed ot read indices")
		}

		dm.Meshes = append(dm.Meshes, DbgMesh{
			Vertices: vertices, Indices: indices,
		})
	}
	return dm, nil
}
