package anm

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

/*
TODO: Rework so it looks like
Clip => Track => Frame

Clip - colleciton of tracks

*/

const ANIMATIONS_MAGIC = 0x00000003

const (
	DATATYPE_SKINNING     = 0  // apply to object (matrices)
	DATATYPE_MATERIAL     = 3  // apply to material (color) or light
	DATATYPE_UNKNOWN5     = 5  // apply to object (show/hide maybe or switch meshes or cameras)
	DATATYPE_TEXUREPOS    = 8  // apply to material (uv)
	DATATYPE_TEXTURESHEET = 9  // apply to material (changes data_id of gfx palette indexes, like gif frame)
	DATATYPE_PARTICLES    = 10 // apply to object (particles), probably additive matricies animation?, or physical affect body affect
	DATATYPE_UNKNOWN11    = 11 // apply to object with sound emitter (? in StonedBRK models and chest model and pushpullblock)
	DATATYPE_UNKNOWN12    = 12 // apply to object (? in flagGrp) flag wind simulation? first mesh static, second - leather? per-vertice animation? rotation-only animation?
	// total - 15 types
)

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

type AnimDatatype struct {
	TypeId          uint16
	Param1          uint8
	TrackSpecsCount uint8

	TrackSpecsStartIndex int
}

type AnimGroup struct {
	Name       string
	IsExternal bool // when true, then Clips located outside of this file
	Clips      []AnimClip

	offset uint32
}

type AnimClip struct {
	// 0x0001 - looped
	// 0x0002 - not looped
	// 0x2000 - more then one frame
	Flags         uint32
	flagsProbably string
	Unk0x4        float32
	Unk0xc        float32
	Unk0x10       int32
	Unk0x20       int32
	Duration      float32
	Name          string

	TrackSpecs []AnimTrackSpec
	TrackTyped []any

	offset uint32
}

type AnimTrackSpec struct {
	Unk0             uint8
	Unk1             uint8
	CountOfSomething uint16
	Unk4             uint16
	Unk6             uint8
	Unk7             uint8
	OffsetToData     uint32
	FrameTime        float32
	Unk16            uint32
}

func u32(d []byte, off uint32) uint32 {
	return binary.LittleEndian.Uint32(d[off : off+4])
}
func u16(d []byte, off uint32) uint16 {
	return binary.LittleEndian.Uint16(d[off : off+2])
}

