package anm

import (
	"math"
)

type AnimState8Texturepos struct {
	ChangedDataIndex uint16
	Byte2            byte
	Byte3            byte
	DatasCount       uint32
	Data             []float32
}

func AnimState8TextureposFromBuf(buf []byte) *AnimState8Texturepos {
	a := &AnimState8Texturepos{
		ChangedDataIndex: u16(buf, 0),
		Byte2:            buf[2],
		Byte3:            buf[3],
		DatasCount:       u32(buf, 4),
	}
	a.Data = make([]float32, a.DatasCount)
	for i := range a.Data {
		a.Data[i] = math.Float32frombits(u32(buf, uint32(0xc+i*4)))
	}
	return a
}
