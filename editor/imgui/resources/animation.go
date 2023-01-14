package resources

import (
	"encoding/binary"

	"github.com/mogaika/god_of_war_browser/editor/imgui/project"
)

type Animations struct {
	*UnknownWadServerInstance
}

func LoadAnimationsPS2(p *project.Project, data []byte) (*Animations, error) {
	b := &Animations{
		UnknownWadServerInstance: &UnknownWadServerInstance{
			p:      p,
			Server: ServerId(binary.LittleEndian.Uint16(data[0:2])),
			Type:   binary.LittleEndian.Uint16(data[2:4]),
			Data:   data,
		},
	}

	return b, nil
}

func (*Animations) Kind() project.Kind { return AnimationsKind }

var AnimationsKind = project.Kind("Animations")
