package vfs

import (
	"fmt"
	"io"
)

func OpenFileAndGetReader(f File, readonly bool) (*io.SectionReader, error) {
	if err := f.Open(readonly); err != nil {
		return nil, fmt.Errorf("Cannot open file '%s': %v", f.Name(), err)
	} else {
		if r, err := f.Reader(); err != nil {
			defer f.Close()
			return nil, fmt.Errorf("Cannot get file '%s' reader: %v", f.Name(), err)
		} else {
			return r, err
		}
	}
}

func OpenFileAndCopy(f File, src io.Reader) error {
	if err := f.Open(false); err != nil {
		return fmt.Errorf("Cannot open file '%s': %v", f.Name(), err)
	} else {
		defer f.Close()
		if err := f.Copy(src); err != nil {
			return fmt.Errorf("Cannot copy data to file '%s': %v", f.Name(), err)
		} else {
			return nil
		}
	}
}

func DirectoryGetFile(d Directory, name string) (File, error) {
	if f, err := d.GetElement(name); err != nil {
		return nil, fmt.Errorf("Cannot open file '%s': %v", name, err)
	} else if f.IsDirectory() {
		return nil, fmt.Errorf("File '%s' is directory, not a file!", name)
	} else {
		return f.(File), nil
	}
}
