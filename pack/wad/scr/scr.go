package scr

import (
	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/scr/store"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/scr/targets"
	"github.com/mogaika/god_of_war_browser/utils"
)

const SCRIPT_MAGIC = 0x00010004
const HEADER_SIZE = 0x24

type ScriptParams struct {
	TargetName string
	Data       interface{}
}

func NewFromData(buf []byte) (*ScriptParams, error) {
	sp := &ScriptParams{
		TargetName: utils.BytesToString(buf[4:20]),
	}

	if loader := store.GetScriptLoader(sp.TargetName); loader != nil {
		sp.Data = loader(buf[HEADER_SIZE:])
	}

	return sp, nil
}

func (sp *ScriptParams) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return sp, nil
}

func init() {
	wad.SetHandler(config.GOW1, SCRIPT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
	wad.SetHandler(config.GOW2, SCRIPT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
