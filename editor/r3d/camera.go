package r3d

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

type Camera interface {
	GetViewMatrix() mgl32.Mat4
}

type OrbitController struct {
	Target   mgl32.Vec3
	Distance float32
	Pitch    float32 // x rotation
	Yaw      float32 // y rotation
}

func NewOrbitController(target mgl32.Vec3, dist, pitch, yaw float32) *OrbitController {
	return &OrbitController{
		Target:   target,
		Distance: dist,
		Pitch:    pitch,
		Yaw:      yaw,
	}
}

func (c *OrbitController) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position(), c.Target, mgl32.Vec3{0, 1, 0})
}

func (c *OrbitController) Position() mgl32.Vec3 {
	return mgl32.Vec3{
		c.Distance * float32(math.Cos(float64(mgl32.DegToRad(c.Pitch)))*math.Sin(float64(mgl32.DegToRad(c.Yaw)))),
		c.Distance * float32(math.Sin(float64(mgl32.DegToRad(c.Pitch)))),
		c.Distance * float32(math.Cos(float64(mgl32.DegToRad(c.Pitch)))*math.Cos(float64(mgl32.DegToRad(c.Yaw)))),
	}.Add(c.Target).Add(mgl32.Vec3{0, 50, 0})
}
