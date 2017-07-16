package flp

import (
	"bytes"
	"io"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

const FLP_MAGIC = 0x21

type FLP struct {
}

func NewFromData(r io.Reader) (*FLP, error) {
	// 0..0x60- headeR?

	// + 0x60 - something start

	return &FLP{}, nil
}

func (flp *FLP) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return flp, nil
}

func init() {
	wad.SetHandler(FLP_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(bytes.NewReader(wrsrc.Tag.Data))
	})
}
