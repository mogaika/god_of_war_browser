package txr

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"log"

	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/psvita/gxt"
	"github.com/mogaika/god_of_war_browser/utils"
)

const PSVITA_MAGIC = "TXTR"

type PsVitaTexture struct {
	g      *gxt.GXT
	images []image.Image
}

type PsVitaTextureAjax struct {
	PsVitaTexture
	Images [][]byte
}

func (t *PsVitaTexture) Images() []image.Image {
	return t.images
}

func (t *PsVitaTexture) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	a := &PsVitaTextureAjax{
		PsVitaTexture: *t,
		Images:        make([][]byte, len(t.g.TextureInfos)),
	}

	for i, img := range t.images {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
		a.Images[i] = buf.Bytes()
	}

	return a, nil
}

func (t *PsVitaTexture) readImages(r io.ReadSeeker) error {
	t.images = make([]image.Image, len(t.g.TextureInfos))
	for i, ti := range t.g.TextureInfos {
		if img, err := ti.ToImage(r); err != nil {
			return errors.Wrapf(err, "Failed to create image from gxt")
		} else {
			t.images[i] = img
		}
	}
	return nil
}

func NewPsVitaTextureFromData(bs *utils.BufStack) (*PsVitaTexture, error) {
	bs.SubBuf("serverId", 0).SetSize(4)

	t := &PsVitaTexture{}

	headerBs := bs.SubBuf("params", 4).SetSize(10)
	magic := headerBs.ReadStringBuffer(4)
	if magic != PSVITA_MAGIC {
		return nil, errors.Errorf("Incorrect magic 0x%x", magic)
	}

	unk01 := headerBs.ReadBU32()
	unk02 := headerBs.ReadBU16()

	log.Printf("unk01:0x%x unk02:0x%x", unk01, unk02)

	gxtBs := headerBs.SubBufFollowing("gxt").Expand()
	if g, err := gxt.Open(bytes.NewReader(gxtBs.Raw())); err != nil {
		return nil, errors.Wrapf(err, "Failed to read gxt")
	} else {
		t.g = g
	}

	if err := t.readImages(bytes.NewReader(gxtBs.Raw())); err != nil {
		return nil, errors.Wrapf(err, "Failed to read images")
	}

	return t, nil
}
