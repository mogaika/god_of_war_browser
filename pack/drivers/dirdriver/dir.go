package dirdriver

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/mogaika/god_of_war_browser/pack"
)

type DirDriver struct {
	Path           string
	Cache          *pack.InstanceCache
	LastOpenedFile *os.File
}

func (d *DirDriver) GetFileNamesList() []string {
	fileinfos, err := ioutil.ReadDir(d.Path)
	if err != nil {
		log.Printf("[pack] Error getting directory '%s' info: %v", d.Path, err)
		return nil
	}
	result := make([]string, 0)
	for _, f := range fileinfos {
		if !f.IsDir() {
			result = append(result, f.Name())
		}
	}
	return result
}

func (d *DirDriver) GetFile(fileName string) (pack.PackFile, error) {
	info, err := os.Stat(path.Join(d.Path, fileName))
	if err != nil {
		return nil, fmt.Errorf("Error file stat: %v", err)
	}
	return info, nil
}

func (d *DirDriver) GetFileReader(fileName string) (pack.PackFile, *io.SectionReader, error) {
	f, err := os.Open(path.Join(d.Path, fileName))
	if err != nil {
		return nil, nil, fmt.Errorf("Error opening file: %v", err)
	}

	if d.LastOpenedFile != nil {
		d.LastOpenedFile.Close()
	}
	d.LastOpenedFile = f

	info, err := f.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("Error file stat: %v", err)
	}
	return info, io.NewSectionReader(f, 0, info.Size()), nil
}

func (d *DirDriver) GetInstance(fileName string) (interface{}, error) {
	return pack.GetInstanceCachedHandler(d, d.Cache, fileName)
}

func (d *DirDriver) UpdateFile(fileName string, in *io.SectionReader) error {
	d.LastOpenedFile.Close()
	d.LastOpenedFile = nil
	f, err := os.OpenFile(path.Join(d.Path, fileName), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("Error opening file: %v", err)
	}
	defer f.Close()

	d.Cache = &pack.InstanceCache{}

	if _, err := io.Copy(f, in); err != nil {
		return fmt.Errorf("Error when copy data: %v", err)
	}

	return nil
}

func NewPackFromDirectory(dirPath string) (*DirDriver, error) {
	p := &DirDriver{
		Path:  dirPath,
		Cache: &pack.InstanceCache{},
	}
	return p, nil
}
