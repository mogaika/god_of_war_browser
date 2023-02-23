package model

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/mogaika/god_of_war_browser/editor/core"
)

type Image struct {
	data []byte
}

type Texture struct {
	Animation core.Ref[Animation]
}

type RenMaterialLayer struct {
	Color   mgl32.Vec4
	Texture *Texture
}

type RenMaterial struct {
	Animation core.Ref[Animation]

	Layers []RenMaterialLayer
}

type RenGeometry struct{}

type RenModel struct{}

type PhyMaterial struct{}

type PhyGeometry struct{}

type PhyStaticGeometry struct{}

type GameObject struct{}

type Animation struct{}

type SoundBank struct{}

type WadArchive struct {
	Variables    map[string]uint32
	Animation    map[string]core.Ref[Animation]
	Textures     map[string]core.Ref[Texture]
	RenMaterials map[string]core.Ref[RenMaterial]
	SoundBank    map[string]core.Ref[SoundBank]
}
