package main

import (
	"log"

	"github.com/mogaika/god_of_war_browser/drivers/psarc"
	"github.com/mogaika/god_of_war_browser/vfs"
)

func main() {
	f := vfs.NewDirectoryDriverFile(`Q:\Downloads\PCSF00438\exec\gow1_psp2_SCEE.psarc`)
	if err := f.Open(true); err != nil {
		log.Panic(err)
	}
	defer f.Close()

	p, err := psarc.NewPsarcDriver(f)
	if err != nil {
		log.Panic(err)
	}
	_ = p

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

	if f, err := vfs.DirectoryGetFile(p, "wad_pand01b.wad_psp2"); err != nil {
		log.Panic(err)
	} else {
		if err := f.Open(true); err != nil {
			log.Panic(err)
		}
		defer f.Close()
	}
}
