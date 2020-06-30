package flp

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
)

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
				slrc.FontScale = float64(int16(binary.LittleEndian.Uint16(data[i+2:]))) / 1024.0
				i += 4
			}
			if op&4 != 0 {
				copy(slrc.BlendColor[:], data[i:i+4])
				i += 4
			}
			if op&2 != 0 {
				slrc.OffsetX = float64(int16(binary.LittleEndian.Uint16(data[i:]))) / 16.0
				i += 2
			}
			if op&1 != 0 {
				slrc.OffsetY = float64(int16(binary.LittleEndian.Uint16(data[i:]))) / 16.0
				i += 2
			}
		} else {
			if slrc.Glyphs == nil {
				slrc.Glyphs = make([]StaticLabelRenderCommandSingleGlyph, 0)
			}
			for j := byte(0); j < op; j++ {
				slrc.Glyphs = append(slrc.Glyphs, StaticLabelRenderCommandSingleGlyph{
					GlyphId: binary.LittleEndian.Uint16(data[i:]),
					Width:   float64(int16(binary.LittleEndian.Uint16(data[i+2:]))) / 16.0,
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

func (d4 *StaticLabel) ParseRenderCommandList(list []byte) []*StaticLabelRenderCommand {
	d4.RenderCommandsList = parseStaticLabelRenderCommandList(list)
	return d4.RenderCommandsList
}

func (d4 *StaticLabel) MarshalRenderCommandList() []byte {
	var o bytes.Buffer
	var buf [4]byte
	for _, cmd := range d4.RenderCommandsList {
		if cmd.Flags != 0 {
			o.WriteByte(cmd.Flags | 0x80)
			if cmd.Flags&8 != 0 {
				binary.LittleEndian.PutUint16(buf[:], cmd.FontHandler)
				binary.LittleEndian.PutUint16(buf[2:], uint16(int16(cmd.FontScale*1024.0)))
				o.Write(buf[:4])
			}
			if cmd.Flags&4 != 0 {
				o.Write(cmd.BlendColor[:4])
			}
			if cmd.Flags&2 != 0 {
				binary.LittleEndian.PutUint16(buf[0:], uint16(int16(cmd.OffsetX*16.0)))
				o.Write(buf[:2])
			}
			if cmd.Flags&1 != 0 {
				binary.LittleEndian.PutUint16(buf[0:], uint16(int16(cmd.OffsetY*16.0)))
				o.Write(buf[:2])
			}
		}

		o.WriteByte(uint8(len(cmd.Glyphs)))
		for _, glyph := range cmd.Glyphs {
			binary.LittleEndian.PutUint16(buf[0:], glyph.GlyphId)
			binary.LittleEndian.PutUint16(buf[2:], uint16(int16(glyph.Width)*16.0))
			o.Write(buf[:4])
		}
	}
	return o.Bytes()
}

func (d4 *StaticLabel) ParseJson(buf []byte) error {
	var unmrshled StaticLabel

	if err := json.Unmarshal(buf, &unmrshled); err != nil {
		log.Println(err)
		return err
	}

	d4.Transformation = unmrshled.Transformation
	d4.RenderCommandsList = unmrshled.RenderCommandsList

	return nil
}
