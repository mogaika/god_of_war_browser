package shg

import (
	"bytes"
	"encoding/binary"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const SHG_MAGIC = 0x00000027

type Object struct {
	Vector1 mgl32.Vec4

	Unk0x10       uint16
	vectors3Count uint16
	vectors4Count uint16
	vectors5Count uint16
	vectors6Count uint16
	vectors7Count uint16
	vectors8Count uint16
	vectors3Start uint16
	vectors4Start uint16
	vectors5Start uint16
	vectors6Start uint16
	vectors7Start uint16
	vectors8Start uint16

	Vectors3 []mgl32.Vec4
	Vectors4 []mgl32.Vec4
	Vectors5 [][2]uint16
	Vectors6 []uint16
	Vectors7 [][2]uint16
	Vectors8 [][2]uint16
}

func (so *Object) Parse(buf []byte) error {
	binary.Read(bytes.NewReader(buf[0x0:0x10]), binary.LittleEndian, &so.Vector1)

	so.Unk0x10 = binary.LittleEndian.Uint16(buf[0x10:])
	so.vectors3Count = binary.LittleEndian.Uint16(buf[0x12:])
	so.vectors4Count = binary.LittleEndian.Uint16(buf[0x14:])
	so.vectors5Count = binary.LittleEndian.Uint16(buf[0x16:])
	so.vectors6Count = binary.LittleEndian.Uint16(buf[0x18:])
	so.vectors7Count = binary.LittleEndian.Uint16(buf[0x1a:])
	so.vectors8Count = binary.LittleEndian.Uint16(buf[0x1c:])
	so.vectors3Start = binary.LittleEndian.Uint16(buf[0x1e:])
	so.vectors4Start = binary.LittleEndian.Uint16(buf[0x20:])
	so.vectors5Start = binary.LittleEndian.Uint16(buf[0x22:])
	so.vectors6Start = binary.LittleEndian.Uint16(buf[0x24:])
	so.vectors7Start = binary.LittleEndian.Uint16(buf[0x26:])
	so.vectors8Start = binary.LittleEndian.Uint16(buf[0x28:])

	so.Vectors3 = make([]mgl32.Vec4, so.vectors3Count)
	so.Vectors4 = make([]mgl32.Vec4, so.vectors4Count)
	so.Vectors5 = make([][2]uint16, so.vectors5Count)
	so.Vectors6 = make([]uint16, so.vectors6Count)
	so.Vectors7 = make([][2]uint16, so.vectors7Count)
	so.Vectors8 = make([][2]uint16, so.vectors8Count)

	if err := binary.Read(bytes.NewReader(buf[so.vectors3Start:]), binary.LittleEndian, so.Vectors3); err != nil {
		return errors.Wrapf(err, "vectors3 read")
	}
	if err := binary.Read(bytes.NewReader(buf[so.vectors4Start:]), binary.LittleEndian, so.Vectors4); err != nil {
		return errors.Wrapf(err, "vectors4 read")
	}
	if err := binary.Read(bytes.NewReader(buf[so.vectors5Start:]), binary.LittleEndian, so.Vectors5); err != nil {
		return errors.Wrapf(err, "vectors5 read")
	}
	if err := binary.Read(bytes.NewReader(buf[so.vectors6Start:]), binary.LittleEndian, so.Vectors6); err != nil {
		return errors.Wrapf(err, "vectors6 read")
	}
	if err := binary.Read(bytes.NewReader(buf[so.vectors7Start:]), binary.LittleEndian, so.Vectors7); err != nil {
		return errors.Wrapf(err, "vectors7 read")
	}
	if err := binary.Read(bytes.NewReader(buf[so.vectors8Start:]), binary.LittleEndian, so.Vectors8); err != nil {
		return errors.Wrapf(err, "vectors8 read")
	}
	return nil
}

type ShadowLod struct {
	Name    string
	Objects []*Object
}

func (sl *ShadowLod) Parse(buf []byte) error {
	sl.Name = utils.BytesToString(buf[4:0x10])

	sl.Objects = make([]*Object, binary.LittleEndian.Uint32(buf[0x10:]))

	offsetsTable := binary.LittleEndian.Uint32(buf[0x14:])
	for i := range sl.Objects {
		sl.Objects[i] = &Object{}
		objOffset := binary.LittleEndian.Uint32(buf[offsetsTable+uint32(i)*4:])
		if err := sl.Objects[i].Parse(buf[objOffset:]); err != nil {
			return errors.Wrapf(err, "Error parsing shadow lod object %d", i)
		}
	}

	return nil
}

func NewFromData(buf []byte) (*ShadowLod, error) {
	sl := &ShadowLod{}

	return sl, nil
}

func (sl *ShadowLod) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return sl, nil
}

func init() {
	wad.SetServerHandler(config.GOW1, SHG_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		sl := &ShadowLod{}
		return sl, sl.Parse(wrsrc.Tag.Data)
	})
}
