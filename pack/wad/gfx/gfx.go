package gfx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

const GFX_MAGIC = 0xc
const HEADER_SIZE = 0x18

type GFX struct {
	Psm        string
	Magic      uint32
	Width      uint32
	Height     uint32
	Encoding   uint32
	Bpi        uint32
	DatasCount uint32
	Data       [][]byte `json:"-"`
}

const (
	GS_PSM_PSMCT32  = 0x00 // 32 bits per pixel.
	GS_PSM_PSMCT24  = 0x01 // 24 bits per pixel.
	GS_PSM_PSMCT16  = 0x02 // 16 bits per pixel.
	GS_PSM_PSMCT16S = 0x0A // 16 bits per pixel.
	GS_PSM_PSGPU24  = 0x12 // 24 bits per pixel.
	GS_PSM_PSMT8    = 0x13 // 8 bits per pixel, palettized.
	GS_PSM_PSMT4    = 0x14 // 4 bits per pixel, palettized.
	GS_PSM_PSMT8H   = 0x1B // 8 bits per pixel, 24 to 32
	GS_PSM_PSMT4HL  = 0x24 // 4 bits per pixel, 28 to 32
	GS_PSM_PSMT4HH  = 0x2C // 4 bits per pixel, 24 to 27
	GS_PSM_PSMZ32   = 0x30 // 32 bits per pixel.
	GS_PSM_PSMZ24   = 0x31 // 24 bits per pixel.
	GS_PSM_PSMZ16   = 0x32 // 16 bits per pixel.
	GS_PSM_PSMZ16S  = 0x3A // 16 bits per pixel.
)

var GsPsm map[int]string = map[int]string{
	GS_PSM_PSMCT32:  "GS_PSM_PSMCT32",
	GS_PSM_PSMCT24:  "GS_PSM_PSMCT24",
	GS_PSM_PSMCT16:  "GS_PSM_PSMCT16",
	GS_PSM_PSMCT16S: "GS_PSM_PSMCT16S",
	GS_PSM_PSGPU24:  "GS_PSM_PSGPU24",
	GS_PSM_PSMT8:    "GS_PSM_PSMT8",
	GS_PSM_PSMT4:    "GS_PSM_PSMT4",
	GS_PSM_PSMT8H:   "GS_PSM_PSMT8H",
	GS_PSM_PSMT4HL:  "GS_PSM_PSMT4HL",
	GS_PSM_PSMT4HH:  "GS_PSM_PSMT4HH",
	GS_PSM_PSMZ32:   "GS_PSM_PSMZ32",
	GS_PSM_PSMZ24:   "GS_PSM_PSMZ24",
	GS_PSM_PSMZ16:   "GS_PSM_PSMZ16",
	GS_PSM_PSMZ16S:  "GS_PSM_PSMZ16S",
}

func (gfx *GFX) AsPallet(idx int) (color.Palette, error) {
	palbuf := gfx.Data[idx]

	colors := gfx.Width * gfx.Height

	pallet := make(color.Palette, colors)
	remap := []int{0, 2, 1, 3}

	for i := range pallet {
		si := i * 4

		clr := color.RGBA{
			R: palbuf[si],
			G: palbuf[si+1],
			B: palbuf[si+2],
			A: byte(float32(palbuf[si+3]) * (255.0 / 128.0)),
			//A: palbuf[si+3],
		}

		switch gfx.Height {
		case 2:
			pallet[i] = clr
		case 32:
			fallthrough
		case 16:
			blockid := i / 8
			blockpos := i % 8

			newpos := blockpos + (remap[blockid%4]+(blockid/4)*4)*8
			pallet[newpos] = clr

		default:
			return nil, fmt.Errorf("Wrong pallet height: %d", gfx.Height)
		}
	}
	return pallet, nil
}

func (gfx *GFX) UnswizzlePosition(x, y uint32) uint32 {
	block_location := (y&(math.MaxUint32^0xf))*gfx.Width + (x&(math.MaxUint32^0xf))*2
	swap_selector := (((y + 2) >> 2) & 0x1) * 4
	posY := (((y & (math.MaxUint32 ^ 3)) >> 1) + (y & 1)) & 0x7
	column_location := posY*gfx.Width*2 + ((x+swap_selector)&0x7)*4

	byte_num := ((y >> 1) & 1) + ((x >> 2) & 2)
	return block_location + column_location + byte_num
}

