package mesh

import (
	"log"
	"net/http"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (mesh *Mesh) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "fbx":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".fbx")
		if err := mesh.ExportFbxDefault(wrsrc).Write(w); err != nil {
			log.Printf("Error when exporting mesh as fbx: %v", err)
		}
	case "obj":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".obj")
		if err := mesh.ExportObj(w, nil); err != nil {
			log.Printf("Error when exporting mesh as obj: %v", err)
		}
	case "gltf":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".glb")

		if doc, err := mesh.ExportGLTFDefault(wrsrc); err != nil {
			log.Printf("Error when exporting mesh as gltf: %v", err)
		} else {
			if err := gltfutils.ExportBinary(w, doc); err != nil {
				log.Printf("Failed to encode gltf: %v", err)
			}
		}
	}
}
