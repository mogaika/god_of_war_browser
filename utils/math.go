package utils

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// result in radians
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

func DegreeToRadiansV3(v mgl32.Vec3) mgl32.Vec3 {
	return v.Mul(1.0 / (2.0 * math.Pi))
}

func RadiansToDegreeV3(v mgl32.Vec3) mgl32.Vec3 {
	return v.Mul(2.0 * math.Pi)
}

// input in radians
func EulerToQuat(v mgl32.Vec3) (q mgl32.Quat) {
	const halfToRad = ((0.5 * math.Pi) / 180.0) * 2.0 * math.Pi
	x := float64(v[0]) * halfToRad
	y := float64(v[1]) * halfToRad
	z := float64(v[2]) * halfToRad

	sx := math.Sin(x)
	cx := math.Cos(x)
	sy := math.Sin(y)
	cy := math.Cos(y)
	sz := math.Sin(z)
	cz := math.Cos(z)

	q.V[0] = float32(sx*cy*cz - cx*sy*sz)
	q.V[1] = float32(cx*sy*cz + sx*cy*sz)
	q.V[2] = float32(cx*cy*sz - sx*sy*cz)
	q.W = float32(cx*cy*cz + sx*sy*sz)

	return q.Normalize()
}

func FloatArray32to64(in []float32) []float64 {
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = float64(v)
	}
	return out
}