func (gfx *GFX) AsPalletIndexes(idx int) []byte {
	data := gfx.Data[idx]

	indexes := make([]byte, gfx.Width*gfx.Height)
	switch gfx.GetPSM() {
	case GS_PSM_PSMT8H:
		fallthrough
	case GS_PSM_PSMT8:
		for y := uint32(0); y < gfx.Height; y++ {
			for x := uint32(0); x < gfx.Width; x++ {
				if gfx.Encoding&2 == 0 {
					indexes[x+y*gfx.Width] = data[gfx.UnswizzlePosition(x, y)]
				} else {
					indexes[x+y*gfx.Width] = data[x+y*gfx.Width]
				}
			}
		}
	case GS_PSM_PSMT4:
		for y := uint32(0); y < gfx.Height; y++ {
			for x := uint32(0); x < gfx.Width; x++ {
				val := data[(x+y*gfx.Width)/2]
				if x&1 == 0 {
					indexes[x+y*gfx.Width] = val & 0xf
				} else {
					indexes[x+y*gfx.Width] = val >> 4
				}
			}
		}
	default:
		panic("Unknown pallete indexes encoding case")
	}
	return indexes
}

func (gfx *GFX) String() string {
	return fmt.Sprintf("GFX Width: %d Height: %d Bpi: %d Encoding: %d Datas: %d\n",
		gfx.Width, gfx.Height, gfx.Bpi, gfx.Encoding, len(gfx.Data))
}

func (gfx *GFX) GetPSM() int {
	switch gfx.Bpi {
	case 32:
		return GS_PSM_PSMCT32
	case 24:
		return GS_PSM_PSMCT24
	case 16:
		return GS_PSM_PSMCT16
	case 8:
		if gfx.Encoding&2 == 0 {
			return GS_PSM_PSMT8
		} else {
			return GS_PSM_PSMT8H
		}
	case 4:
		return GS_PSM_PSMT4
	}
	return -1
}

func NewFromData(fgfx io.ReaderAt) (*GFX, error) {
	buf := make([]byte, HEADER_SIZE)
	if _, err := fgfx.ReadAt(buf, 0); err != nil {
		return nil, err
	}

	gfx := &GFX{
		Magic:      binary.LittleEndian.Uint32(buf[0:4]),
		Width:      binary.LittleEndian.Uint32(buf[4:8]),
		Height:     binary.LittleEndian.Uint32(buf[8:12]),
		Encoding:   binary.LittleEndian.Uint32(buf[12:16]),
		Bpi:        binary.LittleEndian.Uint32(buf[16:20]),
		DatasCount: binary.LittleEndian.Uint32(buf[20:24]),
	}

	gfx.Data = make([][]byte, gfx.DatasCount)

	if gfx.Magic != GFX_MAGIC {
		return nil, errors.New("Wrong magic.")
	}

	dataBlockCount := int(binary.LittleEndian.Uint32(buf[20:24]))

	for iData := 0; iData < dataBlockCount; iData++ {
		data := make([]byte, (gfx.Width*gfx.Height*gfx.Bpi)/8)

		_, err := fgfx.ReadAt(data, HEADER_SIZE)
		if err != nil {
			return nil, err
		}

		gfx.Data[iData] = data
	}

	gfx.Psm = GsPsm[gfx.GetPSM()]

	return gfx, nil
}

func (gfx *GFX) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	return gfx, nil
}

func init() {
	wad.SetHandler(GFX_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		gfx, err := NewFromData(r)
		if err != nil {
			return gfx, err
		}

		/*
			for i := range gfx.Data {
				fpath := filepath.Join("logs", w.Name, fmt.Sprintf("%.4d-%s.gfx.%d.dump", node.Id, node.Name, i))
				os.MkdirAll(filepath.Dir(fpath), 0777)
				dump, _ := os.Create(fpath)
				dump.Write(gfx.Data[i])
				dump.Close()
			}
		*/
		return gfx, err
	})
}
