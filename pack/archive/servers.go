package archive

import (
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/pkg/errors"
)

type ServerId uint16

const (
	SERVER_ID_CXT  ServerId = 0x1
	SERVER_ID_ANMX          = 0x3
	SERVER_ID_SCRX          = 0x4
	SERVER_ID_LGTX          = 0x6
	SERVER_ID_TXRX          = 0x7
	SERVER_ID_MATX          = 0x8
	SERVER_ID_CAMX          = 0x9
	SERVER_ID_GFX           = 0xc
	SERVER_ID_MDLX          = 0xf
	SERVER_ID_COLX          = 0x11
	SERVER_ID_PRTX          = 0x13
	SERVER_ID_WYPX          = 0x14
	SERVER_ID_BHVX          = 0x17
	SERVER_ID_SNDX          = 0x18
	SERVER_ID_EMTX          = 0x1a
	SERVER_ID_WAD           = 0x1b
	SERVER_ID_EEPR          = 0x1c
	SERVER_ID_FX            = 0x1e
	SERVER_ID_FLPX          = 0x21
	SERVER_ID_LINE          = 0x23
	SERVER_ID_SHGX          = 0x27
)

type InstanceType uint16

const (
	INSTANCE_TYPE_SERVER InstanceType = 0x8000
)

type Server interface {
	ServerInstance
	OpenWadTag(ldr *Loader, tag *wad.Tag, instanceType InstanceType) (ServerInstance, error)
}

type ServerInstance interface {
	GetName() string
	AfterGroupEnd(ldr *Loader, instances []GroupStackElement) error
}

type PlaceholderReferencesHolder struct {
	References []GroupStackElement `json:",omitempty"`
}

func (prh *PlaceholderReferencesHolder) AfterGroupEnd(ldr *Loader, group []GroupStackElement) error {
	if prh.References != nil {
		return errors.Errorf("Already used group end")
	}
	prh.References = group
	return nil
}

type PlaceholderInstancesHolder struct {
	Instances []ServerInstance `json:",omitempty"`
}

func (pih *PlaceholderInstancesHolder) OpenWadTag(ldr *Loader, tag *wad.Tag, instanceType InstanceType) (ServerInstance, error) {
	inst := &PlaceholderInstance{
		PlaceholderName: PlaceholderName{Name: tag.Name},
		Flags:           tag.Flags,
		Data:            tag.Data,
	}
	pih.Instances = append(pih.Instances, inst)
	return inst, nil
}

type PlaceholderInstance struct {
	PlaceholderReferencesHolder `yaml:",inline"`
	PlaceholderName             `yaml:",inline"`
	Flags                       uint16
	Data                        []byte `json:",omitempty" yaml:",flow"`
}

type PlaceholderName struct {
	Name string
}

func (pn *PlaceholderName) GetName() string { return pn.Name }

type PlaceholderServer struct {
	PlaceholderReferencesHolder `yaml:",inline"`
	PlaceholderInstancesHolder  `yaml:",inline"`
	PlaceholderName             `yaml:",inline"`
	ServerInit                  []byte `json:",omitempty" yaml:",flow"`
	SubServers                  SubServersHolder
}

type DummyAfterGroupEnd struct{}

func (*DummyAfterGroupEnd) AfterGroupEnd(ldr *Loader, instances []GroupStackElement) error {
	return nil
}

type SubServersHolder []Server

func (sh *SubServersHolder) Insert(s Server) {
	*sh = append(*sh, s)
}

func (sh *SubServersHolder) InsertPlaceholder(name string, serverInit []byte) *PlaceholderServer {
	inst := &PlaceholderServer{}
	inst.Name = name
	inst.ServerInit = serverInit
	sh.Insert(inst)
	return inst
}

var _ Server = (*PlaceholderServer)(nil)
var _ ServerInstance = (*PlaceholderServer)(nil)
var _ ServerInstance = (*PlaceholderInstance)(nil)

func (ps *PlaceholderServer) OpenWadTag(ldr *Loader, tag *wad.Tag, instanceType InstanceType) (ServerInstance, error) {
	if instanceType == INSTANCE_TYPE_SERVER {
		return ps.SubServers.InsertPlaceholder(tag.Name, tag.Data), nil
	}
	return ps.PlaceholderInstancesHolder.OpenWadTag(ldr, tag, instanceType)
}
