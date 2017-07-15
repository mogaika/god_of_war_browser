package mat

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_txr "github.com/mogaika/god_of_war_browser/pack/wad/txr"
	"github.com/mogaika/god_of_war_browser/utils"
)

/*
Info that may help:
https://nccastaff.bournemouth.ac.uk/jmacey/RobTheBloke/www/research/maya/mfnmaterial.htm
https://nccastaff.bournemouth.ac.uk/jmacey/RobTheBloke/www/research/maya/mfnenvmap.htm
https://nccastaff.bournemouth.ac.uk/jmacey/RobTheBloke/www/research/maya/mfnmaterial.htm

*/

type Flags struct {
	FilterLinear            bool // when false, then near filter used. may affect only wheb texture expanded (LOD < 0)
	DisableDepthWrite       bool
	RenderingAdditive       bool
	RenderingUsual          bool // handle transparency
	RenderingSubstract      bool
	RenderingStrangeBlended bool // I'm do not know
	HaveTexture             bool

	AnimationEnabled  bool
	AnimationEnabled2 bool // ATHN04A.WAD/378
}

type Layer struct {
	Texture     string
	Flags       [4]uint32
	BlendColor  [4]float32
	FloatUnk    float32
	GameFlags   uint32
	ParsedFlags Flags
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

func (l *Layer) ParseFlags() error {
	l.ParsedFlags.HaveTexture = (l.Flags[0]>>7)&1 != 0

	l.ParsedFlags.FilterLinear = (l.Flags[0]>>16)&1 != 0

	l.ParsedFlags.DisableDepthWrite = (l.Flags[0]>>19)&1 != 0

	l.ParsedFlags.RenderingStrangeBlended = (l.Flags[0]>>24)&1 != 0
	l.ParsedFlags.RenderingSubstract = (l.Flags[0]>>25)&1 != 0
	l.ParsedFlags.RenderingUsual = (l.Flags[0]>>26)&1 != 0
	l.ParsedFlags.RenderingAdditive = (l.Flags[0]>>27)&1 != 0

	l.ParsedFlags.AnimationEnabled = l.GameFlags&1 != 0
	l.ParsedFlags.AnimationEnabled2 = l.GameFlags&2 != 0

	cnt := 0
	for i := uint(0); i < 4; i++ {
		if (l.Flags[0]>>(24+i))&1 != 0 {
			cnt++
		}
	}
	if cnt > 1 {
		return fmt.Errorf("Too much rendering types in one layer: %+#v", l)
	}

	return nil
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

		mat.Layers[iTex].BlendColor = [4]float32{
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[40:44])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[44:48])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[48:52])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[52:56])),
		}

		mat.Layers[iTex].FloatUnk = math.Float32frombits(binary.LittleEndian.Uint32(tbuf[56:60]))

		mat.Layers[iTex].GameFlags = binary.LittleEndian.Uint32(tbuf[60:64])

		if err := mat.Layers[iTex].ParseFlags(); err != nil {
			return nil, fmt.Errorf("Error paring layer %d: %v", iTex, err)
		}
	}

	return mat, nil
}

type Ajax struct {
	Mat             *Material
	Textures        []interface{}
	TexturesBlended []interface{}
	Refs            map[string]int
}

func (mat *Material) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	res := Ajax{
		Mat:             mat,
		Textures:        make([]interface{}, len(mat.Layers)),
		TexturesBlended: make([]interface{}, len(mat.Layers)),
		Refs:            make(map[string]int),
	}

	for i := range mat.Layers {
		tn := node.FindNode(mat.Layers[i].Texture)
		if tn != nil {
			res.Refs[tn.Name] = tn.Id
			txr, err := wad.Get(tn.Id)
			if err != nil {
				return nil, fmt.Errorf("Error getting texture '%s' for material '%s': %v", tn.Name, node.Name, err)
			}

			if dat, err := txr.(*file_txr.Texture).Marshal(tn.Wad, tn); err != nil {
				return nil, fmt.Errorf("Error marshaling texture '%s' for material '%s': %v", tn.Name, node.Name, err)
			} else {
				res.Textures[i] = dat
			}

			if dat, err := txr.(*file_txr.Texture).MarshalBlend(mat.Layers[i].BlendColor[:], tn.Wad, tn); err != nil {
				return nil, fmt.Errorf("Error marshaling texture '%s' for material '%s': %v", tn.Name, node.Name, err)
			} else {
				res.TexturesBlended[i] = dat
			}
		}
	}
	return res, nil
}

func init() {
	wad.SetHandler(MAT_MAGIC, func(w *wad.Wad, node *wad.WadNode, r *io.SectionReader) (wad.File, error) {
		return NewFromData(r)
	})
}
