package utils

type ColorFloat [4]float32

func (c *ColorFloat) RGBA() (r, g, b, a uint32) {
	const mf = float32(256*256 - 1)
	r = uint32(c[0] * mf)
	g = uint32(c[1] * mf)
	b = uint32(c[2] * mf)
	a = uint32(c[3] * mf)
	return
}

func NewColorFloatA(c []float32) ColorFloat {
	return ColorFloat{c[0], c[1], c[2], c[3]}
}

func NewColorFloat(c []float32) ColorFloat {
	return ColorFloat{c[0], c[1], c[2], 1.0}
}

/*
type Color interface {
	// RGBA returns the alpha-premultiplied red, green, blue and alpha values
	// for the color. Each value ranges within [0, 0xffff], but is represented
	// by a uint32 so that multiplying by a blend factor up to 0xffff will not
	// overflow.
	//
	// An alpha-premultiplied color component c has been scaled by alpha (a),
	// so has valid values 0 <= c <= a.
	RGBA() (r, g, b, a uint32)
}
*/
