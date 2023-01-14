package resources

import (
	"encoding/binary"

	"github.com/inkyblackness/imgui-go/v4"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/editor/imgui/project"
)

type Graphics struct {
	Width      uint32
	Height     uint32
	RealHeight uint32
	Encoding   uint32
	Bpi        uint32
	DataSize   uint32
	Data       [][]byte `json:"-"`
}

func LoadGraphicsPS2(p *project.Project, buf []byte) (*Graphics, error) {
	gfx := &Graphics{
		Width:    binary.LittleEndian.Uint32(buf[4:8]),
		Height:   binary.LittleEndian.Uint32(buf[8:12]),
		Encoding: binary.LittleEndian.Uint32(buf[12:16]),
		Bpi:      binary.LittleEndian.Uint32(buf[16:20]),
		Data:     make([][]byte, binary.LittleEndian.Uint32(buf[20:24])),
	}
	gfx.RealHeight = gfx.Height / uint32(len(gfx.Data))

	if config.GetPlayStationVersion() == config.PS2 {
		pos := uint32(24)
		gfx.DataSize = (((gfx.Width * gfx.RealHeight) * gfx.Bpi) / 8)
		for iData := range gfx.Data {
			gfx.Data[iData] = buf[pos : pos+gfx.DataSize]
			pos += gfx.DataSize
		}
	} else {
		gfx.Data = nil
	}

	return gfx, nil
}

func (g *Graphics) RenderUI() {
	imgui.Textf("Width: %v", g.Width)
	imgui.Textf("Height: %v", g.Height)
	imgui.Textf("RealHeight: %v", g.RealHeight)
	imgui.Textf("Encoding: %v", g.Encoding)
	imgui.Textf("Bpi: %v", g.Bpi)
	imgui.Textf("DataSize: %v", g.DataSize)
	imgui.Textf("DataCount: %v", len(g.Data))
}

func (*Graphics) Kind() project.Kind { return GraphicsKind }

var GraphicsKind = project.Kind("Graphics")
