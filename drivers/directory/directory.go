package directory

import "github.com/mogaika/god_of_war_browser/pack"
import "github.com/mogaika/god_of_war_browser/vfs"

type Directory struct {
	dd vfs.Directory
}

func NewDirectoryDriver(dd vfs.Directory) *Directory {
	return &Directory{dd: dd}
}

func (d *Directory) GetFileNamesList() []string {
	l, _ := d.dd.List()
	return l
}

func (d *Directory) GetFile(fileName string) (vfs.File, error) {
	el, err := d.dd.GetElement(fileName)
	return el.(vfs.File), err
}

func (d *Directory) GetInstance(fileName string) (interface{}, error) {
	return pack.GetInstanceHandler(d, fileName)
}
