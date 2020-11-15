package gltfutils

import (
	"io"
	"os"

	"github.com/mogaika/god_of_war_browser/pack/wad"

	"github.com/qmuntal/gltf"
)

type GLTFCacher struct {
	Doc   *gltf.Document
	Cache map[wad.TagId]interface{}
}

func NewCacher() *GLTFCacher {
	return &GLTFCacher{
		Doc:   gltf.NewDocument(),
		Cache: make(map[wad.TagId]interface{}),
	}
}

func ExportBinary(w io.Writer, doc *gltf.Document) error {
	os.MkdirAll("./lastgltf/", 0777)
	gltf.Save(doc, "./lastgltf/file.gltf")

	encoder := gltf.NewEncoder(w)
	encoder.AsBinary = true
	return encoder.Encode(doc)
}

func (c *GLTFCacher) AddCache(id wad.TagId, data interface{}) { c.Cache[id] = data }

func (c *GLTFCacher) GetCached(id wad.TagId) interface{} {
	if v, e := c.Cache[id]; e {
		return v
	} else {
		return nil
	}
}

func (c *GLTFCacher) GetCachedOr(id wad.TagId, createFunc func() interface{}) interface{} {
	if result := c.GetCached(id); result != nil {
		return result
	} else {
		return createFunc()
	}
}
