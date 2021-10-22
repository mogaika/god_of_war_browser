package main

import (

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
		data, _ := pack.GetInstanceHandler(rootfs, fname)
		if data == nil {
			continue
		}
		switch data.(type) {
		case *file_wad.Wad:
			wad := data.(*file_wad.Wad)
			for _, node := range wad.Nodes {
				if node.Parent != 0 {
					continue
				}
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
			}
		}
	}
}
