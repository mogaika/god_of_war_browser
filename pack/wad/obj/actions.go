package obj

import (
	"archive/zip"
	"bytes"
	"log"
	"net/http"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (obj *Object) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "fbx":
		var buf bytes.Buffer
		log.Printf("Error when exporting obj: %v", obj.ExportFbxDefault(wrsrc).ExportZip(&buf, wrsrc.Tag.Name+".fbx"))
		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Tag.Name+".zip")
	case "zip":
		var buf, objBuf, mtlBuf bytes.Buffer

		z := zip.NewWriter(&buf)

		textures, err := obj.ExportObj(wrsrc, wrsrc.Name()+".mtl", &objBuf, &mtlBuf)
		if err != nil {
			log.Println("exporterr", err)
		}

		wObj, err := z.Create(wrsrc.Name() + ".obj")
		if err != nil {
			log.Println("objerr", err)
		}
		wObj.Write(objBuf.Bytes())

		wMtl, err := z.Create(wrsrc.Name() + ".mtl")
		if err != nil {
			log.Println("mtlerr", err)
		}
		wMtl.Write(mtlBuf.Bytes())

		for tname, t := range textures {
			wTxr, err := z.Create(tname + ".png")
			if err != nil {
				log.Println("txrerr", tname, err)
			}
			wTxr.Write(t)
		}

		if err := z.Close(); err != nil {
			log.Println("zcloseerr", err)
		}

		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Name()+".zip")
	}
}
