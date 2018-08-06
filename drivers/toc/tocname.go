package toc

import (
	"fmt"
)

type TocNamingPolicy struct {
	TocName     string
	UseIndexing bool
	PakPrefix   string
	PakSuffix   string
}

func (tnp *TocNamingPolicy) GetPakName(pi PakIndex) string {
	if tnp.UseIndexing {
		return fmt.Sprintf("%s%d%s", tnp.PakPrefix, int(pi)+1, tnp.PakSuffix)
	} else {
		return fmt.Sprintf("%s%s", tnp.PakPrefix, tnp.PakSuffix)
	}
}

var defaultTocNamePair = []TocNamingPolicy{
	{"GODOFWAR.TOC", true, "PART", ".PAK"},
	// TODO: I'm not sure about this. Need to find related discussion
	{"GODOFWAR.BIN", true, "DATA", ".BIN"},
}
