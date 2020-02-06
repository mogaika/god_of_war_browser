package obj

import (
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mdl "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
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
