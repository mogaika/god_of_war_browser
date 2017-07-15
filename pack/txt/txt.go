package txt

import (
	"io"

	"github.com/mogaika/god_of_war_browser/pack"
)

func init() {
	pack.SetHandler(".VAG", func(p *pack.Pack, pf *pack.PackFile, r *io.SectionReader) (interface{}, error) {

	})
}
