package main

import (
	"log"

	"github.com/mogaika/god_of_war_browser/drivers/psarc"
	"github.com/mogaika/god_of_war_browser/vfs"
)

func main() {
	f := vfs.NewDirectoryDriverFile(`/run/user/1000/gvfs/smb-share:server=192.168.1.147,share=downloads2/BCES00791-[God of War Collection]/PS3_GAME/USRDIR/GOW1/exec/gow1.psarc`)
	if err := f.Open(true); err != nil {
		log.Panic(err)
	}
	defer f.Close()

	p, err := psarc.NewPsarcDriver(f)
	if err != nil {
		log.Panic(err)
	}
	_ = p
}
