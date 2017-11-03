package flp

import (
	"net/http"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (f *FLP) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "staticlabels":
		type StaticLabelAction struct {
			Fonts        []Font
			StaticLabels []StaticLabel
		}
		webutils.WriteJson(w, &StaticLabelAction{
			Fonts:        f.Fonts,
			StaticLabels: f.StaticLabels,
		})
	}
}
