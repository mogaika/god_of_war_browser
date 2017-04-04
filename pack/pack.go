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

func (p *Pack) Close() error {
	for _, s := range p.stream {
		if s != nil {
			return s.Close()
		}
	}
	return nil
}
