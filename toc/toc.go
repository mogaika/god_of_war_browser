package toc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/utils"
)

var ZEROENTRY [0xc]byte
var tocFileNameOverride string
var packFileNamePrefix string
var packFileNamePostfix string
var packFileNameUseIndexing bool

const (
	GOW1_ENTRY_SIZE  = 24
	GOW2_ENTRY_SIZE  = 36
	TOC_FILE_NAME    = "GODOFWAR.TOC"
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

func TocNameOverride(newName string) {
	tocFileNameOverride = newName
}
func PartNamePrefix(newPrefix string) {
	packFileNamePrefix = newPrefix
}

func PartNamePostfix(newPostfix string) {
	packFileNamePostfix = newPostfix
}

func PartNameUseIndexing(useIndexing bool) {
	packFileNameUseIndexing = useIndexing
}

func GetTocFileName() string {
	if tocFileNameOverride != "" {
		return tocFileNameOverride
	} else {
		return TOC_FILE_NAME
	}
}

func GenPartFileName(partIndex int) string {
	if packFileNameUseIndexing {
		return fmt.Sprintf("%s%d%s", packFileNamePrefix, partIndex+1, packFileNamePostfix)
	} else {
		if partIndex > 0 {
			log.Println("WARNING: Your toc use multi-part part storage, but you provide -partindexing=false flag that prevent multi-part usage")
		}
		return packFileNamePrefix + packFileNamePostfix
	}
}

func unmarshalTocEntryGOW1(buffer []byte) Entry {
	return Entry{
		Name: utils.BytesToString(buffer[0:12]),
		Size: int64(binary.LittleEndian.Uint32(buffer[16:20])),
		Enc: FileEncounter{
			Pack:  int(binary.LittleEndian.Uint32(buffer[12:16])),
			Start: int64(binary.LittleEndian.Uint32(buffer[20:24])) * utils.SECTOR_SIZE,
		}}
}

func marshalTocEntryGOW1(e *Entry) []byte {
	buf := make([]byte, GOW1_ENTRY_SIZE)
	copy(buf[:12], utils.StringToBytesBuffer(e.Name, 12, false))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(e.Enc.Pack))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(e.Size))
	binary.LittleEndian.PutUint32(buf[20:24], uint32((e.Enc.Start+utils.SECTOR_SIZE-1)/utils.SECTOR_SIZE))
	return buf
}

func parseEntiesArrayToFiles(entries []Entry) map[string]*File {
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
			log.Printf("[toc] Toc file corrupted! Finded same file but with different size! '%s' %d!=%d", e.Name, e.Size, file.Size())
		}

		file.Encounters = append(file.Encounters, e.Enc)
	}
	return files
}

func parseFilesToEntriesArray(files map[string]*File) []Entry {
	entries := make([]Entry, 0, len(files))
	for fileName, f := range files {
		for _, e := range f.Encounters {
			entries = append(entries, Entry{
				Name: fileName,
				Size: f.size,
				Enc:  e,
			})
		}
	}
	sort.Slice(entries, func(i int, j int) bool {
		return entries[i].Enc.Pack < entries[j].Enc.Pack || entries[i].Enc.Start < entries[j].Enc.Start
	})

	for _, e := range entries {
		log.Printf("entry %24s at 0x%.12x size 0x%.8x", e.Name, e.Enc.Start, e.Size)
	}

	return entries
}

func parseFilesGOW1(tocStream io.Reader) (map[string]*File, []Entry, error) {
	var buffer [GOW1_ENTRY_SIZE]byte
	entries := make([]Entry, 0)

	for {
		if _, err := tocStream.Read(buffer[:]); err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}

		e := unmarshalTocEntryGOW1(buffer[:])
		if e.Name == "" {
			break
		}

		entries = append(entries, e)
	}
	return parseEntiesArrayToFiles(entries), entries, nil
}

