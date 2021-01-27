package twk

import (
	"encoding/binary"
	"log"
	"math"
	"strings"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	TWK_Tag = 0x71
)

type TWK struct {
	MagicHeaderPresened   bool
	HeaderStrangeMagicUid uint32
	Path                  string
}

func (t *TWK) Parse(bsdata *utils.BufStack) error {
	createdNode := false

	path := "unknown"
	for handled := true; handled; {
		cmd := bsdata.ReadByte()
		cmdFlags := cmd & 0xF0
		// log.Printf("cmd 0x%.2x flags 0x%.2x", cmd, cmdFlags)

		switch cmdFlags {
		case 0:
			//log.Printf(" | end ========")
			return nil
		case 0x80:
			t.Path = bsdata.ReadZString(0x100)
			// # label 0x180204
			// log.Printf(" | vfs goto /")
			path = "/"
			if t.Path != "" {
				log.Printf(" | vfs create twk '%s'", t.Path)
				path = t.Path
				createdNode = true
			}
		case 0x10:
			subPath := bsdata.ReadZString(0x100)
			path += "/" + subPath
			log.Printf(" | vfs goto %q", subPath)
		case 0x40:
			parts := strings.Split(path, "/")
			path = strings.Join(parts[:len(parts)-1], "/")
			log.Printf(" | vfs goto ..")
		case 0x20:
			nameHash := bsdata.ReadLU32()
			_ = nameHash
			bufSizeOrIdk := 0
			switch cmd & 0xf {
			case 0:
				bufSizeOrIdk = 4
			case 1:
				bufSizeOrIdk = 0x20
			case 2:
				bufSizeOrIdk = 0x40
			case 3:
				bufSizeOrIdk = 0x100
			case 4:
				bufSizeOrIdk = 0x200
			case 5:
				bufSizeOrIdk = 0x400
			case 6:
				bufSizeOrIdk = 0x800
			case 7:
				bufSizeOrIdk = 0x1000
			}
			cmdData := bsdata.Read(bufSizeOrIdk)

			if !createdNode {
				panic("data without node")
			}
			//log.Printf("  cmd hash %.8x %q size %.4x: %v",
			//	nameHash, utils.GameStringUnhashNodes(nameHash), bufSizeOrIdk, utils.DumpToOneLineString(cmdData))

			if bufSizeOrIdk == 4 {
				log.Printf("  hash(%.8x) path(%s/%q) value(%q) (0x%.8x) (%f)",
					nameHash, path, utils.GameStringUnhashNodes(nameHash), utils.DumpToOneLineString(cmdData),
					binary.LittleEndian.Uint32(cmdData), math.Float32frombits(binary.LittleEndian.Uint32(cmdData)))
			} else {
				log.Printf("  hash(%.8x) path(%s/%q) value(%q)",
					nameHash, path, utils.GameStringUnhashNodes(nameHash), utils.DumpToOneLineString(cmdData))
			}
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

	utils.LogDump(twk)
	log.Printf(" TWK BS TREE +++++++++\n%s", twkrootbs.StringTree())

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
