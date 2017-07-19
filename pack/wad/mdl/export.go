package mdl

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	fmat "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	fmesh "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	ftxr "github.com/mogaika/god_of_war_browser/pack/wad/txr"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (mdl *Model) ExportObj(wrsrc *wad.WadNodeRsrc, bones []mgl32.Mat4, matlibRelativePath string, w io.Writer, wMatlib io.Writer) (map[string][]byte, error) {
	textures := make(map[string][]byte, 0)
	materials := make([]string, 0)

	w.Write([]byte(fmt.Sprintf("mtllib %s\n", filepath.Base(matlibRelativePath))))

	for _, id := range wrsrc.Node.SubGroupNodes {
		node := wrsrc.Wad.GetNodeById(id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id); err == nil {
			switch inst.(type) {
			case *fmat.Material:
				_mat := inst.(*fmat.Material)
				_marshaledMat, err := _mat.Marshal(wrsrc.Wad.GetNodeResourceByNodeId(node.Id))
				if err != nil {
					continue
				}

				marshaledMat := _marshaledMat.(fmat.Ajax)
				var bestImage *ftxr.AjaxImage
				var imgName string
				for iLayer, _txr := range marshaledMat.Textures {
					if _txr == nil {
						continue
					}
					txr := _txr.(*ftxr.Ajax)
					imgName = fmt.Sprintf("%d_%s", iLayer, txr.Data.GfxName)
					if bestImage == nil {
						bestImage = &txr.Images[0]
					} else {
						if marshaledMat.Mat.Layers[iLayer].Flags[0] != 0x01010080 {
							bestImage = &txr.Images[0]
						}
					}
				}

				clr := _mat.Color
				wMatlib.Write([]byte(fmt.Sprintf("newmtl %s\nKd %f %f %f %f\n", node.Tag.Name, clr[0], clr[1], clr[2], clr[3])))
				materials = append(materials, node.Tag.Name)
				if bestImage != nil {
					wMatlib.Write([]byte(fmt.Sprintf("map_Ka %s.png\nmap_Kd %s.png\n\n", imgName, imgName)))
					textures[imgName] = bestImage.Image
				}
			}
		}
	}

	for _, id := range wrsrc.Node.SubGroupNodes {
		node := wrsrc.Wad.GetNodeById(id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id); err == nil {
			switch inst.(type) {
			case *fmesh.Mesh:
				mesh := inst.(*fmesh.Mesh)
				if err := mesh.ExportObj(w, bones, materials); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	return textures, nil
}

func (mdl *Model) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "zip":
		var buf, objBuf, mtlBuf bytes.Buffer

		z := zip.NewWriter(&buf)

		textures, err := mdl.ExportObj(wrsrc, nil, wrsrc.Name()+".mtl", &objBuf, &mtlBuf)
		if err != nil {
			log.Println("exporterr", err)
		}

		wObj, err := z.Create(wrsrc.Name() + ".obj")
		if err != nil {
			log.Println("objerr", err)
		}
		wObj.Write(objBuf.Bytes())

		wMtl, err := z.Create(wrsrc.Name() + ".mtl")
		if err != nil {
			log.Println("mtlerr", err)
		}
		wMtl.Write(mtlBuf.Bytes())

		for tname, t := range textures {
			wTxr, err := z.Create(tname + ".png")
			if err != nil {
				log.Println("txrerr", tname, err)
			}
			wTxr.Write(t)
		}

		if err := z.Close(); err != nil {
			log.Println("zcloseerr", err)
		}

		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Name()+".zip")
	}
}
