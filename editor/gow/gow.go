package gow

import (
	"encoding/binary"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/mogaika/god_of_war_browser/editor/core"
	"github.com/mogaika/god_of_war_browser/pack/wad/gfx"
	"github.com/mogaika/god_of_war_browser/pack/wad/inst"
	"github.com/mogaika/god_of_war_browser/pack/wad/mat"
	"github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	"github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	"github.com/mogaika/god_of_war_browser/pack/wad/obj"
	"github.com/mogaika/god_of_war_browser/pack/wad/txr"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/vfs"
	"github.com/pkg/errors"
)

type wadLoader struct {
	f   vfs.File
	pos int64

	headerParams []byte
	rsrcs        []byte
	variables    map[string]int32
	rawData      map[string][]byte
}

type wadTag struct {
	tag   uint16
	flags uint16
	size  uint32
	name  string
}

func (wl *wadLoader) unmarshalTag(buf []byte) wadTag {
	return wadTag{
		tag:   binary.LittleEndian.Uint16(buf[0:2]),
		flags: binary.LittleEndian.Uint16(buf[2:4]),
		size:  binary.LittleEndian.Uint32(buf[4:8]),
		name:  utils.BytesToString(buf[8:32]),
	}
}

func (wl *wadLoader) readTag() (wadTag, error) {
	var buf [32]byte
	if _, err := wl.f.ReadAt(buf[:], wl.pos); err != nil {
		return wadTag{}, err
	}
	wl.pos += int64(len(buf))
	return wl.unmarshalTag(buf[:]), nil
}

func (wl *wadLoader) readData(size uint32) ([]byte, error) {
	data := make([]byte, size)
	if _, err := wl.f.ReadAt(data, wl.pos); err != nil {
		return nil, err
	}
	wl.pos += int64(size)
	wl.alignPos()
	return data, nil
}

func (wl *wadLoader) alignPos() {
	wl.pos = ((wl.pos + 15) / 16) * 16
}

type wadLoaderLayer struct {
	name      string
	namespace map[string]uuid.UUID
	resources []uuid.UUID // with references
	parent    *wadLoaderLayer
	root      core.Ref[ServerInstanceResource]
}

func (wl *wadLoader) newLayer(parent *wadLoaderLayer, name string) *wadLoaderLayer {
	return &wadLoaderLayer{
		name:      name,
		namespace: make(map[string]uuid.UUID),
		parent:    parent,
	}
}

/*
func (layer *wadLoaderLayer) path() string {
	if layer == nil {
		return ""
	} else {
		return layer.parent.path() + "/" + layer.name
	}
}
*/

type serverKind struct {
	name     string
	subTypes map[uint16]string
}
type snames = map[uint16]string

var serverKinds = map[uint16]serverKind{
	0x1:  {name: "objects", subTypes: snames{0x2: "go", 0x4: "objects"}},
	0x3:  {name: "animations", subTypes: snames{0x0: "animations", 0x1: "external"}},
	0x4:  {name: "scripts", subTypes: snames{0x1: "script"}},
	0x6:  {name: "lights", subTypes: snames{0x0: "lights"}},
	0x7:  {name: "textures", subTypes: snames{0x0: "textures"}},
	0x8:  {name: "materials", subTypes: snames{0x0: "materials"}},
	0x9:  {name: "cameras", subTypes: snames{0x0: "cameras"}},
	0xc:  {name: "graphics", subTypes: snames{0x0: "data"}},
	0xf:  {name: "models", subTypes: snames{0x1: "meshes", 0x2: "models"}},
	0x11: {name: "collisions", subTypes: snames{0x0: "collisions"}},
	0x13: {name: "particles", subTypes: snames{0x0: "particles"}},
	0x14: {name: "waypoints", subTypes: snames{0x0: "waypoints"}},
	0x17: {name: "behaviors", subTypes: snames{0x0: "behaviors"}},
	0x18: {name: "sounds", subTypes: snames{0x0: "banks", 0x4: "vags"}},
	0x1a: {name: "emitters", subTypes: snames{0x0: "semaphore"}},
	0x1b: {name: "wad"},
	0x1c: {name: "eepr"},
	0x1e: {name: "effects", subTypes: snames{
		0x0: "0_effects", 0x1: "1_splash", 0x2: "2_emitters", 0x3: "3_emitters",
		0x4: "4_droolemit", 0x5: "5_emitsplash", 0x6: "6_Gemit", 0x7: "7_waveemitter",
		0x8: "7_smokeemit", 0xa: "A_blackemit", 0xc: "C_gravityFields", 0xd: "D_polySurface"}},
	0x21: {name: "ui", subTypes: snames{0: "ui"}},
	0x23: {name: "line"},
	0x27: {name: "shadows", subTypes: snames{0x0: "group"}},
}

