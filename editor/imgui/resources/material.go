package resources

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/inkyblackness/imgui-go/v4"

	"github.com/mogaika/god_of_war_browser/editor/imgui/project"
	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	MATERIAL_LAYER_FLAG_TEXTURE_PRESENTED = 0x80
	MATERIAL_HEADER_SIZE                  = 0x38
	MATERIAL_LAYER_SIZE                   = 0x40
)

type Material struct {
	p *project.Project

	Color      [3]float32
	Layers     []Layer
	Animations *project.Resource
}

type Layer struct {
	Texture     *project.Resource
	Flags       [4]uint32
	BlendColor  [4]float32
	FloatUnk    float32
	GameFlags   uint32
	ParsedFlags LayerFlags
}

type LayerFlags struct {
	FilterLinear            bool // when false, then near filter used. may affect only when texture expanded (LOD < 0)
	DisableDepthWrite       bool
	RenderingAdditive       bool
	RenderingUsual          bool // handle transparency
	RenderingSubstract      bool
	RenderingStrangeBlended bool // do not know
	HaveTexture             bool

	AnimationUVEnabled    bool // Anim type 8
	AnimationColorEnabled bool // ?? Anim type 3, applyed to mat, not layer
}

func LoadMaterialPS2(p *project.Project, buf []byte, namespace map[string]*project.Resource, childs []*project.Resource) (*Material, error) {
	mat := &Material{
		p:      p,
		Layers: make([]Layer, binary.LittleEndian.Uint32(buf[0x34:0x38])),
	}

	mat.Color = [3]float32{
		math.Float32frombits(binary.LittleEndian.Uint32(buf[0x8:0xc])),
		math.Float32frombits(binary.LittleEndian.Uint32(buf[0xc:0x10])),
		math.Float32frombits(binary.LittleEndian.Uint32(buf[0x10:0x14])),
	}

	for iTex := range mat.Layers {
		start := iTex*MATERIAL_LAYER_SIZE + MATERIAL_HEADER_SIZE
		tbuf := buf[start : start+MATERIAL_LAYER_SIZE]

		mat.Layers[iTex].Flags = [4]uint32{
			binary.LittleEndian.Uint32(tbuf[0x0:0x4]),
			binary.LittleEndian.Uint32(tbuf[0x4:0x8]),
			binary.LittleEndian.Uint32(tbuf[0x8:0xc]),
			binary.LittleEndian.Uint32(tbuf[0xc:0x10]),
		}
		textureName := utils.BytesToString(tbuf[0x10:0x28])
		mat.Layers[iTex].Texture = namespace[textureName]

		mat.Layers[iTex].BlendColor = [4]float32{
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[0x28:0x2c])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[0x2c:0x30])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[0x30:0x34])),
			math.Float32frombits(binary.LittleEndian.Uint32(tbuf[0x34:0x38])),
		}

		mat.Layers[iTex].FloatUnk = math.Float32frombits(binary.LittleEndian.Uint32(tbuf[0x38:0x3c]))
		if mat.Layers[iTex].FloatUnk != 1.0 {
			// Transparency of layer when using multi-layer?
		}

		mat.Layers[iTex].GameFlags = binary.LittleEndian.Uint32(tbuf[0x3c:0x40])

		if err := mat.Layers[iTex].ParseFlags(); err != nil {
			return nil, fmt.Errorf("Error paring layer %d: %v", iTex, err)
		}
	}

	// TODO: better parsing
	if len(childs) != 0 {
		if len(childs) != 1 {
			// expected single animation only
			panic(childs)
		}
		mat.Animations = childs[0]
	}

	return mat, nil
}

func (l *Layer) ParseFlags() error {
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

func (m *Material) RenderUI() {
	imgui.ColorEdit3("BlendColor", &m.Color)
	for i := range m.Layers {
		if imgui.TreeNodeV(fmt.Sprintf("Layer %v", i), imgui.TreeNodeFlagsDefaultOpen) {
			l := &m.Layers[i]

			project.UIReference(m.p, "Texture", &l.Texture)

			// TODO: use custom color edit with HDR support
			imgui.ColorEdit4("BlendColor", &l.BlendColor)
			for i, f := range l.Flags {
				imgui.Textf("Flags %v: 0x%.8x", i, f)
			}

			imgui.Textf("FloatUnk: %v", l.FloatUnk)
			imgui.Textf("GameFlags: 0x%.8x", l.GameFlags)

			imgui.TreePop()
		}
	}

}

func (*Material) Kind() project.Kind { return MaterialKind }

var MaterialKind = project.Kind("Material")
