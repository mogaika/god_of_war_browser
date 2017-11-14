package flp

import "encoding/binary"

type StaticLabelRenderCommandSingleGlyph struct {
	GlyphId uint16
	Width   float64
}

type StaticLabelRenderCommand struct {
	Flags       byte
	FontHandler uint16
	FontScale   float64
	BlendColor  [4]byte
	OffsetX     float64
	OffsetY     float64
	Glyphs      []StaticLabelRenderCommandSingleGlyph
}

func parseStaticLabelRenderCommandList(data []byte) []*StaticLabelRenderCommand {
	var slrc *StaticLabelRenderCommand
	commands := make([]*StaticLabelRenderCommand, 0)

	for i := 0; i < len(data); {
		op := data[i]
		i += 1
		if op&0x80 != 0 {
			if slrc != nil {
				commands = append(commands, slrc)
			}
			slrc = &StaticLabelRenderCommand{Flags: op & 0x7f}
			if op&8 != 0 {
				slrc.FontHandler = binary.LittleEndian.Uint16(data[i:])
				slrc.FontScale = float64(binary.LittleEndian.Uint16(data[i+2:])) / 1024.0
				i += 4
			}
			if op&4 != 0 {
				copy(slrc.BlendColor[:], data[i:i+4])
				i += 4
			}
			if op&2 != 0 {
				slrc.OffsetX = float64(binary.LittleEndian.Uint16(data[i:])) / 16
				i += 2
			}
			if op&1 != 0 {
				slrc.OffsetY = float64(binary.LittleEndian.Uint16(data[i:])) / 16
				i += 2
			}
		} else {
			if slrc.Glyphs == nil {
				slrc.Glyphs = make([]StaticLabelRenderCommandSingleGlyph, 0)
			}
			for j := byte(0); j < op; j++ {
				slrc.Glyphs = append(slrc.Glyphs, StaticLabelRenderCommandSingleGlyph{
					GlyphId: binary.LittleEndian.Uint16(data[i:]),
					Width:   float64(binary.LittleEndian.Uint16(data[i+2:])) / 16,
				})
				i += 4
			}
		}
	}

	if slrc != nil {
		commands = append(commands, slrc)
	}

	return commands
}

func (d4 *StaticLabel) GetRenderCommandList() []*StaticLabelRenderCommand {
	return parseStaticLabelRenderCommandList(d4.RenderCommandsList)
}
