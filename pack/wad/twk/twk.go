package twk

import (
	"encoding/hex"
	"log"

	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	TWK_Tag = 0x71
)

type Value struct {
	Name   string
	Hex    string
	Offset int
}

type Directory struct {
	Name        string
	Values      []Value
	Directories []*Directory
}

type TWK struct {
	MagicHeaderPresened   bool
	HeaderStrangeMagicUid uint32
	Directory             `json:""`
}

var twkBufSizes = [8]int{4, 0x20, 0x40, 0x100, 0x200, 0x400, 0x800, 0x1000}

func (t *TWK) Parse(bsdata *utils.BufStack) error {
	createdNode := false

	var directory *Directory
	dirStack := make([]*Directory, 0)

	//path := "unknown"
	for handled := true; handled; {
		cmd := bsdata.ReadByte()
		cmdFlags := cmd & 0xF0
		// log.Printf("cmd 0x%.2x flags 0x%.2x", cmd, cmdFlags)

		switch cmdFlags {
		case 0:
			log.Printf(" | end ========")
			if len(dirStack) != 0 {
				return errors.Errorf("Non empty directories stack on end: %v", dirStack)
			}
			return nil
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
			directory = &t.Directory

			//path = t.Name
			createdNode = true
		case 0x10:
			subPath := bsdata.ReadZString(0x100)
			//path += "/" + subPath
			//log.Printf(" | vfs goto %q", subPath)

			newDir := &Directory{
				Name: subPath,
			}
			dirStack = append(dirStack, directory)
			directory.Directories = append(directory.Directories, newDir)
			directory = newDir
		case 0x40:
			//parts := strings.Split(path, "/")
			//path = strings.Join(parts[:len(parts)-1], "/")

			directory = dirStack[len(dirStack)-1]
			dirStack = dirStack[:len(dirStack)-1]

			//log.Printf(" | vfs goto ..")
		case 0x20:
			nameHash := bsdata.ReadLU32()
			offset := bsdata.AbsoluteOffset() + bsdata.Pos()
			bufSizeOrIdk := twkBufSizes[cmd&0xf]
			cmdData := bsdata.Read(bufSizeOrIdk)

			unhashedName := utils.GameStringUnhashNodes(nameHash)

			directory.Values = append(directory.Values, Value{
				Name:   unhashedName,
				Hex:    hex.EncodeToString(cmdData),
				Offset: offset,
			})

			if !createdNode {
				panic("data without node")
			}
			//log.Printf("  cmd hash %.8x %q size %.4x: %v",
			//	nameHash, utils.GameStringUnhashNodes(nameHash), bufSizeOrIdk, utils.DumpToOneLineString(cmdData))

			/*if bufSizeOrIdk == 4 {
				log.Printf("  hash(%.8x) path(%s/%q) value(%q) (0x%.8x) (%f)",
					nameHash, path, unhashedName, utils.DumpToOneLineString(cmdData),
					binary.LittleEndian.Uint32(cmdData), math.Float32frombits(binary.LittleEndian.Uint32(cmdData)))
			} else {
				log.Printf("  hash(%.8x) path(%s/%q) value(%q)",
					nameHash, path, unhashedName, utils.DumpToOneLineString(cmdData))
				if bufSizeOrIdk >= 0x40 {
					utils.LogDump(cmdData)
				}
			}*/
		default:
			handled = false
		}

	}

	return nil
}

func NewTwkFromData(twkrootbs *utils.BufStack) (*TWK, error) {
	twk := &TWK{}

	var bsdata *utils.BufStack

	if twkrootbs.LU32(0) == 0xfedcba98 { // -0x1234568
		twk.MagicHeaderPresened = true
		twk.HeaderStrangeMagicUid = twkrootbs.LU32(4)
		bsdata = twkrootbs.SubBuf("twkdata", 0x8).Expand()
	} else {
		bsdata = twkrootbs.SubBuf("twkdata", 0).Expand()
	}

	err := twk.Parse(bsdata)

	//utils.LogDump(twk)
	return twk, err
}

func (twk *TWK) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return twk, nil
}

func init() {
	wad.SetTagHandler(TWK_Tag, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewTwkFromData(utils.NewBufStack("twk", wrsrc.Tag.Data))
	})
}
