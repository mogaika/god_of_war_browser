package main

import (
	"bytes"
	"log"

	"github.com/mogaika/god_of_war_browser/fbx"
)

func main() {
	var b bytes.Buffer
	f := &fbx.FBX{
		GlobalSettings: &fbx.GlobalSettings{
			Properties70: fbx.Properties70{
				P: []*fbx.Propertie70{
					&fbx.Propertie70{Name: "UpAxis", Type: "int", Purpose: "Integer", Value: 2},
					&fbx.Propertie70{Name: "OriginalUnitScaleFactor", Type: "double", Purpose: "Number", Value: 2.54},
					&fbx.Propertie70{Name: "DefaultCamera", Type: "KString", Value: "Producer Perspective"},
					&fbx.Propertie70{Name: "TransparentColor", Type: "Color", Idk: "A", Value: []float32{1.0, 1.0, 1.0}},
				},
			},
		},
		Objects: fbx.Objects{
			Geometry: []*fbx.Geometry{
				&fbx.Geometry{
					Properties70: fbx.Properties70{
						P: []*fbx.Propertie70{
							&fbx.Propertie70{Name: "UpAxis", Type: "int", Purpose: "Integer", Value: 2},
							&fbx.Propertie70{Name: "OriginalUnitScaleFactor", Type: "double", Purpose: "Number", Value: 2.54},
							&fbx.Propertie70{Name: "DefaultCamera", Type: "KString", Value: "Producer Perspective"},
							&fbx.Propertie70{Name: "TransparentColor", Type: "Color", Idk: "A", Value: []float32{1.0, 1.0, 1.0}},
						},
					},
					Vertices: []float64{
						-23.6220474243164, -23.6220474243164, 0, 23.6220474243164, -23.6220474243164, 0, -23.6220474243164, 23.6220474243164, 0, 23.6220474243164,
					},
					PolygonVertexIndex: []int{
						0, 2, 3, -2, 4, 5, 7, -7, 0, 1, 5, -5, 1, 3, 7, -6, 3, 2, 6, -8, 2, 0, 4, -7,
					},
					Edges: []int{
						0, 1, 2, 3, 4, 5, 6, 7, 9, 11, 13, 17,
					},
				},
			},
		},
	}
	f.FBXHeaderExtension = fbx.NewFbx("god of war browser").FBXHeaderExtension

	defer func() {
		log.Println("++++++++++++++ :FBX: ++++++++++++++")
		println(b.String())
	}()

	if err := fbx.Export(f, &b); err != nil {
		log.Panic(err)
	}
}
