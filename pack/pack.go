package pack

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/vfs"
)

type PackDriver interface {
	GetFileNamesList() []string
	GetFile(fileName string) (vfs.File, error)
	GetInstance(fileName string) (interface{}, error)
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

type PackResSrc struct {
	pf vfs.File
	p  PackDriver
}

func (s *PackResSrc) Name() string {
	return s.pf.Name()
}

func (s *PackResSrc) Size() int64 {
	return s.pf.Size()
}
func (s *PackResSrc) Save(in *io.SectionReader) error {
	if f, err := s.p.GetFile(s.pf.Name()); err != nil {
		return fmt.Errorf("[pack] Cannot get file '%s': %v", s.pf.Name(), err)
	} else {
		return f.Copy(in)
	}
}

func GetInstanceHandler(p PackDriver, fileName string) (interface{}, error) {
	f, err := p.GetFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("[pack] Cannot get file '%s': %v", fileName, err)
	}

	r, err := vfs.OpenFileAndGetReader(f, true)
	if err != nil {
		return nil, fmt.Errorf("Cannot get instance of '%s': %v", fileName, err)
	}
	defer f.Close()

	inst, err := CallHandler(&PackResSrc{p: p, pf: f}, r)
	if err != nil {
		return nil, fmt.Errorf("Handler error: %v", err)
	}

	return inst, nil
}
