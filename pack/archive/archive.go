package archive

import (
	"encoding/binary"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/rsrcs"
	"github.com/mogaika/god_of_war_browser/pack/wad/twk"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/pkg/errors"
	"log"
	"strings"
)

type Archive struct {
	HeapSize uint32
	IsLevel  bool

	Servers   map[ServerId]Server
	TWKs      map[string]*twk.TWK // vfs elements
	MCData    map[string][]byte   // probably datas for system
	MCIcon    map[string][]byte   // probably datas to load
	Resources []string            // other R_ wads to load with this one
}

func createPlaceholderServer(name string) *PlaceholderServer {
	return &PlaceholderServer{PlaceholderName: PlaceholderName{Name: name}}
}

func (ar *Archive) initServers() {
	ar.Servers = make(map[ServerId]Server)
	ar.Servers[SERVER_ID_CXT] = &ServerGo{}
	ar.Servers[SERVER_ID_ANMX] = createPlaceholderServer("ANMX")
	ar.Servers[SERVER_ID_SCRX] = createPlaceholderServer("SCRX")
	ar.Servers[SERVER_ID_LGTX] = createPlaceholderServer("LGTX")
	ar.Servers[SERVER_ID_TXRX] = createPlaceholderServer("TXRX")
	ar.Servers[SERVER_ID_MATX] = createPlaceholderServer("MATX")
	ar.Servers[SERVER_ID_CAMX] = createPlaceholderServer("CAMX")
	ar.Servers[SERVER_ID_GFX] = createPlaceholderServer("GFX")
	ar.Servers[SERVER_ID_MDLX] = createPlaceholderServer("MDLX")
	ar.Servers[SERVER_ID_COLX] = createPlaceholderServer("COLX")
	ar.Servers[SERVER_ID_PRTX] = createPlaceholderServer("PRTX")
	ar.Servers[SERVER_ID_WYPX] = createPlaceholderServer("WYPX")
	ar.Servers[SERVER_ID_BHVX] = createPlaceholderServer("BHVX")
	ar.Servers[SERVER_ID_SNDX] = createPlaceholderServer("SNDX")
	ar.Servers[SERVER_ID_EMTX] = createPlaceholderServer("EMTX")
	ar.Servers[SERVER_ID_WAD] = &ServerWad{PlaceholderName: PlaceholderName{Name: "WAD"}}
	ar.Servers[SERVER_ID_EEPR] = createPlaceholderServer("EEPR")
	ar.Servers[SERVER_ID_FX] = createPlaceholderServer("FX")
	ar.Servers[SERVER_ID_FLPX] = createPlaceholderServer("FLPX")
	ar.Servers[SERVER_ID_LINE] = createPlaceholderServer("LINE")
	ar.Servers[SERVER_ID_SHGX] = createPlaceholderServer("SHGX")
}

type Loader struct {
	Servers    map[ServerId]Server
	References map[string]ServerInstance
	RawArrays  map[string][]byte
}

type GroupStackElement struct {
	Name     string
	Instance ServerInstance // may be nil if required from other wad
}

