package anm

import (
	"encoding/binary"
	"math"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const ANIMATIONS_MAGIC = 0x00000003

const (
	DATATYPE_SKINNING     = 0  // apply to object (matrices)
	DATATYPE_MATERIAL     = 3  // apply to material (color)
	DATATYPE_UNKNOWN5     = 5  // apply to object (show/hide maybe)
	DATATYPE_TEXUREPOS    = 8  // apply to material (uv)
	DATATYPE_TEXTURESHEET = 9  // apply to material (changes data_id of gfx palette indexes, like gif frame)
	DATATYPE_PARTICLES    = 10 // apply to object (particles), probably additive matricies animation?, or physical affect body affect
	DATATYPE_UNKNOWN11    = 11 // apply to object (? in StonedBRK models)
	DATATYPE_UNKNOWN12    = 12 // apply to object (? in flagGrp and chest models)
	// total - 15 types
)

type AnimDatatype struct {
	TypeId uint16
	Param1 uint8
	Param2 uint8
}

type AnimActStateDescr struct {
	Unk0             uint16
	OffsetToData     uint32
	CountOfSomething uint16
	ImportantFloat   float32
	Data             interface{}
}

type AnimAct struct {
	Offset      uint32
	Name        string
	StateDescrs []AnimActStateDescr
}

type AnimGroup struct {
	Offset     uint32
	Name       string
	IsExternal bool // when true, then HaventActs in this file
	Acts       []AnimAct
}

type Animations struct {
	ParsedFlags struct {
		Flag0AutoplayProbably bool
		JointRotationAnimated bool
		JointPositionAnimated bool
		JointScaleAnimated    bool
	}

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

	/*
		defer func() {
			if r := recover(); r != nil {
				utils.LogDump("Animation parsing panic: %v", r)
			}
		}()
	*/

	flags := u32(data, 8)
	a.ParsedFlags.Flag0AutoplayProbably = flags&0x1 != 0
	a.ParsedFlags.JointRotationAnimated = flags&0x1000 != 0
	a.ParsedFlags.JointPositionAnimated = flags&0x2000 != 0
	a.ParsedFlags.JointScaleAnimated = flags&0x4000 != 0

	rawFormats := data[0x18+len(a.Groups)*4:]
	for i := range a.DataTypes {
		dt := &a.DataTypes[i]
		rawFmt := rawFormats[i*4 : i*4+4]

		dt.TypeId = u16(rawFmt, 0)
		dt.Param1 = rawFmt[2]
		dt.Param2 = rawFmt[3]
	}

	rawGroupsPointers := data[0x18:]
	for i := range a.Groups {
		g := &a.Groups[i]
		g.Offset = u32(rawGroupsPointers, uint32(i*4))
		rawGroup := data[g.Offset:]

		g.Name = utils.BytesToString(rawGroup[0x14:0x2c])
		g.IsExternal = u32(rawGroup, 8)&0x20000 != 0

		if !g.IsExternal {
			g.Acts = make([]AnimAct, u32(rawGroup, 0xc))
			for j := range g.Acts {
				act := &g.Acts[j]
				act.Offset = u32(rawGroup, uint32(0x30+j*4))

				rawAct := rawGroup[act.Offset:]
				act.Name = utils.BytesToString(rawAct[0x24:0x3c])

				act.StateDescrs = make([]AnimActStateDescr, len(a.DataTypes))
				for iStateDescr := range act.StateDescrs {
					sd := &act.StateDescrs[iStateDescr]
					rawActStateDescr := rawAct[0x64+iStateDescr*0x14:]

					sd.Unk0 = u16(rawActStateDescr, 0)
					sd.CountOfSomething = u16(rawActStateDescr, 2)
					sd.OffsetToData = u32(rawActStateDescr, 8)
					sd.ImportantFloat = math.Float32frombits(u32(rawActStateDescr, 0xc))

					//log.Println(iStateDescr, a.DataTypes, a.DataTypes[iStateDescr].TypeId)
					switch a.DataTypes[iStateDescr].TypeId {
					case DATATYPE_TEXUREPOS:
						data := make([]*AnimState8Texturepos, sd.CountOfSomething)
						for i := 0; i < int(sd.CountOfSomething); i++ {
							data[i] = AnimState8TextureposFromBuf(&a.DataTypes[iStateDescr], rawAct[sd.OffsetToData:], i)
						}
						sd.Data = data
					case DATATYPE_SKINNING:
						data := make([]*AnimState0Skinning, sd.CountOfSomething)
						for i := 0; i < int(sd.CountOfSomething); i++ {
							data[i] = AnimState0SkinningFromBuf(&a.DataTypes[iStateDescr], rawAct[sd.OffsetToData:], i)
						}
						sd.Data = data
					case DATATYPE_TEXTURESHEET:
						buf := rawAct[sd.OffsetToData:]
						data := make([]uint32, u16(buf, 4))
						dataBuf := buf[u16(buf, 0xa):]
						for i := range data {
							data[i] = u32(dataBuf, uint32(i*4))
						}
						sd.Data = data
					}

				}
			}
		}
	}

	return a, nil
}

func (anm *Animations) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return anm, nil
}

func init() {
	wad.SetHandler(ANIMATIONS_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
