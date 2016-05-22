package mat

import (
	"encoding/binary"
	"errors"
	"io"
	"log"

	"github.com/mogaika/god_of_war_tools/files/wad"
	"github.com/mogaika/god_of_war_tools/utils"
)

type Layer struct {
	Texture string
	Flags   uint32
}

const (
	LAYER_FLAG_TEXTURE_PRESENTED = 0x80
)

type Material struct {
	Layers []Layer
}

const MAT_MAGIC = 0x00000008
const HEADER_SIZE = 0x38
const LAYER_SIZE = 0x40

func init() {
	wad.PregisterExporter(MAT_MAGIC, &Material{})
}

func NewFromData(fmat io.ReaderAt) (*Material, error) {
	buf := make([]byte, HEADER_SIZE)
	if _, err := fmat.ReadAt(buf, 0); err != nil {
		return nil, err
	}

	magic := binary.LittleEndian.Uint32(buf[:4])
	if magic != MAT_MAGIC {
		return nil, errors.New("Wrong magic.")
	}

	mat := &Material{
		Layers: make([]Layer, binary.LittleEndian.Uint32(buf[0x34:0x38]))}

	/*
		log.Printf("  %f %f %f %f",
			math.Float32frombits(binary.LittleEndian.Uint32(buf[0x8:0xc])),
			math.Float32frombits(binary.LittleEndian.Uint32(buf[0xc:0x10])),
			math.Float32frombits(binary.LittleEndian.Uint32(buf[0x10:0x14])),
			math.Float32frombits(binary.LittleEndian.Uint32(buf[0x14:0x18])))
	*/

	for iTex := range mat.Layers {
		tbuf := make([]byte, LAYER_SIZE)

		if _, err := fmat.ReadAt(tbuf, int64(iTex*LAYER_SIZE+HEADER_SIZE)); err != nil {
			return nil, err
		}

		flags := binary.LittleEndian.Uint32(tbuf[0x0:0x4])
		texture := utils.BytesToString(tbuf[0x10:0x28])

		//log.Printf("    layer %.8x %.8x %.8x %.8x '%s'",
		//	binary.LittleEndian.Uint32(tbuf[0x0:0x4]), binary.LittleEndian.Uint32(tbuf[0x4:0x8]),
		//	binary.LittleEndian.Uint32(tbuf[0x8:0xc]), binary.LittleEndian.Uint32(tbuf[0xc:0x10]), texture)

		mat.Layers[iTex] = Layer{Flags: flags, Texture: texture}
	}

	return mat, nil
}

func (Material) Cache(nd *wad.WadNode, r io.ReaderAt) error {
	mat, err := NewFromData(r)
	if err != nil {
		return err
	}
	nd.Cache = mat
	return nil
}

func (*Material) ExtractFromNode(nd *wad.WadNode, outfname string) error {
	log.Printf("Mat '%s' extraction", nd.Path)
	reader, err := nd.DataReader()
	if err != nil {
		return err
	}

	return Material{}.Cache(nd, reader)
}
