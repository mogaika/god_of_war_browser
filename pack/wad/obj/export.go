package obj

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mdl "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (obj *Object) ExportObj(wrsrc *wad.WadNodeRsrc, matlibRelativePath string, w io.Writer, wMatlib io.Writer) (map[string][]byte, error) {
	for _, id := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id); err == nil {
			switch inst.(type) {
			case *file_mdl.Model:
				mdl := inst.(*file_mdl.Model)

				bones := make([]mgl32.Mat4, mdl.JointsCount)
				for i, joint := range obj.Joints {
					bones[i] = joint.RenderMat
				}

				return mdl.ExportObj(wrsrc.Wad.GetNodeResourceByNodeId(n.Id), bones, matlibRelativePath, w, wMatlib)
			}
		}
	}
	return nil, fmt.Errorf("Cannot find model :-( .")
}

func (obj *Object) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "zip":
		var buf, objBuf, mtlBuf bytes.Buffer

		z := zip.NewWriter(&buf)

		textures, err := obj.ExportObj(wrsrc, wrsrc.Name()+".mtl", &objBuf, &mtlBuf)
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
