package flp

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (f *FLP) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "staticlabels":
		if strings.ToUpper(r.Method) == "POST" {
			if err := r.ParseForm(); err != nil {
				webutils.WriteError(w, errors.Wrapf(err, "Failed to parse form"))
				return
			}

			id, err := strconv.Atoi(r.PostFormValue("id"))
			if err != nil {
				webutils.WriteError(w, errors.Wrapf(err, "Failed to parse id"))
				return
			}

			if err := f.StaticLabels[id].ParseJson([]byte(r.PostFormValue("sl"))); err != nil {
				webutils.WriteError(w, errors.Wrapf(err, "Failed to load static label"))
				return
			}

			wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
				wrsrc.Tag.Id: f.marshalBufferWithHeader().Bytes(),
			})
		}
	case "importbmfont":
		fZip, hZip, err := r.FormFile("data")
		if err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to open file"))
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
			webutils.WriteError(w, errors.Wrapf(err, "Failed to open zip reader"))
			return
		}
		if err := f.actionImportBmFont(wrsrc, zr, scale); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to import font"))
			return
		}
	case "scriptstring":
		q := r.URL.Query()
		id, err := strconv.ParseInt(q.Get("id"), 10, 32)
		if err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to parse id"))
			return
		}

		decodedStr, err := base64.StdEncoding.DecodeString(q.Get("string"))
		if err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to decode base64"))
			return
		}

		f.scriptPushRefs[id].ChangeString(decodedStr)

		if err := wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
			wrsrc.Tag.Id: f.marshalBufferWithHeader().Bytes(),
		}); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to update tags data"))
			return
		}
	case "transform":
		if strings.ToUpper(r.Method) == "POST" {
			if err := r.ParseForm(); err != nil {
				webutils.WriteError(w, errors.Wrapf(err, "Failed to parse form"))
				return
			}

			id, err := strconv.Atoi(r.PostFormValue("id"))
			if err != nil {
				webutils.WriteError(w, errors.Wrapf(err, "Failed to get id"))
				return
			}

			if err := json.Unmarshal([]byte(r.PostFormValue("data")), &f.Transformations[id]); err != nil {
				webutils.WriteError(w, errors.Wrapf(err, "Failed to unmarshal"))
				return
			}

			wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
				wrsrc.Tag.Id: f.marshalBufferWithHeader().Bytes(),
			})
		}
	case "asjson":
		if data, err := json.MarshalIndent(f, "", "  "); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to unmarshal"))
		} else {
			webutils.WriteFile(w, bytes.NewReader(data), wrsrc.Name()+".json")
		}
	case "fromjson":
		fJson, _, err := r.FormFile("data")
		if err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to get file"))
			return
		}
		defer fJson.Close()

		data, err := ioutil.ReadAll(fJson)
		if err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to read"))
			return
		}

		newFlp := &FLP{}
		if err := json.Unmarshal(data, newFlp); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to unmarshal"))
			return
		}

		newFLPBuf := newFlp.marshalBufferWithHeader()
		// double check validity of file
		if _, err := NewFromData(newFLPBuf.Bytes()); err != nil {
			webutils.WriteError(w, errors.Wrapf(err, "Failed to check validity of file"))
			return
		}
	}
}
