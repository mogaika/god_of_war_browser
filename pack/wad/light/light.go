package light

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"

	"github.com/go-gl/mathgl/mgl32"
)

const LIGHT_MAGIC = 0x6
const FILE_SIZE = 0x58

type Light struct {
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
}

func (l *Light) FromWad(data []byte, gow2 bool) error {
	l.Unk04 = binary.LittleEndian.Uint32(data[0x04:])
	l.Flags = binary.LittleEndian.Uint32(data[0x08:])

	if err := binary.Read(bytes.NewReader(data[0x0c:]), binary.LittleEndian, &l.Position); err != nil {
		return errors.Wrapf(err, "Can't parse Position")
	}
	if err := binary.Read(bytes.NewReader(data[0x1c:]), binary.LittleEndian, &l.Rotation); err != nil {
		return errors.Wrapf(err, "Can't parse Rotation")
	}
	if err := binary.Read(bytes.NewReader(data[0x2c:]), binary.LittleEndian, &l.Color); err != nil {
		return errors.Wrapf(err, "Can't parse Color")
	}

	l.Unk3c = math.Float32frombits(binary.LittleEndian.Uint32(data[0x3c:]))
	l.Unk40 = math.Float32frombits(binary.LittleEndian.Uint32(data[0x40:]))
	l.Unk44 = math.Float32frombits(binary.LittleEndian.Uint32(data[0x44:]))
	l.Unk48 = binary.LittleEndian.Uint32(data[0x48:])
	l.Unk4c = math.Float32frombits(binary.LittleEndian.Uint32(data[0x4c:]))
	l.Unk50 = math.Float32frombits(binary.LittleEndian.Uint32(data[0x50:]))
	if !gow2 {
		l.Unk54 = math.Float32frombits(binary.LittleEndian.Uint32(data[0x54:]))
	}

	return nil
}

func (l *Light) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return l, nil
}

func init() {
	wad.SetHandler(config.GOW1, LIGHT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		light := &Light{}
		return light, light.FromWad(wrsrc.Tag.Data, false)
	})
	wad.SetHandler(config.GOW2, LIGHT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		light := &Light{}
		return light, light.FromWad(wrsrc.Tag.Data, true)
	})
}
