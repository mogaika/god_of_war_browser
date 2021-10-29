package scr

import (
	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/scr/store"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/scr/targets/entity"
	"github.com/mogaika/god_of_war_browser/utils"
)

const SCRIPT_MAGIC = 0x00010004
const HEADER_SIZE = 0x24

type ScriptParams struct {
	TargetName string
	Data       interface{}
}

func NewFromData(buf []byte, wrsrc *wad.WadNodeRsrc) (*ScriptParams, error) {
	sp := &ScriptParams{
		TargetName: utils.BytesToString(buf[4:20]),
	}

	var err error
	if loader := store.GetScriptLoader(sp.TargetName); loader != nil {
		sp.Data, err = loader(buf[HEADER_SIZE:], wrsrc)
	}

	return sp, err
}

func (sp *ScriptParams) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return sp, nil
}

func init() {
	wad.SetHandler(config.GOW1, SCRIPT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data, wrsrc)
	})
	wad.SetHandler(config.GOW2, SCRIPT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data, wrsrc)
	})
}
