package gfx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"log"
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
	DataSize   uint32
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

func IndexSwizzlePalette(i int) int {
	remap := []int{0, 2, 1, 3}
	blockid := i / 8
	blockpos := i % 8
	return blockpos + (remap[blockid%4]+(blockid/4)*4)*8
}

func (gfx *GFX) AsRawPalette(idx int) ([]uint32, error) {
	palbuf := gfx.Data[idx]

	colors := gfx.Width * gfx.Height / gfx.DatasCount

	palette := make([]uint32, colors)

	for i := range palette {
		clr := binary.LittleEndian.Uint32(palbuf[i*4 : i*4+4])
		switch gfx.Height {
		case 2:
			palette[i] = clr
		case 32:
			fallthrough
		case 16:
			palette[IndexSwizzlePalette(i)] = clr
		default:
			return nil, fmt.Errorf("Wrong palette height: %d", gfx.Height)
		}
	}
	return palette, nil
}

func (gfx *GFX) AsPalette(idx int, convertAlphaToPCformat bool) ([]color.NRGBA, error) {
	rawPal, err := gfx.AsRawPalette(idx)
	if err != nil {
		return nil, err
	}

	palette := make([]color.NRGBA, len(rawPal))
	for i, raw := range rawPal {
		clr := color.NRGBA{
			R: uint8(raw),
			G: uint8(raw >> 8),
			B: uint8(raw >> 16),
			A: uint8(raw >> 24),
		}
		if convertAlphaToPCformat {
			clr.A = uint8(float32(clr.A) * (255.0 / 128.0))
		}
		palette[i] = clr
	}

	return palette, nil
}

func IndexUnswizzleTexture(x, y, width uint32) uint32 {
	block_location := (y&(math.MaxUint32^0xf))*width + (x&(math.MaxUint32^0xf))*2
	swap_selector := (((y + 2) >> 2) & 0x1) * 4
	posY := (((y & (math.MaxUint32 ^ 3)) >> 1) + (y & 1)) & 0x7
	column_location := posY*width*2 + ((x+swap_selector)&0x7)*4

	byte_num := ((y >> 1) & 1) + ((x >> 2) & 2)
	return block_location + column_location + byte_num
}

func (gfx *GFX) AsPaletteIndexes(idx int) []byte {
	data := gfx.Data[idx]

	indexes := make([]byte, gfx.Width*gfx.Height)
	switch gfx.GetPSM() {
	case GS_PSM_PSMT8H:
		fallthrough
	case GS_PSM_PSMT8:
		for y := uint32(0); y < gfx.Height; y++ {
			for x := uint32(0); x < gfx.Width; x++ {
				if gfx.Encoding&2 == 0 {
					pos := IndexUnswizzleTexture(x, y, gfx.Width)
					if pos < uint32(len(data)) {
						indexes[x+y*gfx.Width] = data[pos]
					} else {
						//log.Printf("Warning: Texture missed var: len=%v < pos=%v, x=%v, y=%v. w=%v h=%v", len(data), pos, x, y, gfx.Width, gfx.Height)
					}
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
		panic("Unknown palette indexes encoding case")
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

func NewFromData(name string, buf []byte) (*GFX, error) {
	gfx := &GFX{
		Magic:      binary.LittleEndian.Uint32(buf[0:4]),
		Width:      binary.LittleEndian.Uint32(buf[4:8]),
		Height:     binary.LittleEndian.Uint32(buf[8:12]),
		Encoding:   binary.LittleEndian.Uint32(buf[12:16]),
		Bpi:        binary.LittleEndian.Uint32(buf[16:20]),
		DatasCount: binary.LittleEndian.Uint32(buf[20:24]),
	}

	gfx.Data = make([][]byte, gfx.DatasCount)
	gfx.Psm = GsPsm[gfx.GetPSM()]

	if gfx.Magic != GFX_MAGIC {
		return nil, errors.New("Wrong magic.")
	}

	pos := uint32(24)
	gfx.DataSize = ((((gfx.Width * gfx.Height) / gfx.DatasCount) * gfx.Bpi) / 8)
	for iData := uint32(0); iData < gfx.DatasCount; iData++ {
		gfx.Data[iData] = buf[pos : pos+gfx.DataSize]
		pos += gfx.DataSize
	}

	return gfx, nil
}

func (gfx *GFX) MarshalToBinary() ([]byte, error) {
	buf := make([]byte, 24+gfx.DataSize*gfx.DatasCount)

	binary.LittleEndian.PutUint32(buf[0:4], gfx.Magic)
	binary.LittleEndian.PutUint32(buf[4:8], gfx.Width)
	binary.LittleEndian.PutUint32(buf[8:12], gfx.Height)
	binary.LittleEndian.PutUint32(buf[12:16], gfx.Encoding)
	binary.LittleEndian.PutUint32(buf[16:20], gfx.Bpi)
	binary.LittleEndian.PutUint32(buf[20:24], gfx.DatasCount)

	pos := uint32(24)
	for i, data := range gfx.Data {
		if len(data) != int(gfx.DataSize) {
			log.Println("WARNING: GFX.MarshalToBinary: data ", i, "datasize != len(data): ", gfx.DataSize, len(data))
		}
		copy(buf[pos:pos+gfx.DataSize], data)
		pos += gfx.DataSize
	}

	return buf, nil
}

func (gfx *GFX) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return gfx, nil
}

func init() {
	wad.SetHandler(GFX_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		//log.Println(wrsrc.Name())
		gfx, err := NewFromData(wrsrc.Name(), wrsrc.Tag.Data)
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
