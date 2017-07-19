package tok

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	ENTRY_SIZE  = 24
	PARTS_COUNT = 2
	FILE_NAME   = "GODOFWAR.TOC"
)

type File struct {
	name       string
	size       int64
	Encounters []FileEncounter
}

type FileEncounter struct {
	Pack  int
	Start int64
}

func (tf *File) Name() string {
	return tf.name
}

func (tf *File) Size() int64 {
	return tf.size
}

type Entry struct {
	Name string
	Size int64
	Enc  FileEncounter
}

func GenPartFileName(partIndex int) string {
	return fmt.Sprintf("PART%d.PAK", partIndex+1)
}

func UnmarshalTokEntry(buffer []byte) Entry {
	return Entry{
		Name: utils.BytesToString(buffer[0:12]),
		Size: int64(binary.LittleEndian.Uint32(buffer[16:20])),
		Enc: FileEncounter{
			Pack:  int(binary.LittleEndian.Uint32(buffer[12:16])),
			Start: int64(binary.LittleEndian.Uint32(buffer[20:24])) * utils.SECTOR_SIZE,
		}}
}

func MarshalTokEntry(e *Entry) []byte {
	buf := make([]byte, ENTRY_SIZE)
	copy(buf[:12], utils.StringToBytes(e.Name, 12, false))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(e.Enc.Pack))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(e.Size))
	binary.LittleEndian.PutUint32(buf[20:24], uint32((e.Enc.Start+utils.SECTOR_SIZE-1)/utils.SECTOR_SIZE))
	return buf
}

func ParseFiles(tokStream io.Reader) (map[string]*File, error) {
	var buffer [ENTRY_SIZE]byte

	files := make(map[string]*File)

	for {
		if _, err := tokStream.Read(buffer[:]); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		e := UnmarshalTokEntry(buffer[:])
		if e.Name == "" {
			break
		}

		var file *File
		if existFile, ok := files[e.Name]; ok {
			file = existFile
		} else {
			file = &File{
				name:       e.Name,
				size:       e.Size,
				Encounters: make([]FileEncounter, 0),
			}
			files[e.Name] = file
		}

		if e.Size != file.Size() {
			log.Printf("[tok] Tok file corrupted! Finded same file but with different size! '%s' %d!=%d", e.Name, e.Size, file.Size)
		}

		file.Encounters = append(file.Encounters, e.Enc)
	}
	return files, nil
}

func UpdateFile(fTok io.ReadWriteSeeker, partStreams [PARTS_COUNT]io.WriterAt, f *File, in *io.SectionReader) error {
	if in.Size()/utils.SECTOR_SIZE > f.Size()/utils.SECTOR_SIZE {
		return fmt.Errorf("Size increase above sector boundary is not supported yet")
	}

	// update sizes in tok file, if changed
	if in.Size() != f.Size() {
		var buf []byte
		for {
			if _, err := fTok.Read(buf[:]); err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			e := UnmarshalTokEntry(buf[:])
			if e.Name == "" {
				break
			}
			if e.Size != f.Size() {
				log.Printf("[pack] Warning! Tok entry '%s': incorrect file size, file may be unconsistent: %d != %d",
					e.Name, e.Size, f.Size())
			}
			if e.Name == f.Name() {
				if _, err := fTok.Seek(-ENTRY_SIZE, os.SEEK_CUR); err != nil {
					return err
				}
				e.Size = in.Size()
				if _, err := fTok.Write(MarshalTokEntry(&e)); err != nil {
					return err
				}
			}
		}
	}

	var fileBuffer bytes.Buffer
	if _, err := io.Copy(&fileBuffer, in); err != nil {
		return err
	}

	for iPart, fPart := range partStreams {
		for _, enc := range f.Encounters {
			if enc.Pack == iPart {
				if fPart == nil {
					return fmt.Errorf("For updating file required all part streams that contain this file. Stream %d = nil", iPart)
				}
				if _, err := fPart.WriteAt(fileBuffer.Bytes(), enc.Start); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
