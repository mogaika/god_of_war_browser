package main

import (
	"flag"
	"log"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/toc"
	"github.com/mogaika/god_of_war_browser/web"

	"github.com/mogaika/god_of_war_browser/pack/drivers/dirdriver"
	"github.com/mogaika/god_of_war_browser/pack/drivers/isodriver"
	"github.com/mogaika/god_of_war_browser/pack/drivers/tocdriver"

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
	var addr, tocpath, dir, iso, partprefix, partpostfix, tocname string
	var partindexing bool
	var gowversion int
	flag.StringVar(&addr, "i", ":8000", "Address of server")
	flag.StringVar(&tocpath, "toc", "", "Path to folder with toc file")
	flag.StringVar(&dir, "dir", "", "Path to unpacked wads and other stuff")
	flag.StringVar(&iso, "iso", "", "Pack to iso file")
	flag.StringVar(&partprefix, "partprefix", "PART", "PAK name prefix override")
	flag.StringVar(&partpostfix, "partpostfix", ".PAK", "PAK name postfix override")
	flag.BoolVar(&partindexing, "partindexing", true, "use -partprefix%index%.PAK naming, or use -partprefix name instead")
	flag.StringVar(&tocname, "tocname", "", "GODOFWAR.TOC name override")
	flag.IntVar(&gowversion, "gowversion", 0, "0 - auto, 1 - 'gow1 ps2', 2 - 'gow2 ps2'")
	flag.Parse()

	var p pack.PackDriver
	var err error

	config.SetGOWVersion(config.GOWVersion(gowversion))

	toc.PartNamePrefix(partprefix)
	toc.PartNamePostfix(partpostfix)
	toc.PartNameUseIndexing(partindexing)
	toc.TocNameOverride(tocname)

	if iso != "" {
		p, err = isodriver.NewPackFromIso(iso)
	} else if tocpath != "" {
		p, err = tocdriver.NewPackFromToc(tocpath)
	} else if dir != "" {
		p, err = dirdriver.NewPackFromDirectory(dir)
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
