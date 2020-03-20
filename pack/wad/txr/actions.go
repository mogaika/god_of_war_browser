package txr

import (
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_gfx "github.com/mogaika/god_of_war_browser/pack/wad/gfx"
)

func (txr *Texture) changeTexturePS2(wrsrc *wad.WadNodeRsrc, img image.Image, createNewPal bool) error {
	if txr.GfxName == "" || txr.PalName == "" {
		return fmt.Errorf("Do not support texture with lod levels")
	}

	gfxcn := wrsrc.Wad.GetNodeByName(txr.GfxName, wrsrc.Node.Id, false)
	palcn := wrsrc.Wad.GetNodeByName(txr.PalName, wrsrc.Node.Id, false)

	gfxcw, _, gfxErr := wrsrc.Wad.GetInstanceFromNode(gfxcn.Id)
	palcw, _, palErr := wrsrc.Wad.GetInstanceFromNode(palcn.Id)

	gfxc := gfxcw.(*file_gfx.GFX)
	palc := palcw.(*file_gfx.GFX)

	if gfxErr != nil || palErr != nil {
		return fmt.Errorf("Cannot get gfx or pal instance: %v, %v", gfxErr, palErr)
	}

	if len(gfxc.Data) != 1 {
		return fmt.Errorf("Do not support gfx with DatasCount != 1")
	}

	b := img.Bounds().Max
	log.Println("Calculating palette...")
	newPal, newIdx := imgToPaletteAndIndex(img, gfxc.Encoding == 0)
	//log.Println("done")

	gfxc.Data[0] = newIdx
	gfxc.DataSize = uint32(len(gfxc.Data[0]))
	gfxc.Bpi = 8
	// gfxc.Encoding = do not change
	gfxc.Width = uint32(b.X)
	gfxc.Height = uint32(b.Y)

	palc.Data[0] = paletteToBytearray(newPal)
	palc.Width = 16
	palc.Height = (uint32(len(newPal)) / palc.Width) * uint32(len(palc.Data))
	palc.DataSize = uint32(len(palc.Data[0]))
	palc.Encoding = 0
	palc.Bpi = 32

	if len(palc.Data) == 2 {
		log.Println("Detected grayscale palette. Calculating new grayscale palette...")
		if err := gfxSecondPaletteToGrayscale(palc); err != nil {
			return fmt.Errorf("Error when calculating grayscale palette: %v", err)
		}
	}

	gfxBinRaw, err := gfxc.MarshalToBinary()
	if err != nil {
		return fmt.Errorf("gfxc.MarshalToBinary(): %v", err)
	}

	palBinRaw, err := palc.MarshalToBinary()
	if err != nil {
		return fmt.Errorf("palc.MarshalToBinary(): %v", err)
	}

	if err := wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
		gfxcn.Tag.Id: gfxBinRaw,
	}); err != nil {
		return fmt.Errorf("Update gfx tag error: %v", err)
	}

	if createNewPal {
		newPalName := wrsrc.Wad.GenerateName(txr.PalName)

		txr.PalName = newPalName

		log.Printf("Creating new palette '%s'", newPalName)

		if err := wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
			wrsrc.Tag.Id: txr.MarshalToBinary(),
		}); err != nil {
			return fmt.Errorf("Update txr tag error: %v", err)
		}

		if err := wrsrc.Wad.InsertNewTags(palcn.Tag.Id, []wad.Tag{
			{Tag: wad.GetServerInstanceTag(), Flags: palcn.Tag.Flags, Name: newPalName, Data: palBinRaw},
		}); err != nil {
			return fmt.Errorf("Insert pal tag error: %v", err)
		}
	} else {
		if err := wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
			palcn.Tag.Id: palBinRaw,
		}); err != nil {
			return fmt.Errorf("Update pal tag error: %v", err)
		}
	}

	return nil
}

func (txr *Texture) ChangeTexture(wrsrc *wad.WadNodeRsrc, fNewImage io.Reader, createNewPal bool) error {
	img, _, err := image.Decode(fNewImage)
	if err != nil {
		return err
	}

	switch config.GetPlayStationVersion() {
	case config.PS2:
		return txr.changeTexturePS2(wrsrc, img, createNewPal)
	case config.PS3:
		return txr.changeTexturePS3(wrsrc, img)
	default:
		return fmt.Errorf("Unsupported playstation version")
	}

}

func gfxSecondPaletteToGrayscale(palc *file_gfx.GFX) error {
	if len(palc.Data) != 2 {
		return fmt.Errorf("DatasCount != 2 (%d)", len(palc.Data))
	}

	pal, err := palc.AsPalette(0, false)
	if err != nil {
		return fmt.Errorf("Getting palette fail: %v", err)
	}
	d := palc.Data[1]
	for i := range pal {
		c := pal[file_gfx.IndexSwizzlePalette(i)]

		y := byte(0.299*float32(c.R) + 0.587*float32(c.G) + 0.114*float32(c.B))
		d[i*4] = y
		d[i*4+1] = y
		d[i*4+2] = y
		d[i*4+3] = 0x80
	}
	return nil
}

func (txr *Texture) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "upload":
		q := r.URL.Query()
		createNewPal := strings.ToLower(q.Get("create_new_pal")) == "true"

		fImg, _, err := r.FormFile("img")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		defer fImg.Close()
		if err := txr.ChangeTexture(wrsrc, fImg, createNewPal); err != nil {
			log.Printf("[txr] Error changing texture: %v", err)
			fmt.Fprintln(w, "change texture error:", err)
		}
	}
}
