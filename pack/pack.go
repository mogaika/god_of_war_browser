package pack

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	TOK_ENCOUNTER_SIZE = 24
	TOK_PARTS_COUNT    = 2
)

type PackDriver interface {
	GetFileNamesList() []string
	GetFile(fileName string) (PackFile, error)
	GetFileReader(fileName string) (PackFile, *io.SectionReader, error)
	GetInstance(fileName string) (interface{}, error)
	UpdateFile(fileName string, in *io.SectionReader) error
}

type PackFile interface {
	Name() string
	Size() int64
}

type FileLoader func(src utils.ResourceSource, r *io.SectionReader) (interface{}, error)

var gHandlers map[string]FileLoader = make(map[string]FileLoader, 0)

func SetHandler(format string, ldr FileLoader) {
	gHandlers[strings.ToUpper(format)] = ldr
}

func CallHandler(s utils.ResourceSource, r *io.SectionReader) (interface{}, error) {
	ext := strings.ToUpper(filepath.Ext(s.Name()))

	if h, found := gHandlers[ext]; found {
		return h(s, r)
	} else {
		return nil, fmt.Errorf("Cannot find handler for '%s' extension", ext)
	}
}

type InstanceCacheEntry struct {
	Name     string
	Instance interface{}
}

type InstanceCache struct {
	Cache [16]*InstanceCacheEntry
	Pos   int
}

func (ic *InstanceCache) getEntry(name string) *InstanceCacheEntry {
	for _, ce := range ic.Cache {
		if ce != nil && ce.Name == name {
			return ce
		}
	}
	return nil
}

func (ic *InstanceCache) Get(name string) interface{} {
	if e := ic.getEntry(name); e != nil {
		return e.Instance
	}
	return nil
}

func (ic *InstanceCache) Add(name string, val interface{}) {
	if e := ic.getEntry(name); e != nil {
		e.Instance = val
	} else {
		ic.Cache[ic.Pos] = &InstanceCacheEntry{
			Name:     name,
			Instance: val,
		}
		ic.Pos = (ic.Pos + 1) % len(ic.Cache)
		if ic.Pos == 0 {
			runtime.GC()
		}
	}
}

type PackResSrc struct {
	pf PackFile
	p  PackDriver
}

func (s *PackResSrc) Name() string {
	return s.pf.Name()
}

func (s *PackResSrc) Size() int64 {
	return s.pf.Size()
}
func (s *PackResSrc) Save(in *io.SectionReader) error {
	return s.p.UpdateFile(s.pf.Name(), in)
}

func defaultGetInstanceCachedHandler(p PackDriver, cache *InstanceCache, fileName string) (interface{}, error) {
	f, r, err := p.GetFileReader(fileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot get instance of '%s': %v", fileName, err)
	}

	if cached := cache.Get(fileName); cached != nil {
		return cached, nil
	}

	inst, err := CallHandler(&PackResSrc{p: p, pf: f}, r)
	if err != nil {
		return nil, fmt.Errorf("Handler error: %v", err)
	}
	cache.Add(fileName, inst)
	return inst, nil
}

type TokFile struct {
	name       string
	size       int64
	Encounters []TokFileEncounter
}

func (tf *TokFile) Name() string {
	return tf.name
}

func (tf *TokFile) Size() int64 {
	return tf.size
}

type TokFileEncounter struct {
	Pack  int
	Start int64
}

func unmarshalTokEntry(buffer []byte) (name string, size int64, enc TokFileEncounter) {
	name = utils.BytesToString(buffer[0:12])
	size = int64(binary.LittleEndian.Uint32(buffer[16:20]))
	enc = TokFileEncounter{
		Pack:  int(binary.LittleEndian.Uint32(buffer[12:16])),
		Start: int64(binary.LittleEndian.Uint32(buffer[20:24])) * utils.SECTOR_SIZE,
	}
	return
}

func marshalTokEntry(name string, size int64, enc TokFileEncounter) []byte {
	buf := make([]byte, TOK_ENCOUNTER_SIZE)
	copy(buf[:12], utils.StringToBytes(name, 12, false))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(enc.Pack))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(size))
	binary.LittleEndian.PutUint32(buf[20:24], uint32((enc.Start+utils.SECTOR_SIZE-1)/utils.SECTOR_SIZE))
	return buf
}

func tokPartsParseFiles(tokStream io.Reader) (map[string]*TokFile, error) {
	var buffer [TOK_ENCOUNTER_SIZE]byte

	files := make(map[string]*TokFile)

	for {
		if _, err := tokStream.Read(buffer[:]); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		name, size, enc := unmarshalTokEntry(buffer[:])
		if name == "" {
			break
		}

		var file *TokFile
		if existFile, ok := files[name]; ok {
			file = existFile
		} else {
			file = &TokFile{
				name:       name,
				size:       size,
				Encounters: make([]TokFileEncounter, 0),
			}
			files[name] = file
		}

		if size != file.Size() {
			log.Printf("[pack] Finded same file but with different size! '%s' %d!=%d", name, size, file.Size)
		}

		file.Encounters = append(file.Encounters, enc)
	}
	return files, nil
}
