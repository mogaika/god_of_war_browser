package cxt

import (
	"bytes"
	"log"
	"net/http"
	"strings"

	"github.com/mogaika/god_of_war_browser/fbx"
	"github.com/mogaika/god_of_war_browser/fbx/cache"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (cxt *Chunk) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "fbx":
		var buf bytes.Buffer
		// Export zip
		log.Printf("Error when exporting cxt: %v", cxt.ExportFbxDefault(wrsrc).ExportZip(&buf, wrsrc.Name()+".fbx"))
		// log.Printf("Error when exporting cxt: %v", cxt.ExportFbxDefault(wrsrc).Export(&buf))
		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Tag.Name+".zip")
	case "fbx_all":
		f := fbx.NewFbx()
		fbxCache := cache.NewCache()

		for _, node := range wrsrc.Wad.Nodes {
			if !strings.HasPrefix(node.Tag.Name, "CXT_") {
				continue
			}
			inst, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id)
			if err != nil {
				log.Panicf("Can't load cxt %s: %v", node.Tag.Name, err)
			}

			fe := inst.(*Chunk).ExportFbx(wrsrc.Wad.GetNodeResourceByNodeId(node.Id), f, fbxCache)
			f.Connections.C = append(f.Connections.C, fbx.Connection{
				Type: "OO", Parent: 0, Child: fe.FbxModelId,
			})
		}
		f.CountDefinitions()

		var buf bytes.Buffer
		// Export zip
		log.Printf("Error when exporting wad(cxt array): %v", f.ExportZip(&buf, wrsrc.Wad.Name()+".fbx"))
		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Wad.Name()+".zip")
	}
}
