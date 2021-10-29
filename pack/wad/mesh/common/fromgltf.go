package common

import (
	"github.com/qmuntal/gltf"
)

func FromGLTF(d *gltf.Document) (*Mesh, error) {
	scene := d.Scenes[*d.Scene]

	_ = scene
	return nil, nil
}
