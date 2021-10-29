package cxt

import (
	"bytes"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/mogaika/fbx/builders/bfbx73"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils/fbxbuilder"
	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (cxt *Chunk) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "gltf":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".glb")
		if doc, err := cxt.ExportGLTFDefault(wrsrc); err != nil {
			log.Printf("Error when exporting object as gltf: %v", err)
		} else {
			if err := gltfutils.ExportBinary(w, doc); err != nil {
				log.Printf("Failed to encode gltf: %v", err)
			}
		}
	case "gltf_all":
		gltfCacher := gltfutils.NewCacher()
		doc := gltfCacher.Doc

		for _, node := range wrsrc.Wad.Nodes {
			if !strings.HasPrefix(node.Tag.Name, "CXT_") {
				continue
			}
			inst, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id)
			if err != nil {
				log.Panicf("Failed to load cxt %s: %v", node.Tag.Name, err)
			}

			_, err = inst.(*Chunk).ExportGLTF(wrsrc.Wad.GetNodeResourceByNodeId(node.Id), gltfCacher)
			if err != nil {
				log.Panicf("Failed to encode %q: %v", node.Tag.Name, err)
			}
		}

		webutils.WriteFileHeaders(w, wrsrc.Wad.Name()+".glb")
		if err := gltfutils.ExportBinary(w, doc); err != nil {
			log.Printf("Failed to encode gltf: %v", err)
		}
	case "fbx":
		var buf bytes.Buffer
		// Export zip
		log.Printf("Error when exporting cxt: %v", cxt.ExportFbxDefault(wrsrc).WriteZip(&buf, wrsrc.Name()+".fbx"))
		// log.Printf("Error when exporting cxt: %v", cxt.ExportFbxDefault(wrsrc).Export(&buf))
		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Tag.Name+".zip")
	case "fbx_all":
		f := fbxbuilder.NewFBXBuilder(filepath.Join(wrsrc.Wad.Name(), wrsrc.Name()))

		for _, node := range wrsrc.Wad.Nodes {
			if !strings.HasPrefix(node.Tag.Name, "CXT_") {
				continue
			}
			inst, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id)
			if err != nil {
				log.Panicf("Can't load cxt %s: %v", node.Tag.Name, err)
			}

			fe := inst.(*Chunk).ExportFbx(wrsrc.Wad.GetNodeResourceByNodeId(node.Id), f)
			f.AddConnections(bfbx73.C("OO", fe.FbxModelId, 0))
		}

		var buf bytes.Buffer
		// Export zip
		log.Printf("Error when exporting wad(cxt array): %v", f.WriteZip(&buf, wrsrc.Wad.Name()+".fbx"))
		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Wad.Name()+".zip")
	}
}
