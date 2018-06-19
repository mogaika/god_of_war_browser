package toc

import (
	"sort"

	"github.com/mogaika/god_of_war_browser/vfs"

	"github.com/mogaika/god_of_war_browser/utils"
)

func encounterSortFunc(ei *Encounter, ej *Encounter) bool {
	return ei.Pak < ej.Pak || (ei.Pak == ej.Pak && ei.Offset < ej.Offset)
}

func sortFilesByEncounters(files map[string]*File) []*File {
	result := make([]*File, 0, len(files))
	for _, f := range files {
		result = append(result, f)
	}

	sort.Slice(result, func(i int, j int) bool {
		// if one of encounter of i file earlier then
		// all encounters of j file
		lesser := result[i]
		bigger := result[j]

		for _, le := range lesser.encounters {
			isAllEncountersLess := true
			for _, be := range bigger.encounters {
				if le.Pak > be.Pak || (le.Pak == be.Pak && le.Offset >= be.Offset) {
					isAllEncountersLess = false
				}
			}
			if isAllEncountersLess {
				return true
			}
		}
		return false
	})

	return result
}

func sortedEncountersFromFiles(files map[string]*File) []Encounter {
	result := make([]Encounter, 0, len(files)*2)
	for _, f := range files {
		for i := range f.encounters {
			result = append(result, f.encounters[i])
		}
	}
	sort.Slice(result, func(i int, j int) bool {
		return encounterSortFunc(&result[i], &result[j])
	})
	return result
}

type FreeSpace struct {
	Start, End int64
	Pak        PakIndex
}

func constructFreeSpaceArray(files map[string]*File, paks []vfs.File) []FreeSpace {
	encounters := sortedEncountersFromFiles(files)
	lastPak := PakIndex(0)
	lastPos := int64(0)
	result := make([]FreeSpace, 0, len(encounters))

	handlePakChange := func(newPak PakIndex) {
		if paks[lastPak].Size() > lastPos {
			result = append(result, FreeSpace{
				Start: lastPos,
				End:   paks[lastPak].Size(),
				Pak:   lastPak})
		}
		for curPak := lastPak + 1; curPak < newPak; curPak++ {
			result = append(result, FreeSpace{
				Start: 0,
				End:   paks[curPak].Size(),
				Pak:   curPak})
		}
		lastPak = newPak
		lastPos = 0
	}

	for iEncounter := range encounters {
		e := &encounters[iEncounter]

		if e.Pak != lastPak {
			handlePakChange(e.Pak)
		}

		if e.Offset > lastPos {
			result = append(result, FreeSpace{
				Start: lastPos,
				End:   e.Offset,
				Pak:   lastPak})
		}

		lastPos = e.Offset + utils.GetRequiredSectorsCount(e.Size)*utils.SECTOR_SIZE
	}
	handlePakChange(PakIndex(len(paks)))

	return result
}

func paksAsFreeSpaces(paks []vfs.File) []FreeSpace {
	result := make([]FreeSpace, len(paks))
	for iPak, f := range paks {
		result[iPak] = FreeSpace{
			Pak:   PakIndex(iPak),
			Start: 0,
			End:   f.Size(),
		}
	}
	return result
}
