package main

import (
	"log"

	"github.com/mogaika/god_of_war_browser/drivers/psarc"
	"github.com/mogaika/god_of_war_browser/vfs"
)

func main() {
	f := vfs.NewDirectoryDriverFile(`E:\Downloads\BCES00791-[God of War Collection]/PS3_GAME/USRDIR/GOW1/exec/gow1.psarc`)
	if err := f.Open(true); err != nil {
		log.Panic(err)
	}
	defer f.Close()

	p, err := psarc.NewPsarcDriver(f)
	if err != nil {
		log.Panic(err)
	}
	_ = p

	/*
		if files, err := p.List(); err != nil {
			log.Panic(err)
		} else {
			for _, f := range files {
				if fel, err := p.GetElement(f); err != nil {
					log.Panic(err)
				} else {
					log.Printf("%-48s %+#v", f, fel.(vfs.File).Size())
				}
			}
		}
	*/
	if f, err := vfs.DirectoryGetFile(p, "wad/r_ah3.wad_ps3"); err != nil {
		log.Panic(err)
	} else {
		if err := f.Open(true); err != nil {
			log.Panic(err)
		}
		defer f.Close()
	}
}
