package main

import (
	"github.com/mogaika/god_of_war_browser/pack/wad/twk"
	"github.com/mogaika/god_of_war_browser/pack/wad/twk/twktree"

	//"encoding/binary"
	"log"
	"sort"
	"strings"

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
		if !strings.HasSuffix(fname, ".WAD") && !strings.HasSuffix(fname, ".wad_psp2") {
			continue
		}
		log.Printf("Parsecheck %q", fname)
		data, _ := pack.GetInstanceHandler(rootfs, fname)
		if data == nil {
			continue
		}
		switch data.(type) {
		case *file_wad.Wad:
			wad := data.(*file_wad.Wad)
			for _, node := range wad.Nodes {
				if node.Parent != file_wad.NODE_INVALID {
					continue
				}

				if strings.HasPrefix(node.Tag.Name, "TWK_") {
					// log.Printf("Parsing %s:%s: %v", wad.Name(), node.Tag.Name, err)
					fi, _, err := wad.GetInstanceFromNode(node.Id)
					if err != nil {
						// log.Printf("Failed to parse %s:%s: %v", wad.Name(), node.Tag.Name, err)
					} else if tw, ok := fi.(*twk.TWK); !ok {
						log.Printf("Failed to get twk from %s:%s", wad.Name(), node.Tag.Name)
					} else {
						if _, err := twktree.Root().UnmarshalTWK(tw.Tree); err != nil {
							log.Printf("Failed to parse abstarct tree for %s:%s: %v", wad.Name(), node.Tag.Name, err)
						}
					}

				}

				/*
					if strings.HasPrefix(node.Tag.Name, "ESC_") {
						log.Printf("Parsing %s:%s: %v", wad.Name(), node.Tag.Name, err)
						_, _, err := wad.GetInstanceFromNode(node.Id)
						if err != nil {
							log.Printf("Failed to parse %s:%s: %v", wad.Name(), node.Tag.Name, err)
						}
					}
				*/

				/*
					if len(node.Tag.Data) != 0 {
						for _, onode := range wad.Nodes {
							if onode.Parent != file_wad.NODE_INVALID || len(onode.Tag.Data) == 0 {
								continue
							}

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
				*/
			}
		}
	}
}
