package resources

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/editor/imgui/project"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

/*
	SoundBanks         []*project.Resource
	Lights             []*project.Resource
	Textures           []*project.Resource
	Animations         []*project.Resource
	Materials          []*project.Resource
	Models             []*project.Resource
	Particles          []*project.Resource
	ParticleEmitters   []*project.Resource
	SoundEmitters      []*project.Resource
	Shadow             []*project.Resource
	CollisionVolumes   []*project.Resource
	CollisionDebugs    []*project.Resource
	Objects            []*project.Resource
	GameObjects        []*project.Resource
	Waypoints          []*project.Resource
	RootObjects        []*project.Resource
	CollisionStatic    []*project.Resource
	Contexts           []*project.Resource
	OSData             []*project.Resource // 110
	RawFile            []*project.Resource // 111
	UserInterface      []*project.Resource
	Tweaks             Tweaks
	RequiredResources  []string
	ExternalAnimations []*project.Resource
*/

type WAD struct {
	p *project.Project

	Variables       []WADVariable
	ServerConfigs   []*WadServerConfig
	CollisionStatic []*project.Resource
	Contexts        []*project.Resource
	UserInterface   []*project.Resource
	Tweaks          Tweaks

	RootResources []*project.Resource

	RequiredResources []string
}

type WADVariable struct {
	Name  string
	Value int32
}

var WADKind = project.Kind("WAD")

func (w *WAD) Kind() project.Kind { return WADKind }

func (w *WAD) RenderUI() {
	if imgui.TreeNode("ServerConfigs") {
		for i, c := range w.ServerConfigs {
			imgui.PushIDInt(i)
			imgui.Text(fmt.Sprintf("%.5s:%x", c.Server, c.Data[4:]))
			imgui.PopID()
		}
		imgui.Separator()
		imgui.TreePop()
	}
	if imgui.TreeNode("Variables") {
		for i, v := range w.Variables {
			imgui.PushIDInt(i)
			imgui.InputInt(v.Name, &w.Variables[i].Value)
			imgui.PopID()
		}
		imgui.Separator()
		imgui.TreePop()
	}
	if imgui.TreeNodeV("Root resources", imgui.TreeNodeFlagsDefaultOpen) {
		for i, r := range w.RootResources {
			imgui.PushIDInt(i)
			imgui.Text(string(r.GetKind()))
			imgui.SameLine()
			if imgui.Button(fmt.Sprintf("%s", r.GetName())) {
				w.p.OpenResource(r)
			}
			imgui.PopID()
		}
		imgui.Separator()
		imgui.TreePop()
	}
}

type Tweaks struct {
}

type wadSource struct {
	name string
	size int64
}

var _ utils.ResourceSource = (*wadSource)(nil)

func (ws *wadSource) Name() string                    { return ws.name }
func (ws *wadSource) Save(in *io.SectionReader) error { return errors.New("Not implemented") }
func (ws *wadSource) Size() int64                     { return ws.size }

func (w *WAD) tagAsServerInstance(tag wad.Tag, childs []*project.Resource, namespace map[string]*project.Resource) project.IResource {
	serverId := ServerId(binary.LittleEndian.Uint16(tag.Data[0:2]))
	instanceType := binary.LittleEndian.Uint16(tag.Data[2:4])

	type index struct {
		id ServerId
		t  uint16
	}

	noNeedChilds := func() {
		if len(childs) != 0 {
			var names []string
			for _, r := range childs {
				names = append(names, r.GetName())
			}
			log.Panicf("Tag %q has childs %v", tag.Name, names)
		}
	}

	i := index{serverId, instanceType}
	switch i {
	case index{0x3, 0}:
		noNeedChilds()
		anm, err := LoadAnimationsPS2(w.p, tag.Data)
		if err != nil {
			panic(err)
		}
		return anm
	case index{0x4, 1}:
		noNeedChilds()
		s, err := LoadScriptPS2(w.p, tag.Data)
		if err != nil {
			panic(err)
		}
		return s
	case index{0x6, 0}:
		l, err := LoadLightsPS2(w.p, tag.Data, childs)
		if err != nil {
			panic(err)
		}
		return l
	case index{0x7, 0}:
		noNeedChilds()
		txr, err := LoadTexturePS2(w.p, tag.Data, namespace)
		if err != nil {
			panic(err)
		}
		return txr
	case index{0x8, 0}:
		mat, err := LoadMaterialPS2(w.p, tag.Data, namespace, childs)
		if err != nil {
			panic(err)
		}
		return mat
	case index{0xc, 0}:
		noNeedChilds()
		gfx, err := LoadGraphicsPS2(w.p, tag.Data)
		if err != nil {
			panic(err)
		}
		return gfx
	case index{0x17, 0}:
		noNeedChilds()
		gfx, err := LoadBehaviorPS2(w.p, tag.Data)
		if err != nil {
			panic(err)
		}
		return gfx
	default:
		return &UnknownWadServerInstance{
			p: w.p, Server: serverId, Type: instanceType, Data: tag.Data, Childs: childs,
		}
	}

}

