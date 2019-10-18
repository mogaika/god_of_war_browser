package collision

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

func NewFromData(f io.ReaderAt, wrtr io.Writer) (c *Collision, err error) {
	var buf [16]byte
	if _, err := f.ReadAt(buf[:], 0); err != nil {
		return nil, err
	}

	c = &Collision{
		Magic: binary.LittleEndian.Uint32(buf[0:4]),
	}

	for _, sh := range []struct {
		Offset int
		Name   string
	}{{8, "BallHull"}, {4, "SheetHdr"}, {4, "mCDbgHdr"}} {
		if utils.BytesToString(buf[sh.Offset:sh.Offset+8]) == sh.Name {
			c.ShapeName = sh.Name
			break
		}
	}

	switch c.ShapeName {
	case "SheetHdr":
		c.Shape, err = NewRibSheet(f, wrtr)
	case "BallHull":
		c.Shape, err = NewBallHull(f, wrtr)
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

		return NewFromData(bytes.NewReader(wrsrc.Tag.Data), f)
	})
}
