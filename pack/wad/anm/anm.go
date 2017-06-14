package anm

import (
	"encoding/binary"
	"io"
	"log"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const ANIMATIONS_MAGIC = 0x00000003

const (
	DATATYPE_SKINNING  = 0  // apply to object (matrices)
	DATATYPE_MATERIAL  = 3  // apply to material (color)
	DATATYPE_UNKNOWN5  = 5  // apply to object (show/hide maybe)
	DATATYPE_TEXUREPOS = 8  // apply to material (uv)
	DATATYPE_UNKNOWN9  = 9  // apply to material
	DATATYPE_PARTICLES = 10 // apply to object (particles), probably additive matricies animation?, or physical affect body affect
	DATATYPE_UNKNOWN11 = 11 // apply to object (? in StonedBRK models)
	DATATYPE_UNKNOWN12 = 12 // apply to object (? in flagGrp models)
	// total - 15 types
)

type AnimDatatype struct {
	TypeId uint16
	Param1 uint8
	Param2 uint8
}

type AnimAct struct {
	Name   string
	Offset uint32
}

type AnimGroup struct {
	Name   string
	Acts   []AnimAct
	Offset uint32
}

type Animations struct {
	DataTypes []AnimDatatype
	Groups    []AnimGroup
}

func u32(d []byte, off uint32) uint32 {
	return binary.LittleEndian.Uint32(d[off : off+4])
}
func u16(d []byte, off uint32) uint16 {
	return binary.LittleEndian.Uint16(d[off : off+2])
}

func NewFromData(data []byte) (*Animations, error) {
	a := &Animations{
		DataTypes: make([]AnimDatatype, u16(data, 0x10)),
		Groups:    make([]AnimGroup, u16(data, 0x12)),
	}

	rawGroupsPointers := data[0x18:]
	for i := range a.Groups {
		g := &a.Groups[i]
		g.Offset = u32(rawGroupsPointers, uint32(i*4))
		rawGroup := data[g.Offset:]

		g.Name = utils.BytesToString(rawGroup[0x14:0x2c])
		g.Acts = make([]AnimAct, u32(rawGroup, 0xc))

		for j := range g.Acts {
			act := &g.Acts[j]
			act.Offset = u32(rawGroup, uint32(0x30+j*4))
			log.Println(i, j, act, len(rawGroup))

			// TODO : Thereis issue with size cheking
			if act.Offset+0x30 < uint32(len(rawGroup)) {
				rawAct := rawGroup[act.Offset:]

				act.Name = utils.BytesToString(rawAct[0x24:0x3c])
			}
		}
	}

	rawFormats := data[0x18+len(a.Groups)*4:]
	for i := range a.DataTypes {
		dt := &a.DataTypes[i]
		rawFmt := rawFormats[i*4 : i*4+4]

		dt.TypeId = u16(rawFmt, 0)
		dt.Param1 = rawFmt[2]
		dt.Param2 = rawFmt[3]
	}

	return a, nil
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