func (w *WAD) tagAsUnknownServerInstance(tag wad.Tag, childs []*project.Resource) *UnknownWadServerInstance {
	serverId := ServerId(binary.LittleEndian.Uint16(tag.Data[0:2]))
	instanceType := binary.LittleEndian.Uint16(tag.Data[2:4])
	return &UnknownWadServerInstance{
		p: w.p, Server: serverId, Type: instanceType, Data: tag.Data, Childs: childs,
	}
}

func LoadWadFromReader(p *project.Project, r io.ReadSeeker, name string) (*project.Resource, error) {
	size, _ := r.Seek(0, io.SeekEnd)
	r.Seek(0, io.SeekStart)

	w, err := wad.NewWad(r, &wadSource{name: name, size: int64(size)}, false)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse wad")
	}

	const (
		state_init = iota
		state_header
		state_header_done
		state_resources
	)
	state := state_init
	result := &WAD{p: p}

	type group struct {
		mainTag wad.Tag
		childs  []*project.Resource
		parent  *group
	}

	var currentGroup *group = nil
	isGroupStart := false
	namespace := make(map[string]*project.Resource)

	addNodeNamespaced := func(r *project.Resource) {
		if currentGroup != nil {
			currentGroup.childs = append(currentGroup.childs, r)
		} else if strings.HasPrefix(r.GetName(), " ") {
			// do not include in namespace
		} else {
			namespace[r.GetName()] = r
			result.RootResources = append(result.RootResources, r)
		}
	}

	for _, tag := range w.Tags {
		switch state {
		case state_init:
			if tag.Tag != wad.TAG_GOW1_HEADER_START {
				return nil, errors.Errorf("Expected first tag to be header")
			}
			state = state_header
		case state_header:
			switch tag.Tag {
			case wad.TAG_GOW1_FILE_GROUP_START, wad.TAG_GOW1_FILE_GROUP_END:
				// can ignore since nesting in header is hardcoded
			case wad.TAG_GOW1_SERVER_INSTANCE:
				result.ServerConfigs = append(result.ServerConfigs, &WadServerConfig{
					UnknownWadServerInstance: result.tagAsUnknownServerInstance(tag, nil),
				})
			case wad.TAG_GOW1_HEADER_POP:
				state = state_header_done
			}
		case state_header_done:
			if tag.Tag != wad.TAG_GOW1_DATA_START1 {
				return nil, errors.Errorf("Expected data start tag")
			}
			state = state_resources
		case state_resources:
			switch tag.Tag {
			case wad.TAG_GOW1_FILE_GROUP_START:
				isGroupStart = true
			case wad.TAG_GOW1_FILE_GROUP_END:
				if currentGroup == nil {
					return nil, errors.Errorf("Invalid group end tag placement")
				}

				inst := p.AddResource(currentGroup.mainTag.Name,
					result.tagAsServerInstance(currentGroup.mainTag, currentGroup.childs, namespace))

				currentGroup = currentGroup.parent

				addNodeNamespaced(inst)
			case wad.TAG_GOW1_SERVER_INSTANCE:
				if isGroupStart {
					isGroupStart = false
					currentGroup = &group{
						mainTag: tag,
						parent:  currentGroup,
					}
				} else {
					var inst *project.Resource
					if tag.Size == 0 {
						inst = namespace[tag.Name]
					} else {
						inst = p.AddResource(tag.Name, result.tagAsServerInstance(tag, nil, namespace))
					}

					addNodeNamespaced(inst)
				}
			case wad.TAG_GOW1_ENTITY_COUNT:
				result.Variables = append(result.Variables,
					WADVariable{
						Name:  tag.Name,
						Value: int32(w.HeapSizes[tag.Name]),
					})
			default:
				log.Printf("Failed to load unknown tag %x:%q", tag.Tag, tag.Name)
			}
		default:
			panic(state)
		}
	}

	return p.AddResource(name, result), nil
}

type WadServerConfig struct {
	*UnknownWadServerInstance
}

func (c *WadServerConfig) RenderUI() {}

var WadServerConfigKind = project.Kind("WadServerConfig")
