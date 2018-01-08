package flp

import (
	"archive/zip"
	"net/http"
	"strconv"
	"strings"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (f *FLP) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "staticlabels":
		if strings.ToUpper(r.Method) == "POST" {
			if err := r.ParseForm(); err != nil {
				webutils.WriteError(w, err)
			}

			id, err := strconv.Atoi(r.PostFormValue("id"))
			if err != nil {
				webutils.WriteError(w, err)
			}

			if err := f.StaticLabels[id].ParseJson([]byte(r.PostFormValue("sl"))); err != nil {
				webutils.WriteError(w, err)
			}

			wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
				wrsrc.Tag.Id: f.marshalBufferWithHeader().Bytes(),
			})
		}
	case "importbmfont":
		fZip, hZip, err := r.FormFile("data")
		if err != nil {
			webutils.WriteError(w, err)
			return
		}
		defer fZip.Close()

		zr, err := zip.NewReader(fZip, hZip.Size)
		if err != nil {
			webutils.WriteError(w, err)
			return
		}
		if err := f.ImportBmpFont(wrsrc, zr); err != nil {
			webutils.WriteError(w, err)
		}
	}
}
