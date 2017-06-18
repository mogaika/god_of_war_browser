package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mogaika/god_of_war_browser/utils"
)

var motd = `#
#  ▄████  ▒█████  ▓█████▄  ▒█████    █████▒█     █░ ▄▄▄       ██▀███  
# ██▒ ▀█▒▒██▒  ██▒▒██▀ ██▌▒██▒  ██▒▓██   ▒▓█░ █ ░█░▒████▄    ▓██ ▒ ██▒
#▒██░▄▄▄░▒██░  ██▒░██   █▌▒██░  ██▒▒████ ░▒█░ █ ░█ ▒██  ▀█▄  ▓██ ░▄█ ▒
#░▓█  ██▓▒██   ██░░▓█▄   ▌▒██   ██░░▓█▒  ░░█░ █ ░█ ░██▄▄▄▄██ ▒██▀▀█▄  
#░▒▓███▀▒░ ████▓▒░░▒████▓ ░ ████▓▒░░▒█░   ░░██▒██▓  ▓█   ▓██▒░██▓ ▒██▒
# ░▒   ▒ ░ ▒░▒░▒░  ▒▒▓  ▒ ░ ▒░▒░▒░  ▒ ░   ░ ▓░▒ ▒   ▒▒   ▓▒█░░ ▒▓ ░▒▓░
#  ░   ░   ░ ▒ ▒░  ░ ▒  ▒   ░ ▒ ▒░  ░       ▒ ░ ░    ▒   ▒▒ ░  ░▒ ░ ▒░
#      ░     ░ ░     ░        ░ ░             ░          ░  ░   ░     
#
# <=======> Wad meta file <=======>
#
# All numbers in hex 
# Lines format:
# tag | flags | name | saved_filename | [size | ]
# [size] required only if it not same as in saved_filename
# 
# in names use @ for spaces " SCR_Sky" become @SCR_Sky
# if saved filename empty (''), then size field required
# special case: 
# EntityCount tag (id=0x18) has size = entity count
# and real size always = 0
`

func UnpackWad(f io.ReadSeeker, outDir string) {
	if err := os.MkdirAll(outDir, 0776); err != nil {
		panic(err)
	}
	wadMeta, err := os.Create(filepath.Join(outDir, "_wad_meta_.txt"))
	if err != nil {
		panic(err)
	}

	fmt.Fprint(wadMeta, motd)

	var groupLevel = 0

	id := 0
	var itembuf [32]byte
	for {
		id += 1
		if id > 10000 {
			panic(id)
		}
		_, err := f.Read(itembuf[:])
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}

		tag := binary.LittleEndian.Uint16(itembuf[0:2])
		flags := binary.LittleEndian.Uint16(itembuf[2:4])
		size := binary.LittleEndian.Uint32(itembuf[4:8])
		name := utils.BytesToString(itembuf[8:32])

		log.Println(name)

		name = strings.Replace(name, " ", "@", -1)

		savedFileName := name
		if len(savedFileName) > 3 && savedFileName[3] == '_' {
			savedFileName += "." + savedFileName[:3]
		}
		savedFileName = fmt.Sprintf("%.3d_%s", id, savedFileName)

		fmt.Fprintf(wadMeta, "%-4x | %-4x | %34s |", tag, flags, name)
		if size == 0 || tag == 0x18 {
			fmt.Fprintf(wadMeta, " %-34s | %-6x", "", size)
			savedFileName = ""
		} else {
			fmt.Fprintf(wadMeta, " %-34s         ", savedFileName)
		}

		switch tag {
		case 0x1e, 0x28, 0x32, 0x6e,
			0x6f, 0x70, 0x71, 0x72,
			0x01f4, 0x029a, 0x0378, 0x03e7:
			if savedFileName != "" {
				of, err := os.Create(filepath.Join(outDir, savedFileName))
				if err != nil {
					panic(err)
				}

				if _, err := io.CopyN(of, f, int64(size)); err != nil {
					panic(err)
				}
			}

		case 0x18:
		}

		if pos, _ := f.Seek(0, os.SEEK_CUR); pos%16 != 0 {
			if _, err := f.Seek(((pos+15)/16)*16, os.SEEK_SET); err != nil {
				panic(err)
			}
		}

		switch tag {
		case 0x28:
			fmt.Fprintf(wadMeta, "# group start >")
			groupLevel++
		case 0x32:
			groupLevel--
			fmt.Fprintf(wadMeta, "# group end < ")
		case 0x18:
			fmt.Fprintf(wadMeta, "# special entity count tag. Just dont touch it, ok?")
		default:
			if groupLevel != 0 {
				fmt.Fprintf(wadMeta, "# ")
				for i := 0; i < groupLevel; i++ {
					fmt.Fprintf(wadMeta, " -- ")
				}
			}
		}

		fmt.Fprintf(wadMeta, "\n")
	}
}

func main() {
	var inWad, outDir string
	flag.StringVar(&inWad, "wad", "", "Path to wad file to unpack")
	flag.StringVar(&outDir, "out", "wad_content", "Patch where to unpack wad file")
	flag.Parse()

	f, err := os.Open(inWad)
	if err != nil {
		panic(err)
	}
	defer func() {
		f.Close()
	}()

	UnpackWad(f, outDir)
}
