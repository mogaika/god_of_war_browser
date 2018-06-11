package config

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
	godOfWarVersion = v
}
