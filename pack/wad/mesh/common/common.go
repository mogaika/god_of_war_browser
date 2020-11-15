package common

import (
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
)

type Position = mgl32.Vec3
type UV = mgl32.Vec2
type Normal = mgl32.Vec3
type RGBA = color.NRGBA

type Vertex struct {
	Position      Position
	Weight        float32
	JointsIndexes [2]uint16
}

type Object struct {
	Vertices    []Vertex
	BlendColors [][]RGBA // rgba color per layer or instance
	UVs         [][]UV   // one uv per material layer
	Normals     []Normal
	Indexes     []int
	JointMaps   [][]uint32 // one map per instance (each instance has it's own jointsmap)

	MaterialIndex int
	PartIndex     int
	LodGroupIndex int
	ObjectIndex   int

	InstancesCount int
	LayersCount    int
}

type LodGroup struct {
	Objects      []*Object
	HideDistance float32
}

type Part struct {
	LodGroups []*LodGroup
}

type Mesh struct {
	Parts []*Part
}
