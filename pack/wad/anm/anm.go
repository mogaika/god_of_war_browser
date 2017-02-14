package anm

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const ANIMATIONS_MAGIC = 0x00000003

const (
	DATATYPE_SKINNING  = 0
	DATATYPE_UNKNOWN1  = 1
	DATATYPE_UNKNOWN2  = 2
	DATATYPE_MATERIAL  = 3
	DATATYPE_UNKNOWN6  = 6
	DATATYPE_UNKNOWN7  = 7
	DATATYPE_TEXUREPOS = 8
	DATATYPE_PARTICLES = 10
)

type RawDataType struct {
	TypeId uint16
	Param1 uint8
	Param2 uint8
}

type RawAnim struct {
	Offset uint32
	//	DatasCount
	Name string
}

type RawHeader struct {
	UnkFloat  float32
	FileSize  uint32
	DataTypes []RawDataType
	Anims     []RawAnim
	Data      []byte `json:"-"`
}

type Animations struct {
	Raw RawHeader
}

func NewFromData(data []byte) (*Animations, error) {
	anm := &Animations{}
	if err := anm.Raw.parseRawData(data); err != nil {
		return nil, err
	}
	return anm, nil
}

func (rh *RawHeader) parseRawData(data []byte) error {
	rh.Data = data
	rh.FileSize = binary.LittleEndian.Uint32(data[0xc:])
	if rh.FileSize != uint32(len(data)) {
		return fmt.Errorf("Filesize field not match real file size")
	}
	rh.UnkFloat = math.Float32frombits(binary.LittleEndian.Uint32(data[0x8:]))

	animsCount := int(binary.LittleEndian.Uint16(data[0x12:]))
	rh.Anims = make([]RawAnim, animsCount)
	for i := range rh.Anims {
		ptr := binary.LittleEndian.Uint32(data[0x18+i*4:])
		rh.Anims[i] = RawAnim{
			Offset: ptr,
			Name:   utils.BytesToString(data[ptr+0x14:]),
		}
	}

	rh.DataTypes = make([]RawDataType, binary.LittleEndian.Uint16(data[0x10:]))
	dataTypesStart := animsCount*4 + 0x18
	for i := range rh.DataTypes {
		ptr := dataTypesStart + i*4
		rh.DataTypes[i] = RawDataType{
			TypeId: binary.LittleEndian.Uint16(data[ptr:]),
			Param1: data[ptr+2],
			Param2: data[ptr+3],
		}
	}

	return nil
}

func (anm *Animations) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	return anm, nil
}

func init() {
	wad.SetHandler(ANIMATIONS_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		data := make([]byte, node.Size)
		_, err := r.ReadAt(data, 0)
		if err != nil {
			return nil, err
		}
		return NewFromData(data)
	})
}
