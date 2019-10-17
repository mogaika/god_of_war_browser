package txr

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_gfx "github.com/mogaika/god_of_war_browser/pack/wad/gfx"
)

func (txr *Texture) imagePS2(gfx *file_gfx.GFX, pal *file_gfx.GFX, igfx int, ipal int) (*image.RGBA, error) {
	width := int(gfx.Width)
	height := int(gfx.RealHeight)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	palette, err := pal.AsPalette(ipal, true)

	if err != nil {
		return nil, err
	}

	palidx := gfx.AsPaletteIndexes(igfx)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, palette[palidx[x+y*width]])
		}
	}

	return img, nil
}

func (txr *Texture) marshalBlendPS2(clrBlend []float32, wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	res := &Ajax{
		Data:         txr,
		FilterLinear: txr.Flags&1 == 0,
	}

	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("Panic when marshaling texture %s: %v", wrsrc.Tag.Name, r))
		}
	}()

	if txr.GfxName != "" && txr.PalName != "" {
		gfxn := wrsrc.Wad.GetNodeByName(txr.GfxName, wrsrc.Node.Id, false)
		paln := wrsrc.Wad.GetNodeByName(txr.PalName, wrsrc.Node.Id, false)
		if gfxn == nil {
			return nil, fmt.Errorf("Cannot find gfx: %s", txr.GfxName)
		}

		if paln == nil {
			return nil, fmt.Errorf("Cannot find pal: %s", txr.PalName)
		}

		gfxc, _, err := wrsrc.Wad.GetInstanceFromNode(gfxn.Id)
		if err != nil {
			return nil, fmt.Errorf("Error getting gfx %s: %v", txr.GfxName, err)
		}

		palc, _, err := wrsrc.Wad.GetInstanceFromNode(paln.Id)
		if err != nil {
			return nil, fmt.Errorf("Error getting pal %s: %v", txr.PalName, err)
		}

		gfx := gfxc.(*file_gfx.GFX)
		pal := palc.(*file_gfx.GFX)

		res.Images = make([]AjaxImage, len(gfx.Data)*len(pal.Data))

		i := 0
		for iGfx := range gfx.Data {
			for iPal := range pal.Data {
				img, err := txr.Image(gfx, pal, iGfx, iPal)
				if err != nil {
					return nil, err
				}

				blendImg(img, clrBlend)

				var bufImage bytes.Buffer
				png.Encode(&bufImage, img)

				res.Images[i].Gfx = iGfx
				res.Images[i].Pal = iPal
				res.Images[i].Image = bufImage.Bytes()

				i++
			}
		}
	}
	return res, nil
}
