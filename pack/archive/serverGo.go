package archive

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_obj "github.com/mogaika/god_of_war_browser/pack/wad/obj"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/pkg/errors"
)

type ServerGo struct {
	PlaceholderReferencesHolder
	SubServers SubServersHolder
	Objects    []*ServerGoInstanceObject
	Instances  []*ServerGoInstanceGO
}

type ServerGoInstanceGO struct {
	PlaceholderName
	DummyAfterGroupEnd

	Object         *ServerGoInstanceObject
	Id             uint16
	Params         uint16
	Position       mgl32.Vec4
	Rotation       mgl32.Vec3
	Scale          float32
	CenterProbably mgl32.Vec4
	Unk            [3]uint32

	Scripts []*ServerScriptInstance
}

type ServerGoInstanceObject struct {
	PlaceholderName
	DummyAfterGroupEnd
	*file_obj.Object
}

func (sgo *ServerGo) GetName() string { return "ServerGO" }

func (sgo *ServerGo) OpenWadTag(ldr *Loader, tag *wad.Tag, instanceType InstanceType) (ServerInstance, error) {
	if instanceType == INSTANCE_TYPE_SERVER {
		return sgo.SubServers.InsertPlaceholder(tag.Name, tag.Data), nil
	}

	switch instanceType {
	case 2:
		inst := &ServerGoInstanceGO{}
		inst.FromData(ldr, tag)
		sgo.Instances = append(sgo.Instances, inst)
		return inst, nil
	case 4:
		obj, err := file_obj.NewFromData(tag.Data, tag.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create go object")
		}

		wrapped := &ServerGoInstanceObject{Object: obj}
		wrapped.Name = tag.Name

		sgo.Objects = append(sgo.Objects, wrapped)

		return wrapped, nil
	default:
		return nil, errors.Errorf("Unknown type %v", instanceType)
	}
}

func (inst *ServerGoInstanceGO) FromData(ldr *Loader, tag *wad.Tag) {
	inst.Name = tag.Name

	bs := utils.NewBufStack("goinst", tag.Data)

	objectName := utils.BytesToString(bs.Raw()[0x4:0x1c])

	inst.Object = ldr.References[objectName].(*ServerGoInstanceObject)
	inst.Id = bs.LU16(0x1c)
	inst.Params = bs.LU16(0x1e)
	for i := range inst.Position {
		inst.Position[i] = bs.LF(0x20 + i*4)
	}
	for i := range inst.Rotation {
		inst.Rotation[i] = bs.LF(0x30 + i*4)
	}
	inst.Scale = bs.LF(0x3c)
	for i := range inst.CenterProbably {
		inst.CenterProbably[i] = bs.LF(0x40 + i*4)
	}

	inst.Unk[0] = bs.LU32(0x50)
	inst.Unk[0] = bs.LU32(0x54)
	inst.Unk[0] = bs.LU32(0x58)
}

func (inst *ServerGoInstanceGO) AfterGroupEnd(ldr *Loader, group []GroupStackElement) error {
	for _, grel := range group {
		switch child := grel.Instance.(type) {
		case *ServerScriptInstance:
			inst.Scripts = append(inst.Scripts, child)
		default:
			return errors.Errorf("Unknown child %T", grel.Instance)
		}
	}
	return nil
}
