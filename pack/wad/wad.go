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

type FileLoader func(wad *Wad, node *WadNode, r io.ReaderAt) (interface{}, error)

var cacheHandlers map[uint32]FileLoader = make(map[uint32]FileLoader, 0)

func SetHandler(format uint32, ldr FileLoader) {
	cacheHandlers[format] = ldr
}

type Wad struct {
	Name   string
	Reader io.ReaderAt `json:"-"`
	Nodes  []*WadNode
	Roots  []int
}

func (wad *Wad) Get(id int) (interface{}, error) {
	node := wad.Nodes[id]
	if node.Cache != nil {
		return node.Cache, nil
	}

	if han, ex := cacheHandlers[node.Format]; ex {
		rdr, err := wad.GetFileReader(id)
		if err != nil {
			return nil, fmt.Errorf("Error getting wad '%s' node %d(%s)reader: %v", wad.Name, id, node.Name, err)
		}
		cache, err := han(wad, node, rdr)
		node.Cache = cache
		return cache, err
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
	Wad      *Wad `json:"-"`
	Size     uint32
	Start    int64

	Cached bool        `json:"-"`
	Cache  interface{} `json:"-"`

	Format uint32 // first 4 bytes of data
}

func (wad *Wad) NewNode(name string, isLink bool, parent int) *WadNode {
	node := &WadNode{
		Id:     len(wad.Nodes),
		Name:   name,
		IsLink: isLink,
		Parent: parent,
		Wad:    wad,
	}

	wad.Nodes = append(wad.Nodes, node)
	if parent >= 0 {
		pnode := wad.Nodes[parent]
		pnode.SubNodes = append(pnode.SubNodes, node.Id)
	} else {
		wad.Roots = append(wad.Roots, node.Id)
	}
	return node
}

func (wad *Wad) FindNode(name string, start int) *WadNode {
	if start < 0 {
		for _, n := range wad.Roots {
			if wad.Nodes[n].Name == name {
				return wad.Nodes[n]
			}
		}
		return nil
	} else {
		if wad.Nodes != nil {
			for _, n := range wad.Nodes[start].SubNodes {
				if wad.Nodes[n].Name == name {
					return wad.Nodes[n]
				}
			}
		}
		return wad.FindNode(name, wad.Nodes[start].Parent)
	}
}

func (wad *Wad) GetFileReader(id int) (*io.SectionReader, error) {
	node := wad.Nodes[id]
	return io.NewSectionReader(wad.Reader, node.Start, int64(node.Size)), nil
}

func NewWad(r io.ReaderAt, name string) (*Wad, error) {
	wad := &Wad{
		Reader: r,
		Name:   name,
	}

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
		size := binary.LittleEndian.Uint32(item[4:8])
		name := utils.BytesToString(item[8:32])

		switch tag {
		case 0x1e: // file data packet
			data_pos := pos + WAD_ITEM_SIZE
			var node *WadNode
			// minimal size of data == 4, for storing data format
			if size == 0 {
				node = wad.NewNode(name, true, currentNode)
			} else {
				node = wad.NewNode(name, false, currentNode)

				var bfmt [4]byte
				_, err := r.ReadAt(bfmt[:], data_pos)
				if err != nil {
					return nil, err
				}
				node.Format = binary.LittleEndian.Uint32(bfmt[0:4])
			}
			node.Size = size
			node.Start = data_pos

			if newGroupTag {
				newGroupTag = false
				currentNode = node.Id
			}
		case 0x28: // file data group start
			newGroupTag = true
		case 0x32: // file data group end
			newGroupTag = false
			if currentNode < 0 {
				return nil, errors.New("Trying to end not started group")
			} else {
				currentNode = wad.Nodes[currentNode].Parent
			}
		case 0x18: // entity count
			size = 0

		// TODO: use this tags
		case 0x006e: // MC_DATA   < R_PERM.WAD
		case 0x006f: // MC_ICON   < R_PERM.WAD
		case 0x0070: // MSH_BDepoly6Shape
		case 0x0071: // TWK_Cloth_195
		case 0x0072: // TWK_CombatFile_328
		case 0x01f4: // RSRCS
		case 0x0378: // file header start
		case 0x03e7: // file header pop heap
		case 0x029a: // file data start
		default:
			log.Printf("unknown wad tag %.4x size %.8x name %s", tag, size, name)
			return nil, fmt.Errorf("unknown tag")
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
