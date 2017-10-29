package flp

import (
	"encoding/json"
	"fmt"
	"os"

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
	GlobalHandlersIndexes []GlobalHandlerIndex `json:"-"`
	MeshPartReferences    []MeshPartReference
	Fonts                 []Font `json:"-"`
	StaticLabels          []StaticLabel
	DynamicLabels         []DynamicLabel
	Datas6                []Data6
	Datas7                []Data6Subtype1
	Data8                 Data6Subtype1    // Root logic node
	Transformations       []Transformation `json:"-"`
	BlendColors           []BlendColor     `json:"-"`
	Strings               []string         `json:"-"`
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

	Flag2Datas2             []MeshPartReference
	Flag4Datas2             []MeshPartReference
	SymbolWidths            []int16
	CharNumberToSymbolIdMap []int16 // Char to glyph map?
	Float020                float32
}

type StaticLabel struct {
	RenderCommandsList []byte `json:"-"`
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
	// Frame
	LabelNameSecOff int16
	LabelName       string
	Subs            []Data6Subtype1Subtype2Subtype1
}

type Data6Subtype1Subtype2Subtype1 struct {
	Script  *Script
	payload []byte
}

type Data6Subtype2 struct {
	Script  *Script
	payload []byte
}

type Transformation struct {
	Ints  [4]int32 // used as floats, and divided by 65536.0
	Half1 uint16   // used as float also
	Half2 uint16   // used as float too
}

type BlendColor struct {
	// in range [0, 256]. used 16 bits to better multiply
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

func (f *FLP) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return f, nil
}

func init() {
	wad.SetHandler(FLP_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		inst, err := NewFromData(wrsrc.Tag.Data)
		if err != nil {
			return nil, err
		}

		marshaled := inst.marshalBuffer()
		if f, err := os.Create(wrsrc.Name() + "_my.FLP"); err == nil {
			defer f.Close()
			f.Write(marshaled.Bytes())
		}

		return inst, nil
	})
}
