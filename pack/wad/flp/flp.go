package flp

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
)

const (
	FLP_MAGIC   = 0x21
	HEADER_SIZE = 0x60

	DATA1_ELEMENT_SIZE                            = 0x4
	DATA2_ELEMENT_SIZE                            = 0x8
	DATA2_SUBTYPE1_ELEMENT_SIZE                   = 0x8
	DATA3_ELEMENT_SIZE                            = 0x24
	DATA4_ELEMENT_SIZE                            = 0x24
	DATA5_ELEMENT_SIZE                            = 0x20
	DATA6_ELEMENT_SIZE                            = 0xc
	DATA6_SUBTYPE1_ELEMENT_SIZE                   = 0x18
	DATA6_SUBTYPE2_ELEMENT_SIZE                   = 0x10
	DATA6_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE          = 0x8
	DATA6_SUBTYPE1_SUBTYPE1_SUBTYPE1_ELEMENT_SIZE = 0xa
	DATA6_SUBTYPE1_SUBTYPE2_ELEMENT_SIZE          = 0xc
	DATA6_SUBTYPE1_SUBTYPE2_SUBTYPE1_ELEMENT_SIZE = 0x8
	DATA9_ELEMENT_SIZE                            = 0x14
	DATA10_ELEMENT_SIZE                           = 0x8
)

func posPad4(pos int) int {
	if pos%4 != 0 {
		newPos := pos + 4 - pos%4
		if newPos&3 != 0 {
			panic(fmt.Sprintf("How it even possible? %x + 4 - %x = %x", pos, pos%4, newPos))
		}
		return newPos
	} else {
		return pos
	}
}

var currentFlpInstance *FLP

type FLP struct {
	Unk04                 uint32
	Unk08                 uint32
	GlobalHandlersIndexes []GlobalHandlerIndex
	MeshPartReferences    []MeshPartReference `json:"-"`
	Fonts                 []Font
	StaticLabels          []StaticLabel
	DynamicLabels         []DynamicLabel `json:"-"`
	Datas6                []Data6
	Datas7                []Data6Subtype1 `json:"-"`
	Data8                 Data6Subtype1   // Root logic node
	Transformations       []Transformation
	BlendColors           []BlendColor `json:"-"`
	Strings               []string     `json:"-"`

	scriptPushRefs []ScriptOpcodeStringPushReference
}

type GlobalHandler uint16

func (gh GlobalHandler) MarshalJSON() ([]byte, error) {
	return json.Marshal(currentFlpInstance.GlobalHandlersIndexes[gh])
}

// Mesh instance linkage?
type GlobalHandlerIndex struct {
	TypeArrayId       uint16
	IdInThatTypeArray uint16
}

type MeshPartReference struct {
	MeshPartIndex int16
	Materials     []MeshPartMaterialSlot // Count equals to objects count in mesh part group
}

type MeshPartMaterialSlot struct {
	// Texture Linkage
	Color             uint32
	TextureNameSecOff uint32
	TextureName       string
}

type Font struct {
	CharsCount uint32
	Unk04      uint16
	Size       int16
	Unk08      uint16
	Unk0a      uint16
	// Flags
	// & 1 != 0 => CharNumberToSymbolIdMap contain 0x100 elements of symbol=>char map
	// & 1 == 0 => CharNumberToSymbolIdMap contain CharsCount elements of char=>symbol map
	Flags uint16

	MeshesRefs              []MeshPartReference
	SymbolWidths            []int16
	CharNumberToSymbolIdMap []int16 // Char to glyph map?
	Float020                float32
}

type StaticLabel struct {
	Transformation          Transformation
	RenderCommandsList      []*StaticLabelRenderCommand
	tempRenderCommandBuffer []byte
}

type DynamicLabel struct {
	ValueNameSecOff   uint16
	ValueName         string
	PlaceholderSecOff uint16
	Placeholder       string
	FontHandler       GlobalHandler
	Width1            uint16
	BlendColor        uint32
	StringLengthLimit uint16
	OffsetX1          uint16
	Width2            uint16
	OffsetX2          uint16
	Unk01a            uint16
	Unk01e            uint16
}

type Data6 struct {
	Sub1  Data6Subtype1
	Sub2s []Data6Subtype2
}

type Data6Subtype1 struct {
	TotalFramesCount  uint16
	ElementsAnimation []ElementAnimation
	FrameScriptLables []FrameScriptLabel
	Width             uint16
}

