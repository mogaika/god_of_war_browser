package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/mogaika/udf"
)

const PACK_FILE_SIZE = 24
const PACK_PARTS_COUNT = 2

// Actually structs for second layer starts from 0x0fdf98000
// Because boot-space for iso/udf fs is 0x8000 we get 0x0fdf90000 as layer start
// But 0x0fdf92000 is end of part1 on my iso
// What i'm doin wrong?
var IsoSecondLayerStart int64 = 0x0fdf90000

const (
	PACK_STORE_ISO = iota // Iso
	PACK_STORE_TOK        // Directory with tok and packs
	PACK_STORE_DIR        // Wads in directory
)

type FileLoader func(pk *Pack, pf *PackFile, r io.ReaderAt) (interface{}, error)

var gHandlers map[string]FileLoader = make(map[string]FileLoader, 0)

func SetHandler(format string, ldr FileLoader) {
	gHandlers[format] = ldr
}

type PackFileEncounter struct {
	Pack  uint8
	Start int64
}

type PackFile struct {
	Name       string
	Size       int64
	Encounters []PackFileEncounter
	Instance   interface{} `json:"-"`
}

type Pack struct {
	Files     map[string]*PackFile
	GamePath  string `json:"-"`
	StoreType int8

	stream [PACK_PARTS_COUNT]io.ReaderAt
}

func (p *Pack) Get(name string) (interface{}, error) {
	file, ex := p.Files[name]
	if !ex {
		return nil, fmt.Errorf("File %s not exists in pack", name)
	}
	if file.Instance != nil {
		return file.Instance, nil
	}

	if han, ex := gHandlers[name[len(name)-4:]]; ex {
		rdr, err := p.GetFileReader(name)
		if err != nil {
			return nil, fmt.Errorf("Error getting file reader: %v", err)
		}
		instance, err := han(p, file, rdr)
		file.Instance = instance
		return instance, err
	} else {
		return nil, utils.ErrHandlerNotFound
	}
}

func NewPackFromDirectory(filesPath string) (*Pack, error) {
	files := make(map[string]*PackFile, 0)

	dirfiles, err := ioutil.ReadDir(filesPath)
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
		Files:     files,
		GamePath:  filesPath,
		StoreType: PACK_STORE_DIR,
	}

	return pack, nil
}

func (pack *Pack) parseTokAndParts(tokStream io.ReaderAt) error {
	var buffer [PACK_FILE_SIZE]byte

	for pos := int64(0); ; pos += PACK_FILE_SIZE {
		if _, err := tokStream.ReadAt(buffer[:], pos); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		name := utils.BytesToString(buffer[0:12])
		if name == "" {
			break
		}

		fileSize := int64(binary.LittleEndian.Uint32(buffer[16:20]))

		var file *PackFile
		if existFile, ok := pack.Files[name]; ok {
			file = existFile
		} else {
			file = &PackFile{
				Name:       name,
				Size:       int64(binary.LittleEndian.Uint32(buffer[16:20])),
				Encounters: make([]PackFileEncounter, 0),
			}
			pack.Files[name] = file
		}

		if fileSize != file.Size {
			log.Printf("[pack] Finded same file but with different size! '%s' %d!=%d", name, fileSize, file.Size)
		}

		file.Encounters = append(file.Encounters, PackFileEncounter{
			Pack:  uint8(binary.LittleEndian.Uint32(buffer[12:16])),
			Start: int64(binary.LittleEndian.Uint32(buffer[20:24])) * utils.SECTOR_SIZE,
		})
	}
	return nil
}

func NewPackFromTok(gamePath string) (*Pack, error) {
	pack := &Pack{
		Files:     make(map[string]*PackFile, 0),
		StoreType: PACK_STORE_TOK,
	}

	tokStream, err := os.Open(path.Join(gamePath, "GODOFWAR.TOC"))
	if err != nil {
		return nil, err
	}
	defer tokStream.Close()

	log.Printf("[pack] Using gamedir: '%s'", gamePath)
	for i := 0; i < PACK_PARTS_COUNT; i++ {
		pio, err := os.Open(filepath.Join(gamePath, fmt.Sprintf("PART%d.PAK", i+1)))
		if err != nil {
			log.Printf("[pack] WARNING! Cannot open part%d.pak: %v", i+1, err)
		} else {
			pack.stream[i] = pio
		}
	}

	return pack, pack.parseTokAndParts(tokStream)
}

func NewPackFromIso(isoPath string) (*Pack, error) {
	pack := &Pack{
		Files:     make(map[string]*PackFile, 0),
		StoreType: PACK_STORE_TOK,
	}

	fiso, err := os.Open(isoPath)
	if err != nil {
		return nil, err
	}

	var layers [2]*udf.Udf
	layers[0] = udf.NewUdfFromReader(fiso)

	isoinfo, err := fiso.Stat()
	if err == nil {
		if isoinfo.Size() > IsoSecondLayerStart {
			r := io.NewSectionReader(fiso, IsoSecondLayerStart, isoinfo.Size()-IsoSecondLayerStart)
			log.Printf("[pack] Detected second layer of iso file (size:%d)", r.Size())
			layers[1] = udf.NewUdfFromReader(r)
		}
	}

	tryOpen := func(fname string) (*io.SectionReader, error) {
		for _, u := range layers {
			for _, fe := range u.ReadDir(nil) {
				if strings.ToLower(fe.Name()) == strings.ToLower(fname) {
					return fe.NewReader(), nil
				}
			}
		}
		return nil, os.ErrNotExist
	}

	for i := 0; i < PACK_PARTS_COUNT; i++ {
		pio, err := tryOpen(fmt.Sprintf("PART%d.PAK", i+1))
		if err != nil {
			log.Printf("[pack] WARNING! Cannot open part%d.pak: %v", i+1, err)
		} else {
			pack.stream[i] = pio
		}
	}

	tokStream, err := tryOpen("GODOFWAR.TOC")
	if err != nil {
		return nil, err
	}

	return pack, pack.parseTokAndParts(tokStream)
}

func (p *Pack) GetFileReader(fname string) (*io.SectionReader, error) {
	if file, ex := p.Files[fname]; ex {
		switch p.StoreType {
		case PACK_STORE_DIR:
			arr, err := ioutil.ReadFile(filepath.Join(p.GamePath, fname))
			if err != nil {
				return nil, err
			}
			file.Size = int64(len(arr))
			return io.NewSectionReader(bytes.NewReader(arr), 0, int64(len(arr))), nil
		case PACK_STORE_TOK:
			for iFenc := range file.Encounters {
				fec := &file.Encounters[iFenc]
				if p.stream[fec.Pack] != nil {
					return io.NewSectionReader(p.stream[fec.Pack], fec.Start, file.Size), nil
				}
			}
			return nil, fmt.Errorf("[pack] Cannot find source for '%s' file :(")
		}
	}
	return nil, errors.New("[pack] Cannot find specifed file")

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
