package model

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Image struct {
	data []byte
}

type Texture struct {
	Animation Ref[Animation]
}

type MaterialLayer struct {
	mgl32.Vec4
}

type Material struct {
	Animation Ref[Animation]

	Layers []MaterialLayer
}

type Animation struct {
}

type SoundBank struct {
}

type WadArchive struct {
	Variables map[string]uint32
	Animation map[string]Ref[Animation]
	Textures  map[string]Ref[Texture]
	Materials map[string]Ref[Material]
	SoundBank map[string]Ref[SoundBank]
}
