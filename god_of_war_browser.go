package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/mogaika/god_of_war_browser/status"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/vfs"
	"github.com/mogaika/god_of_war_browser/web"

	"github.com/mogaika/god_of_war_browser/drivers/iso"
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
	_ "github.com/mogaika/god_of_war_browser/pack/wad/light"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/obj"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/rsrcs"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/sbk"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/scr"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/shg"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/twk"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/txr"
)

func main() {
	var addr, tocpath, dirpath, isopath, psarcpath, psversion string
	var gowversion int
	var parsecheck bool
	flag.StringVar(&addr, "i", ":8000", "Address of server")
	flag.StringVar(&tocpath, "toc", "", "Path to folder with toc file")
	flag.StringVar(&dirpath, "dir", "", "Path to unpacked wads and other stuff")
	flag.StringVar(&isopath, "iso", "", "Path to iso file")
	flag.StringVar(&psarcpath, "psarc", "", "Path to ps3 psarc file")
	flag.StringVar(&psversion, "ps", "ps2", "Playstation version (ps2, ps3, psvita)")
	flag.IntVar(&gowversion, "gowversion", 0, "0 - auto, 1 - 'gow1', 2 - 'gow2'")
	flag.BoolVar(&parsecheck, "parsecheck", false, "Check every file for parse errors (for devs)")
	flag.Parse()

	var err error
	var gameDir vfs.Directory
	var driverDir vfs.Directory

	switch psversion {
	case "ps2":
		config.SetPlayStationVersion(config.PS2)
	case "ps3":
		config.SetPlayStationVersion(config.PS3)
	case "psvita":
		config.SetPlayStationVersion(config.PSVita)
	default:
		log.Fatalf("Provide correct 'ps' parameter (ps2, ps3, psvita)")
	}

	config.SetGOWVersion(config.GOWVersion(gowversion))

	if psarcpath != "" {
		if config.GetPlayStationVersion() != config.PS3 && config.GetPlayStationVersion() != config.PSVita {
			log.Fatalf("Cannot use psarcpath when 'ps' is not ps3 or psvita")
		}
		f := vfs.NewDirectoryDriverFile(psarcpath)
		if err = f.Open(true); err == nil {
			gameDir, err = psarc.NewPsarcDriver(f)
		}
	} else if isopath != "" {
		f := vfs.NewDirectoryDriverFile(isopath)
		if err = f.Open(false); err == nil {
			if driverDir, err = iso.NewIsoDriver(f); err == nil {
				gameDir, err = toc.NewTableOfContent(driverDir)
			}
		}
	} else if tocpath != "" {
		gameDir, err = toc.NewTableOfContent(vfs.NewDirectoryDriver(tocpath))
	} else if dirpath != "" {
		gameDir = vfs.NewDirectoryDriver(dirpath)
		if gowversion == 0 {
			log.Fatalf("You must provide 'gowversion' argument if you use directory driver")
		}
	} else {
		flag.PrintDefaults()
		return
	}

	if err != nil {
		log.Fatalf("Cannot start god of war browser: %v", err)
	}

	if f, err := setLogging(); err != nil {
		log.Printf("Wasn't able to setup logs dup: %v", err)
	} else {
		defer f.Close()
	}

	//parsecheck = true
	if parsecheck {
		parseCheck(gameDir)
	}
	status.Info("Starting web server on address '%s'", addr)

	if err := web.StartServer(addr, gameDir, driverDir, "web"); err != nil {
		log.Fatalf("Cannot start web server: %v", err)
	}
}

func setLogging() (io.Closer, error) {
	os.MkdirAll("applogs", 0777)
	f, err := os.Create(fmt.Sprintf("applogs/%s.log", time.Now().Format("2006.Jan.2_15.04.05")))
	if err != nil {
		return nil, err
	}

	log.SetOutput(io.MultiWriter(os.Stdout, f))
	return f, nil
}
