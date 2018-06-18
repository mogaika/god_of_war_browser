package vfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	path_ "path"
)

type DirectoryDriver struct {
	path string
}

func (dd *DirectoryDriver) Init(parent Directory) {}

func (dd *DirectoryDriver) Name() string {
	return path_.Base(dd.path)
}

func (dd *DirectoryDriver) IsDirectory() bool {
	return true
}

func (dd *DirectoryDriver) List() ([]string, error) {
	if fileinfos, err := ioutil.ReadDir(dd.path); err != nil {
		return nil, fmt.Errorf("Error getting directory '%s' info: %v", dd.path, err)
	} else {
		result := make([]string, 0, 32)
		for _, f := range fileinfos {
			result = append(result, f.Name())
		}
		return result, nil
	}
}

func (dd *DirectoryDriver) GetElement(name string) (Element, error) {
	newPath := path_.Join(dd.path, name)
	if s, err := os.Stat(newPath); err != nil {
		return nil, fmt.Errorf("Stat error: %v", err)
	} else {
		var e Element
		if s.IsDir() {
			e = NewDirectoryDriver(newPath)
		} else {
			e = NewDirectoryDriverFile(newPath)
		}
		e.Init(dd)
		return e, nil
	}
}

func (dd *DirectoryDriver) Add(e Element) error {
	path := path_.Join(dd.path, e.Name())
	if e.IsDirectory() {
		return os.Mkdir(path, os.ModePerm)
	} else {
		if f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666); err != nil {
			return fmt.Errorf("file '%s' creation failure: %v", path, err)
		} else {
			f.Close()
			return nil
		}
	}
}

func (dd *DirectoryDriver) Remove(name string) error {
	return os.Remove(path_.Join(dd.path, name))
}

func (dd *DirectoryDriver) Path() string {
	return dd.path
}

func NewDirectoryDriver(path string) *DirectoryDriver {
	return &DirectoryDriver{path: path}
}

type DirectoryDriverFile struct {
	path string
	f    *os.File
}

func NewDirectoryDriverFile(path string) *DirectoryDriverFile {
	return &DirectoryDriverFile{
		path: path,
	}
}

func (ddf *DirectoryDriverFile) Init(parent Directory) {
	if dd, ok := parent.(*DirectoryDriver); ok {
		ddf.path = path_.Join(dd.path, path_.Base(ddf.path))
	}
}

func (ddf *DirectoryDriverFile) Name() string {
	return path_.Base(ddf.path)
}

func (ddf *DirectoryDriverFile) IsDirectory() bool {
	return false
}

func (ddf *DirectoryDriverFile) Size() int64 {
	if stat, err := os.Stat(ddf.path); err != nil {
		return 0
	} else {
		return stat.Size()
	}
}

func (ddf *DirectoryDriverFile) Open(readonly bool) error {
	if ddf.f == nil {
		flags := 0
		if readonly {
			flags = os.O_RDONLY
		} else {
			flags = os.O_RDWR
		}

		f, err := os.OpenFile(ddf.path, flags, 0)
		if err != nil {
			return fmt.Errorf("os.Open('%s'): %v", ddf.path, err)
		}
		ddf.f = f
		return nil
	} else {
		return fmt.Errorf("File already opened")
	}
}

func (ddf *DirectoryDriverFile) Close() error {
	if ddf.f != nil {
		if err := ddf.f.Close(); err != nil {
			return fmt.Errorf("os.File.Close(): %v", err)
		}
		ddf.f = nil
	}
	return nil
}

func (ddf *DirectoryDriverFile) Reader() (*io.SectionReader, error) {
	if ddf.f == nil {
		return nil, fmt.Errorf("First you need to open file")
	} else {
		return io.NewSectionReader(ddf.f, 0, ddf.Size()), nil
	}
}

func (ddf *DirectoryDriverFile) ReadAt(b []byte, off int64) (n int, err error) {
	if ddf.f == nil {
		return 0, fmt.Errorf("First you need to open file")
	} else {
		return ddf.f.ReadAt(b, off)
	}
}

func (ddf *DirectoryDriverFile) Copy(src io.Reader) error {
	ddf.Close()

	f, err := os.Create(ddf.path)
	if err != nil {
		return fmt.Errorf("os.Create('%s'): %v", ddf.path, err)
	}
	if _, err := io.Copy(f, src); err != nil {
		return fmt.Errorf("os.Copy(...): %v", err)
	}
	f.Close()
	return nil
}

func (ddf *DirectoryDriverFile) WriteAt(b []byte, off int64) (n int, err error) {
	if ddf.f == nil {
		return 0, fmt.Errorf("First you need to open file")
	} else {
		return ddf.f.WriteAt(b, off)
	}
}

func (ddf *DirectoryDriverFile) Sync() error {
	if ddf.f == nil {
		return fmt.Errorf("First you need to open file")
	} else {
		return ddf.f.Sync()
	}
}
