package vag

import (
	"io"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/ps2/vagp"
	"github.com/mogaika/god_of_war_browser/utils"
)

func init() {
	h := func(p utils.ResourceSource, r *io.SectionReader) (interface{}, error) {
		return vagp.NewVAGPFromReader(r)
	}
	pack.SetHandler(".VAG", h)
	pack.SetHandler(".VA1", h)
	pack.SetHandler(".VA2", h)
	pack.SetHandler(".VA3", h)
	pack.SetHandler(".VA4", h)
	pack.SetHandler(".VA5", h)
}
