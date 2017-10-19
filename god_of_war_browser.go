package main

import (
	"flag"
	"log"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/toc"
	"github.com/mogaika/god_of_war_browser/web"

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
	var addr, tocpath, dir, iso, part1name, part2name, tocname string
	var secondLayerStart int64
	flag.StringVar(&addr, "i", ":8000", "Address of server")
	flag.StringVar(&tocpath, "toc", "", "Path to folder with toc file")
	//flag.StringVar(&dir, "dir", "", "Path to unpacked wads and other stuff")
	flag.StringVar(&iso, "iso", "", "Pack to iso file")
	flag.StringVar(&part1name, "partname1", "", "PART1.PAK name override")
	flag.StringVar(&part2name, "partname2", "", "PART2.PAK name override")
	flag.StringVar(&tocname, "tocname", "", "GODOFWAR.TOC name override")
	flag.Int64Var(&secondLayerStart, "secondLayerStart", pack.IsoSecondLayerStart, "Start of second layer of DVD disk in iso. Use 0 if you dont want use second layer. Default value valid for 8 522 792 960 bytes iso")
	flag.Parse()

	var p pack.PackDriver
	var err error

	toc.PartNameOverride(0, part1name)
	toc.PartNameOverride(1, part2name)
	toc.TocNameOverride(tocname)

	if iso != "" {
		p, err = pack.NewPackFromIso(iso, secondLayerStart)
	} else if tocpath != "" {
		p, err = pack.NewPackFromToc(tocpath)
	} else if dir != "" {
		//s	p, err = pack.NewPackFromDirectory(dir)
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
