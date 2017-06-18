package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/mogaika/god_of_war_browser/pack"
)

func main() {
	var inPath, outDir string
	flag.StringVar(&inPath, "i", "", "Path to files folder")
	flag.StringVar(&outDir, "o", "", "Output directory")
	flag.Parse()

	var p *pack.Pack
	var err error

	if inPath != "" {
		p, err = pack.NewPackUnpacked(inPath)
	} else {
		log.Fatal("Provide path to folder with files. Use --help if you stuck.")
	}
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting saving to catalog ", outDir)
	log.Println("This can take a lot of time. Please be patient...")
	outTok, errTok := os.Create(path.Join(outDir, "GODOFWAR.TOC"))
	outPack, errPak := os.Create(path.Join(outDir, "PART1.PAK"))
	if errTok != nil || errPak != nil {
		log.Fatal(errTok, errPak)
	}
	defer func() {
		outPack.Close()
		outTok.Close()
	}()

	if err := p.SaveWithReplacement(outTok, outPack, nil, nil); err != nil {
		log.Fatal(err)
	}
	log.Println("Complete !")
}
