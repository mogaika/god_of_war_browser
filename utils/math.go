package utils

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

func QuatToEuler(q mgl32.Quat) (e mgl32.Vec3) {
	sinr_cosp := float64(2 * (q.W*q.X() + q.Y()*q.Z()))
	cosr_cosp := float64(1 - 2*(q.X()*q.X()+q.Y()*q.Y()))

	e[0] = float32(math.Atan2(sinr_cosp, cosr_cosp))

	sinp := float64(2 * (q.W*q.Y() - q.Z()*q.X()))
	if math.Abs(sinp) >= 1 {
		e[1] = math.Pi / 2
		if sinp < 0 {
			e[1] *= -1
		}
	} else {
		e[1] = float32(math.Asin(sinp))
	}

	siny_cosp := float64(2 * (q.W*q.Z() + q.X()*q.Y()))
	cosy_cosp := float64(1 - 2*(q.Y()*q.Y()+q.Z()*q.Z()))
	e[2] = float32(math.Atan2(siny_cosp, cosy_cosp))

	return e
}

func FloatArray32to64(in []float32) []float64 {
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = float64(v)
	}
	return out
}
