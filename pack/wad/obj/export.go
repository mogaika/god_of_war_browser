package obj

/*
import (
	"archive/zip"
	"bytes"
	"io"
	"strings"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/mat"
	"github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (obj *Object) ExportObj(w io.Writer) error {
	zwr := zip.NewWriter(w)
	defer zwr.Close()

	for _, id := range wrsrc.Node.SubGroupNodes {
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(wrsrc.Wad.GetNodeById(id).Id); err == nil {
if inst.(type) {

}
		}
	}

}

func (obj *Object) SubfileGetter(w http.ResponseWriter, r *http.Request, wrsrc *wad.WadNodeRsrc, subfile string) {
	switch {
	case strings.HasSuffix(subfile, "@obj@"):
		var buf bytes.Buffer
		obj.ExportObj(&buf)
		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Name()+".zip")
	}
}

func (obj *Object) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	mrshl := &ObjMarshal{Data: obj}
	for _, id := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id); err == nil {
			switch inst.(type) {
			case *mdl.Model, *scr.ScriptParams, *collision.Collision:
				if subFileMarshled, err := inst.Marshal(wrsrc.Wad.GetNodeResourceByNodeId(n.Id)); err != nil {
					panic(err)
				} else {
					switch inst.(type) {
					case *mdl.Model:
						mrshl.Model = subFileMarshled
					case *collision.Collision:
						mrshl.Collision = subFileMarshled
					case *scr.ScriptParams:
						mrshl.Script = subFileMarshled
					}
				}
			}
		}
	}

	return mrshl, nil
}
*/
