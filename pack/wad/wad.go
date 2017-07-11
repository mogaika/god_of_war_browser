package wad

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/utils"
)

const WAD_ITEM_SIZE = 0x20

type File interface {
	Marshal(wad *Wad, node *WadNode) (interface{}, error)
}

type FileLoader func(wad *Wad, node *WadNode, r io.ReaderAt) (File, error)

var cacheHandlers map[uint32]FileLoader = make(map[uint32]FileLoader, 0)
var cacheTagHandlers map[uint16]FileLoader = make(map[uint16]FileLoader, 0)

func SetHandler(format uint32, ldr FileLoader) {
	cacheHandlers[format] = ldr
}

func SetTagHandler(tag uint16, ldr FileLoader) {
	cacheTagHandlers[tag] = ldr
}

type Wad struct {
	Name   string
	Reader io.ReaderAt `json:"-"`
	Nodes  []*WadNode
	Roots  []int
}

func (wad *Wad) Node(id int) *WadNode {
	if id > len(wad.Nodes) || id < 0 {
		return nil
	} else {
		nd := wad.Nodes[id]
		return nd
	}
}

func (link *WadNode) ResolveLink() *WadNode {
	for link != nil && link.IsLink {
		link = link.FindNode(link.Name)
	}
	return link
}

func (wad *Wad) Get(id int) (File, error) {
	node := wad.Node(id).ResolveLink()

	if node == nil {
		return nil, errors.New("Node not found")
	}

	if node.Cache != nil {
		return node.Cache, nil
	}

	evaulateHandler := func(han FileLoader) (File, error) {
		rdr, err := wad.GetFileReader(node.Id)
		if err != nil {
			return nil, fmt.Errorf("Error getting wad '%s' node %d(%s)reader: %v", wad.Name, node.Id, node.Name, err)
		}

		cache, err := han(wad, node, rdr)
		if err == nil {
			node.Cache = cache
		}
		return cache, err
	}

	if han, ex := cacheTagHandlers[node.Tag]; ex {
		return evaulateHandler(han)
	} else if han, ex := cacheHandlers[node.Format]; ex {
		return evaulateHandler(han)
	} else {
		return nil, utils.ErrHandlerNotFound
	}
}

type WadNode struct {
	Id       int
	Name     string // can be empty
	IsLink   bool
	Parent   int
	SubNodes []int
	Flags    uint16
	Wad      *Wad `json:"-"`
	Size     uint32
	Start    int64
	Tag      uint16

	Cached bool `json:"-"`
	Cache  File `json:"-"`

	Format uint32 // first 4 bytes of data
}

func (wad *Wad) NewNode(name string, isLink bool, parent int, flags uint16) *WadNode {
	node := &WadNode{
		Id:     len(wad.Nodes),
		Name:   name,
		IsLink: isLink,
		Parent: parent,
		Wad:    wad,
		Flags:  flags,
	}

	wad.Nodes = append(wad.Nodes, node)
	if parent >= 0 {
		pnode := wad.Node(parent)
		pnode.SubNodes = append(pnode.SubNodes, node.Id)
	} else {
		wad.Roots = append(wad.Roots, node.Id)
	}
	return node
}

func (wad *Wad) FindNode(name string, parent int, end_at int) *WadNode {
	var result *WadNode = nil
	if parent < 0 {
		for _, n := range wad.Roots {
			if n >= end_at {
				return result
			}
			nd := wad.Node(n)
			if nd.Name == name {
				result = nd
			}
		}
		return result
	} else {
		if wad.Nodes != nil {
			for _, n := range wad.Node(parent).SubNodes {
				if n >= end_at {
					if result != nil {
						return result
					} else {
						break
					}
				}
				nd := wad.Node(n)
				if nd.Name == name {
					result = nd
				}
			}
		}
		return wad.FindNode(name, wad.Node(parent).Parent, parent)
	}
}

func (wad *Wad) GetFileReader(id int) (*io.SectionReader, error) {
	node := wad.Node(id)
	return io.NewSectionReader(wad.Reader, node.Start, int64(node.Size)), nil
}

