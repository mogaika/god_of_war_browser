package main

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mogaika/udf"
)

func GetFileEntryInIso(isoPath string, fileName string) (int64, map[int64]uint32) {
	f, err := os.Open(isoPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		f.Close()
	}()

	u := udf.NewUdfFromReader(f)
	for _, file := range u.ReadDir(nil) {
		if strings.ToUpper(file.Name()) == strings.ToUpper(fileName) {
			return FindFileStart(&file), FindFileSizePoses(f, &file)
		}
	}
	log.Fatalf("Cannot find file with name '%s' in iso", fileName)
	return 0, nil
}

func FindFileSizePoses(iso *os.File, f *udf.File) map[int64]uint32 {
	//f.FileEntry().FileEntryPosition
	//f.FileEntry().AllocationDescriptors[0].Length
	result := make(map[int64]uint32, 0)

	fePos := udf.SECTOR_SIZE * (int64(f.GetFileEntryPosition()) + int64(f.Udf.PartitionStart()))
	result[fePos+56] = 0

	allocDescStart := uint32(fePos) + 176 + f.FileEntry().LengthOfExtendedAttributes
	result[int64(allocDescStart+0)] = 0
	for i := range result {
		var s uint32
		if _, err := iso.Seek(i, os.SEEK_SET); err != nil {
			panic(err)
		}
		if err := binary.Read(iso, binary.LittleEndian, &s); err != nil {
			panic(err)
		}
		result[i] = s
	}
	return result
}

func FindFileStart(f *udf.File) int64 {
	return udf.SECTOR_SIZE * (int64(f.FileEntry().AllocationDescriptors[0].Location) + int64(f.Udf.PartitionStart()))
}

func main() {
	var isoPath string

	flag.StringVar(&isoPath, "iso", "", "Iso file name")
	flag.Parse()

	if isoPath == "" {
		log.Fatalf("Provide -iso argument")
	}

	for _, filePath := range flag.Args() {
		fileName := filepath.Base(filePath)

		filePos, fileSizePoses := GetFileEntryInIso(isoPath, fileName)

		log.Printf("File '%s' located at pos %v\n", fileName, filePos)

		isoFile, err := os.OpenFile(isoPath, os.O_WRONLY, 0777)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			isoFile.Close()
		}()

		if _, err := isoFile.Seek(filePos, os.SEEK_SET); err != nil {
			panic(err)
		}

		sourceFile, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			sourceFile.Close()
		}()

		bytesCopyed, err := io.Copy(isoFile, sourceFile)
		log.Printf("Copyed %v bytes", bytesCopyed)
		if err != nil {
			log.Fatal(err)
		}

		for sizePose, oldSize := range fileSizePoses {
			if _, err := isoFile.Seek(int64(sizePose), os.SEEK_SET); err != nil {
				panic(err)
			}
			log.Printf("Updating file size in %v pos. From %v => %v", sizePose, oldSize, bytesCopyed)
			binary.Write(isoFile, binary.LittleEndian, uint32(bytesCopyed))
		}
	}
}
