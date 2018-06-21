package txr

import (
	"fmt"
	"image"
	"image/color"
	"sort"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	file_wad "github.com/mogaika/god_of_war_browser/pack/wad"
	file_gfx "github.com/mogaika/god_of_war_browser/pack/wad/gfx"
)

func CreateNewTextureInWad(wad *file_wad.Wad, baseTextureName string, insertAfterTag file_wad.TagId, img image.Image) error {
	var gfxc, palc file_gfx.GFX

	b := img.Bounds().Max
	newPal, newIdx := imgToPaletteAndIndex(img, true)

	gfxc.Magic = file_gfx.GFX_MAGIC
	gfxc.Data = make([][]byte, 1)
	gfxc.Data[0] = newIdx
	gfxc.DataSize = uint32(len(gfxc.Data[0]))
	gfxc.Bpi = 8
	gfxc.Encoding = 0
	gfxc.Width = uint32(b.X)
	gfxc.Height = uint32(b.Y)

	palc.Magic = file_gfx.GFX_MAGIC
	palc.Data = make([][]byte, 1)
	palc.Data[0] = paletteToBytearray(newPal)
	palc.Width = 16
	palc.Height = (uint32(len(newPal)) / palc.Width) * uint32(len(palc.Data))
	palc.DataSize = uint32(len(palc.Data[0]))
	palc.Encoding = 0
	palc.Bpi = 32

	gfxBinRaw, err := gfxc.MarshalToBinary()
	if err != nil {
		return fmt.Errorf("gfxc.MarshalToBinary(): %v", err)
	}

	palBinRaw, err := palc.MarshalToBinary()
	if err != nil {
		return fmt.Errorf("palc.MarshalToBinary(): %v", err)
	}

	txr := &Texture{
		Magic:         TXR_MAGIC,
		GfxName:       "GFX_" + baseTextureName,
		PalName:       "PAL_" + baseTextureName,
		LODParamK:     -160,
		LODMultiplier: 96,
		Flags:         0x510000,
	}

	return wad.InsertNewTags(insertAfterTag, []file_wad.Tag{
		// flags are same for gow1 and gow2
		{Tag: file_wad.GetServerInstanceTag(), Flags: 3, Name: txr.GfxName, Data: gfxBinRaw},
		{Tag: file_wad.GetServerInstanceTag(), Flags: 3, Name: txr.PalName, Data: palBinRaw},
		{Tag: file_wad.GetServerInstanceTag(), Flags: 0, Name: "TXR_" + baseTextureName, Data: txr.MarshalToBinary()},
	})
}

func clrToUint32(c color.Color) uint32 {
	r, g, b, a := c.RGBA()
	r /= 0x101
	g /= 0x101
	b /= 0x101
	a /= 0x101
	return r | g<<8 | b<<16 | a<<24
}

func imgToPaletteAndIndex(img image.Image, swizzleGfx bool) (color.Palette, []byte) {
	type clrCounter struct {
		uc     uint32
		c      color.Color
		counts int
	}

	counter := make([]clrCounter, 0)
	findClr := func(uc uint32) *clrCounter {
		for i := range counter {
			if counter[i].uc == uc {
				return &counter[i]
			}
		}
		return nil
	}
	// log.Println("Construct counters array")
	b := img.Bounds().Max
	for y := 0; y < b.Y; y++ {
		for x := 0; x < b.X; x++ {
			c := img.At(x, y)
			uc := clrToUint32(c)
			if clr := findClr(uc); clr != nil {
				clr.counts++
			} else {
				counter = append(counter, clrCounter{c: c, uc: uc, counts: 1})
			}
		}
	}
	// log.Println("Sorting")
	sort.Slice(counter, func(i, j int) bool { return counter[i].counts > counter[j].counts })
	pal := make(color.Palette, 256)

	for i := range pal {
		if i < len(counter) {
			pal[i] = counter[i].c
		} else {
			pal[i] = counter[len(counter)-1].c
		}
	}

	// log.Println("Generating img indexes")
	idx := make([]byte, b.X*b.Y)
	for y := 0; y < b.Y; y++ {
		for x := 0; x < b.X; x++ {
			if swizzleGfx {
				idx[file_gfx.IndexUnswizzleTexture(uint32(x), uint32(y), uint32(b.X))] = byte(pal.Index(img.At(x, y)))
			} else {
				idx[y*b.X+x] = byte(pal.Index(img.At(x, y)))
			}
		}
	}

	// log.Println("Swizzle palette")
	swizzledpal := make(color.Palette, 256)
	for i := range pal {
		swizzledpal[i] = pal[file_gfx.IndexSwizzlePalette(i)]
	}

	return swizzledpal, idx
}

func paletteToBytearray(p color.Palette) []byte {
	buf := make([]byte, len(p)*4)
	pos := 0
	for _, c := range p {
		r, g, b, a := c.RGBA()
		buf[pos+0] = byte(r / 0x101)
		buf[pos+1] = byte(g / 0x101)
		buf[pos+2] = byte(b / 0x101)
		buf[pos+3] = byte(float32(a/0x101) * (128.0 / 255.0))
		pos += 4
	}
	return buf
}
