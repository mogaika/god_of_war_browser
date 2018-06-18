package iso

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mogaika/udf"

	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/vfs"
)

type IsoDriver struct {
	f                vfs.File
	layers           [2]*udf.Udf
	secondLayerStart int64
}

func (iso *IsoDriver) Init(parent vfs.Directory) {}
func (iso *IsoDriver) Name() string              { return iso.f.Name() }
func (iso *IsoDriver) IsDirectory() bool         { return true }

func (iso *IsoDriver) List() ([]string, error) {
	result := make([]string, 0, 48)
	for _, layer := range iso.layers {
		if layer != nil {
			files := layer.ReadDir(nil)
			for i := range files {
				result[i] = files[i].Name()
			}
		}
	}
	return result, nil
}

func (iso *IsoDriver) GetElement(name string) (vfs.Element, error) {
	for _, layer := range iso.layers {
		if layer != nil {
			dir := layer.ReadDir(nil)
			for i := range dir {
				if strings.ToLower(dir[i].Name()) == strings.ToLower(name) {
					return &IsoDriverFile{
						iso: iso,
						f:   &dir[i]}, nil
				}
			}
		}
	}
	return nil, os.ErrNotExist
}
func (iso *IsoDriver) Add(e vfs.Element) error  { panic("Not implemented") }
func (iso *IsoDriver) Remove(name string) error { panic("Not implemented") }
func (iso *IsoDriver) Sync() error {
	if s, ok := iso.f.(vfs.Syncer); ok {
		return s.Sync()
	}
	return nil
}

func (iso *IsoDriver) OpenStreams() error {
	iso.layers[0] = udf.NewUdfFromReader(iso.f)

	var volSizeBuf [4]byte
	// primary volume description sector + offset of volume space size
	if _, err := iso.f.ReadAt(volSizeBuf[:], 0x10*2048+80); err != nil {
		log.Printf("[vfs] [iso] Error when detecting second layer: Read vol size buf error: %v", err)
	} else {
		// minus 16 boot sectors, because they do not replicated over layers (volumes)
		volumeSize := int64(binary.LittleEndian.Uint32(volSizeBuf[:])-16) * utils.SECTOR_SIZE
		if volumeSize+32*utils.SECTOR_SIZE < iso.f.Size() {
			iso.layers[1] = udf.NewUdfFromReader(io.NewSectionReader(iso.f, volumeSize, iso.f.Size()-volumeSize))
			log.Printf("[vfs] [iso] Detected second layer of disk. Start: 0x%x (0x%x)", volumeSize+16*utils.SECTOR_SIZE, volumeSize)
			iso.secondLayerStart = volumeSize
		}
	}
	return nil
}

func NewIsoDriver(f vfs.File) (*IsoDriver, error) {
	iso := &IsoDriver{f: f}
	return iso, iso.OpenStreams()
}

type IsoDriverFile struct {
	iso *IsoDriver
	f   *udf.File
}

func (f *IsoDriverFile) Init(parent vfs.Directory) {}
func (f *IsoDriverFile) Name() string              { return f.f.Name() }
func (f *IsoDriverFile) IsDirectory() bool         { return f.f.IsDir() }
func (f *IsoDriverFile) Size() int64               { return f.f.Size() }
func (f *IsoDriverFile) Open(readonly bool) error  { return nil }
func (f *IsoDriverFile) Close() error              { return f.Sync() }
func (f *IsoDriverFile) Reader() (*io.SectionReader, error) {
	return f.f.NewReader(), nil
}
func (f *IsoDriverFile) ReadAt(b []byte, off int64) (n int, err error) {
	return f.f.NewReader().ReadAt(b, off)
}
func (f *IsoDriverFile) Copy(src io.Reader) error {
	var b bytes.Buffer
	if _, err := io.Copy(&b, src); err != nil {
		return err
	}
	if int64(b.Len()) != f.Size() {
		return fmt.Errorf("[vfs] [iso] Do not support file size changing")
	}
	_, err := f.WriteAt(b.Bytes(), 0)
	return err
}
func (f *IsoDriverFile) WriteAt(b []byte, off int64) (n int, err error) {
	if off+int64(len(b)) > f.Size() {
		return 0, fmt.Errorf("[vfs] [iso] Do not support file size increasing")
	}

	if f.f.Udf == f.iso.layers[1] {
		off += f.iso.secondLayerStart
	}
	return f.iso.f.WriteAt(b, f.f.GetFileOffset()+off)
}
func (f *IsoDriverFile) Sync() error {
	return f.iso.Sync()
}
