package tok

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/mogaika/god_of_war_browser/utils"
)

var ZEROENTRY [0xc]byte

const (
	ENTRY_SIZE       = 24
	PARTS_COUNT      = 2
	FILE_NAME        = "GODOFWAR.TOC"
	SANITY_FILE_NAME = "SANITY.TXT"
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

func ParseEntiesArray(entries []Entry) map[string]*File {
	files := make(map[string]*File)
	for _, e := range entries {
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
			log.Printf("[tok] Tok file corrupted! Finded same file but with different size! '%s' %d!=%d", e.Name, e.Size, file.Size())
		}

		file.Encounters = append(file.Encounters, e.Enc)
	}
	return files
}

func ParseFiles(tokStream io.Reader) (map[string]*File, []Entry, error) {
	var buffer [ENTRY_SIZE]byte
	entries := make([]Entry, 0)

	for {
		if _, err := tokStream.Read(buffer[:]); err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}

		e := UnmarshalTokEntry(buffer[:])
		if e.Name == "" {
			break
		}

		entries = append(entries, e)
	}
	return ParseEntiesArray(entries), entries, nil
}

// Copy data between or packs inside one pack
// Allowing overlap destination and source
// size parameter must be in bytes count, not in sectors count
func CopyDataBetweenPackStreams(to FileEncounter, from FileEncounter, size int64, partStreams [PARTS_COUNT]utils.ReaderWriterAt) error {
	if to.Pack == from.Pack && to.Start == from.Start {
		// Data already in place
		return nil
	}

	sizeInSectors := utils.GetRequiredSectorsCount(size)

	var bunchSize int64
	if to.Pack != from.Pack || to.Start+sizeInSectors*utils.SECTOR_SIZE < from.Start {
		// There is no overlaps, just copy data
		// using bunches of 512 sectors
		bunchSize = 512 * utils.SECTOR_SIZE
	} else {
		// Source and target overlapped
		// Copy sector by sector
		bunchSize = 1
	}

	var realBuf = make([]byte, bunchSize*utils.SECTOR_SIZE)
	var pos int64 = 0
	var left int64 = size
	for left > 0 {
		buf := realBuf[:]
		if left < int64(len(buf)) {
			buf = buf[:left]
		}
		if _, err := partStreams[from.Pack].ReadAt(buf, from.Start+pos); err != nil {
			return fmt.Errorf("Error when reading (size:%v): %v", len(buf), err)
		}
		if writtenbytes, err := partStreams[to.Pack].WriteAt(buf, to.Start+pos); err != nil {
			return fmt.Errorf("Error when writing (size:%v): %v", len(buf), err)
		} else {
			pos += int64(writtenbytes)
			left -= int64(writtenbytes)
		}
	}
	return nil
}

// Removes files duplicates (except sanity.txt)
// Shrink files in pack
// Move files in begin of packs (first pack have priority over next)
// Reading on dvd disk ps2 version slowed after shrinkig :(
func shrinkPackFiles(originalToksEntries []Entry, partStreams [PARTS_COUNT]utils.ReaderWriterAt) ([]Entry, error) {
	var streamOffsetsSectors [PARTS_COUNT]int64
	var partsTokens [PARTS_COUNT][]Entry
	for i := range partsTokens {
		partsTokens[i] = make([]Entry, 0)
	}
	alreadyProcessedFiles := make(map[string]Entry)

	for _, oldE := range originalToksEntries {
		if _, already := alreadyProcessedFiles[oldE.Name]; !already || oldE.Name == SANITY_FILE_NAME {
			// find place where we can move file
			targetPack := -1
			for i := range partStreams {
				if partStreams[i].Size()-streamOffsetsSectors[i]*utils.SECTOR_SIZE >= oldE.Size {
					targetPack = i
					break
				}
			}
			if targetPack == -1 {
				return nil, fmt.Errorf("Cannot find place for file. Size: %v, StreamOffsets: %v", oldE.Size, streamOffsetsSectors)
			}

			e := oldE
			e.Enc.Pack = targetPack
			e.Enc.Start = streamOffsetsSectors[targetPack] * utils.SECTOR_SIZE

			if err := CopyDataBetweenPackStreams(e.Enc, oldE.Enc, e.Size, partStreams); err != nil {
				return nil, fmt.Errorf("Error when copying data between packs (file %v): %v", oldE.Name, err)
			}

			streamOffsetsSectors[targetPack] += utils.GetRequiredSectorsCount(e.Size)
			alreadyProcessedFiles[oldE.Name] = e
			partsTokens[targetPack] = append(partsTokens[targetPack], e)
		}
	}

	resultTokens := make([]Entry, 0)
	for _, partTokens := range partsTokens {
		resultTokens = append(resultTokens, partTokens...)
	}
	return resultTokens, nil
}

