package vag

import (
	"io"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/ps2/vagp"
)

func init() {
	pack.SetHandler(".VAG", func(p pack.PackFile, r *io.SectionReader) (interface{}, error) {
		return vagp.NewVAGPFromReader(r)
	})
}
