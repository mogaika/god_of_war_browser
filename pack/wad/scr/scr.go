package scr

import (
	"io"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/scr/store"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/scr/targets"
	"github.com/mogaika/god_of_war_browser/utils"
)

const SCRIPT_MAGIC = 0x00010004
const HEADER_SIZE = 0x20

type ScriptParams struct {
	TargetName string
	Data       interface{}
}

func NewFromData(r io.ReaderAt, size uint32) (*ScriptParams, error) {
	var data [0x24]byte
	_, err := r.ReadAt(data[:], 0)
	if err != nil {
		return nil, err
	}

	sp := &ScriptParams{
		TargetName: utils.BytesToString(data[4:20]),
	}

	if loader := store.GetScriptLoader(sp.TargetName); loader != nil {
		buf := make([]byte, size-0x24)
		if _, err := r.ReadAt(buf, 0x24); err != nil {
			panic(err)
		}
		sp.Data = loader(buf)
	}

	return sp, nil
}

func (sp *ScriptParams) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	return sp, nil
}

func init() {
	wad.SetHandler(SCRIPT_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		return NewFromData(r, node.Size)
	})
}
