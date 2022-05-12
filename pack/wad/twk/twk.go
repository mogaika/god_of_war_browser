package twk

import (
	"encoding/binary"
	"io"
	"strings"

	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/twk/twktree"
	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	TWK_Tag = 0x71
)

type TWK struct {
	Name                  string
	MagicHeaderPresened   bool
	HeaderStrangeMagicUid uint32

	Tree         *twktree.VFSNode
	AbstractTree *twktree.VFSAbstractNode
}

var twkBufSizes = [8]int{4, 0x20, 0x40, 0x100, 0x200, 0x400, 0x800, 0x1000}

func (t *TWK) getPathNode(root *twktree.VFSNode, path string) *twktree.VFSNode {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	node := root

	// log.Printf("Path parts %+#v", parts)
	for _, pathPart := range parts {
		found := false
		for _, subNode := range node.Fields {
			if subNode.Name == pathPart {
				node = subNode
				found = true
				break
			}
		}
		if !found {
			newNode := twktree.NewVFSNode(pathPart)
			node.Fields = append(node.Fields, newNode)
			node = newNode
		}
	}

	return node
}

func (t *TWK) Parse(bsdata *utils.BufStack) error {
	createdNode := false

	t.Tree = twktree.NewVFSNode("/")
	directory := t.Tree
	dirStack := make([]*twktree.VFSNode, 0)

	// path := "unknown"
cmdLoop:
	for handled := true; handled; {
		cmd := bsdata.ReadByte()
		cmdFlags := cmd & 0xF0
		// log.Printf("cmd 0x%.2x flags 0x%.2x", cmd, cmdFlags)

		switch cmdFlags {
		case 0:
			//log.Printf(" | end ========")
			if len(dirStack) != 0 {
				return errors.Errorf("Non empty directories stack on end: %v", dirStack)
			}
			break cmdLoop
		case 0x80:
			if t.Name != "" {
				return errors.Errorf("Multi root paths in twk")
			}
			t.Name = bsdata.ReadZString(0x100)

			// # label 0x180204
			//log.Printf(" | vfs goto /")
			if t.Name == "" {
				return errors.Errorf("Empty twk root")
			}
			//log.Printf(" | vfs create twk '%s'", t.Name)
			directory = t.getPathNode(t.Tree, t.Name)

			// path = t.Name
			createdNode = true
		case 0x10:
			subPath := bsdata.ReadZString(0x100)
			// path += "/" + subPath
			//log.Printf(" | vfs goto %q", subPath)

			dirStack = append(dirStack, directory)

			directory = t.getPathNode(directory, subPath)
		case 0x40:
			// parts := strings.Split(path, "/")
			// path = strings.Join(parts[:len(parts)-1], "/")

			directory = dirStack[len(dirStack)-1]
			dirStack = dirStack[:len(dirStack)-1]

			//log.Printf(" | vfs goto ..")
		case 0x20:
			nameHash := bsdata.ReadLU32()
			bufSizeOrIdk := twkBufSizes[cmd&0xf]
			cmdData := bsdata.Read(bufSizeOrIdk)

			unhashedName := utils.GameStringUnhashNodes(nameHash)

			node := twktree.NewVFSNode(unhashedName)
			node.Value = cmdData
			directory.Fields = append(directory.Fields, node)

			if !createdNode {
				panic("data without node")
			}
		default:
			handled = false
		}

	}

	return nil
}

func (t *TWK) Produce(w io.Writer) error {
	le := binary.LittleEndian
	if t.MagicHeaderPresened {
		binary.Write(w, le, uint32(0xfedcba98))
		binary.Write(w, le, uint32(t.HeaderStrangeMagicUid))
	}

	binary.Write(w, le, uint8(0x80))

	w.Write(utils.StringToBytes(t.Name, true))

	directory := t.getPathNode(t.Tree, t.Name)
	if directory == nil {
		return errors.Errorf("Dir should contain name entry")
	}

	var encodeDirectory func(d *twktree.VFSNode)
	encodeDirectory = func(d *twktree.VFSNode) {
		for _, field := range d.Fields {
			if len(field.Value) != 0 {
				var sizeFlag uint8 = 0xf
				for i, size := range twkBufSizes {
					if len(field.Value) > size {
						sizeFlag = uint8(i)
						continue
					}
					sizeFlag = uint8(i)
					break
				}

				binary.Write(w, le, uint8(0x20|sizeFlag))
				binary.Write(w, le, utils.GameStringHashNodes(field.Name, 0))
				fieldBuf := make([]byte, twkBufSizes[sizeFlag])
				copy(fieldBuf, field.Value)
				w.Write(fieldBuf)
			} else {
				binary.Write(w, le, uint8(0x10))
				w.Write(utils.StringToBytes(field.Name, true))
				encodeDirectory(field)
				binary.Write(w, le, uint8(0x40))

			}
		}
	}

	encodeDirectory(directory)

	binary.Write(w, le, uint8(0))

	return nil
}

func NewTwkFromData(twkrootbs *utils.BufStack) (*TWK, error) {
	t := &TWK{}

	var bsdata *utils.BufStack

	if twkrootbs.LU32(0) == 0xfedcba98 { // -0x1234568
		t.MagicHeaderPresened = true
		t.HeaderStrangeMagicUid = twkrootbs.LU32(4)
		bsdata = twkrootbs.SubBuf("twkdata", 0x8).Expand()
	} else {
		bsdata = twkrootbs.SubBuf("twkdata", 0).Expand()
	}

	err := t.Parse(bsdata)

	// utils.LogDump(twk)
	return t, err
}

func (twk *TWK) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return twk, nil
}

func init() {
	wad.SetTagHandler(TWK_Tag, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewTwkFromData(utils.NewBufStack("twk", wrsrc.Tag.Data))
	})
}
