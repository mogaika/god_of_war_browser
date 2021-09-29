package collision

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

func (c *Collision) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	log.Println(c.ShapeName, action)
	switch c.ShapeName {
	case "SheetHdr":
		c.Shape.(*ShapeRibSheet).HttpAction(wrsrc, w, r, action)
	}
}

func (rib *ShapeRibSheet) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "frommodel":
		gltfReader, _, err := r.FormFile("model")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		defer gltfReader.Close()
		if err := rib.FromModel(wrsrc, gltfReader); err != nil {
			log.Printf("[rib] Error updating col mesh: %v", err)
			fmt.Fprintln(w, "col mesh update error:", err)
		}
	}
}

func (rib *ShapeRibSheet) FromModel(wrsrc *wad.WadNodeRsrc, gltfReader io.Reader) error {
	switch config.GetPlayStationVersion() {
	case config.PS2:
	default:
		return fmt.Errorf("Unsupported playstation version")
	}

	doc := &gltf.Document{}
	if err := gltf.NewDecoder(gltfReader).Decode(doc); err != nil {
		return errors.Wrapf(err, "Failed to read gltf")
	}

	vertices := make([][3]float32, 0)
	indices := make([][3]uint16, 0)

	var showTree func(id uint32, tab string)
	showTree = func(id uint32, tab string) {
		node := doc.Nodes[id]
		mesh := -1
		if node.Mesh != nil {
			mesh = int(*node.Mesh)
		}
		log.Printf("%s node %q (mesh %v)", tab, node.Name, mesh)
		for _, c := range node.Children {
			showTree(c, tab+" -=")
		}
	}
	for _, iNode := range doc.Scenes[0].Nodes {
		showTree(iNode, "+ ")
	}

	for _, iNode := range doc.Scenes[0].Nodes {
		node := doc.Nodes[iNode]
		log.Printf("Loaded name %q", node.Name)

		if node.Name == "CONTEXT_BBOX" {
			continue
		}

		if node.Mesh == nil {
			log.Printf("Node %q without mesh", node.Name)
			continue
		}
		mesh := doc.Meshes[*node.Mesh]
		for _, primitive := range mesh.Primitives {
			if primitive.Indices == nil {
				log.Printf("One of mesh %q primitive has no indices", mesh.Name)
				continue
			}

			primitiveVertices2 := make([][3]float32, 0)
			primitiveIndices2 := make([]uint32, 0)

			primitiveVertices, err := modeler.ReadPosition(
				doc, doc.Accessors[primitive.Attributes["POSITION"]], primitiveVertices2)
			if err != nil {
				return errors.Wrapf(err, "Failed to read mesh vertices")
			}
			primitiveIndices, err := modeler.ReadIndices(
				doc, doc.Accessors[*primitive.Indices], primitiveIndices2)
			if err != nil {
				return errors.Wrapf(err, "Failed to read mesh indices")
			}

			utils.LogDump(primitiveIndices, primitiveVertices)

			indicesOffset := uint32(len(vertices))
			vertices = append(vertices, primitiveVertices...)
			for i := 0; i < len(primitiveIndices)/3; i++ {
				indices = append(indices, [3]uint16{
					uint16(primitiveIndices[i*3+0] + indicesOffset),
					uint16(primitiveIndices[i*3+1] + indicesOffset),
					uint16(primitiveIndices[i*3+2] + indicesOffset),
				})
			}
		}
	}

	utils.LogDump(indices, vertices)

	bbox := getBboxForVertices(vertices)
	rib.LevelBBox[0] = mgl32.Vec4{bbox[0][0], bbox[0][1], bbox[0][2]}
	rib.LevelBBox[1] = mgl32.Vec4{bbox[1][0], bbox[1][1], bbox[1][2]}

	rib.Some1 = make([]RibKDTreeNode, 1)
	rib.Some1[0] = RibKDTreeNode{
		IsPolygon:     true,
		PolygonFlag:   0,
		PolygonIndex:  0,
		PolygonsCount: uint16(len(indices)),
		PolygonUnk0x6: 0,
	}

	rib.Some6 = make([]RibPolygon, len(indices))
	for i := range indices {
		rib.Some6[i] = RibPolygon{
			IsQuad:              false,
			QuadOrTriangleIndex: uint16(i),
		}
	}

	rib.Some7TrianglesIndex = make([]RibTriangle, len(indices))
	for i := range indices {
		rib.Some7TrianglesIndex[i] = RibTriangle{
			RibPolygonBase: RibPolygonBase{
				Flags:         0,
				MaterialIndex: 0,
			},
			Indexes: indices[i],
		}
	}

	rib.Some8QuadsIndex = make([]RibQuad, 0)

	rib.Some9Points = make([]mgl32.Vec3, len(vertices))
	for i, v := range vertices {
		rib.Some9Points[i] = mgl32.Vec3(v)
	}

	wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
		wrsrc.Tag.Id: rib.Marshal(),
	})

	return nil
}

func getBboxForVertices(vertices [][3]float32) [2][3]float32 {
	result := [2][3]float32{
		{-100000, -100000, -100000},
		{100000, 100000, 100000},
	}

	for _, v := range vertices {
		if v[0] < result[0][0] {
			result[0][0] = v[0]
		}
		if v[1] < result[0][1] {
			result[0][1] = v[1]
		}
		if v[2] < result[0][2] {
			result[0][2] = v[2]
		}
		if v[0] > result[1][0] {
			result[1][0] = v[0]
		}
		if v[1] > result[1][1] {
			result[1][1] = v[1]
		}
		if v[2] > result[1][2] {
			result[1][2] = v[2]
		}
	}

	return result
}
