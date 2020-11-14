package gltfutils

import (
	"io"
	"os"

	"github.com/qmuntal/gltf"
)

func NewDocument() *gltf.Document {
	return gltf.NewDocument()
}

func ExportBinary(w io.Writer, doc *gltf.Document) error {
	for iNode := range doc.Nodes {
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(iNode))
	}

	os.MkdirAll("./lastgltf/", 0777)
	gltf.Save(doc, "./lastgltf/file.gltf")

	encoder := gltf.NewEncoder(w)
	encoder.AsBinary = true
	return encoder.Encode(doc)
}
