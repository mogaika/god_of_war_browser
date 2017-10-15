package utils

import (
	"io"
	"os"
)

type ReaderWriterAt interface {
	io.ReaderAt
	io.WriterAt
	Size() int64
}

type ReaderWriterAtImplementation struct {
	r    io.ReaderAt
	w    io.WriterAt
	size int64
}

func (rwi *ReaderWriterAtImplementation) Size() int64 {
	return rwi.size
}

func (rwi *ReaderWriterAtImplementation) WriteAt(p []byte, off int64) (n int, err error) {
	return rwi.w.WriteAt(p, off)
}

func (rwi *ReaderWriterAtImplementation) ReadAt(p []byte, off int64) (n int, err error) {
	return rwi.r.ReadAt(p, off)
}

func NewReaderWriterAt(r io.ReaderAt, w io.WriterAt, size int64) *ReaderWriterAtImplementation {
	return &ReaderWriterAtImplementation{r: r, w: w}
}

func NewReaderWriterAtFromFile(f *os.File) *ReaderWriterAtImplementation {
	if s, err := f.Stat(); err != nil {
		panic(err)
	} else {
		return &ReaderWriterAtImplementation{r: f, w: f, size: s.Size()}
	}
}
