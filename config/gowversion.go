package config

import (
	"log"
)

const (
	GOWunknown = iota
	GOW1
	GOW2
)

type GOWVersion int

var godOfWarVersion GOWVersion = GOWunknown
var playStationVersion int = 2

func GetGOWVersion() GOWVersion {
	return godOfWarVersion
}

func SetGOWVersion(v GOWVersion) {
	switch v {
	default:
		log.Panicf("Unknown gow version '%v'", v)
	case GOWunknown:
	case GOW1:
	case GOW2:
	}
	godOfWarVersion = v
}

func GetPlayStationVersion() int {
	return playStationVersion
}

func SetPlayStationVersion(psVersion int) {
	switch psVersion {
	default:
		log.Panicf("Unknown ps version '%v'", psVersion)
	case 2:
	case 3:
	}
	playStationVersion = psVersion
}
