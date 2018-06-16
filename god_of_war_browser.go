package main

import (
	"flag"
	"log"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/vfs"
	"github.com/mogaika/god_of_war_browser/web"

	"github.com/mogaika/god_of_war_browser/drivers/directory"
	"github.com/mogaika/god_of_war_browser/drivers/psarc"
	"github.com/mogaika/god_of_war_browser/drivers/toc"

	_ "github.com/mogaika/god_of_war_browser/pack/txt"
	_ "github.com/mogaika/god_of_war_browser/pack/vag"
	_ "github.com/mogaika/god_of_war_browser/pack/vpk"
	_ "github.com/mogaika/god_of_war_browser/pack/wad"

	_ "github.com/mogaika/god_of_war_browser/pack/wad/anm"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/collision"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/cxt"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/flp"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/gfx"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/inst"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/obj"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/sbk"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/scr"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/txr"
)

func main() {
	var addr, tocpath, dirpath, isopath, psarcpath string
	var gowversion int
	flag.StringVar(&addr, "i", ":8000", "Address of server")
	flag.StringVar(&tocpath, "toc", "", "Path to folder with toc file")
	flag.StringVar(&dirpath, "dir", "", "Path to unpacked wads and other stuff")
	flag.StringVar(&isopath, "iso", "", "Path to iso file")
	flag.StringVar(&psarcpath, "psarc", "", "Path to ps3 psarc file")
	flag.IntVar(&gowversion, "gowversion", 0, "0 - auto, 1 - 'gow1 ps2', 2 - 'gow2 ps2'")
	flag.Parse()

	var p pack.PackDriver
	var err error

	config.SetGOWVersion(config.GOWVersion(gowversion))

	if psarcpath != "" {
		var arch *psarc.Psarc
		f := vfs.NewDirectoryDriverFile(psarcpath)
		if err = f.Open(true); err == nil {
			if arch, err = psarc.NewPsarcDriver(f); err == nil {
				p = directory.NewDirectoryDriver(arch)
			}
		}
	} else if isopath != "" {
	} else if tocpath != "" {
		var t *toc.TableOfContent
		t, err = toc.NewTableOfContent(vfs.NewDirectoryDriver(tocpath))
		if err == nil {
			p = directory.NewDirectoryDriver(t)
		}
	} else if dirpath != "" {
		p = directory.NewDirectoryDriver(vfs.NewDirectoryDriver(dirpath))
	} else {
		flag.PrintDefaults()
		return
	}

	if err != nil {
		log.Fatal(err)
	}

	if err := web.StartServer(addr, p, "web"); err != nil {
		log.Fatal(err)
	}

}
