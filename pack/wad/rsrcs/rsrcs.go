package rsrcs

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const RSRCS_Tag = 500

type RSRCS struct {
	Wads []string
}

func NewRSRCSFromData(bs *utils.BufStack) (*RSRCS, error) {
	rsrcs := &RSRCS{
		Wads: make([]string, bs.Size()/24),
	}

	for i := range rsrcs.Wads {
		rsrcs.Wads[i] = bs.ReadStringBuffer(24)
	}

	return rsrcs, nil
}

func (rsrcs *RSRCS) MarshalData() *bytes.Buffer {
	var b bytes.Buffer
	for _, wad := range rsrcs.Wads {
		n, _ := b.WriteString(wad)
		b.Write(make([]byte, 24-n))
	}
	return &b
}

func (rsrcs *RSRCS) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return rsrcs, nil
}

func (rsrcs *RSRCS) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "update":
		rsrcs.Wads = make([]string, 0)

		for i := 0; i < 6; i++ {
			if wad := r.URL.Query().Get(fmt.Sprintf("wad%d", i)); wad != "" {
				rsrcs.Wads = append(rsrcs.Wads, wad)
			}
		}

		wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
			wrsrc.Tag.Id: rsrcs.MarshalData().Bytes(),
		})
	}
}

func init() {
	wad.SetTagHandler(RSRCS_Tag, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewRSRCSFromData(utils.NewBufStack("rsrcs", wrsrc.Tag.Data))
	})
}
