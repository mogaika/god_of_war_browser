package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/drivers/toc"
	"github.com/mogaika/god_of_war_browser/utils"
)

func copyFileToWriter(filePath string, w io.Writer) (int64, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("Cannot open file '%s':%v", filePath, err)
	}
	defer f.Close()

	return io.Copy(w, f)
}

func main() {
	var inPath, outDir string
	var gowVersion int
	var useIndexing bool
	flag.StringVar(&inPath, "i", "", "Path to files folder")
	flag.StringVar(&outDir, "o", "", "Output directory")
	flag.IntVar(&gowVersion, "gowversion", 1, "1 - 'gow1', 2 - 'gow2'")
	flag.BoolVar(&useIndexing, "arrayindexingabsolute", false, "Use absolute offset in array. Used in some GOW2 rips")
	flag.Parse()

	var err error

	config.SetGOWVersion(config.GOWVersion(gowVersion))
	if gowVersion != 1 && gowVersion != 2 {
		log.Fatal("Incorrect gow version")
	}

	if inPath == "" {
		log.Fatal("Provide path to folder with files. Use --help if you stuck.")
	}

	files, err := ioutil.ReadDir(inPath)
	if err != nil {
		log.Fatalf("Cannot read dir '%s': %v", inPath, err)
	}

	log.Println("Starting saving to catalog ", outDir)
	log.Println("This can take a lot of time. Please be patient...")

	tb := toc.NewTableOfContentBuilder()
	if useIndexing {
		tb.SetPackArrayIndexing(toc.PACK_ADDR_ABSOLUTE)
	} else {
		tb.SetPackArrayIndexing(toc.PACK_ADDR_INDEX)
	}

	outPack, err := os.Create(filepath.Join(outDir, "PART1.PAK"))
	if err != nil {
		log.Fatalf("Cannot create file '%s' in dir '%s':%v", "PART1.PAK", outDir, err)
	}
	defer func() {
		outPack.Close()
	}()

	var offset int64
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		n, err := copyFileToWriter(filepath.Join(inPath, file.Name()), outPack)
		if err != nil {
			log.Printf("Error when copying '%s': %v", file.Name(), err)
		}
		offset += n

		if offset%utils.SECTOR_SIZE != 0 {
			// pad to cd sector size
			offset, err = outPack.Seek(utils.SECTOR_SIZE-(offset%utils.SECTOR_SIZE), os.SEEK_CUR)
			if err != nil {
				log.Fatalf("Error when seeking: %v", err)
			}
		}

		tb.AddFile(file.Name(), file.Size(), toc.Encounter{
			Offset: offset, Pak: 0, Size: file.Size()})

		log.Printf("Packed '%s'", file.Name())
	}

	tokFileName := path.Join(outDir, "GODOFWAR.TOC")
	if err := ioutil.WriteFile(tokFileName, tb.Marshal(), 0777); err != nil {
		log.Fatal("Failed to write tok file %q: %v", tokFileName, err)
	}

	log.Println("Packing completed!")
}