func LoadWadFromReader(project *core.Project, wadFile vfs.File) error {
	if err := wadFile.Open(true); err != nil {
		return errors.Wrap(err, "failed to open wad file")
	}

	wl := wadLoader{
		f:         wadFile,
		variables: make(map[string]int32),
		rawData:   make(map[string][]byte),
	}

	fileSize := wadFile.Size()
	rootLayer := wl.newLayer(nil, wadFile.Name())
	currentLayer := rootLayer

	searchLayersRecursive := func(lowerLayer *wadLoaderLayer, name string) (uuid.UUID, bool) {
		for layer := lowerLayer; layer != nil; layer = layer.parent {
			/*
				var keys []string
				for key := range layer.namespace {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				log.Printf("Searching %q in %v", name, keys)
			*/

			if id, ok := layer.namespace[name]; ok {
				return id, true
			}
		}
		return uuid.Nil, false
	}

	for {
		if wl.pos == fileSize {
			break
		}

		tag, err := wl.readTag()
		if err != nil {
			return errors.Wrapf(err, "Failed to read wad tag")
		}
		// log.Printf("pos 0x%x tag %+#v", wl.pos, tag)

		var tagData []byte
		if tag.size == 0 || tag.tag == 24 {
			// do not load
		} else {
			if data, err := wl.readData(tag.size); err != nil {
				return errors.Wrapf(err, "Failed to load tag %q data", tag.name)
			} else {
				tagData = data
			}
		}

		//directory := ""
		directory := wadFile.Name()

		switch tag.tag {
		case 888: // wad header start
			wl.headerParams = tagData
		case 999: // wad header end
		case 666: // wad data start
		case 500: // RSRCS
			wl.rsrcs = tagData
		case 24: // variable
			wl.variables[tag.name] = int32(tag.size)
		case 30: // inst
			//log.Printf("loading inst %q", tag.name)
			if tag.size == 0 {
				// reference
				id, found := searchLayersRecursive(currentLayer, tag.name)
				if !found {
					unresolvedId := project.AddResource(directory+"/unresolved/"+tag.name, &UnresolvedReference{})
					currentLayer.resources = append(currentLayer.resources, unresolvedId)
					// return errors.Errorf("Failed to find tag reference to %q", tag.name)
					log.Printf("Failed to find tag reference to %q", tag.name)
				} else {
					currentLayer.resources = append(currentLayer.resources, id)
				}
			} else {
				// instance
				if currentLayer.parent != nil && !currentLayer.root.Exists() {
					currentLayer.name = tag.name
				}

				serverId := binary.LittleEndian.Uint16(tagData[0:2])
				kindId := binary.LittleEndian.Uint16(tagData[2:4])
				server, ok := serverKinds[serverId]
				if !ok {
					return errors.Errorf("Unknown server id 0x%x for tag %q", serverId, tag.name)
				}

				directory += "/" + server.name
				if kindId == 0x8000 {
					directory += "/_servers"
				} else {
					if kind, ok := server.subTypes[kindId]; !ok {
						//return errors.Errorf("Unknown kind id 0x%x for serverid 0x%x tag %q", kindId, serverId, tag.name)
						log.Printf("Unknown kind id 0x%x for serverid 0x%x tag %q", kindId, serverId, tag.name)
						directory += "/unknown"
					} else {
						directory += "/" + kind
					}
				}
				if !ok {
					return errors.Errorf("Unknown server id 0x%x for tag %q", serverId, tag.name)
				}

				var newResource core.Resource = &DefaultServerInstanceResource{
					Data: tagData,
					Name: tag.name,
				}
				type t = [2]uint16
				searchCode := t{serverId, kindId}
				switch searchCode {
				case t{0xc, 0x0}:
					if r, err := gfx.NewFromData(tagData); err != nil {
						panic(err)
					} else {
						newResource = &GFX{OG: r}
					}
				case t{0x7, 0x0}:
					if r, err := txr.NewFromData(tagData); err != nil {
						panic(err)
					} else {
						gfxUid, _ := searchLayersRecursive(currentLayer, r.GfxName)
						palUid, _ := searchLayersRecursive(currentLayer, r.PalName)
						subUid, _ := searchLayersRecursive(currentLayer, r.SubTxrName)
						newResource = &Texture{
							OG:  r,
							GFX: core.NewRef[*GFX](gfxUid),
							PAL: core.NewRef[*GFX](palUid),
							LOD: core.NewRef[*Texture](subUid),
						}
					}
				case t{0x8, 0x0}:
					if r, err := mat.NewFromData(tagData); err != nil {
						panic(err)
					} else {
						newMat := &Material{
							OG:    r,
							Color: [4]float32(r.Color),
						}
						for _, layer := range r.Layers {
							textureUid, _ := searchLayersRecursive(currentLayer, layer.Texture)
							newMat.Layers = append(newMat.Layers, &MaterialLayer{
								og:      layer,
								texture: core.NewRef[*Texture](textureUid),
								color:   [4]float32(layer.BlendColor),
							})
						}
						newResource = newMat
					}
				case t{0xf, 0x1}:
					if r, err := mesh.NewFromData(tagData, nil); err != nil {
						panic(err)
					} else {
						newResource = &Meshes{OG: r}
					}
				case t{0xf, 0x2}:
					if r, err := mdl.NewFromData(tagData); err != nil {
						panic(err)
					} else {
						newMdl := &Model{
							OG: r,
						}
						newResource = newMdl
					}
				case t{0x1, 0x4}:
					if r, err := obj.NewFromData(tagData, tag.name); err != nil {
						panic(err)
					} else {
						newObj := &Object{
							OG: r,
						}
						newResource = newObj
					}
				case t{0x1, 0x2}:
					if r, err := inst.NewFromData(tagData); err != nil {
						panic(err)
					} else {
						newInst := &GameObject{
							OG: r,
						}
						objectUid, _ := searchLayersRecursive(currentLayer, r.Object)
						newInst.Object = core.NewRef[*Object](objectUid)
						newResource = newInst
					}
				case t{0x1, 0x8000}:
					newContext := &GameContext{
						Data: tagData,
						Name: tag.name,
					}
					newResource = newContext
				}

				id := project.AddResource(directory+"/"+tag.name, newResource)

				if currentLayer.parent != nil && !currentLayer.root.Exists() {
					// this is first instance inside group
					currentLayer.root = core.NewRef[ServerInstanceResource](id)
				} else {
					currentLayer.resources = append(currentLayer.resources, id)
				}

				if !strings.HasPrefix(tag.name, " ") { // not private
					currentLayer.namespace[tag.name] = id
				}
			}
		case 40: // group start
			//log.Printf("group start %vq", tag.name)
			currentLayer = wl.newLayer(currentLayer, tag.name)
		case 50: // group end
			//log.Printf("group end %q %q %q", tag.name, currentLayer.name, currentLayer.root)
			parent := currentLayer.parent
			root := currentLayer.root
			if root.Exists() {
				parent.namespace[currentLayer.name] = root.Uid()
				parent.resources = append(parent.resources, root.Uid())
				root.Resolve(project).WadGroupEnd(project, currentLayer.resources)
			}
			currentLayer = parent
		case 110, 111, 112: // mc_data
			// case 111: // mc_icon
			// case 112: // raw_data
			wl.rawData[tag.name] = tagData
		case 113, 114: // twk_typeA
			// case 114: // twk_typeB
			project.AddResource(directory+"/tweaks/"+tag.name, &Tweak{
				Tag:  tag.tag,
				Data: tagData,
			})
		default:
			return errors.Errorf("Unknown tag 0x%x on pos 0x%x", tag.tag, wl.pos)
		}
	}

	arhive := &Archive{}
	project.AddResource("archive_"+wadFile.Name(), arhive)
	arhive.WadGroupEnd(project, rootLayer.resources)

	return nil
}
