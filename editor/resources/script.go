package resources

import (
	"encoding/binary"

	"github.com/mogaika/god_of_war_browser/editor/project"
	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/inkyblackness/imgui-go/v4"
)

const SCRIPT_HEADER_SIZE = 0x24

type Script struct {
	p *project.Project

	ScriptName string
	Unk1e      uint16
	Unk1c      uint16
	Unk20      uint16
	Unk22      uint16
	Data       []byte
}

func LoadScriptPS2(p *project.Project, buf []byte) (*Script, error) {
	s := &Script{
		p: p,
	}

	s.ScriptName = utils.BytesToString(buf[0x4:0x14])
	s.Unk1c = binary.LittleEndian.Uint16(buf[0x1c:])
	s.Unk1e = binary.LittleEndian.Uint16(buf[0x1e:])
	s.Unk20 = binary.LittleEndian.Uint16(buf[0x20:])
	s.Unk22 = binary.LittleEndian.Uint16(buf[0x22:])
	s.Data = buf[SCRIPT_HEADER_SIZE:]

	return s, nil
}

func (s *Script) RenderUI() {
	imgui.Textf("ScriptName: %q", s.ScriptName)
	imgui.Textf("Unk1e: 0x%.4x", s.Unk1e)
	imgui.Textf("Unk1c: 0x%.4x", s.Unk1c)
	imgui.Textf("Unk20: 0x%.4x", s.Unk20)
	imgui.Textf("Unk22: 0x%.4x", s.Unk22)
	imgui.Textf("Data size: %v", len(s.Data))
}

func (*Script) Kind() project.Kind { return ScriptKind }

var ScriptKind = project.Kind("Script")
