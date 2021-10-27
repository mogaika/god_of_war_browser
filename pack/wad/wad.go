package wad

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mogaika/god_of_war_browser/pack/wad/scr/entitycontext"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/utils"
)

const WAD_ITEM_SIZE = 0x20

type File interface {
	Marshal(rsrc *WadNodeRsrc) (interface{}, error)
}

type FileLoader func(rsrc *WadNodeRsrc) (File, error)

var gHandlers map[uint64]FileLoader = make(map[uint64]FileLoader, 0)

func SetHandler(version config.GOWVersion, serverId uint32, ldr FileLoader) {
	key := (uint64(version) << 32) | uint64(serverId)
	if _, ok := gHandlers[key]; ok {
		log.Panicf("Trying to override handler %d(0x%.x) for %v version",
			serverId, serverId, version)
	}
	gHandlers[key] = ldr
}

var gTagHandlers map[uint16]FileLoader = make(map[uint16]FileLoader, 0)

func SetTagHandler(tag uint16, ldr FileLoader) {
	if _, ok := gTagHandlers[tag]; ok {
		log.Panicf("Trying to override tag handler %d(0x%.x)", tag, tag)
	}
	gTagHandlers[tag] = ldr
}

type NodeId int
type TagId int

const NODE_INVALID = -1

type Wad struct {
	Source utils.ResourceSource `json:"-"`
	Tags   []Tag
	Nodes  []*Node
	Roots  []NodeId

	entityContext entitycontext.EntityLevelContext

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

	DebugPos uint32
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
	} else if n.Tag.Tag == GetServerInstanceTag() {
		if n.Tag.Data != nil && len(n.Tag.Data) >= 4 {
			serverId = binary.LittleEndian.Uint32(n.Tag.Data)
			if han, ex := gHandlers[(uint64(config.GetGOWVersion())<<32)|uint64(serverId)]; ex {
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
	copy(buf[8:32], utils.StringToBytesBuffer(t.Name, 24, false))
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
	w.Tags = make([]Tag, 0)
	w.HeapSizes = make(map[string]uint32)
	var buf [WAD_ITEM_SIZE]byte
	pos := int64(0)

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
		t.DebugPos = uint32(pos)

		if isZeroSizedTag(&t) {
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

		if pos, err = r.Seek(0, os.SEEK_CUR); err == nil {
			pos = int64(alignToWadTag(int(pos)))
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

	switch config.GetGOWVersion() {
	case config.GOW1:
		for id := range w.Tags {
			if err := w.gow1parseTag(&w.Tags[id], &currentNode, &newGroupTag, addNode); err != nil {
				return fmt.Errorf("Error parsing gow1 tag: %v", err)
			}
		}
	case config.GOW2:
		for id := range w.Tags {
			if err := w.gow2parseTag(&w.Tags[id], &currentNode, &newGroupTag, addNode); err != nil {
				return fmt.Errorf("Error parsing gow2 tag: %v", err)
			}
		}
	default:
		panic("not implemented")
	}

	if len(w.Tags) == 0 {
		return fmt.Errorf("Empty wad file")
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

func (w *Wad) GenerateName(prefix string) string {
	// generates name by first free hex suffix
	// l - hex suffix byes count
	for l := 1; ; l += 1 {
		lenlimit := 24 - l*2
		if lenlimit < len(prefix) {
			prefix = prefix[:24-l*2] // get space for hex number
		}

		for i := 0; i < 0x100*l; i++ {
			name := fmt.Sprintf("%s%x", prefix, i)
			if w.GetTagByName(name, 0, true) == nil {
				return name
			}
		}
	}
}

func (w *Wad) GetNodeById(nodeId NodeId) *Node {
	if nodeId == NODE_INVALID {
		return nil
	}
	n := w.Nodes[nodeId]

	if n.Tag.Tag == GetServerInstanceTag() {
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

func alignToWadTag(pos int) int {
	return ((pos + 15) / 16) * 16
}

func (w *Wad) Save(tags []Tag) error {
	var buf bytes.Buffer

	for _, t := range tags {
		if isZeroSizedTag(&t) {
			t.Size = w.HeapSizes[t.Name]
		} else {
			t.Size = uint32(len(t.Data))
		}

		buf.Write(MarshalTag(&t))
		if t.Data != nil {
			buf.Write(t.Data)
			buf.Write(make([]byte, alignToWadTag(buf.Len())-buf.Len()))
		}
	}

	w.flushCache()
	// sanity check for not corrupting wad and also update wad structure to collect changes
	if err := w.loadTags(io.NewSectionReader(bytes.NewReader(buf.Bytes()), 0, int64(buf.Len()))); err != nil {
		return fmt.Errorf("Error when perfoming reload sanity check: %v", err)
	}
	if err := w.parseTags(); err != nil {
		return fmt.Errorf("Error when parsing tags: %v", err)
	}

	return w.Source.Save(io.NewSectionReader(bytes.NewReader(buf.Bytes()), 0, int64(buf.Len())))
}

func (w *Wad) InsertNewTags(insertAfterId TagId, newTags []Tag) error {
	updatedTagsArray := append(w.Tags[:insertAfterId], append(newTags, w.Tags[insertAfterId:]...)...)
	return w.Save(updatedTagsArray)
}

func (w *Wad) UpdateTagInfo(updateTags map[TagId]Tag) error {
	for i, newTag := range updateTags {
		t := &w.Tags[i]
		log.Printf("Updating tag %x-%s to %x-%s", t.Id, t.Name, newTag.Id, newTag.Name)
		w.Tags[i] = newTag
	}
	return w.Save(w.Tags)
}

func (w *Wad) UpdateTagsData(updateData map[TagId][]byte) error {
	for i, newData := range updateData {
		t := &w.Tags[i]
		log.Println("Changing size at ", t.Name, " from ", t.Size, " to ", len(newData))
		t.Data = newData
		t.Size = uint32(len(newData))
	}
	return w.Save(w.Tags)
}

func (w *Wad) flushCache() {
	w.Tags = nil
	w.Nodes = nil
	w.Roots = nil
}

func (w *Wad) GetEntityContext() *entitycontext.EntityLevelContext {
	return &w.entityContext
}

func NewWad(r io.ReadSeeker, rsrc utils.ResourceSource) (*Wad, error) {
	w := &Wad{
		Source:        rsrc,
		entityContext: entitycontext.NewContext(),
	}

	if err := w.loadTags(r); err != nil {
		return nil, fmt.Errorf("Error when loading tags: %v", err)
	}

	if err := w.parseTags(); err != nil {
		return nil, fmt.Errorf("Error when parsing tags: %v", err)
	}

	if config.GetGOWVersion() == config.GOW1 {
		// load scripts so we have filled variables
		for _, n := range w.Nodes {
			if len(n.Tag.Data) > 40 && binary.LittleEndian.Uint32(n.Tag.Data) == 0x00010004 {
				if _, _, err := w.GetInstanceFromNode(n.Id); err != nil {
					log.Printf("[levelinit] Failed to load script %q: %v", n.Tag.Name, err)
				}
				n.Cache = nil
			}
		}
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
	pack.SetHandler(".wad_ps3", func(p utils.ResourceSource, r *io.SectionReader) (interface{}, error) {
		return NewWad(r, p)
	})
	pack.SetHandler(".wad_psp2", func(p utils.ResourceSource, r *io.SectionReader) (interface{}, error) {
		return NewWad(r, p)
	})
}
