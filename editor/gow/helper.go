package gow

import "github.com/go-gl/mathgl/mgl32"

type helperBBox [2]mgl32.Vec3

func (bbox *helperBBox) ExpandToPoint(pos mgl32.Vec3) {
	for i, coord := range pos {
		if coord < bbox[0][i] {
			bbox[0][i] = coord
		}
	}
	for i, coord := range pos {
		if coord > bbox[1][i] {
			bbox[1][i] = coord
		}
	}
}

func (bbox *helperBBox) Size() float32 {
	return bbox[1].Sub(bbox[0]).Len()
}

func (bbox *helperBBox) Center() mgl32.Vec3 {
	return bbox[0].Add(bbox[1]).Mul(0.5)
}
