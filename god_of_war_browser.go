package main

import (
	"flag"
	"log"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/web"

	_ "github.com/mogaika/god_of_war_browser/pack/wad"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/gfx"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/txr"
)

func main() {
	var addr, game string
	flag.StringVar(&addr, "i", ":8000", "Address of server")
	flag.StringVar(&game, "game", "", "Path to game folder")
	flag.Parse()

	pack, err := pack.NewPack(game)
	defer pack.Close()
	if err != nil {
		log.Fatal(err)
	}

	if err := web.StartServer(addr, pack); err != nil {
		log.Fatal(err)
	}
}
