package textureformats

import (
	"encoding/binary"
	"image"
	"image/color"
)

// Based on github.com/xdanieldzd/GXTConvert

func decompressBlockDXT1(blockData []byte, outColors []color.NRGBA) {
	color0 := binary.LittleEndian.Uint16(blockData[0:])
	color1 := binary.LittleEndian.Uint16(blockData[2:])
	code := binary.LittleEndian.Uint32(blockData[4:])

	r0, g0, b0 := rgb565fromUint16(color0)
	r1, g1, b1 := rgb565fromUint16(color1)

	for y := uint32(0); y < 4; y++ {
		for x := uint32(0); x < 4; x++ {
			positionCode := (code >> (2 * ((4 * y) + x))) & 3

			r, g, b := dxColorFromPosition(
				positionCode, color0, color1,
				r0, g0, b0, r1, g1, b1)

			outColors[x+y*4] = color.NRGBA{R: r, G: g, B: b, A: 0xff}
		}
	}
}

func DecompressImageDX1(data []byte, w, h int) *image.NRGBA {
	return decomporessImageDX(len(data)/8, w, h,
		func(blockIndex int, outColors []color.NRGBA) {
			decompressBlockDXT1(data[blockIndex*8:], outColors)
		})
}
