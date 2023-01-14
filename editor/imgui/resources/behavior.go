package resources

import (
	"encoding/binary"

	"github.com/mogaika/god_of_war_browser/editor/imgui/project"
)

type Behavior struct {
	*UnknownWadServerInstance
}

func LoadBehaviorPS2(p *project.Project, data []byte) (*Behavior, error) {
	b := &Behavior{
		UnknownWadServerInstance: &UnknownWadServerInstance{
			p:      p,
			Server: ServerId(binary.LittleEndian.Uint16(data[0:2])),
			Type:   binary.LittleEndian.Uint16(data[2:4]),
			Data:   data,
		},
	}

	return b, nil
}

func (*Behavior) Kind() project.Kind { return BehaviorKind }

var BehaviorKind = project.Kind("Behavior")