func (ar *Archive) loadPs2(p Provider, wd *wad.Wad) error {
	ar.initServers()
	ar.TWKs = make(map[string]*twk.TWK)
	groupStack := make([][]GroupStackElement, 0)
	popHeap := false

	loader := &Loader{
		Servers:    ar.Servers,
		References: make(map[string]ServerInstance),
		RawArrays:  make(map[string][]byte),
	}

	for tagId, tag := range wd.Tags {
		// log.Printf("%q %v 0x%.4x, %v", tag.Name, tag.Tag, tag.Flags, tag.Size)

		switch tag.Tag {
		case wad.TAG_GOW1_DATA_START1:
		case wad.TAG_GOW1_ENTITY_COUNT:
			log.Printf("Variable %q=%v %+#v", tag.Name, wd.HeapSizes[tag.Name], tag)
		case wad.TAG_GOW1_HEADER_POP:
			popHeap = true
			_ = popHeap
		case wad.TAG_GOW1_HEADER_START:
			if tagId != 0 {
				return errors.Errorf("Expected heap size as first tag")
			}
			sz1 := binary.LittleEndian.Uint32(tag.Data[0:4])
			sz2 := binary.LittleEndian.Uint32(tag.Data[4:8])
			if sz1 != sz2 {
				return errors.Errorf("Wad heap sizes do not match")
			}
			ar.HeapSize = sz1
		case wad.TAG_GOW1_FILE_GROUP_START:
			groupStack = append(groupStack, make([]GroupStackElement, 0))
		case wad.TAG_GOW1_SERVER_INSTANCE:
			var inst ServerInstance
			if len(tag.Data) != 0 {
				serverId := ServerId(binary.LittleEndian.Uint16(tag.Data))
				instanceType := InstanceType(binary.LittleEndian.Uint16(tag.Data[2:]))

				server, ok := ar.Servers[serverId]
				if !ok {
					return errors.Errorf("Wasn't able to find server for id %v", serverId)
				}

				if newInst, err := server.OpenWadTag(loader, &tag, instanceType); err != nil {
					return errors.Wrapf(err, "Failed to open server instance")
				} else {
					inst = newInst
					if !strings.HasPrefix(tag.Name, " ") {
						loader.References[tag.Name] = inst
					}
				}
			} else {
				inst = loader.References[tag.Name]
			}
			if len(groupStack) != 0 {
				pos := len(groupStack) - 1
				groupStack[pos] = append(groupStack[pos], GroupStackElement{Name: tag.Name, Instance: inst})
			}
			// log.Printf("inst %+#v", groupStack)
		case wad.TAG_GOW1_FILE_GROUP_END:
			if len(groupStack) == 0 {
				return errors.Errorf("Unexpected group end when there is no group")
			}
			// log.Printf("pop %+#v", groupStack)
			pos := len(groupStack) - 1
			group := groupStack[pos]
			group[0].Instance.AfterGroupEnd(loader, group[1:])
			groupStack = groupStack[:pos]
			if pos != 0 {
				pos--
				groupStack[pos] = append(groupStack[pos], group[0])
			}
		case wad.TAG_GOW1_FILE_RAW_DATA:
			loader.RawArrays[tag.Name] = tag.Data
		case wad.TAG_GOW1_TWK_INSTANCE:
			entry, err := twk.NewTwkFromData(utils.NewBufStack("twk", tag.Data))
			if err != nil {
				return errors.Wrapf(err, "Failed to load TWK %q", tag.Name)
			}
			ar.TWKs[tag.Name] = entry
		case wad.TAG_GOW1_TWK_OBJECT:
			entry, err := twk.NewTwkFromCombatFile(utils.NewBufStack("twkcb", tag.Data))
			if err != nil {
				return errors.Wrapf(err, "Failed to load TWK cb %q", tag.Name)
			}
			ar.TWKs[tag.Name] = entry
		case wad.TAG_GOW1_FILE_MC_DATA:
			if ar.MCData == nil {
				ar.MCData = make(map[string][]byte)
			}
			ar.MCData[tag.Name] = tag.Data
		case wad.TAG_GOW1_FILE_MC_ICON:
			if ar.MCIcon == nil {
				ar.MCIcon = make(map[string][]byte)
			}
			ar.MCIcon[tag.Name] = tag.Data
		case wad.TAG_GOW1_RSRCS:
			res, err := rsrcs.NewRSRCSFromData(utils.NewBufStack("rsrcs", tag.Data))
			if err != nil {
				return errors.Wrapf(err, "Failed to load resources")
			}
			ar.Resources = res.Wads
			ar.IsLevel = true
		default:
			return errors.Errorf("Unknown tag %+#v", tag)
		}
	}
	return nil
}

func LoadArchive(p Provider, name string) (*Archive, error) {
	wd, err := p.GetArchive(name)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read file")
	}

	ar := &Archive{}

	return ar, ar.loadPs2(p, wd)
}

func (ar *Archive) ConvertToJSON() (interface{}, error) {
	instReferences := make(map[interface{}]interface{})
	_ = instReferences
	return nil, nil
}
