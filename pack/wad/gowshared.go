package wad

import "github.com/mogaika/god_of_war_browser/config"

func GetServerInstanceTag() uint16 {
	switch config.GetGOWVersion() {
	case config.GOW1:
		return TAG_GOW1_SERVER_INSTANCE
	case config.GOW2:
		return TAG_GOW2_SERVER_INSTANCE
	case config.GOW2018:
		return TAG_GOW2018_SERVER_INSTANCE
	default:
		panic("unknwn")
	}
}

func isZeroSizedTag(tag *Tag) bool {
	switch config.GetGOWVersion() {
	case config.GOW1:
		return tag.Tag == TAG_GOW1_ENTITY_COUNT
	case config.GOW2:
		return tag.Tag == TAG_GOW2_ENTITY_COUNT
	case config.GOW2018:
		return false
	default:
		panic("unknwn")
	}
}