func NewFromData(animInstanceRawData []byte) (*Animations, error) {
	bs := utils.NewBufStack("anm", animInstanceRawData).SetSize(len(animInstanceRawData))

	bsHeader := bs.SubBuf("header", 0).SetSize(0x18)

	a := &Animations{
		DataTypes: make([]AnimDatatype, bsHeader.LU16(0x10)),
		Groups:    make([]AnimGroup, bsHeader.LU16(0x12)),
	}

	defer func() {
		/*if r := recover(); r != nil {
			utils.LogDump("Animation parsing panic: %v", r)
		}*/
	}()

	defer func() {
		log.Printf("\n%s", bs.StringTree())
	}()

	flags := bsHeader.LU16(8)
	a.ParsedFlags.Flag0AutoplayProbably = flags&0x1 != 0
	a.ParsedFlags.JointRotationAnimated = flags&0x1000 != 0
	a.ParsedFlags.JointPositionAnimated = flags&0x2000 != 0
	a.ParsedFlags.JointScaleAnimated = flags&0x4000 != 0

	var _l *utils.Logger
	var fff *os.File
	defer fff.Close()

	bsGroupPointers := bsHeader.SubBufFollowing("groups pointers").SetSize(len(a.Groups) * 4)
	bsTrackTypes := bsGroupPointers.SubBufFollowing("track types").SetSize(len(a.DataTypes) * 4)

	trackSpecsCount := 0
	for iTrackType := range a.DataTypes {
		dt := &a.DataTypes[iTrackType]
		bsTrackType := bsTrackTypes.SubBuf("track type", iTrackType*4).SetSize(4)

		dt.TypeId = bsTrackType.LU16(0)
		dt.Param1 = bsTrackType.Byte(2)
		dt.TrackSpecsCount = bsTrackType.Byte(3)

		dt.TrackSpecsStartIndex = trackSpecsCount
		trackSpecsCount += int(dt.TrackSpecsCount)

		bsTrackType.SetName(fmt.Sprintf("%d %d-%d", dt.TypeId, dt.Param1, dt.TrackSpecsCount))

		if dt.TypeId == DATATYPE_SKINNING {
			if iTrackType != 0 {
				log.Panicf("Animation datatype skinning requirerd to be first because of gow engine internals (static offsets). Happend on %d", iTrackType)
			}
			if dt.TrackSpecsCount < 3 {
				log.Panicf("Animation datatype skinning expected at least 3 tracks for pos/rot/scale. Got %v", dt.TrackSpecsCount)
			}
			fff, _ = os.Create(`currentanim.log`)
			_l = &utils.Logger{fff}
		}
	}

	for iGroup := range a.Groups {
		a.Groups[iGroup].offset = bsGroupPointers.LU32(iGroup * 4)
	}

	for iGroup := range a.Groups {
		g := &a.Groups[iGroup]

		bsGroup := bs.SubBuf("group", int(g.offset))
		if iGroup == len(a.Groups)-1 {
			bsGroup.Expand()
		} else {
			bsGroup.SetSize(int(a.Groups[iGroup+1].offset - g.offset))
		}

		bsGroupHeader := bsGroup.SubBuf("header", 0).SetSize(0x30)

		g.Name = utils.BytesToString(bsGroupHeader.Raw()[0x14:0x2c])
		bsGroup.SetName(g.Name)

		g.IsExternal = bsGroupHeader.LU32(8)&0x20000 != 0

		_l.Printf("++++++++++++++ GROUP '%s' +++++++++++++++++++++++++++++++++++++++++++++++++++++", g.Name)

		if !g.IsExternal {
			g.Clips = make([]AnimClip, bsGroupHeader.LU32(0xc))

			bsClipPointers := bsGroupHeader.SubBufFollowing("clip pointers").SetSize(len(g.Clips) * 4)

			for iClip := range g.Clips {
				g.Clips[iClip].offset = bsClipPointers.LU32(iClip * 4)
			}

			for iClip := range g.Clips {
				clip := &g.Clips[iClip]

				bsClip := bsGroup.SubBuf("clip", int(clip.offset))
				if iClip == len(g.Clips)-1 {
					bsClip.Expand()
				} else {
					bsClip.SetSize(int(g.Clips[iClip+1].offset - clip.offset))
				}

				bsClipHeader := bsClip.SubBuf("header", 0).SetSize(0x64)

				clip.Flags = bsClipHeader.LU32(0x0)
				clip.flagsProbably = fmt.Sprintf("%.32b", clip.Flags)
				clip.Unk0x4 = math.Float32frombits(bsClipHeader.LU32(0x4))
				clip.Unk0xc = math.Float32frombits(bsClipHeader.LU32(0xc))
				clip.Unk0x10 = int32(bsClipHeader.LU32(0xc))
				clip.Unk0x20 = int32(bsClipHeader.LU32(0xc))
				clip.Duration = math.Float32frombits(bsClipHeader.LU32(0x1c))
				clip.Name = utils.BytesToString(bsClipHeader.Raw()[0x24:0x3c])
				bsClip.SetName(clip.Name)

				_l.Printf("======= CLIP '%s'  (datatypes cnt: %d  duration: %f) ======================================================",
					clip.Name, trackSpecsCount, clip.Duration)
				_l.Printf("---- SDUMP: %s", utils.SDump(bsClipHeader.Raw()))
				_l.Printf("---- SDUMP: %s", utils.SDump(clip))

				clip.TrackSpecs = make([]AnimTrackSpec, trackSpecsCount)
				bsTrackSpecs := bsClipHeader.SubBufFollowing("track specs").SetSize(len(clip.TrackSpecs) * 0x14)

				for iTrackSpec := range clip.TrackSpecs {
					trackSpec := &clip.TrackSpecs[iTrackSpec]

					bsTargetSpec := bsTrackSpecs.SubBuf("track spec", iTrackSpec*0x14).SetSize(0x14)

					trackSpec.Unk0 = bsTargetSpec.Byte(0)
					trackSpec.Unk1 = bsTargetSpec.Byte(1)
					trackSpec.CountOfSomething = bsTargetSpec.LU16(2)
					trackSpec.Unk4 = bsTargetSpec.LU16(4)
					trackSpec.Unk6 = bsTargetSpec.Byte(6)
					trackSpec.Unk7 = bsTargetSpec.Byte(7)
					trackSpec.OffsetToData = bsTargetSpec.LU32(8)
					trackSpec.FrameTime = bsTargetSpec.LF(0xc)
					trackSpec.Unk16 = bsTargetSpec.LU32(0x10)
				}

				_l.Printf("---- track specs:\n%s", utils.SDump(clip.TrackSpecs))
				_l.Printf("---- track specs:\n%s", utils.SDump(bsTrackSpecs.Raw()))

				bsTracksData := bsTrackSpecs.SubBufFollowing("tracks data").Expand()
				bsTracksDataOffset := uint32(bsTracksData.RelativeOffset())

				for iDataType := range a.DataTypes {
					dt := &a.DataTypes[iDataType]

					var trackTypeData any

					switch dt.TypeId {
					case DATATYPE_SKINNING:
						skinningTracks := &AnimStateSkinningTracks{}

						trackSpecRotation := &clip.TrackSpecs[dt.TrackSpecsStartIndex+0]
						for i := 0; i < int(trackSpecRotation.CountOfSomething); i++ {
							skinningTracks.Rotation = append(skinningTracks.Rotation,
								ParseSkinningAttributeTrackRotation(bsTracksData.Raw()[trackSpecRotation.OffsetToData-bsTracksDataOffset:][i*0xc:], _l))
						}

						trackSpecPosition := &clip.TrackSpecs[dt.TrackSpecsStartIndex+1]
						for i := 0; i < int(trackSpecPosition.CountOfSomething); i++ {
							skinningTracks.Position = append(skinningTracks.Position,
								ParseSkinningAttributeTrackPosition(bsTracksData.Raw()[trackSpecPosition.OffsetToData-bsTracksDataOffset:][i*0xc:], _l),
							)
						}

						trackSpecScale := &clip.TrackSpecs[dt.TrackSpecsStartIndex+2]
						for i := 0; i < int(trackSpecScale.CountOfSomething); i++ {
							skinningTracks.Scale = append(skinningTracks.Scale,
								ParseSkinningAttributeTrackScale(bsTracksData.Raw()[trackSpecScale.OffsetToData-bsTracksDataOffset:][i*0xc:], _l),
							)
						}

						if dt.TrackSpecsCount > 3 {
							trackSpecAttachments := &clip.TrackSpecs[dt.TrackSpecsStartIndex+3]
							for i := 0; i < int(trackSpecAttachments.CountOfSomething); i++ {
								skinningTracks.Attachments = append(skinningTracks.Attachments,
									ParseSkinningAttributeTrackAttachments(bsTracksData.Raw()[trackSpecAttachments.OffsetToData-bsTracksDataOffset:][i*0xc:], _l))
							}
						}
						//_l.Printf("---- track spec position:\n%s", utils.SDump(trackSpecPosition))
						//_l.Printf("---- tracks position:\n%s", utils.SDump(skinningTracks.Position))
						trackTypeData = skinningTracks
					case DATATYPE_TEXUREPOS:
						trackSpecTexsturePos := &clip.TrackSpecs[dt.TrackSpecsStartIndex]

						data := make([]*AnimState8Texturepos, trackSpecTexsturePos.CountOfSomething)
						for i := 0; i < int(trackSpecTexsturePos.CountOfSomething); i++ {
							data[i] = AnimState8TextureposFromBuf(bsTracksData.Raw()[trackSpecTexsturePos.OffsetToData-bsTracksDataOffset:][i*0xc:])
						}
						trackTypeData = data
					case DATATYPE_TEXTURESHEET:
						trackSpecTexstureSheet := &clip.TrackSpecs[dt.TrackSpecsStartIndex]

						buf := bsTracksData.Raw()[trackSpecTexstureSheet.OffsetToData-bsTracksDataOffset:]
						data := make([]uint32, u16(buf, 4))
						dataBuf := buf[u16(buf, 0xa):]
						for i := range data {
							data[i] = u32(dataBuf, uint32(i*4))
						}
						trackTypeData = data
					}

					clip.TrackTyped = append(clip.TrackTyped, trackTypeData)
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
	wad.SetServerHandler(config.GOW1, ANIMATIONS_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
