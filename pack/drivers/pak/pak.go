package pak

import (
	"fmt"
	"io"
)

const (
	PACK_ADDR_ABSOLUTE = iota // absolute offset of file relative to pak1
	PACK_ADDR_INDEX           // size of previous paks + offset
)

type Pak interface {
	io.ReaderAt
	io.WriterAt
	Size() int64
	Sync() error
}

type PaksArray struct {
	paks     []Pak
	addrType int
}

func NewPaksArray(paks []Pak, addrType int) *PaksArray {
	pa := &PaksArray{
		paks:     make([]Pak, len(paks)),
		addrType: addrType,
	}
	copy(pa.paks, paks)
	return pa
}
func (pa *PaksArray) ReadAt(p []byte, off int64) (n int, err error) {
	return pa.absoluteReadWriteAt(p, off, true)
}

func (pa *PaksArray) WeadAt(p []byte, off int64) (n int, err error) {
	return pa.absoluteReadWriteAt(p, off, false)
}

func (pa *PaksArray) absoluteReadWriteAt(p []byte, off int64, doRead bool) (n int, err error) {
	estimatedBytes := int64(len(p))
	bufOff := int64(0)
	for _, pak := range pa.paks {
		pakSize := pak.Size()
		if off < pakSize {
			leftToPocess := pakSize - off
			processAmount := estimatedBytes
			if processAmount > leftToPocess {
				processAmount = leftToPocess
			}

			processedN := 0
			if doRead {
				processedN, err = pak.ReadAt(p[bufOff:bufOff+processAmount], off)
			} else {
				processedN, err = pak.WriteAt(p[bufOff:bufOff+processAmount], off)
			}
			if err != nil {
				return n, fmt.Errorf("[pak] absolute readwrite error: %v", err)
			}
			if int64(processedN) != processAmount {
				return n, fmt.Errorf("[pak] absolute readwrite N calculation error: %v", err)
			}
			off += processAmount
		}
		if estimatedBytes == 0 {
			return n, nil
		}
		off -= pakSize
	}
	return n, io.EOF
}

func (pa *PaksArray) NewReader(e Encounter) *PakEncounterReader {
	return &PakEncounterReader{pa: pa, e: e}
}

func (pa *PaksArray) Copy(from, to Encounter) error {
	if from.Size != to.Size {
		return fmt.Errorf("[pak] Wrong size amount %d != %d", from.Size, to.Size)
	}

}
