package mesh

import (
	"bytes"
	"log"
	"net/http"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (mesh *Mesh) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "obj":
		var buf bytes.Buffer
		log.Printf("Error when exporting mesh: %v", mesh.ExportObj(&buf, nil, nil))
		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Tag.Name+".obj")
	case "fbx":
		var buf bytes.Buffer
		log.Printf("Error when exporting mesh: %v", mesh.ExportFbxDefault(wrsrc).Export(&buf))
		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Tag.Name+".fbx")
	}
}
