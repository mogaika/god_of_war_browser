package config

import (
	"log"
)

const (
	GOWunknown = iota
	GOW1ps2
	GOW2ps2
)

type GOWVersion int

var godOfWarVersion GOWVersion

func GetGOWVersion() GOWVersion {
	return godOfWarVersion
}

func SetGOWVersion(v GOWVersion) {
	switch v {
	default:
		log.Panicf("Unknown gow version '%v'", v)
	case GOWunknown:
	case GOW1ps2:
	case GOW2ps2:
	}
	godOfWarVersion = v
}
