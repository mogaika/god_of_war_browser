package main

import (
	"strings"
	//"encoding/binary"
	"log"
	"sort"

	"github.com/mogaika/god_of_war_browser/pack"
	file_wad "github.com/mogaika/god_of_war_browser/pack/wad"

	//	"github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	"github.com/mogaika/god_of_war_browser/vfs"
)

func parseCheck(rootfs vfs.Directory) {
	packList, err := rootfs.List()
	if err != nil {
		log.Fatal(err)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(packList)))

	for _, fname := range packList {
		if !strings.HasSuffix(fname, "WAD") {
			continue
		}
		data, _ := pack.GetInstanceHandler(rootfs, fname)
		if data == nil {
			continue
		}
		switch data.(type) {
		case *file_wad.Wad:
			wad := data.(*file_wad.Wad)
			contexts := make([]string, 0)
			after := false
			for _, node := range wad.Nodes {

				/*if node.Parent == 0 {
					continue
				}*/
				/*
					for _, onode := range wad.Nodes {
						if onode.Tag.Id >= node.Tag.Id {
							break
						}

						if node.Tag.Name == onode.Tag.Name {
							log.Printf("Conflicting name %q [%d:%q:%d] [%d:%q:%d]",
								wad.Name(),
								node.Tag.Id, node.Tag.Name, node.Tag.Size,
								onode.Tag.Id, onode.Tag.Name, node.Tag.Size)
							break
						}
					}
				*/
				if len(node.Tag.Data) != 0 && node.Tag.Name == "RIB_sheet" {
					log.Println(wad.Name())
					wad.GetInstanceFromNode(node.Id)
					after = true
				}
				if after && len(node.Tag.Data) == 52 && strings.HasPrefix(node.Tag.Name, "CXT_") {
					contexts = append(contexts, node.Tag.Name)
					//log.Println(node.Tag.Name)
				}

				// if len(node.Tag.Data) != 0 && binary.LittleEndian.Uint32(node.Tag.Data) == mesh.MESH_MAGIC {
				// 	_, _, err := wad.GetInstanceFromNode(node.Id)
				// 	if err != nil {
				// 		// errStr := err.Error()
				// 		/* if !strings.Contains(errStr, "Cannot find handler for tag ") &&
				// 			!strings.Contains(errStr, "Handler return error: Unknown enz shape type mCDbgHdr") {
				// 			log.Printf("E %.16s %.5d %.15s: %v", fname, node.Tag.Id, node.Tag.Name, err)
				// 		} */
				// 	}
				// }
			}
			if after {
				//log.Printf("CONTEXTS: %d %v", len(contexts)-1, contexts)
			}
		}
	}
}
