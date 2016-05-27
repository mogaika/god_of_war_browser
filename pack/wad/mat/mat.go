package mat

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_txr "github.com/mogaika/god_of_war_browser/pack/wad/txr"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Layer struct {
	Texture string
	Flags   [4]uint32
	Floats  [5]float32
	Unkn    uint32
}

const (
	LAYER_FLAG_TEXTURE_PRESENTED = 0x80
)

type Material struct {
	Color  utils.ColorFloat
	Layers []Layer
}

const MAT_MAGIC = 0x00000008
const HEADER_SIZE = 0x38
const LAYER_SIZE = 0x40

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
		Layers: make([]Layer, binary.LittleEndian.Uint32(buf[0x34:0x38])),
	}

	mat.Color = utils.NewColorFloat([]float32{
		math.Float32frombits(binary.LittleEndian.Uint32(buf[8:12])),
		math.Float32frombits(binary.LittleEndian.Uint32(buf[12:16])),
		math.Float32frombits(binary.LittleEndian.Uint32(buf[16:20])),
	})

	for iTex := range mat.Layers {
		tbuf := make([]byte, LAYER_SIZE)

		if _, err := fmat.ReadAt(tbuf, int64(iTex*LAYER_SIZE+HEADER_SIZE)); err != nil {
			return nil, err
		}

		mat.Layers[iTex].Flags = [4]uint32{
			binary.LittleEndian.Uint32(tbuf[0:4]),
			binary.LittleEndian.Uint32(tbuf[4:8]),
			binary.LittleEndian.Uint32(tbuf[8:12]),
			binary.LittleEndian.Uint32(tbuf[12:16]),
		}
		mat.Layers[iTex].Texture = utils.BytesToString(tbuf[16:40])

		mat.Layers[iTex].Floats = [5]float32{
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[40:44])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[44:48])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[48:52])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[52:56])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[56:60])),
		}

		mat.Layers[iTex].Unkn = binary.LittleEndian.Uint32(tbuf[60:64])
	}

	return mat, nil
}

type Ajax struct {
	Mat      *Material
	Textures []string
}

func (mat *Material) AjaxMarshal(wad *wad.Wad, node *wad.WadNode) ([]byte, error) {
	res := Ajax{
		Mat:      mat,
		Textures: make([]string, len(mat.Layers)),
	}

	for i := range mat.Layers {
		tn := wad.FindNode(mat.Layers[i].Texture, node.Parent)
		if tn != nil {
			txr, err := wad.Get(tn.Id)
			if err != nil {
				return nil, fmt.Errorf("Error getting texture '%s' for material '%s': %v", tn.Name, node.Name, err)
			}

			dat, err := txr.(*file_txr.Texture).AjaxMarshal(tn.Wad, tn)
			if err != nil {
				return nil, fmt.Errorf("Error marshaling texture '%s' for material '%s': %v", tn.Name, node.Name, err)
			}

			res.Textures[i] = string(dat)
		}
	}
	return json.Marshal(res)
}

func init() {
	wad.SetHandler(MAT_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (interface{}, error) {
		return NewFromData(r)
	})
}
