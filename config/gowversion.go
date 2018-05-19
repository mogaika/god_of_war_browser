package config

const (
	GOWunknown = iota
	GOW1ps2
	GOW2ps2
)

type GOWVersion int

var GodOfWarVersion GOWVersion

func GetGOWVersion() GOWVersion {
	return GodOfWarVersion
}

func SetGOWVersion(v GOWVersion) {
	GodOfWarVersion = v
}
