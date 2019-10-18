package main

import (
	"log"
	"sort"
	"strings"

	"github.com/mogaika/god_of_war_browser/pack"
	file_wad "github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/vfs"
)

func parseCheck(rootfs vfs.Directory) {
	packList, err := rootfs.List()
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(packList)
	for _, fname := range packList {
		data, _ := pack.GetInstanceHandler(rootfs, fname)
		if data == nil {
			continue
		}
		switch data.(type) {
		case *file_wad.Wad:
			wad := data.(*file_wad.Wad)
			for _, node := range wad.Nodes {
				_, _, err := wad.GetInstanceFromNode(node.Id)
				if err != nil {
					errStr := err.Error()
					if !strings.Contains(errStr, "Cannot find handler for tag ") &&
						!strings.Contains(errStr, "Handler return error: Unknown enz shape type mCDbgHdr") {
						log.Printf("E %.16s %.5d %.15s: %v", fname, node.Id, node.Tag.Name, err)
					}
				}
			}
		}
	}
}
