package wad

import "github.com/mogaika/god_of_war_browser/config"

func GetServerInstanceTag() uint16 {
	switch config.GetGOWVersion() {
	case config.GOW1ps2:
		return TAG_GOW1_SERVER_INSTANCE
	case config.GOW2ps2:
		return TAG_GOW2_SERVER_INSTANCE
	default:
		panic("unknwn")
	}
}

func isZeroSizedTag(tag *Tag) bool {
	switch config.GetGOWVersion() {
	case config.GOW1ps2:
		return tag.Tag == TAG_GOW1_ENTITY_COUNT
	case config.GOW2ps2:
		return tag.Tag == TAG_GOW2_ENTITY_COUNT
	default:
		panic("unknwn")
	}
}
