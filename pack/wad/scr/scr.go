package scr

import (
	"encoding/binary"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/scr/store"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/scr/targets/entity"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/webutils"
)

const SCRIPT_MAGIC = 0x00010004
const HEADER_SIZE = 0x24

type ScriptParams struct {
	TargetName string
	Unk_0x1e   uint16
	Unk_0x1c   uint16
	Unk_0x20   uint16
	Unk_0x22   uint16
	Data       interface{}
}

func NewFromData(buf []byte, wrsrc *wad.WadNodeRsrc) (*ScriptParams, error) {
	sp := &ScriptParams{
		TargetName: utils.BytesToString(buf[0x4:0x14]),
		Unk_0x1c:   binary.LittleEndian.Uint16(buf[0x1c:]),
		Unk_0x1e:   binary.LittleEndian.Uint16(buf[0x1e:]),
		Unk_0x20:   binary.LittleEndian.Uint16(buf[0x20:]),
		Unk_0x22:   binary.LittleEndian.Uint16(buf[0x22:]),
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

func (sp *ScriptParams) MarshalBufHeader() []byte {
	result := make([]byte, HEADER_SIZE)
	binary.LittleEndian.PutUint32(result[0x00:], SCRIPT_MAGIC)
	copy(result[0x4:0x14], utils.StringToBytesBuffer(sp.TargetName, 0x10, true))
	binary.LittleEndian.PutUint16(result[0x1c:], sp.Unk_0x1c)
	binary.LittleEndian.PutUint16(result[0x1e:], sp.Unk_0x1e)
	binary.LittleEndian.PutUint16(result[0x20:], sp.Unk_0x20)
	binary.LittleEndian.PutUint16(result[0x22:], sp.Unk_0x22)
	return result
}

func (sp *ScriptParams) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "dataasjson":
		webutils.WriteJsonFile(w, sp.Data, wrsrc.Name())
	case "datafromjson":
		ldr, ok := sp.Data.(store.ScriptContentJsonLoader)
		if !ok {
			webutils.WriteError(w, errors.Errorf("Unsupported loader for %T", sp.Data))
			return
		}

		jsonData, err := webutils.ReadFile(r, "data")
		if err != nil {
			webutils.WriteError(w, errors.Wrap(err, "Failed to read file"))
			return
		}

		scrData, err := ldr.FromJSON(wrsrc, jsonData)
		if err != nil {
			webutils.WriteError(w, errors.Wrap(err, "Failed to parse json"))
		}

		if err := wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
			wrsrc.Tag.Id: append(sp.MarshalBufHeader(), scrData...),
		}); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to write tag"))
			return
		}
	}
}

func init() {
	wad.SetServerHandler(config.GOW1, SCRIPT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data, wrsrc)
	})
	wad.SetServerHandler(config.GOW2, SCRIPT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data, wrsrc)
	})
}
