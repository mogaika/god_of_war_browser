package textureformats

import (
	"encoding/binary"
	"image"
	"image/color"
)

// Based on github.com/xdanieldzd/GXTConvert

func decompressBlockDXT5(blockData []byte, outColors []color.NRGBA) {
	alpha0 := uint32(blockData[0])
	alpha1 := uint32(blockData[1])

	// uint48 splitted to uint32 + uint16
	alphaCode1 := binary.LittleEndian.Uint32(blockData[4:])
	alphaCode2 := uint32(binary.LittleEndian.Uint16(blockData[2:]))

	color0 := binary.LittleEndian.Uint16(blockData[8:])
	color1 := binary.LittleEndian.Uint16(blockData[10:])

	colorCode := binary.LittleEndian.Uint32(blockData[12:])

	r0, g0, b0 := rgb565fromUint16(color0)
	r1, g1, b1 := rgb565fromUint16(color1)

	for y := uint32(0); y < 4; y++ {
		for x := uint32(0); x < 4; x++ {
			alphaCodeIndex := 3 * (4*y + x)
			var alphaCode uint32

			if alphaCodeIndex <= 12 {
				alphaCode = (alphaCode2 >> alphaCodeIndex) & 7
			} else if alphaCodeIndex == 15 {
				alphaCode = (alphaCode2 >> 15) | ((alphaCode1 << 1) & 6)
			} else {
				alphaCode = (alphaCode1 >> (alphaCodeIndex - 16)) & 7
			}

			var finalAlpha byte
			if alphaCode == 0 {
				finalAlpha = byte(alpha0)
			} else if alphaCode == 1 {
				finalAlpha = byte(alpha1)
			} else {
				if alpha0 > alpha1 {
					finalAlpha = byte(((8-alphaCode)*alpha0 + (alphaCode-1)*alpha1) / 7)
				} else {
					if alphaCode == 6 {
						finalAlpha = 0
					} else if alphaCode == 7 {
						finalAlpha = 0xff
					} else {
						finalAlpha = byte(((6-alphaCode)*alpha0 + (alphaCode-1)*alpha1) / 5)
					}
				}
			}

			positionCode := (colorCode >> (2 * ((4 * y) + x))) & 3

			r, g, b := dxColorFromPosition(
				positionCode, color0, color1,
				r0, g0, b0, r1, g1, b1)

			outColors[x+y*4] = color.NRGBA{R: r, G: g, B: b, A: finalAlpha}
		}
	}
}

func DecompressImageDX5(data []byte, w, h int) *image.NRGBA {
	return decomporessImageDX(len(data)/16, w, h,
		func(blockIndex int, outColors []color.NRGBA) {
			decompressBlockDXT5(data[blockIndex*0x10:], outColors)
		})
}
