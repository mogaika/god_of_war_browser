package flp

import (
	"fmt"
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
		fZip, _, err := r.FormFile("bmfont")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		defer fZip.Close()

		//zip.NewReader(fZip, fZip.)
	}
}
