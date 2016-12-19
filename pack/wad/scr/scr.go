package scr

import (
	"io"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const SCRIPT_MAGIC = 0x00010004
const HEADER_SIZE = 0x20

type ScriptParams struct {
	TargetScript string
}

func NewFromData(r io.ReaderAt) (*ScriptParams, error) {
	var data [20]byte
	_, err := r.ReadAt(data[:], 0)
	if err != nil {
		return nil, err
	}
	return &ScriptParams{
		TargetScript: utils.BytesToString(data[4:20]),
	}, nil
}

func (sp *ScriptParams) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	return sp, nil
}

func init() {
	wad.SetHandler(SCRIPT_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		return NewFromData(r)
	})
}