func (node *WadNode) FindNode(name string) *WadNode {
	return node.Wad.FindNode(name, node.Parent, node.Id)
}

func NewWad(r io.ReaderAt, wadName string) (*Wad, error) {
	wad := &Wad{
		Reader: r,
		Name:   wadName,
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Wad parsing error: %v", err)
		}
	}()

	item := make([]byte, WAD_ITEM_SIZE)

	newGroupTag := false
	currentNode := -1

	pos := int64(0)
	for {
		_, err := r.ReadAt(item, pos)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, fmt.Errorf("Error reading from wad: %v", err)
			}
		}

		tag := binary.LittleEndian.Uint16(item[0:2])
		flags := binary.LittleEndian.Uint16(item[2:4])
		size := binary.LittleEndian.Uint32(item[4:8])
		name := utils.BytesToString(item[8:32])

		nd := wad.FindNode(name, currentNode, len(wad.Nodes))

		addNode := func(isLink bool, hasFormat bool) *WadNode {
			data_pos := pos + WAD_ITEM_SIZE
			node := wad.NewNode(name, isLink, currentNode, flags)
			if !isLink {
				if hasFormat {
					var bfmt [4]byte
					if _, err := r.ReadAt(bfmt[:], data_pos); err != nil {
						panic(err)
					}
					node.Format = binary.LittleEndian.Uint32(bfmt[0:4])
				}
			}
			node.Size = size
			node.Start = data_pos
			node.Tag = tag
			return node
		}

		switch tag {
		case 0x1e: // file data packet
			// Tell file server (server determined by first uint16)
			// that new file is loaded
			// if name start with space, then name ignored (unnamed)
			// overwrite previous instance with same name
			if nd != nil && nd.Parent == currentNode {
				log.Printf("Finded copy of %s->%d", nd.Name, nd.Id)
			}

			// size cannot be 0, because game store server id in first uint16
			node := addNode(size == 0, true)

			if newGroupTag {
				newGroupTag = false
				currentNode = node.Id
			}
		case 0x28: // file data group start
			newGroupTag = true // same realization in game
		case 0x32: // file data group end
			// Tell server about group ended
			newGroupTag = false
			if currentNode < 0 {
				return nil, errors.New("Trying to end not started group")
			} else {
				currentNode = wad.Nodes[currentNode].Parent
			}
		case 0x18: // entity count
			// Game also adding empty named node to nodedirectory
			log.Printf("%s Entity count: %v", wadName, size)
			size = 0
		case 0x006e: // MC_DATA   < R_PERM.WAD
			// Just add node to nodedirectory
			addNode(false, false)
		case 0x006f: // MC_ICON   < R_PERM.WAD
			// Like 0x006e, but also store size of data
			addNode(false, false)
		case 0x0070: // MSH_BDepoly6Shape
			// Add node to nodedirectory only if
			// another node with this name not exists
			if nd == nil {
				addNode(false, false)
			}
		case 0x0071: // TWK_Cloth_195
			// Tweaks affect VFS of game
			// AI logics, animation
			// exec twk asap?
			addNode(false, false)
		case 0x0072: // TWK_CombatFile_328
			// Affect VFS too
			// store twk in mem?
			addNode(false, false)
		case 0x01f4: // RSRCS
			// probably affect WadReader
			// (internally transformed to R_RSRCS)
			addNode(false, false)
		case 0x029a: // file data start
			// synonyms - 0x50, 0x309
			// PopBatchServerStack of server from first uint16
		case 0x0378: // file header start
			// create new memory namespace and push to memorystack
			// create new nodedirectory and push to nodestack
			// data loading init
		case 0x03e7: // file header pop heap
			// data loading structs cleanup
		default:
			log.Printf("unknown wad tag %.4x size %.8x name %s at %v", tag, size, name, pos)
			//return nil, fmt.Errorf("unknown tag")
		}

		off := (size + 15) & (15 ^ math.MaxUint32)
		pos = int64(off) + pos + 0x20
	}

	return wad, nil
}

func init() {
	pack.SetHandler(".WAD", func(p *pack.Pack, pf *pack.PackFile, r io.ReaderAt) (interface{}, error) {
		return NewWad(r, pf.Name)
	})
}