func parseFilesGOW2(tocStream io.ReadSeeker) (map[string]*File, []Entry, error) {
	var countOfFiles uint32
	if err := binary.Read(tocStream, binary.LittleEndian, &countOfFiles); err != nil {
		return nil, nil, fmt.Errorf("[gow2] Cannot read files count: %v", err)
	}

	filesBuffer := make([]byte, GOW2_ENTRY_SIZE*countOfFiles)
	if _, err := tocStream.Read(filesBuffer); err != nil {
		return nil, nil, fmt.Errorf("[gow2] Error reading files buffer: %v", err)
	}

	savedSpace := int64(0)

	offsetsArrayOffset, err := tocStream.Seek(0, os.SEEK_CUR)
	if err != nil {
		return nil, nil, fmt.Errorf("[gow2] Cannot get position of array offsets: %v", err)
	}

	files := make(map[string]*File)
	for iFile := uint32(0); iFile < countOfFiles; iFile++ {
		currentFileBuf := filesBuffer[iFile*GOW2_ENTRY_SIZE:]
		currentFile := &File{
			name:       utils.BytesToString(currentFileBuf),
			size:       int64(binary.LittleEndian.Uint32(currentFileBuf[0x18:])),
			Encounters: make([]FileEncounter, binary.LittleEndian.Uint32(currentFileBuf[0x1c:])),
		}
		encountersStartIndex := binary.LittleEndian.Uint32(currentFileBuf[0x20:])
		if len(currentFile.Encounters) > 1 {
			savedSpace += (int64(len(currentFile.Encounters)) - 1) * (((currentFile.size + utils.SECTOR_SIZE - 1) / utils.SECTOR_SIZE) * utils.SECTOR_SIZE)
		}
		for iEncounter := range currentFile.Encounters {
			if _, err := tocStream.Seek(offsetsArrayOffset+int64(encountersStartIndex)*4+int64(iEncounter)*4, os.SEEK_SET); err != nil {
				return nil, nil, fmt.Errorf("[gow2] Error when reading encounter %d(%d) of file %d '%s': %v", int(encountersStartIndex)+iEncounter, iEncounter, iFile, currentFile.name)
			}
			var fileOffset uint32
			binary.Read(tocStream, binary.LittleEndian, &fileOffset)
			currentFile.Encounters[iEncounter].Pack = -1
			currentFile.Encounters[iEncounter].Start = int64(fileOffset) * utils.SECTOR_SIZE
		}

		files[currentFile.Name()] = currentFile
	}

	log.Printf("SAVED SPACE: %db %x %dkB %dMb", savedSpace, savedSpace, savedSpace/1024, savedSpace/1024/1024)

	return files, parseFilesToEntriesArray(files), nil
}

func ParseFiles(tocStream io.ReadSeeker) (map[string]*File, []Entry, error) {
	if config.GodOfWarVersion == config.GOWunknown {
		var buf [4]byte
		if _, err := tocStream.Read(buf[:]); err != nil {
			return nil, nil, fmt.Errorf("Error when detecting gow version: %v", err)
		}
		if buf[2] != 0 {
			log.Println("[toc] Detected gow version: GOW1ps1")
			config.SetGOWVersion(config.GOW1ps2)
		} else {
			log.Println("[toc] Detected gow version: GOW2ps1")
			config.SetGOWVersion(config.GOW2ps2)
		}
		tocStream.Seek(0, os.SEEK_SET)
	}
	if config.GodOfWarVersion == config.GOW2ps2 {
		return parseFilesGOW2(tocStream)
	} else {
		return parseFilesGOW1(tocStream)
	}
}

func GetPacksCount(entries []Entry) int {
	packs := 0
	for i := range entries {
		if entries[i].Enc.Pack >= packs {
			packs = entries[i].Enc.Pack + 1
		}
	}
	return packs
}

var dataCopyPreallocatedBuffer [512 * utils.SECTOR_SIZE]byte

