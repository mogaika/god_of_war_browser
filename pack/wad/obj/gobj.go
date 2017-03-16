package obj

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/mat"
	"github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	"github.com/mogaika/god_of_war_browser/pack/wad/mesh"
)

type GobjExportRaw struct {
	Object *Object
	Model  *mdl.Model
	Meshes []*mesh.Mesh
}

func (obj *Object) ExportGobj(offile string, wd *wad.Wad, node *wad.WadNode) error {
	exRaw := &GobjExportRaw{
		Object: obj,
		Meshes: make([]*mesh.Mesh, 0),
	}

	for _, id := range node.SubNodes {
		nd := wd.Node(id).ResolveLink()
		if nd.Format == mdl.MODEL_MAGIC {
			modelFile, err := wd.Get(id)
			if err != nil {
				return fmt.Errorf("Error fetching mdl: %v", err)
			}
			exRaw.Model = modelFile.(*mdl.Model)
			for _, i := range nd.SubNodes {
				sn, err := wd.Get(i)
				if err != nil {
					return fmt.Errorf("Error when extracting node %d->%s mdl info: %v", i, wd.Node(i).Name, err)
				} else {
					switch sn.(type) {
					case *mesh.Mesh:
						exRaw.Meshes = append(exRaw.Meshes, sn.(*mesh.Mesh))
					case *mat.Material:
						// TODO
					}
				}
			}
		}
	}

	data, err := json.MarshalIndent(exRaw, "", "\t")
	if err != nil {
		return fmt.Errorf("Marshal error: %v", err)
	}

	if err = ioutil.WriteFile(offile, data, 0777); err != nil {
		return fmt.Errorf("Write file error: %v", err)
	}
	return nil
}
