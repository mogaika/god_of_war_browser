package mat

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_anm "github.com/mogaika/god_of_war_browser/pack/wad/anm"
	file_txr "github.com/mogaika/god_of_war_browser/pack/wad/txr"
	"github.com/mogaika/god_of_war_browser/utils"
)

/*
Info that may help:
https://nccastaff.bournemouth.ac.uk/jmacey/RobTheBloke/www/research/maya/mfnenvmap.htm
https://nccastaff.bournemouth.ac.uk/jmacey/RobTheBloke/www/research/maya/mfnmaterial.htm
*/

type Flags struct {
	FilterLinear            bool // when false, then near filter used. may affect only when texture expanded (LOD < 0)
	DisableDepthWrite       bool
	RenderingAdditive       bool
	RenderingUsual          bool // handle transparency
	RenderingSubstract      bool
	RenderingStrangeBlended bool // do not know
	HaveTexture             bool
	NOTSUREThisIsReflection bool

	AnimationUVEnabled    bool // Anim type 8
	AnimationColorEnabled bool // ?? Anim type 3, applyed to mat, not layer
}

type Layer struct {
	Texture     string
	Flags       [4]uint32
	BlendColor  utils.ColorFloat
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
	l.ParsedFlags.NOTSUREThisIsReflection = (l.Flags[0]>>1)&1 != 0

	l.ParsedFlags.HaveTexture = (l.Flags[0]>>7)&1 != 0

	l.ParsedFlags.FilterLinear = (l.Flags[0]>>16)&1 != 0

	l.ParsedFlags.DisableDepthWrite = (l.Flags[0]>>19)&1 != 0

	l.ParsedFlags.RenderingStrangeBlended = (l.Flags[0]>>24)&1 != 0
	l.ParsedFlags.RenderingSubstract = (l.Flags[0]>>25)&1 != 0
	l.ParsedFlags.RenderingUsual = (l.Flags[0]>>26)&1 != 0
	l.ParsedFlags.RenderingAdditive = (l.Flags[0]>>27)&1 != 0

	l.ParsedFlags.AnimationUVEnabled = l.GameFlags&1 != 0
	l.ParsedFlags.AnimationColorEnabled = l.GameFlags&2 != 0

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

func NewFromData(buf []byte) (*Material, error) {
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
		start := iTex*LAYER_SIZE + HEADER_SIZE
		tbuf := buf[start : start+LAYER_SIZE]

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
		if mat.Layers[iTex].FloatUnk != 1.0 {
			// Transparency of layer when using multi-layer?
		}

		mat.Layers[iTex].GameFlags = binary.LittleEndian.Uint32(tbuf[60:64])

		if err := mat.Layers[iTex].ParseFlags(); err != nil {
			return nil, fmt.Errorf("Error paring layer %d: %v", iTex, err)
		}
	}

	return mat, nil
}

type Ajax struct {
	Mat             *Material
	Textures        map[int]interface{}
	TexturesBlended map[int]interface{}
	Animations      interface{}
}

func (mat *Material) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	res := Ajax{
		Mat:             mat,
		Textures:        make(map[int]interface{}),
		TexturesBlended: make(map[int]interface{}),
	}

	for iLayerId := range mat.Layers {
		if mat.Layers[iLayerId].Texture != "" {
			n := wrsrc.Wad.GetNodeByName(mat.Layers[iLayerId].Texture, wrsrc.Node.Id-1, false)
			if n != nil {
				txr, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id)
				if err != nil {
					return nil, fmt.Errorf("Error getting texture '%s' for material '%s': %v", n.Tag.Name, wrsrc.Tag.Name, err)
				}

				if dat, err := txr.(*file_txr.Texture).Marshal(wrsrc.Wad.GetNodeResourceByNodeId(n.Id)); err != nil {
					return nil, fmt.Errorf("Error marshaling texture '%s' for material '%s': %v", n.Tag.Name, wrsrc.Tag.Name, err)
				} else {
					res.Textures[iLayerId] = dat
				}

				if dat, err := txr.(*file_txr.Texture).MarshalBlend(mat.Layers[iLayerId].BlendColor[:], wrsrc.Wad.GetNodeResourceByNodeId(n.Id)); err != nil {
					return nil, fmt.Errorf("Error marshaling texture '%s' for material '%s': %v", n.Tag.Name, wrsrc.Tag.Name, err)
				} else {
					res.TexturesBlended[iLayerId] = dat
				}
			}
		}
	}

	for _, i := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(i)
		name := n.Tag.Name
		sn, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id)
		if err != nil {
			// TODO: improve
			// return nil, fmt.Errorf("Error when extracting node %d->%s mat info: %v", i, name, err)
		} else {
			switch sn.(type) {
			case *file_anm.Animations:
				anims, err := sn.(*file_anm.Animations).Marshal(wrsrc.Wad.GetNodeResourceByNodeId(n.Id))
				if err != nil {
					return nil, fmt.Errorf("Error when getting script info %d-'%s': %v", i, name, err)
				}
				res.Animations = anims
			}
		}
	}
	return res, nil
}

func init() {
	h := func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	}
	wad.SetServerHandler(config.GOW1, MAT_MAGIC, h)
	wad.SetServerHandler(config.GOW2, MAT_MAGIC, h)
}
