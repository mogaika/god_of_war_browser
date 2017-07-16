package vag

import (
	"io"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/ps2/vagp"
	"github.com/mogaika/god_of_war_browser/utils"
)

func init() {
	pack.SetHandler(".VAG", func(p utils.ResourceSource, r *io.SectionReader) (interface{}, error) {
		return vagp.NewVAGPFromReader(r)
	})
}
