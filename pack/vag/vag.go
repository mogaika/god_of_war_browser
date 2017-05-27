package vag

import (
	"io"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/ps2/vagp"
)

func init() {
	pack.SetHandler(".VAG", func(p *pack.Pack, pf *pack.PackFile, r io.ReaderAt) (interface{}, error) {
		return vagp.NewVAGPFromReader(io.NewSectionReader(r, 0, pf.Size))
	})
}
