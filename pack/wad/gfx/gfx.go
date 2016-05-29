package gfx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

const GFX_MAGIC = 0xc
const HEADER_SIZE = 0x18

type GFX struct {
	Magic    uint32
	Width    uint32
	Height   uint32
	Encoding uint32
	Bpi      uint32
	Data     [][]byte
}

func (gfx *GFX) GetPallet(idx int) (color.Palette, error) {
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

func (gfx *GFX) String() string {
	return fmt.Sprintf("GFX Width: %d Height: %d Bpi: %d Encoding: %d Datas: %d\n",
		gfx.Width, gfx.Height, gfx.Bpi, gfx.Encoding, len(gfx.Data))
}

func NewFromData(fgfx io.ReaderAt) (*GFX, error) {
	buf := make([]byte, HEADER_SIZE)
	if _, err := fgfx.ReadAt(buf, 0); err != nil {
		return nil, err
	}

	gfx := &GFX{
		Magic:    binary.LittleEndian.Uint32(buf[0:4]),
		Width:    binary.LittleEndian.Uint32(buf[4:8]),
		Height:   binary.LittleEndian.Uint32(buf[8:12]),
		Encoding: binary.LittleEndian.Uint32(buf[12:16]),
		Bpi:      binary.LittleEndian.Uint32(buf[16:20]),
		Data:     make([][]byte, binary.LittleEndian.Uint32(buf[20:24])),
	}

	if gfx.Magic != GFX_MAGIC {
		return nil, errors.New("Wrong magic.")
	}

	dataBlockCount := int(binary.LittleEndian.Uint32(buf[20:24]))

	for iData := 0; iData < dataBlockCount; iData++ {
		rawData := make([]byte, (gfx.Width*gfx.Height*gfx.Bpi)/8)
		var data []byte

		_, err := fgfx.ReadAt(rawData, HEADER_SIZE)
		if err != nil {
			return nil, err
		}

		switch gfx.Bpi {
		case 4:
			data = make([]byte, gfx.Width*gfx.Height)
			for i, v := range rawData {
				data[i*2] = v & 0xf
				data[i*2+1] = (v >> 4) & 0xf
			}
		case 8:
			data = rawData
		case 32:
			data = rawData
		default:
			return nil, errors.New("Unknown gfx bpi")
		}

		gfx.Data[iData] = data
	}

	return gfx, nil
}

func init() {
	wad.SetHandler(GFX_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (interface{}, error) {
		gfx, err := NewFromData(r)
		if err != nil {
			return gfx, err
		}

		for i := range gfx.Data {
			fpath := filepath.Join("logs", w.Name, fmt.Sprintf("%.4d-%s.gfx.%d.dump", node.Id, node.Name, i))
			os.MkdirAll(filepath.Dir(fpath), 0777)
			dump, _ := os.Create(fpath)
			dump.Write(gfx.Data[i])
			dump.Close()
		}

		return gfx, err
	})
}
