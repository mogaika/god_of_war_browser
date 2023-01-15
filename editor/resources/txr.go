package resources

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/inkyblackness/imgui-go/v4"

	"github.com/mogaika/god_of_war_browser/editor/project"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Texture struct {
	p *project.Project

	GFX *project.Resource
	PAL *project.Resource
	LOD *project.Resource

	LODParamK     int32
	LODMultiplier float32
	Flags         uint32
}

func LoadTexturePS2(p *project.Project, buf []byte, namespace map[string]*project.Resource) (*Texture, error) {
	gfxName := utils.BytesToString(buf[4:28])
	palName := utils.BytesToString(buf[28:52])
	lodName := utils.BytesToString(buf[52:76])

	tex := &Texture{
		p:             p,
		GFX:           namespace[gfxName],
		PAL:           namespace[palName],
		LOD:           namespace[lodName],
		LODParamK:     int32(binary.LittleEndian.Uint32(buf[76:80])),
		LODMultiplier: math.Float32frombits(binary.LittleEndian.Uint32(buf[80:84])),
		Flags:         binary.LittleEndian.Uint32(buf[84:88]),
	}

	// flags
	// 0x010000 - 3d usual/additive/alpha
	// 0x018000 - 3d usual/alpha
	// 0x510000 - 3d additive? billboard?
	// 0x510000 - 2d nontransparent (except for fonts)
	// 0x5d0000 - 2d transparent
	flags1 := tex.Flags & 0xffff
	if flags1 != 0 && flags1 != 0x8000 {
		return nil, fmt.Errorf("Unknown unkFlags 0x%.4x != 0", flags1)
	}

	flags2 := tex.Flags >> 16
	if flags2 != 1 && flags2 != 0x41 && flags2 != 0x5d && flags2 != 0x51 && flags2 != 0x11 {
		return nil, fmt.Errorf("Unknown unkFlags2 0x%.4x (0x1,0x41,0x5d,0x51,0x11)", flags2)
	}

	return tex, nil
}

func (t *Texture) RenderUI() {
	project.UIReference(t.p, "GFX", &t.GFX)
	project.UIReference(t.p, "PAL", &t.PAL)
	project.UIReference(t.p, "LOD", &t.LOD)

	imgui.Textf("LOD Param K: %v", t.LODParamK)
	imgui.Textf("LOD Multiplier: %v", t.LODMultiplier)
	imgui.Textf("Flags: 0x%.8x", t.Flags)
}

func (*Texture) Kind() project.Kind { return TextureKind }

var TextureKind = project.Kind("Texture")
