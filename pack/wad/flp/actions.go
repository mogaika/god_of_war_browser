package flp

import (
	"archive/zip"
	"log"
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

		scale := float32(1.0)
		if strScale := r.FormValue("scale"); len(strScale) > 0 {
			possibleScale, err := strconv.ParseFloat(strScale, 32)
			if err != nil {
				log.Printf("Error parsing scale param: %v", err)
			} else {
				scale = float32(possibleScale)
				log.Printf("Used scale %v", scale)
			}
		} else {
			log.Println("Scale parameter not provided")
		}

		zr, err := zip.NewReader(fZip, hZip.Size)
		if err != nil {
			webutils.WriteError(w, err)
			return
		}
		if err := f.actionImportBmFont(wrsrc, zr, scale); err != nil {
			webutils.WriteError(w, err)
		}
	}
}
