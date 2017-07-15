package flp

import (
	"io"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

const FLP_MAGIC = 0x21

type FLP struct {
}

func NewFromData(r *io.SectionReader) (*FLP, error) {
	// 0..0x60- headeR?

	// + 0x60 - something start

	return &FLP{}, nil
}

func (flp *FLP) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	return flp, nil
}

func init() {
	wad.SetHandler(FLP_MAGIC, func(w *wad.Wad, node *wad.WadNode, r *io.SectionReader) (wad.File, error) {
		return NewFromData(r)
	})
}
