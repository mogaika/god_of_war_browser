package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mogaika/god_of_war_browser/utils"
)

func CopyFileToStream(f io.Writer, fileName string) int64 {
	src, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	defer func() {
		src.Close()
	}()

	if written, err := io.Copy(f, src); err != nil {
		panic(err)
	} else {
		return written
	}
}

func MakeWad(metaDir string, inMeta *bufio.Reader, outWad io.Writer) {
	for {
		s, err := inMeta.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		s = strings.Split(s, "#")[0]
		s = strings.Trim(s, " \t\n")
		if s != "" {
			var tag, flags uint16
			var name, fileName string
			var size uint32

			params := strings.Split(s, "|")
			fmt.Sscanf(params[0], "%x", &tag)
			fmt.Sscanf(params[1], "%x", &flags)
			name = params[2]
			fileName = params[3]

			name = strings.Trim(name, " \t")
			name = strings.Replace(name, "@", " ", -1)
			fileName = strings.Trim(fileName, " \t")
			fmt.Println(tag, flags, name, fileName, size)

			hasSize := len(params) > 4
			if hasSize {
				fmt.Sscanf(params[4], "%x", &size)
			}

			hasFile := fileName != ""

			if !hasSize {
				if hasFile {
					stats, err := os.Stat(filepath.Join(metaDir, fileName))
					if err != nil {
						panic(err)
					}
					size = uint32(stats.Size())
				}
			}

			var buf [32]byte
			binary.LittleEndian.PutUint16(buf[0:], tag)
			binary.LittleEndian.PutUint16(buf[2:], flags)
			binary.LittleEndian.PutUint32(buf[4:], size)
			copy(buf[8:], utils.StringToBytes(name, 24, false))

			outWad.Write(buf[:])

			written := int64(0)

			if hasFile {
				written = CopyFileToStream(outWad, filepath.Join(metaDir, fileName))

				var zeroes [16]byte
				if written%16 != 0 {
					if _, err := outWad.Write(zeroes[:16-written%16]); err != nil {
						panic(err)
					}
				}
			}

		}
	}
}

func main() {
	var inMeta, outWad string
	flag.StringVar(&inMeta, "meta", "", "Wad meta file")
	flag.StringVar(&outWad, "out", "", "Output wad file")
	flag.Parse()

	if inMeta == "" || outWad == "" {
		panic("Provide -meta and -out flags")
	}

	fMeta, err := os.Open(inMeta)
	if err != nil {
		panic(err)
	}

	fWad, err := os.Create(outWad)
	if err != nil {
		panic(err)
	}

	inMetaDir, _ := filepath.Split(inMeta)
	MakeWad(inMetaDir, bufio.NewReader(fMeta), fWad)
}
