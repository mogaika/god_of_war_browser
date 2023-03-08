package r3d

import (
	"github.com/go-gl/mathgl/mgl32"
)

/*
transform canonic
transform canonicRotation
transform matrix
*/

type Node struct {
	Transform mgl32.Mat4

	Position mgl32.Vec3
	Rotation mgl32.Quat
	Scale    mgl32.Vec3

	Parent *Node
	Childs []*Node
}
