package toc

import (
	"fmt"
	"io"

	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/vfs"
)

const (
	PACK_ADDR_UNKNOWN  = iota
	PACK_ADDR_ABSOLUTE // absolute offset to file relative to pak1
	PACK_ADDR_INDEX    // size of previous paks + offset
)

type PaksArray struct {
	paks     []vfs.File
	addrType int
}

func NewPaksArray(paks []vfs.File, addrType int) *PaksArray {
	pa := &PaksArray{
		paks:     make([]vfs.File, len(paks)),
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
				return n, fmt.Errorf("[pak] Absolute readwrite error: %v", err)
			}
			if int64(processedN) != processAmount {
				return n, fmt.Errorf("[pak] Absolute readwrite N calculation error: %v", err)
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

func (pa *PaksArray) NewReaderWriter(e Encounter) *EncounterReaderWriter {
	return &EncounterReaderWriter{pa: pa, e: e}
}

func (pa *PaksArray) Move(from, to Encounter) error {
	if from.Size != to.Size {
		return fmt.Errorf("[pak] Wrong size amount %d != %d", from.Size, to.Size)
	}

	if from.Pak == to.Pak && from.Offset == to.Offset {
		return nil
	}

	frw := pa.NewReaderWriter(from)
	trw := pa.NewReaderWriter(to)

	sizeInSectors := utils.GetRequiredSectorsCount(from.Size)
	sizeInBytes := sizeInSectors * utils.SECTOR_SIZE

	forwardCopy := true
	bunchSize := int64(512)
	if to.Pak == from.Pak {
		// if we collide, then use memmove logic
		if from.Offset < to.Offset+sizeInBytes && from.Offset+sizeInBytes >= to.Offset {
			bunchSize = 1
			if to.Offset > from.Offset {
				forwardCopy = false
			}
		}
	}

	bigBuffer := make([]byte, bunchSize*utils.SECTOR_SIZE)

	pos := int64(0)
	if !forwardCopy {
		pos += sizeInBytes
	}

	left := sizeInSectors
	for left != int64(0) {
		readSectorsAmount := left
		if left > bunchSize {
			left = bunchSize
		}
		readBytesAmount := readSectorsAmount * utils.SECTOR_SIZE

		// if we do reverse copy, then pos pointed to upper bound
		ioOffset := pos
		if !forwardCopy {
			ioOffset -= readBytesAmount
		}

		b := bigBuffer[:readBytesAmount]
		if _, err := frw.ReadAt(b, ioOffset); err != nil {
			return fmt.Errorf("[pak] Move() ReadAt (forwardCopy: %v, bunchSize: %v) error: %v",
				forwardCopy, bunchSize, err)
		}
		if _, err := trw.WriteAt(b, ioOffset); err != nil {
			return fmt.Errorf("[pak] Move() WriteAt (forwardCopy: %v, bunchSize: %v) error: %v",
				forwardCopy, bunchSize, err)
		}

		if forwardCopy {
			pos += readBytesAmount
		} else {
			pos -= readBytesAmount
		}
		left -= readSectorsAmount
	}

	return nil
}
