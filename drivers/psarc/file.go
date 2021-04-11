package psarc

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/mogaika/god_of_war_browser/vfs"
)

type File struct {
	p   *Psarc
	e   Entry
	buf *bytes.Buffer
}

func (f *File) initBuf() error {
	buf := &bytes.Buffer{}

	if f.e.OriginalSize == 0 {
		return nil
	}

	blockIndex := f.e.BlockListStart
	compressedFileOffset := int64(0)
	compressedBuf := make([]byte, f.p.h.BlockSize)

	for int64(buf.Len()) < f.e.OriginalSize {
		compressedArchiveOffset := f.e.StartOffset + compressedFileOffset
		compressedBlockSize := f.p.blockSizes[blockIndex]

		if compressedBlockSize == 0 {
			if _, err := f.p.f.ReadAt(compressedBuf, compressedArchiveOffset); err != nil {
				return fmt.Errorf("[psarc.File.initBuf blockSize=0 ReadAt] %v", err)
			}

			if _, err := buf.Write(compressedBuf); err != nil {
				panic(err)
			}

			compressedFileOffset += int64(f.p.h.BlockSize)
		} else {
			compressedBlock := compressedBuf[:compressedBlockSize]
			if _, err := f.p.f.ReadAt(compressedBlock, compressedArchiveOffset); err != nil {
				return fmt.Errorf("[psarc.File.initBuf blockSize!=0 ReadAt] %v", err)
			}

			if config.GetPlayStationVersion() == config.PSVita &&
				// last block marked as compressed to take less space?
				int64(buf.Len())+int64(compressedBlockSize) == f.e.OriginalSize {
				if _, err := buf.Write(compressedBlock); err != nil {
					panic(err)
				}
			} else {
				// TODO: improve something here pls
				if compressedBlock[0] != 0x78 || compressedBlock[1] != 0xda {
					panic(fmt.Sprintf("incorrect compression header: 0x%x 0x%x 0x%x 0x%x",
						compressedBlock[0], compressedBlock[1], compressedBlock[2], compressedBlock[3]))
				}

				if zr, err := zlib.NewReader(bytes.NewReader(compressedBlock)); err != nil {
					panic(err)
				} else {
					if _, err := io.Copy(buf, zr); err != nil {
						panic(err)
					}
				}
			}

			compressedFileOffset += int64(compressedBlockSize)
		}

		blockIndex++
	}
	if int64(buf.Len()) != f.e.OriginalSize {
		return fmt.Errorf("File sizes not equals: %v != %v", buf.Len(), f.e.OriginalSize)
	}
	f.buf = buf
	return nil
}

// interface vfs.Element
func (f *File) Init(parent vfs.Directory) {}
func (f *File) Name() string              { return f.e.Name }
func (f *File) IsDirectory() bool         { return false }

// interface vfs.File
func (f *File) Size() int64 { return f.e.OriginalSize }
func (f *File) Open(readonly bool) error {
	if readonly != true {
		return fmt.Errorf("[psarc.File.Open] write not supported")
	}
	if f.buf == nil {
		return f.initBuf()
	} else {
		return nil
	}
}
func (f *File) Close() error {
	f.buf = nil
	return nil
}
func (f *File) Reader() (*io.SectionReader, error) {
	return io.NewSectionReader(bytes.NewReader(f.buf.Bytes()), 0, f.e.OriginalSize), nil
}
func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	return copy(b, f.buf.Bytes()[off:]), nil
}
func (f *File) Copy(src io.Reader) error                       { panic("read-only") }
func (f *File) WriteAt(b []byte, off int64) (n int, err error) { panic("read-only") }
