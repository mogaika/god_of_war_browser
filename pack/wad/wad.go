package wad

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/utils"
)

const WAD_ITEM_SIZE = 0x20

type File interface {
	Marshal(rsrc *WadNodeRsrc) (interface{}, error)
}

type FileLoader func(rsrc *WadNodeRsrc) (File, error)

var gHandlers map[uint32]FileLoader = make(map[uint32]FileLoader, 0)

func SetHandler(format uint32, ldr FileLoader) {
	gHandlers[format] = ldr
}

var gTagHandlers map[uint16]FileLoader = make(map[uint16]FileLoader, 0)

func SetTagHandler(tag uint16, ldr FileLoader) {
	gTagHandlers[tag] = ldr
}

type NodeId int
type TagId int

const NODE_INVALID = -1

type Wad struct {
	Source utils.ResourceSource `json:"-"`
	Tags   []Tag                `json:"-"`
	Nodes  []*Node
	Roots  []NodeId

	HeapSizes map[string]uint32
}

type Tag struct {
	Id     TagId // our internal id to identify tags
	Tag    uint16
	Flags  uint16
	Size   uint32
	Name   string
	Data   []byte `json:"-"`
	NodeId NodeId
}

type Node struct {
	Id             NodeId // our internal id to identify nodes
	Tag            *Tag
	Parent         NodeId
	SubGroupNodes  []NodeId
	Cache          File `json:"-"`
	CachedServerId uint32
}

func (w *Wad) CallHandler(id NodeId) (File, uint32, error) {
	var h FileLoader
	var serverId uint32

	n := w.GetNodeById(id)
	if han, ex := gTagHandlers[n.Tag.Tag]; ex {
		h = han
	} else if n.Tag.Tag == TAG_SERVER_INSTANCE {
		if n.Tag.Data != nil && len(n.Tag.Data) >= 4 {
			serverId = binary.LittleEndian.Uint32(n.Tag.Data)
			if han, ex := gHandlers[serverId]; ex {
				h = han
			}
		}
	}
	if h == nil {
		return nil, serverId, fmt.Errorf("Cannot find handler for tag %.4x (%s)", n.Tag.Tag, n.Tag.Name)
	}

	instance, err := h(w.GetNodeResourceByNodeId(n.Id))
	if err != nil {
		return nil, serverId, fmt.Errorf("Handler return error: %v", err)
	}
	n.Cache = instance
	n.CachedServerId = serverId
	return instance, serverId, nil
}

func (w *Wad) GetInstanceFromTag(tagId TagId) (File, uint32, error) {
	tag := w.GetTagById(tagId)
	return w.GetInstanceFromNode(tag.NodeId)
}

func (w *Wad) GetInstanceFromNode(nodeId NodeId) (File, uint32, error) {
	node := w.GetNodeById(nodeId)
	if node.Cache != nil {
		return node.Cache, node.CachedServerId, nil
	} else {
		return w.CallHandler(node.Id)
	}
}

func UnmarshalTag(buf []byte) Tag {
	return Tag{
		Tag:    binary.LittleEndian.Uint16(buf[0:2]),
		Flags:  binary.LittleEndian.Uint16(buf[2:4]),
		Size:   binary.LittleEndian.Uint32(buf[4:8]),
		Name:   utils.BytesToString(buf[8:32]),
		NodeId: NODE_INVALID,
	}
}

func MarshalTag(t *Tag) []byte {
	buf := make([]byte, WAD_ITEM_SIZE)
	binary.LittleEndian.PutUint16(buf[0:2], t.Tag)
	binary.LittleEndian.PutUint16(buf[2:4], t.Flags)
	binary.LittleEndian.PutUint32(buf[4:8], t.Size)
	copy(buf[8:32], utils.StringToBytes(t.Name, 24, false))
	return buf
}

func (wad *Wad) Name() string {
	return wad.Source.Name()
}

func (w *Wad) GetNodeResourceByNodeId(id NodeId) *WadNodeRsrc {
	return w.GetNodeResourceByTagId(w.GetNodeById(id).Tag.Id)
}

func (w *Wad) GetNodeResourceByTagId(id TagId) *WadNodeRsrc {
	tag := w.GetTagById(id)
	return &WadNodeRsrc{Wad: w, Node: w.GetNodeById(tag.NodeId), Tag: tag}
}

