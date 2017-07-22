package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/tok"
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
	flag.StringVar(&inPath, "i", "", "Path to files folder")
	flag.StringVar(&outDir, "o", "", "Output directory")
	flag.Parse()

	var err error

	if inPath == "" {
		log.Fatal("Provide path to folder with files. Use --help if you stuck.")
	}

	files, err := ioutil.ReadDir(inPath)
	if err != nil {
		log.Fatalf("Cannot read dir '%s': %v", inPath, err)
	}

	log.Println("Starting saving to catalog ", outDir)
	log.Println("This can take a lot of time. Please be patient...")
	outTok, err := os.Create(filepath.Join(outDir, tok.FILE_NAME))
	if err != nil {
		log.Fatalf("Cannot create file '%s' in dir '%s':%v", tok.FILE_NAME, outDir, err)
	}
	outPack, err := os.Create(filepath.Join(outDir, tok.GenPartFileName(0)))
	if err != nil {
		log.Fatalf("Cannot create file '%s' in dir '%s':%v", tok.GenPartFileName(0), outDir, err)
	}
	defer func() {
		outPack.Close()
		outTok.Close()
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

		outTok.Write(tok.MarshalTokEntry(&tok.Entry{
			Name: file.Name(),
			Size: file.Size(),
			Enc: tok.FileEncounter{
				Pack:  0,
				Start: offset,
			},
		}))

		log.Printf("Packed '%s'", file.Name())
	}

	log.Println("Packing completed!")
}
