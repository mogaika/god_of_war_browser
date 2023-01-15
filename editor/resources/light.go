package resources

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/project"
	"github.com/pkg/errors"
)

type Light struct {
	p *project.Project

	Unk04 uint32 // == 0 ?

	// 0 - ambient (do not care about rotation+position?)
	// 1 - point light (do care about rotation?)
	// 2/6 - dir light (do care about position?)
	Flags uint32

	Position mgl32.Vec4
	Rotation mgl32.Vec4
	Color    mgl32.Vec4 // can contain negative values, last item is intensity??
	Unk3c    float32
	Unk40    float32
	Unk44    float32
	Unk48    uint32  // uint16  either 1 or 0
	Unk4c    float32 // == 0 ?
	Unk50    float32 // == 0 ?
	Unk54    float32 // == 0 ?

	Animations *project.Resource
}

func LoadLightsPS2(p *project.Project, data []byte, childs []*project.Resource) (*Light, error) {
	l := &Light{
		p: p,
	}

	l.Unk04 = binary.LittleEndian.Uint32(data[0x04:])
	l.Flags = binary.LittleEndian.Uint32(data[0x08:])

	if err := binary.Read(bytes.NewReader(data[0x0c:]), binary.LittleEndian, &l.Position); err != nil {
		return nil, errors.Wrapf(err, "Can't parse Position")
	}
	if err := binary.Read(bytes.NewReader(data[0x1c:]), binary.LittleEndian, &l.Rotation); err != nil {
		return nil, errors.Wrapf(err, "Can't parse Rotation")
	}
	if err := binary.Read(bytes.NewReader(data[0x2c:]), binary.LittleEndian, &l.Color); err != nil {
		return nil, errors.Wrapf(err, "Can't parse Color")
	}

	l.Unk3c = math.Float32frombits(binary.LittleEndian.Uint32(data[0x3c:]))
	l.Unk40 = math.Float32frombits(binary.LittleEndian.Uint32(data[0x40:]))
	l.Unk44 = math.Float32frombits(binary.LittleEndian.Uint32(data[0x44:]))
	l.Unk48 = binary.LittleEndian.Uint32(data[0x48:])
	l.Unk4c = math.Float32frombits(binary.LittleEndian.Uint32(data[0x4c:]))
	l.Unk50 = math.Float32frombits(binary.LittleEndian.Uint32(data[0x50:]))
	l.Unk54 = math.Float32frombits(binary.LittleEndian.Uint32(data[0x54:]))

	// TODO: better way of parsing
	if len(childs) != 0 {
		if len(childs) == 1 {
			l.Animations = childs[0]
		} else {
			panic(childs)
		}
	}

	return l, nil
}

func (l *Light) RenderUI() {
	imgui.Textf("Unk04: %v", l.Unk04)
	imgui.Textf("Position: %v", l.Position)
	imgui.Textf("Rotation: %v", l.Rotation)
	imgui.Textf("Color: %v", l.Color)
	imgui.Textf("Unk3c: %v", l.Unk3c)
	imgui.Textf("Unk40: %v", l.Unk40)
	imgui.Textf("Unk44: %v", l.Unk44)
	imgui.Textf("Unk48: 0x%.8x", l.Unk48)
	imgui.Textf("Unk4c: %v", l.Unk4c)
	imgui.Textf("Unk50: %v", l.Unk50)
	imgui.Textf("Unk54: %v", l.Unk54)

	project.UIReference(l.p, "Animations", &l.Animations)
}

func (*Light) Kind() project.Kind { return LightKind }

var LightKind = project.Kind("Light")
