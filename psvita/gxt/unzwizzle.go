package gxt

import (
	"image"
	"math/bits"
)

// Based on github.com/xdanieldzd/GXTConvert

func compact1By1(pos uint32) uint32 {
	pos &= 0x5555_5555                     // x = -f-e -d-c -b-a -9-8 -7-6 -5-4 -3-2 -1-0
	pos = (pos | (pos >> 1)) & 0x3333_3333 // x = --fe --dc --ba --98 --76 --54 --32 --10
	pos = (pos | (pos >> 2)) & 0x0f0f_0f0f // x = ---- fedc ---- ba98 ---- 7654 ---- 3210
	pos = (pos | (pos >> 4)) & 0x00ff_00ff // x = ---- ---- fedc ba98 ---- ---- 7654 3210
	pos = (pos | (pos >> 8)) & 0x0000_ffff // x = ---- ---- ---- ---- fedc ba98 7654 3210
	return pos
}

func IndexUnSwizzle(i, width, height uint32) (x, y uint32) {
	min := width
	if height < width {
		min = height
	}

	k := uint32(bits.TrailingZeros32(min))
	j := i & ((^uint32(0)) << (2 * k))

	mx := compact1By1(i>>0) & (min - 1)
	my := compact1By1(i>>1) & (min - 1)

	if height < width {
		j |= (my << k) | mx
		x, y = j>>k, j&(min-1)
	} else {
		j |= (mx << k) | my
		y, x = j>>k, j&(min-1)
	}

	return
}

func ImageUnSwizzle(img *image.NRGBA) *image.NRGBA {
	newImage := image.NewNRGBA(img.Rect)
	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			newX, newY := IndexUnSwizzle(uint32(y*width+x), uint32(width), uint32(height))
			newImage.Set(int(newX), int(newY), img.At(x, y))
		}
	}
	return newImage
}
