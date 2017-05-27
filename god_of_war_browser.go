package main

import (
	"flag"
	"log"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/web"

	_ "github.com/mogaika/god_of_war_browser/pack/vag"
	_ "github.com/mogaika/god_of_war_browser/pack/vpk"
	_ "github.com/mogaika/god_of_war_browser/pack/wad"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/anm"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/collision"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/cxt"
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
	var addr, game string
	var unpacked bool
	flag.StringVar(&addr, "i", ":8000", "Address of server")
	flag.StringVar(&game, "game", "", "Path to game folder")
	flag.BoolVar(&unpacked, "unpacked", false, "Pack data already unpacked")
	flag.Parse()

	var p *pack.Pack
	var err error

	if game != "" {
		if unpacked {
			p, err = pack.NewPackUnpacked(game)
		} else {
			p, err = pack.NewPack(game)
		}
	} else {
		p, err = pack.NewPackUnpacked("./wads")
	}

	defer p.Close()
	if err != nil {
		log.Fatal(err)
	}

	if err := web.StartServer(addr, p); err != nil {
		log.Fatal(err)
	}
}