func updateFileWithIncreacingSize(fTokOriginal io.Reader, fTokNew io.Writer, partStreams [PARTS_COUNT]utils.ReaderWriterAt, filename string, in *io.SectionReader) error {
	log.Println("Updating tok+parts with increacing required sectors count")
	_, originalEntries, err := ParseFiles(fTokOriginal)
	if err != nil {
		return fmt.Errorf("Error when parsing tok: %v", err)
	}

	// delete our file from entries array
	entriesWithoutOurFile := make([]Entry, 0)
	for _, e := range originalEntries {
		if e.Name != filename {
			entriesWithoutOurFile = append(entriesWithoutOurFile, e)
		}
	}

	findPlaceInPartWithSuchFreeSectors := func(entries []Entry, sectors int64) *FileEncounter {
		// return last entry in pack
		var packLasts [PARTS_COUNT]int64
		for iPack := range partStreams {
			packLasts[iPack] = int64(0)
			for _, e := range entries {
				if e.Enc.Pack == iPack {
					// log.Printf("--------------: %v;  lastSector %v sectors %v < diff %v", e, packLasts[iPack], sectors, e.Enc.Start/utils.SECTOR_SIZE-packLasts[iPack])
					if e.Enc.Start/utils.SECTOR_SIZE-packLasts[iPack] >= sectors {
						return &FileEncounter{Pack: iPack, Start: packLasts[iPack] * utils.SECTOR_SIZE}
					}
					packLasts[iPack] = e.Enc.Start/utils.SECTOR_SIZE + utils.GetRequiredSectorsCount(e.Size)
				}
			}
		}
		for iPack, packStream := range partStreams {
			log.Println("<= Detecting free space in ", iPack, " pack =>")
			log.Println("Pack stream size: ", packStream.Size())
			log.Println("Pack stream sectors: ", packStream.Size()/utils.SECTOR_SIZE)
			log.Println("Free pack stream sectors: ", packStream.Size()/utils.SECTOR_SIZE-packLasts[iPack])
			log.Println("Needed sectors: ", sectors)
			if packStream.Size()/utils.SECTOR_SIZE-packLasts[iPack] >= sectors {
				return &FileEncounter{Pack: iPack, Start: packLasts[iPack] * utils.SECTOR_SIZE}
			}
		}
		return nil
	}

	newEncounter := findPlaceInPartWithSuchFreeSectors(entriesWithoutOurFile, utils.GetRequiredSectorsCount(in.Size()))
	if newEncounter == nil {
		log.Println("Cannot find place for new file, trying to shrink...")
		entries, err := shrinkPackFiles(entriesWithoutOurFile, partStreams)
		if err != nil {
			return fmt.Errorf("Error when shrinking: %v", err)
		} else {
			entriesWithoutOurFile = entries
		}
		newEncounter = findPlaceInPartWithSuchFreeSectors(entriesWithoutOurFile, utils.GetRequiredSectorsCount(in.Size()))
		if newEncounter == nil {
			return fmt.Errorf("Cannot find free place even after shrinking")
		}
	}
	log.Println("Place for new file finded at", newEncounter, *newEncounter)

	buf := make([]byte, in.Size())
	if _, err := in.Read(buf); err != nil {
		return fmt.Errorf("Error when reading new file: %v", err)
	}
	log.Println("Writing ", len(buf), "bytes at ", newEncounter.Start)
	if _, err := partStreams[newEncounter.Pack].WriteAt(buf, newEncounter.Start); err != nil {
		return fmt.Errorf("Error when writing new file: %v", err)
	}

	writed := false
	marshaledEntry := MarshalTokEntry(&Entry{
		Name: filename,
		Size: in.Size(),
		Enc:  *newEncounter})
	for _, e := range entriesWithoutOurFile {
		if !writed {
			// Write our file entry to tok file
			if e.Enc.Pack > newEncounter.Pack || (e.Enc.Pack == newEncounter.Pack && e.Enc.Start > newEncounter.Start) {
				if _, err := fTokNew.Write(marshaledEntry); err != nil {
					return fmt.Errorf("Error when writing new tok entry: %v", err)
				}
				writed = true
			}
		}
		if _, err := fTokNew.Write(MarshalTokEntry(&e)); err != nil {
			return fmt.Errorf("Error when writing old tok entry: %v", err)
		}
	}
	if !writed {
		if _, err := fTokNew.Write(marshaledEntry); err != nil {
			return fmt.Errorf("Error when writing new tok entry at end of file: %v", err)
		}
	}

	if _, err := fTokNew.Write(ZEROENTRY[:]); err != nil {
		return fmt.Errorf("Error when writing last zeros in tok: %v", err)
	}

	return nil
}

// do not reorder files in pack
func updateFileWithoutIncreacingSize(fTokOriginal io.Reader, fTokNew io.Writer, partStreams [PARTS_COUNT]utils.ReaderWriterAt, filename string, in *io.SectionReader) error {
	log.Println("Updating tok+parts without increacing required sectors count")
	// update sizes in tok file, if changed
	files, entries, err := ParseFiles(fTokOriginal)
	if err != nil {
		return fmt.Errorf("Error when parsing tok: %v", err)
	}
	f := files[filename]
	if in.Size() != f.Size() {
		log.Println("Updating file size because its changed.")
		for _, e := range entries {
			if e.Name == filename {
				e.Size = in.Size()
			}
			fTokNew.Write(MarshalTokEntry(&e))
		}
	}
	log.Println("Reading file")
	var fileBuffer bytes.Buffer
	if _, err := io.Copy(&fileBuffer, in); err != nil {
		return err
	}
	log.Println("Changing encounters")
	for iPart, fPart := range partStreams {
		for _, enc := range f.Encounters {
			if enc.Pack == iPart {
				if fPart == nil {
					return fmt.Errorf("For updating file required all pack streams that contain this file. Stream %d = nil", iPart)
				}
				log.Println(enc)
				if _, err := fPart.WriteAt(fileBuffer.Bytes(), enc.Start); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func UpdateFile(fTokOriginal io.Reader, fTokNew io.Writer, partStreams [PARTS_COUNT]utils.ReaderWriterAt, f *File, in *io.SectionReader) error {
	if in.Size()/utils.SECTOR_SIZE > f.Size()/utils.SECTOR_SIZE {
		return updateFileWithIncreacingSize(fTokOriginal, fTokNew, partStreams, f.Name(), in)
	} else {
		return updateFileWithoutIncreacingSize(fTokOriginal, fTokNew, partStreams, f.Name(), in)
	}
}