// Copy data between or packs inside one pack
// Allowing overlap destination and source
// size parameter must be in bytes count, not in sectors count
func CopyDataBetweenPackStreams(to FileEncounter, from FileEncounter, size int64, partStreams []utils.ReaderWriterAt) error {
	if to.Pack == from.Pack && to.Start == from.Start {
		// Data already in place
		return nil
	}

	sizeInSectors := utils.GetRequiredSectorsCount(size)

	var bunchSize int64
	if to.Pack != from.Pack || to.Start+sizeInSectors*utils.SECTOR_SIZE < from.Start {
		// There is no overlaps, just copy data
		// using bunches of 512 sectors
		bunchSize = 512
	} else {
		// Source and target overlapped
		// Copy sector by sector
		bunchSize = 1
	}

	var realBuf = dataCopyPreallocatedBuffer[:bunchSize*utils.SECTOR_SIZE]
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
func shrinkPackFiles(originalTocsEntries []Entry, partStreams []utils.ReaderWriterAt) ([]Entry, error) {
	streamOffsetsSectors := make([]int64, len(partStreams))
	partsTocens := make([][]Entry, len(partStreams))
	for i := range partsTocens {
		partsTocens[i] = make([]Entry, 0)
	}
	alreadyProcessedFiles := make(map[string]Entry)

	for _, oldE := range originalTocsEntries {
		if _, already := alreadyProcessedFiles[oldE.Name]; !already || oldE.Name == SANITY_FILE_NAME {
			log.Println("Shrinking: ", oldE.Name)
			// find place where we can move file
			targetPack := -1
			for i := range partStreams {
				if partStreams[i] != nil && partStreams[i].Size()-streamOffsetsSectors[i]*utils.SECTOR_SIZE >= oldE.Size {
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
			partsTocens[targetPack] = append(partsTocens[targetPack], e)
		} else {
			log.Println("Already shrinked: ", oldE.Name)
		}
	}

	resultTocens := make([]Entry, 0)
	for _, partTocens := range partsTocens {
		resultTocens = append(resultTocens, partTocens...)
	}
	return resultTocens, nil
}

func marshalTocFileGOW1(fTocNew io.Writer, entries []Entry) error {
	for _, e := range entries {
		if _, err := fTocNew.Write(marshalTocEntryGOW1(&e)); err != nil {
			return fmt.Errorf("Error when writing toc entry %v: %v", e, err)
		}
	}
	if _, err := fTocNew.Write(ZEROENTRY[:]); err != nil {
		return fmt.Errorf("Error when writing last zeros in toc: %v", err)
	}
	return nil
}

func MarshalTocFile(fTocNew io.Writer, entries []Entry) error {
	if config.GodOfWarVersion == config.GOW2ps2 {
		panic("NOT IMPLEMENTED")
		//return marshalTocFileGOW2(fTocNew, entries)
	} else {
		return marshalTocFileGOW1(fTocNew, entries)
	}
}

func updateFileWithIncreacingSize(fTocOriginal io.ReadSeeker, fTocNew io.Writer, partStreams []utils.ReaderWriterAt, filename string, in *io.SectionReader) error {
	log.Println("Updating toc+parts with increacing required sectors count")
	_, originalEntries, err := ParseFiles(fTocOriginal)
	if err != nil {
		return fmt.Errorf("Error when parsing toc: %v", err)
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
		packLasts := make([]int64, len(partStreams))
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
			if packStream == nil {
				continue
			}
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

	added := false
	newEntries := make([]Entry, len(entriesWithoutOurFile)+1)
	newEntriesIndex := 0
	newEntry := Entry{
		Name: filename,
		Size: in.Size(),
		Enc:  *newEncounter,
	}
	for _, e := range entriesWithoutOurFile {
		if !added {
			// Write our file entry to toc file
			if e.Enc.Pack > newEncounter.Pack || (e.Enc.Pack == newEncounter.Pack && e.Enc.Start > newEncounter.Start) {
				newEntries[newEntriesIndex] = newEntry
				newEntriesIndex += 1
				added = true
			}
			newEntries[newEntriesIndex] = e
			newEntriesIndex += 1
		}
	}
	if !added {
		newEntries[newEntriesIndex] = newEntry
	}

	return MarshalTocFile(fTocNew, newEntries)
}

// do not reorder files in pack
func updateFileWithoutIncreacingSize(fTocOriginal io.ReadSeeker, fTocNew io.Writer, partStreams []utils.ReaderWriterAt, filename string, in *io.SectionReader) error {
	log.Println("Updating toc+parts without increacing required sectors count")
	// update sizes in toc file, if changed
	files, entries, err := ParseFiles(fTocOriginal)
	if err != nil {
		return fmt.Errorf("Error when parsing toc: %v", err)
	}
	f := files[filename]
	if in.Size() != f.Size() {
		log.Println("Updating file size because its changed.")
		for _, e := range entries {
			if e.Name == filename {
				e.Size = in.Size()
			}
		}
	}

	if err := MarshalTocFile(fTocNew, entries); err != nil {
		return fmt.Errorf("Error when creating new toc file: %v", err)
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

func UpdateFile(fTocOriginal io.ReadSeeker, fTocNew io.Writer, partStreams []utils.ReaderWriterAt, f *File, in *io.SectionReader) error {
	if in.Size()/utils.SECTOR_SIZE > f.Size()/utils.SECTOR_SIZE {
		return updateFileWithIncreacingSize(fTocOriginal, fTocNew, partStreams, f.Name(), in)
	} else {
		return updateFileWithoutIncreacingSize(fTocOriginal, fTocNew, partStreams, f.Name(), in)
	}
}

type GOW2PackArrayDVD5Rader struct {
	readers []*io.SectionReader
}

func (ar *GOW2PackArrayDVD5Rader) ReadAt(b []byte, off int64) (n int, err error) {
	estimatedToRead := len(b)
	log.Printf("reading off %d, size %d", off, estimatedToRead)
	for _, r := range ar.readers {
		if r == nil || r.Size() == 0 {
			// we cannot continue because we need size of part
			log.Printf("breaking because nil, or r.Size == 0")
			break
		}
		if off < r.Size() {
			if off+int64(estimatedToRead) <= r.Size() {
				log.Printf("reading entire data at %d", off)
				rN, err := r.ReadAt(b[n:n+estimatedToRead], off)
				if err == nil && rN != estimatedToRead {
					log.Panicf("[gow2] Something wrong with read entire array calc: %d != %d", rN, estimatedToRead)
				}
				return n + rN, err
			} else {
				log.Println("reading part of data at %d", off)
				currentReadSize := off + int64(estimatedToRead) - r.Size()
				rN, err := r.ReadAt(b[n:int64(n)+currentReadSize], off)
				n += rN
				estimatedToRead -= int(currentReadSize)
				if err != nil {
					return n, fmt.Errorf("[gow2] Error reading pack array: %v", err)
				}
				if rN != int(currentReadSize) {
					log.Panicf("[gow2] Something wrong with read array calc: %d != %d", rN, currentReadSize)
				}
				off = 0
			}
		} else {
			log.Printf("skipping r because too large %d = %d - %d", off-r.Size(), off, r.Size())
			off -= r.Size()
		}
	}
	return n, io.EOF
}

func NewGOW2PackArrayDVD5Rader(readers []*io.SectionReader) *GOW2PackArrayDVD5Rader {
	return &GOW2PackArrayDVD5Rader{readers: readers}
}