func (w *Wad) loadTags(r io.ReadSeeker) error {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Wad parsing error: %v", err)
		}
	}()

	w.Tags = make([]Tag, 0)
	w.HeapSizes = make(map[string]uint32)
	var buf [WAD_ITEM_SIZE]byte
	for id := TagId(0); ; id++ {
		_, err := r.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return fmt.Errorf("Error reading from wad: %v", err)
			}
		}
		t := UnmarshalTag(buf[:])
		t.Id = id

		if t.Tag == 0x18 {
			// entity count
			w.HeapSizes[t.Name] = t.Size
			t.Size = 0
		}

		if t.Size != 0 {
			t.Data = make([]byte, t.Size)
			if _, err := r.Read(t.Data); err != nil {
				return fmt.Errorf("Error when reading tag '%s' body: %v", t.Name, err)
			}
		}

		w.Tags = append(w.Tags, t)

		if pos, err := r.Seek(0, os.SEEK_CUR); err == nil {
			pos = ((pos + 15) / 16) * 16
			if _, err := r.Seek(pos, os.SEEK_SET); err != nil {
				return fmt.Errorf("Error when seek_set: %v", err)
			}
		} else {
			return fmt.Errorf("Error when seek_cur: %v", err)
		}

	}

	return nil
}

func (w *Wad) addNode(tag *Tag) *Node {
	n := &Node{
		Tag: tag,
		Id:  NodeId(len(w.Nodes))}
	n.Parent = NODE_INVALID
	w.Nodes = append(w.Nodes, n)
	tag.NodeId = n.Id
	return n
}

func (w *Wad) parseTags() error {
	newGroupTag := false
	currentNode := NodeId(NODE_INVALID)
	w.Nodes = make([]*Node, 0)
	w.Roots = make([]NodeId, 0)

	addNode := func(tag *Tag) *Node {
		n := w.addNode(tag)
		n.Parent = currentNode
		if n.Parent == NODE_INVALID {
			w.Roots = append(w.Roots, n.Id)
		} else {
			p := w.GetNodeById(n.Parent)
			if p.SubGroupNodes == nil {
				p.SubGroupNodes = make([]NodeId, 0)
			}
			p.SubGroupNodes = append(p.SubGroupNodes, n.Id)
		}
		return n
	}

	for id := range w.Tags {
		tag := w.GetTagById(TagId(id))
		switch tag.Tag {
		case TAG_SERVER_INSTANCE: // file data packet
			// Tell file server (server determined by first uint16)
			// that new file is loaded
			// if name start with space, then name ignored (unnamed)
			// overwrite previous instance with same name
			n := addNode(tag)
			if newGroupTag {
				newGroupTag = false
				currentNode = n.Id
			}
		case TAG_FILE_GROUP_START: // file data group start
			newGroupTag = true // same realization in game
		case TAG_FILE_GROUP_END: // file data group end
			// Tell server about group ended
			if !newGroupTag {
				// theres been some nodes and we change currentNode
				if currentNode == NODE_INVALID {
					return fmt.Errorf("Trying to end not started group id%d-%s", tag.Id, tag.Name)
				}
				currentNode = w.GetNodeById(currentNode).Parent
			} else {
				newGroupTag = false
			}
		case TAG_ENTITY_COUNT: // entity count
			// Game also add empty named node to nodedirectory?
		case TAG_FILE_MC_DATA: // MC_DATA   < R_PERM.WAD
			// Just add node to nodedirectory
			addNode(tag)
		case TAG_FILE_MC_ICON: // MC_ICON   < R_PERM.WAD
			// Like 0x006e, but also store size of data
			addNode(tag)
		case TAG_FILE_RAW_DATA: // MSH_BDepoly6Shape
			// Add node to nodedirectory only if
			// another node with this name not exists
			addNode(tag)
		case TAG_TWK_INSTANCE: // TWK_Cloth_195
			// Tweaks affect VFS of game
			// AI logics, animation
			// exec twk asap?
			addNode(tag)
		case TAG_TWK_OBJECT: // TWK_CombatFile_328
			// Affect VFS too
			// store twk in mem?
			addNode(tag)
		case TAG_RSRCS: // RSRCS
			// probably affect WadReader
			// (internally transformed to R_RSRCS)
			addNode(tag)
		case TAG_DATA_START1, TAG_DATA_START2, TAG_DATA_START3: // file data start
			// PopBatchServerStack of server from first uint16
			addNode(tag)
		case TAG_HEADER_START: // file header start
			// create new memory namespace and push to memorystack
			// create new nodedirectory and push to nodestack
			// data loading init
			addNode(tag)
		case TAG_HEADER_POP: // file header pop heap
			// data loading structs cleanup
			addNode(tag)
		default:
			return fmt.Errorf("unknown tag id%.4x-tag%.4x-%s", tag.Id, tag.Tag, tag.Name)
		}
	}

	return nil
}

