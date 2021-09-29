package collision

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const COLLISION_MAGIC = 0x00000011

type Collision struct {
	Magic     uint32
	ShapeName string
	FileSize  uint32
	Shape     interface{}
}

func NewFromData(bs *utils.BufStack, wrtr io.Writer) (c *Collision, err error) {
	head := bs.Raw()[:16]

	c = &Collision{
		Magic: bs.LU32(0),
	}

	for _, sh := range []struct {
		Offset int
		Name   string
	}{{8, "BallHull"}, {4, "SheetHdr"}, {4, "mCDbgHdr"}} {
		if utils.BytesToString(head[sh.Offset:sh.Offset+8]) == sh.Name {
			c.ShapeName = sh.Name
			break
		}
	}

	bs.SetName(c.ShapeName)

	switch c.ShapeName {
	case "SheetHdr":
		var rib *ShapeRibSheet
		if rib, err = NewRibSheet(bs, wrtr); err == nil {
			c.Shape = rib
			if _, err := NewRibSheet(utils.NewBufStack("heh", rib.Marshal()), ioutil.Discard); err != nil {
				log.Printf("Failed to reopen remarshaled file: %v", err)
			}
		}
	case "BallHull":
		c.Shape, err = NewBallHull(bs, wrtr)
	default:
		return nil, fmt.Errorf("Unknown enz shape type %s", c.ShapeName)
	}

	return
}

func (c *Collision) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return c, nil
}

func init() {
	wad.SetHandler(config.GOW1, COLLISION_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		fpath := filepath.Join("logs", wrsrc.Wad.Name(), fmt.Sprintf("%.4d-%s.enz.obj", wrsrc.Tag.Id, wrsrc.Tag.Name))
		os.MkdirAll(filepath.Dir(fpath), 0777)
		f, _ := os.Create(fpath)
		defer f.Close()

		// f := ioutil.Discard

		bs := utils.NewBufStack("collision", wrsrc.Tag.Data)

		return NewFromData(bs, f)
	})
}
