package txr

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net/http"
	"sort"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_gfx "github.com/mogaika/god_of_war_browser/pack/wad/gfx"
)

func clrToUint32(c color.Color) uint32 {
	r, g, b, a := c.RGBA()
	r /= 0x101
	g /= 0x101
	b /= 0x101
	a /= 0x101
	return r | g<<8 | b<<16 | a<<24
}

func imgToPalleteAndIndex(img image.Image) (color.Palette, []byte) {
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
	log.Println("Construct counters array")
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
	log.Println("Sorting")
	sort.Slice(counter, func(i, j int) bool { return counter[i].counts > counter[j].counts })
	pal := make(color.Palette, 256)

	for i := range pal {
		if i < len(counter) {
			pal[i] = counter[i].c
		} else {
			pal[i] = counter[len(counter)-1].c
		}
	}

	log.Println("Generating img indexes")
	idx := make([]byte, b.X*b.Y)
	for y := 0; y < b.Y; y++ {
		for x := 0; x < b.X; x++ {
			idx[y*b.X+x] = byte(pal.Index(img.At(x, y)))
		}
	}

	log.Println("Swizzle pallete")
	remap := []int{0, 2, 1, 3}
	swizzledpal := make(color.Palette, 256)
	for i := range pal {
		blockid := i / 8
		blockpos := i % 8

		newpos := blockpos + (remap[blockid%4]+(blockid/4)*4)*8
		swizzledpal[i] = pal[newpos]
	}

	return swizzledpal, idx
}

func palleteToBytearray(p color.Palette) []byte {
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

func (txr *Texture) ChangeTexture(wrsrc *wad.WadNodeRsrc, fNewImage io.Reader) error {
	img, _, err := image.Decode(fNewImage)
	if err != nil {
		return err
	}

	if txr.GfxName == "" || txr.PalName == "" {
		return fmt.Errorf("Do not support texture with lod levels")
	}

	gfxcn := wrsrc.Wad.GetNodeByName(txr.GfxName, wrsrc.Node.Id, false)
	palcn := wrsrc.Wad.GetNodeByName(txr.PalName, wrsrc.Node.Id, false)

	gfxcw, _, err := wrsrc.Wad.GetInstanceFromNode(gfxcn.Id)
	palcw, _, err := wrsrc.Wad.GetInstanceFromNode(palcn.Id)

	gfxc := gfxcw.(*file_gfx.GFX)
	palc := palcw.(*file_gfx.GFX)

	if gfxc.DatasCount != 1 || palc.DatasCount != 1 {
		return fmt.Errorf("Do not support pal or gfx with DatasCount != 1")
	}

	b := img.Bounds().Max
	log.Println("Calculating palette...")
	newPal, newIdx := imgToPalleteAndIndex(img)
	log.Println("done")
	gfxc.Data[0] = newIdx
	gfxc.Bpi = 8
	gfxc.Encoding = 2
	gfxc.Width = uint32(b.X)
	gfxc.Height = uint32(b.Y)

	palc.Data[0] = palleteToBytearray(newPal)
	palc.Height = 16
	palc.Width = uint32(len(newPal)) / palc.Height
	palc.Encoding = 0
	palc.Bpi = 32

	gfxBinRaw, err := gfxc.MarshalToBinary()
	if err != nil {
		return fmt.Errorf("gfxc.MarshalToBinary(): %v", err)
	}

	log.Println("Updating wad...")
	palBinRaw, err := palc.MarshalToBinary()
	log.Println("done")
	if err != nil {
		return fmt.Errorf("palc.MarshalToBinary(): %v", err)
	}

	return wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
		gfxcn.Tag.Id: gfxBinRaw,
		palcn.Tag.Id: palBinRaw,
	})
}

func (txr *Texture) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "upload":
		log.Println(r.Method)
		fImg, _, err := r.FormFile("img")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		defer fImg.Close()
		if err := txr.ChangeTexture(wrsrc, fImg); err != nil {
			log.Println(err)
			fmt.Fprintln(w, "change texture error:", err)
		}
	}
}