type ElementAnimation struct {
	FramesCount uint16
	KeyFrames   []KeyFrame
}

type KeyFrame struct {
	WhenThisFrameEnds uint16 // in frameNumberUnits
	ElementHandler    GlobalHandler
	TransformationId  uint16
	ColorId           uint16
	NameSecOff        int16
	Name              string
}

type FrameScriptLabel struct {
	TriggerFrameNumber uint16
	LabelNameSecOff    int16
	LabelName          string
	Subs               []Data6Subtype1Subtype2Subtype1
}

type Data6Subtype1Subtype2Subtype1 struct {
	Script           *Script
	scriptDataLength uint32
}

type Data6Subtype2 struct {
	EventKeysMask    uint32
	EventUnkMask     uint16
	Script           *Script
	scriptDataLength uint32
}

type Transformation struct {
	// 2d transformation matrix in fx 1:15:16 format
	Matrix [4]float64

	OffsetX float64
	OffsetY float64
}

type BlendColor struct {
	// in range [0, 256]. used 16 bits for better multiply
	Color [4]uint16 // rgba
}

func NewFromData(buf []byte) (*FLP, error) {
	f := &FLP{}
	currentFlpInstance = f
	if err := f.fromBuffer(buf); err != nil {
		return nil, fmt.Errorf("Error when reading flp header: %v", err)
	}
	return f, nil
}

type Marshaled struct {
	FLP             *FLP
	Model           interface{}
	FontCharAliases config.FontCharToAsciiByteAssoc
	Textures        map[string]interface{}
	ScriptPushRefs  []ScriptOpcodeStringPushReference
}

func (f *FLP) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	mrsh := &Marshaled{
		FLP:            f,
		Textures:       make(map[string]interface{}),
		ScriptPushRefs: f.scriptPushRefs,
	}
	if fontaliases, err := config.GetFontAliases(); err == nil {
		mrsh.FontCharAliases = fontaliases
	} else {
		log.Printf("Error loading fontaliases: %v", err)
	}

	mdln := wrsrc.Wad.GetNodeByName(strings.Replace(wrsrc.Name(), "FLP_", "MDL_", 1), wrsrc.Node.Id, false)
	if mdln != nil {
		if wfile, _, err := wrsrc.Wad.GetInstanceFromNode(mdln.Id); err == nil {
			if marshaledNode, err := wfile.Marshal(wrsrc.Wad.GetNodeResourceByNodeId(mdln.Id)); err == nil {
				mrsh.Model = marshaledNode
			} else {
				log.Printf("Cannot marshal mdl instance %s for %s: %v", mdln.Tag.Name, wrsrc.Name(), err)
			}
		} else {
			log.Printf("Cannot get mdl instance %s for %s: %v", mdln.Tag.Name, wrsrc.Name(), err)
		}
	} else {
		log.Printf("Cannot find mdl_ for %s", wrsrc.Name())
	}

	for _, font := range f.Fonts {
		marshalData2 := func(d2 *MeshPartReference) {
			for _, ref := range d2.Materials {
				if ref.TextureName != "" {
					if _, ok := mrsh.Textures[ref.TextureName]; !ok {
						txr := wrsrc.Wad.GetNodeByName(ref.TextureName, wrsrc.Node.Id, false)
						if txr != nil {
							if wfile, _, err := wrsrc.Wad.GetInstanceFromNode(txr.Id); err == nil {
								if marshaledNode, err := wfile.Marshal(wrsrc.Wad.GetNodeResourceByNodeId(txr.Id)); err == nil {
									mrsh.Textures[ref.TextureName] = marshaledNode
								} else {
									log.Printf("Cannot marshal txr instance %s for %s: %v", txr.Tag.Name, wrsrc.Name(), err)
								}
							} else {
								log.Printf("Cannot get txr instance %s for %s: %v", txr.Tag.Name, wrsrc.Name(), err)
							}
						}

					}
				}
			}
		}
		for i := range font.MeshesRefs {
			marshalData2(&font.MeshesRefs[i])
		}
	}

	return mrsh, nil
}

func init() {
	wad.SetHandler(config.GOW1ps2, FLP_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		inst, err := NewFromData(wrsrc.Tag.Data)
		if err != nil {
			return nil, err
		}

		return inst, nil
	})
}
