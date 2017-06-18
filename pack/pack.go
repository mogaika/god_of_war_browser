package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/utils"
)

const PACK_FILE_SIZE = 24

type FileLoader func(pk *Pack, pf *PackFile, r io.ReaderAt) (interface{}, error)

var cacheHandlers map[string]FileLoader = make(map[string]FileLoader, 0)

func SetHandler(format string, ldr FileLoader) {
	cacheHandlers[format] = ldr
}

type PackFile struct {
	Name  string
	Pack  uint32
	Size  int64
	Start int64

	Count int
	Cache interface{} `json:"-"`
}

type Pack struct {
	Files           map[string]*PackFile
	AlreadyUnpacked bool
	GamePath        string `json:"-"`
	stream          [2]*os.File
}

func (p *Pack) Get(name string) (interface{}, error) {
	file, ex := p.Files[name]
	if !ex {
		return nil, fmt.Errorf("File %s not exists in pack", name)
	}
	if file.Cache != nil {
		return file.Cache, nil
	}

	if han, ex := cacheHandlers[name[len(name)-4:]]; ex {
		rdr, err := p.GetFileReader(name)
		if err != nil {
			return nil, fmt.Errorf("Error getting file reader: %v", err)
		}
		cache, err := han(p, file, rdr)
		file.Cache = cache
		return cache, err
	} else {
		return nil, utils.ErrHandlerNotFound
	}
}

func NewPackUnpacked(gamePath string) (*Pack, error) {
	files := make(map[string]*PackFile, 0)

	dirfiles, err := ioutil.ReadDir(gamePath)
	if err != nil {
		return nil, err
	}

	for _, file := range dirfiles {
		if file.IsDir() {
			continue
		}
		files[file.Name()] = &PackFile{
			Name: file.Name(),
			Size: file.Size(),
		}
	}

	pack := &Pack{
		Files:           files,
		GamePath:        gamePath,
		AlreadyUnpacked: true,
	}

	return pack, nil
}

func NewPack(gamePath string) (*Pack, error) {
	var buffer [PACK_FILE_SIZE]byte
	files := make(map[string]*PackFile, 0)

	tokStream, err := os.Open(path.Join(gamePath, "GODOFWAR.TOC"))
	if err != nil {
		return nil, err
	}

	for pos := int64(0); ; pos += PACK_FILE_SIZE {
		_, err := tokStream.ReadAt(buffer[:], pos)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		name := utils.BytesToString(buffer[0:12])
		if name == "" {
			break
		}

		file := &PackFile{
			Name:  name,
			Pack:  binary.LittleEndian.Uint32(buffer[12:16]),
			Size:  int64(binary.LittleEndian.Uint32(buffer[16:20])),
			Start: int64(binary.LittleEndian.Uint32(buffer[20:24])) * utils.SECTOR_SIZE,
			Count: 1,
		}

		if _, ok := files[name]; ok {
			files[name].Count++
			if files[name].Size != file.Size {
				return nil, fmt.Errorf("Files has same names, but different sizes: %s", name)
			}
		} else {
			files[name] = file
		}
	}

	pack := &Pack{
		Files:           files,
		GamePath:        gamePath,
		AlreadyUnpacked: false,
	}

	return pack, nil
}

func (p *Pack) GetFileReader(fname string) (*io.SectionReader, error) {
	if file, ex := p.Files[fname]; ex {
		if p.AlreadyUnpacked {
			arr, err := ioutil.ReadFile(filepath.Join(p.GamePath, fname))
			if err != nil {
				return nil, err
			}
			file.Size = int64(len(arr))
			return io.NewSectionReader(bytes.NewReader(arr), 0, int64(len(arr))), nil
		} else {
			if p.stream[file.Pack] == nil {
				pio, err := os.Open(filepath.Join(p.GamePath, fmt.Sprintf("PART%d.PAK", file.Pack+1)))
				if err != nil {
					return nil, err
				} else {
					p.stream[file.Pack] = pio
				}
			}
		}

		return io.NewSectionReader(p.stream[file.Pack], file.Start, file.Size), nil
	} else {
		return nil, errors.New("Cannot find specifed file")
	}
}

// Function support only 1-pack
func (p *Pack) SaveWithReplacement(outTok, outPack io.Writer, filesToReplaceOrAdd map[string]io.Reader, filesToRemove []string) error {
	outPackPos := int64(0)
	alreadyWrittenFiles := make([]string, 0)

	var zerobytes [utils.SECTOR_SIZE]byte

	writeFile := func(name string, source io.Reader) (err error) {
		for _, writtenFname := range alreadyWrittenFiles {
			if name == writtenFname {
				return nil
			}
		}

		startPos := outPackPos

		if packWritten, err := io.Copy(outPack, source); err != nil {
			return err
		} else {
			outPackPos += packWritten
		}

		outTok.Write(utils.StringToBytes(name, 12, false))
		var outBuf [12]byte
		binary.LittleEndian.PutUint32(outBuf[:4], 0)
		binary.LittleEndian.PutUint32(outBuf[4:8], uint32(outPackPos-startPos))
		if packWritten, err := outPack.Write(zerobytes[:utils.SECTOR_SIZE-outPackPos%utils.SECTOR_SIZE]); err != nil {
			return err
		} else {
			outPackPos += int64(packWritten)
		}

		binary.LittleEndian.PutUint32(outBuf[8:], uint32(startPos/int64(utils.SECTOR_SIZE)))
		alreadyWrittenFiles = append(alreadyWrittenFiles, name)
		if _, err := outTok.Write(outBuf[:]); err != nil {
			return err
		}

		return nil
	}

	processFile := func(fileName string) error {
		for _, remFileName := range filesToRemove {
			if remFileName == fileName {
				return nil
			}
		}

		var fileSource io.Reader
		if newReader, hasNew := filesToReplaceOrAdd[fileName]; hasNew {
			fileSource = newReader
		} else {
			if reader, err := p.GetFileReader(fileName); err != nil {
				return err
			} else {
				fileSource = reader
			}
		}

		return writeFile(fileName, fileSource)
	}

	for fileName := range p.Files {
		if err := processFile(fileName); err != nil {
			return err
		}
	}

	for newFileName, source := range filesToReplaceOrAdd {
		if err := writeFile(newFileName, source); err != nil {
			return err
		}
	}

	outTok.Write(zerobytes[:0x10])

	return nil
}

func (p *Pack) Close() error {
	for _, s := range p.stream {
		if s != nil {
			return s.Close()
		}
	}
	return nil
}
