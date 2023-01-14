package archive

import (
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/pkg/errors"
)

type ServerBehavior struct {
	PlaceholderReferencesHolder
	SubServers SubServersHolder
	Behaviors  []*ServerBehaviorInstance
}

type ServerBehaviorInstance PlaceholderInstance

func (sb *ServerBehavior) GetName() string { return "ServerBehavior" }

func (sb *ServerBehavior) OpenWadTag(ldr *Loader, tag *wad.Tag, instanceType InstanceType) (ServerInstance, error) {
	if instanceType == INSTANCE_TYPE_SERVER {
		return sb.SubServers.InsertPlaceholder(tag.Name, tag.Data), nil
	}

	switch instanceType {
	case 0:
		inst := &ServerBehaviorInstance{
			PlaceholderName: PlaceholderName{
				Name: tag.Name,
			},
			Flags: tag.Flags,
			Data:  tag.Data,
		}
		sb.Behaviors = append(sb.Behaviors, inst)
		return inst, nil
	default:
		return nil, errors.Errorf("Unknown type %v", instanceType)
	}
}
