package collision

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_tools/utils"
)

const ENZ_MAGIC = 0x00000011
const ENZ_HEADER_SIZE = 0x30

type ENZ struct {
	Magic     uint32
	ShapeName string
	FileSize  uint32
	Shape     interface{}
}

func NewFromData(f io.ReaderAt, wrtr io.Writer) (enz *ENZ, err error) {
	buf := make([]byte, ENZ_HEADER_SIZE)
	if _, err := f.ReadAt(buf, 0); err != nil {
		return nil, err
	}

	enz = &ENZ{
		Magic:     binary.LittleEndian.Uint32(buf[0:4]),
		ShapeName: utils.BytesToString(buf[4:12]),
		FileSize:  binary.LittleEndian.Uint32(buf[12:16]),
	}

	if enz.FileSize > 0x100000 {
		return nil, fmt.Errorf("Invalid cz file")
	}

	switch enz.ShapeName {
	case "SheetHdr":
		enz.Shape, err = NewRibSheet(f, wrtr)
	default:
		panic("Unknown enz shape type")
	}

	return
}

func (enz *ENZ) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	return enz, nil
}

func init() {
	wad.SetHandler(ENZ_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		fpath := filepath.Join("logs", w.Name, fmt.Sprintf("%.4d-%s.enz.obj", node.Id, node.Name))
		os.MkdirAll(filepath.Dir(fpath), 0777)
		f, _ := os.Create(fpath)
		defer f.Close()

		enz, err := NewFromData(r, f)
		return enz, err
	})
}
