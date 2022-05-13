package archive

import "github.com/mogaika/god_of_war_browser/pack/wad"
import "github.com/pkg/errors"

type ServerScript struct {
	PlaceholderReferencesHolder
	SubServers SubServersHolder
	Scripts    []*ServerScriptInstance
}

type ServerScriptInstance PlaceholderInstance

func (ss *ServerScript) GetName() string { return "ServerScript" }

func (ss *ServerScript) OpenWadTag(ldr *Loader, tag *wad.Tag, instanceType InstanceType) (ServerInstance, error) {
	if instanceType == INSTANCE_TYPE_SERVER {
		return ss.SubServers.InsertPlaceholder(tag.Name, tag.Data), nil
	}

	switch instanceType {
	case 1:
		inst := &ServerScriptInstance{
			Flags:           tag.Flags,
			PlaceholderName: PlaceholderName{Name: tag.Name},
		}
		ss.Scripts = append(ss.Scripts, inst)
		return inst, nil
	default:
		return nil, errors.Errorf("Unknown type %v", instanceType)
	}
}
