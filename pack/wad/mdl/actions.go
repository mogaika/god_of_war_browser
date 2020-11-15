package mdl

import (
	"log"
	"net/http"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (mdl *Model) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "fbx":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".fbx")
		if err := mdl.ExportFbxDefault(wrsrc).Write(w); err != nil {
			log.Printf("Error when exporting model as fbx: %v", err)
		}
	case "gltf":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".glb")

		if doc, err := mdl.ExportGLTFDefault(wrsrc); err != nil {
			log.Printf("Error when exporting model as gltf: %v", err)
		} else {
			if err := gltfutils.ExportBinary(w, doc); err != nil {
				log.Printf("Failed to encode gltf: %v", err)
			}
		}
	}
}