func (w *Wad) GetTagById(id TagId) *Tag {
	return &w.Tags[id]
}

func (w *Wad) GetTagByName(name string, searchStart TagId, searchForward bool) *Tag {
	if searchForward {
		for i := searchStart; int(i) < len(w.Tags); i++ {
			if t := w.GetTagById(i); t.Name == name {
				return t
			}
		}
	} else {
		for i := searchStart; i >= 0; i-- {
			if t := w.GetTagById(i); t.Name == name {
				return t
			}
		}
	}
	return nil
}

func (w *Wad) GetNodeById(nodeId NodeId) *Node {
	if nodeId == NODE_INVALID {
		return nil
	}
	n := w.Nodes[nodeId]
	if n.Tag.Tag == TAG_SERVER_INSTANCE {
		if n.Tag.Size == 0 {
			linked := w.GetNodeByName(n.Tag.Name, n.Id-1, false)
			if linked != nil && linked.Tag.Data != nil && len(linked.Tag.Data) >= 4 {
				return linked
			}
		}
	}
	return n
}

func (w *Wad) GetNodeByName(name string, searchStart NodeId, searchForward bool) *Node {
	if searchForward {
		for i := searchStart; int(i) < len(w.Nodes); i++ {
			if w.Nodes[i].Tag.Name == name {
				if n := w.GetNodeById(i); n.Parent == NODE_INVALID {
					return n
				}
			}
		}
	} else {
		for i := searchStart; i >= 0; i-- {
			if w.Nodes[i].Tag.Name == name {
				if n := w.GetNodeById(i); n.Parent == NODE_INVALID {
					return n
				}
			}
		}
	}
	return nil
}

func (w *Wad) UpdateTagsData(updateData map[TagId][]byte) error {
	var buf bytes.Buffer

	for _, t := range w.Tags {
		data := t.Data
		if t.Tag == TAG_ENTITY_COUNT {
			t.Size = w.HeapSizes[t.Name]
		} else {
			if newdata, ex := updateData[t.Id]; ex {
				data = newdata
				log.Println("Changing size at ", t.Name, " from ", t.Size, " to ", len(data))
			}
			t.Size = uint32(len(data))
		}

		buf.Write(MarshalTag(&t))
		if data != nil && len(data) != 0 {
			buf.Write(data)

			targetPos := ((buf.Len() + 15) / 16) * 16
			buf.Write(make([]byte, targetPos-buf.Len()))
		}
	}

	// To prevent usage of this wad instance
	w.Tags = nil
	w.Nodes = nil
	w.Roots = nil

	return w.Source.Save(io.NewSectionReader(bytes.NewReader(buf.Bytes()), 0, int64(buf.Len())))
}

func NewWad(r io.ReadSeeker, rsrc utils.ResourceSource) (*Wad, error) {
	w := &Wad{
		Source: rsrc,
	}

	if err := w.loadTags(r); err != nil {
		return nil, fmt.Errorf("Error when loading tags: %v", err)
	}

	if err := w.parseTags(); err != nil {
		return nil, fmt.Errorf("Error when parsing tags: %v", err)
	}

	return w, nil
}

type WadNodeRsrc struct {
	Node *Node
	Wad  *Wad
	Tag  *Tag
}

func (r *WadNodeRsrc) Name() string {
	return r.Node.Tag.Name
}

func (r *WadNodeRsrc) Size() int64 {
	return int64(r.Node.Tag.Size)
}

func init() {
	pack.SetHandler(".WAD", func(p utils.ResourceSource, r *io.SectionReader) (interface{}, error) {
		return NewWad(r, p)
	})
}
