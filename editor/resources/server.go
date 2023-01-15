package resources

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/project"
)

type ServerId uint16

var serverIDToString = map[ServerId]string{
	0x1:  "CXT",
	0x3:  "ANM",
	0x4:  "SCR",
	0x6:  "LGT",
	0x7:  "TXR",
	0x8:  "MAT",
	0x9:  "CAM",
	0xc:  "GFX",
	0xf:  "MDL",
	0x11: "COL",
	0x13: "PRT",
	0x14: "WYP",
	0x17: "BHV",
	0x18: "SND",
	0x1a: "EMT",
	0x1b: "WAD",
	0x1c: "EEPR",
	0x1e: "FX",
	0x21: "FLP",
	0x23: "LINE",
	0x27: "SHG",
}

func (sid ServerId) String() string {
	if s, ok := serverIDToString[sid]; ok {
		return s
	} else {
		return fmt.Sprintf("server(%v)", sid)
	}
}

type UnknownWadServerInstance struct {
	p      *project.Project
	Server ServerId
	Type   uint16
	Data   []byte
	Childs []*project.Resource
}

func (*UnknownWadServerInstance) Kind() project.Kind { return UnknownWadServerInstanceKind }

func (uwi *UnknownWadServerInstance) RenderUI() {
	imgui.Textf("Data size: %v", len(uwi.Data))
	if imgui.TreeNodeV("Childs", imgui.TreeNodeFlagsDefaultOpen) {
		for i, r := range uwi.Childs {
			imgui.PushIDInt(i)
			project.UIReference(uwi.p, fmt.Sprintf("Child %v", i), &r)
			imgui.PopID()
		}
		imgui.Separator()
		imgui.TreePop()
	}
}

var UnknownWadServerInstanceKind = project.Kind("UnknownWadServerInstance")
