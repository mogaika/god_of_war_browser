package textureformats

import (
	"image"
	"image/color"
)

// Based on github.com/xdanieldzd/GXTConvert

var dxtReplacement = []int{
	0, 2, 8, 10,
	1, 3, 9, 11,
	4, 6, 12, 14,
	5, 7, 13, 15,
}

func rgb565fromUint16(v uint16) (r, g, b uint16) {
	r = (v >> 11) & 0x1f
	g = (v >> 5) & 0x3f
	b = (v >> 0) & 0x1f

	// r = uint16(((uint32(r) * 255) + 15) / 31)
	// g = uint16(((uint32(g) * 255) + 31) / 63)
	// b = uint16(((uint32(b) * 255) + 15) / 31)

	r = (r << 3) | (r >> 2)
	g = (g << 2) | (g >> 4)
	b = (b << 3) | (b >> 2)

	return
}

func dxColorFromPosition(positionCode uint32, color0, color1 uint16, r0, g0, b0 uint16, r1, g1, b1 uint16) (r, g, b byte) {
	switch positionCode {
	case 0:
		r, g, b = byte(r0), byte(g0), byte(b0)
	case 1:
		r, g, b = byte(r1), byte(g1), byte(b1)
	case 2:
		if color0 > color1 {
			r, g, b = byte((2*r0+r1)/3), byte((2*g0+g1)/3), byte((2*b0+b1)/3)
		} else {
			r, g, b = byte((r0+r1)/2), byte((g0+g1)/2), byte((b0+b1)/2)
		}
	case 3:
		if color0 > color1 {
			r, g, b = byte((r0+2*r1)/3), byte((g0+2*g1)/3), byte((b0+2*b1)/3)
		} else {
			r, g, b = 0, 0, 0
		}
	}
	return
}

func roundToPow2(x int) int {
	result := 1
	for result < x {
		result *= 2
	}
	return result
}

func decomporessImageDX(blocks int, w, h int, blockmethod func(blockIndex int, colors []color.NRGBA)) *image.NRGBA {
	roundedW := roundToPow2(w)
	roundedH := roundToPow2(h)

	img := image.NewNRGBA(image.Rect(0, 0, roundedW, roundedH))

	colors := make([]color.NRGBA, 4*4)

	for iBlock := 0; iBlock < blocks; iBlock++ {
		blockmethod(iBlock, colors)

		for iColor, c := range colors {
			pos := iBlock*4*4 + dxtReplacement[iColor]

			img.SetNRGBA(pos%roundedW, pos/roundedW, c)
		}
	}

	return img
}
