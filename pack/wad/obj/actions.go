package obj

import (
	"log"
	"net/http"

	"github.com/mogaika/god_of_war_browser/utils/gltfutils"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (obj *Object) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "gltf":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".glb")

		if doc, err := obj.ExportGLTFDefault(wrsrc); err != nil {
			log.Printf("Error when exporting object as gltf: %v", err)
		} else {
			if err := gltfutils.ExportBinary(w, doc); err != nil {
				log.Printf("Failed to encode gltf: %v", err)
			}
		}
	case "fbx":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".fbx")
		if err := obj.ExportFbxDefault(wrsrc).Write(w); err != nil {
			log.Printf("Error when exporting object as fbx: %v", err)
		}
	}
}
