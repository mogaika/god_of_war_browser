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

	Vector2  mgl32.Vec4
	Vectors3 []mgl32.Vec4
	Vectors4 []mgl32.Vec4
}

func (so *Object) Parse(buf []byte) error {
	binary.Read(bytes.NewReader(buf[0x0:0x10]), binary.LittleEndian, &so.Vector1)
	binary.Read(bytes.NewReader(buf[0x20:0x30]), binary.LittleEndian, &so.Vector2)

	vectors3Count := binary.LittleEndian.Uint16(buf[0x12:])
	so.Vectors3 = make([]mgl32.Vec4, vectors3Count)
	vectors4Count := binary.LittleEndian.Uint16(buf[0x14:])
	so.Vectors4 = make([]mgl32.Vec4, vectors4Count)

	r := bytes.NewReader(buf[0x30:])
	if err := binary.Read(r, binary.LittleEndian, so.Vectors3); err != nil {
		return errors.Wrapf(err, "vectors3 read")
	}
	if err := binary.Read(r, binary.LittleEndian, so.Vectors4); err != nil {
		return errors.Wrapf(err, "vectors4 read")
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
