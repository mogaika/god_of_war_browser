package main

import (
	"flag"
	"log"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/tools/browser/web"

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

	"net/http"
	_ "net/http/pprof"
)

func main() {
	var addr, tok, dir, iso string
	flag.StringVar(&addr, "i", ":8000", "Address of server")
	flag.StringVar(&tok, "tok", "", "Path to folder with tok file")
	flag.StringVar(&dir, "dir", "", "Path to unpacked wads and other stuff")
	flag.StringVar(&iso, "iso", "", "Pack to iso file")
	flag.Parse()

	var p pack.PackDriver
	var err error

	if iso != "" {
		//		p, err = pack.NewPackFromIso(iso)
	} else if tok != "" {
		p, err = pack.NewPackFromTok(tok)
	} else if dir != "" {
		//s	p, err = pack.NewPackFromDirectory(dir)
	} else {
		flag.PrintDefaults()
		return
	}

	if err != nil {
		log.Fatal(err)
	}

	go http.ListenAndServe(":7777", http.DefaultServeMux)

	if err := web.StartServer(addr, p, "web"); err != nil {
		log.Fatal(err)
	}
}
