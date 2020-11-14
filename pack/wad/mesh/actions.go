package mesh

import (
	"log"
	"net/http"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (mesh *Mesh) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "fbx":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".fbx")
		if err := mesh.ExportFbxDefault(wrsrc).Write(w); err != nil {
			log.Printf("Error when exporting mesh as fbx: %v", err)
		}
	}
}
