package txr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_gfx "github.com/mogaika/god_of_war_browser/pack/wad/gfx"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Texture struct {
	Magic         uint32
	GfxName       string
	PalName       string
	SubTxrName    string
	LODParamK     int32
	LODMultiplier float32
	Flags         uint32
}

const FILE_SIZE = 0x58
const TXR_MAGIC = 0x00000007

func NewFromData(buf []byte) (*Texture, error) {
	tex := &Texture{
		Magic:         binary.LittleEndian.Uint32(buf[0:4]),
		GfxName:       utils.BytesToString(buf[4:28]),
		PalName:       utils.BytesToString(buf[28:52]),
		SubTxrName:    utils.BytesToString(buf[52:76]),
		LODParamK:     int32(binary.LittleEndian.Uint32(buf[76:80])),
		LODMultiplier: math.Float32frombits(binary.LittleEndian.Uint32(buf[80:84])),
		Flags:         binary.LittleEndian.Uint32(buf[84:88]),
	}

	if tex.Magic != TXR_MAGIC {
		return nil, errors.New("Wrong magic.")
	}

	// 0 - any; 8000 - alpha channel
	flags1 := tex.Flags & 0xffff
	if flags1 != 0 && flags1 != 0x8000 {
		return nil, fmt.Errorf("Unknown unkFlags 0x%.4x != 0", flags1)
	}

	// 1 - no alpha; 5d - alpha channel; 51 - font
	flags2 := tex.Flags >> 16
	if flags2 != 1 && flags2 != 0x41 && flags2 != 0x5d && flags2 != 0x51 && flags2 != 0x11 {
		return nil, fmt.Errorf("Unknown unkFlags2 0x%.4x (0x1,0x41,0x5d,0x51,0x11)", flags2)
	}

	return tex, nil
}

func (txr *Texture) MarshalToBinary() []byte {
	var buf [FILE_SIZE]byte
	binary.LittleEndian.PutUint32(buf[0:4], txr.Magic)
	copy(buf[4:28], utils.StringToBytesBuffer(txr.GfxName, 24, true))
	copy(buf[28:52], utils.StringToBytesBuffer(txr.PalName, 24, true))
	copy(buf[52:76], utils.StringToBytesBuffer(txr.SubTxrName, 24, true))
	binary.LittleEndian.PutUint32(buf[76:80], uint32(txr.LODParamK))
	binary.LittleEndian.PutUint32(buf[80:84], math.Float32bits(txr.LODMultiplier))
	binary.LittleEndian.PutUint32(buf[84:88], txr.Flags)
	return buf[:]
}

type ImageType int

const (
	TEXTURE_IMAGE_RGBA ImageType = iota
	TEXTURE_IMAGE_RGB
	TEXTURE_IMAGE_A
)

func (txr *Texture) image(gfx *file_gfx.GFX, pal *file_gfx.GFX, igfx int, ipal int, image_type ImageType) (*image.RGBA, error) {
	width := int(gfx.Width)
	height := int(gfx.RealHeight)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	palette, err := pal.AsPalette(ipal, image_type != TEXTURE_IMAGE_A)

	if err != nil {
		return nil, err
	}

	palidx := gfx.AsPaletteIndexes(igfx)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			clr := palette[palidx[x+y*width]]
			switch image_type {
			case TEXTURE_IMAGE_RGB:
				clr.A = 255
			case TEXTURE_IMAGE_A:
				clr.B = clr.A
				if clr.A < 128 {
					clr.G = clr.A
				} else {
					clr.G = 255
				}
				if clr.A < 129 {
					clr.R = clr.A
				} else {
					clr.R = 255
				}
				clr.A = 255
			}
			img.Set(x, y, clr)
		}
	}

	return img, nil
}

func (txr *Texture) Image(gfx *file_gfx.GFX, pal *file_gfx.GFX, igfx int, ipal int) (*image.RGBA, error) {
	return txr.image(gfx, pal, igfx, ipal, TEXTURE_IMAGE_RGBA)
}

func (txr *Texture) ImageOnlyColor(gfx *file_gfx.GFX, pal *file_gfx.GFX, igfx int, ipal int) (*image.RGBA, error) {
	return txr.image(gfx, pal, igfx, ipal, TEXTURE_IMAGE_RGB)
}

func (txr *Texture) ImageOnlyAlpha(gfx *file_gfx.GFX, pal *file_gfx.GFX, igfx int, ipal int) (*image.RGBA, error) {
	return txr.image(gfx, pal, igfx, ipal, TEXTURE_IMAGE_A)
}

type AjaxImage struct {
	Gfx, Pal  int
	Image     []byte
	AlphaOnly []byte
	ColorOnly []byte
}

type Ajax struct {
	Data         *Texture
	Images       []AjaxImage
	FilterLinear bool
}

func blendImg(img *image.RGBA, clrBlend []float32) {
	if clrBlend != nil {
		bounds := img.Bounds()
		width, height := bounds.Max.X, bounds.Max.Y
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				r, g, b, a := img.At(x, y).RGBA()

				clamp := func(a float32) uint8 {
					if a > 255.0 {
						return 0xff
					} else {
						return uint8(a)
					}
				}

				img.Set(x, y, color.RGBA{
					R: clamp(float32(r/0x101) / 255.0 * clrBlend[0] * 255.0),
					G: clamp(float32(g/0x101) / 255.0 * clrBlend[1] * 255.0),
					B: clamp(float32(b/0x101) / 255.0 * clrBlend[2] * 255.0),
					A: clamp(float32(a/0x101) / 255.0 * clrBlend[3] * 255.0),
				})
			}
		}
	}
}

func (txr *Texture) MarshalBlend(clrBlend []float32, wrsrc *wad.WadNodeRsrc) (interface{}, error) {
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

				if clrOnly, _ := txr.ImageOnlyColor(gfx, pal, iGfx, iPal); clrOnly != nil {
					blendImg(clrOnly, clrBlend)
					var bufImageClrOnly bytes.Buffer
					png.Encode(&bufImageClrOnly, clrOnly)
					res.Images[i].ColorOnly = bufImageClrOnly.Bytes()
				}
				if alphaOnly, _ := txr.ImageOnlyAlpha(gfx, pal, iGfx, iPal); alphaOnly != nil {
					blendImg(alphaOnly, clrBlend)
					var bufImageAlphaOnly bytes.Buffer
					png.Encode(&bufImageAlphaOnly, alphaOnly)
					res.Images[i].AlphaOnly = bufImageAlphaOnly.Bytes()
				}

				i++
			}
		}
	}
	return res, nil
}

func (t *Texture) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return t.MarshalBlend(nil, wrsrc)
}

func init() {
	h := func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	}
	wad.SetHandler(config.GOW1ps2, TXR_MAGIC, h)
	wad.SetHandler(config.GOW2ps2, TXR_MAGIC, h)
}
