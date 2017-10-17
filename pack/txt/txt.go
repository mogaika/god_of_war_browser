package vag

import (
	"io"
	"io/ioutil"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Txt string

func init() {
	pack.SetHandler(".TXT", func(p utils.ResourceSource, r *io.SectionReader) (interface{}, error) {
		if data, err := ioutil.ReadAll(r); err != nil {
			return nil, err
		} else {
			return Txt(utils.BytesToString(data)), nil
		}
	})
}
