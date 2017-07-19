package mesh

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

func (m *Mesh) ExportObj(_w io.Writer, bones []mgl32.Mat4, materials []string) error {
	iV := 0
	iT := 0
	iN := 0
	var facesBuff bytes.Buffer

	w := func(format string, args ...interface{}) {
		_w.Write(([]byte)(fmt.Sprintf(format+"\n", args...)))
	}

	wi := func(format string, args ...interface{}) {
		facesBuff.WriteString(fmt.Sprintf(format+"\n", args...))
	}
	lastb := uint32(0)
	for iPart, part := range m.Parts {
		for iGroup, group := range part.Groups {
			wi("o p%.2dg%.2d", iPart, iGroup)
			for iObject, object := range group.Objects {
				wi("g p%.2dg%.2do%.2d", iPart, iGroup, iObject)
				if materials != nil && int(object.MaterialId) < len(materials) {
					wi("usemtl %s", materials[object.MaterialId])
				}

				for i, _ := range object.Blocks {
					for _, b := range object.Blocks[i] {
						haveUV := b.Uvs.U != nil
						haveNorm := b.Norms.X != nil

						for iVertex := range b.Trias.X {
							vertex := mgl32.Vec3{b.Trias.X[iVertex], b.Trias.Y[iVertex], b.Trias.Z[iVertex]}

							if bones != nil && b.Joints != nil && object.JointMapper != nil {
								jointId := int(b.Joints[iVertex])
								bone := bones[object.JointMapper[jointId]]
								if lastb != object.JointMapper[jointId] {
									log.Println(jointId, object.JointMapper[jointId], part.JointId)
									log.Println(bone)
									lastb = object.JointMapper[jointId]
								}
								vertex = mgl32.TransformCoordinate(vertex, bone)
							}

							w("v %f %f %f", vertex[0], vertex[1], vertex[2])
							iV++
							if haveUV {
								w("vt %f %f", b.Uvs.U[iVertex], 1.0-b.Uvs.V[iVertex])
								iT++
							}
							if haveNorm {
								w("vn %f %f %f", b.Norms.X[iVertex], b.Norms.Y[iVertex], b.Norms.Z[iVertex])
								iN++
							}
							if !b.Trias.Skip[iVertex] {
								if haveNorm {
									if haveUV {
										wi("f %d/%d/%d %d/%d/%d %d/%d/%d", iV-1, iT-1, iN-1, iV-2, iT-2, iN-2, iV, iT, iN)
									} else {
										wi("f %d//%d %d//%d %d//%d", iV-1, iN-1, iV-2, iN-2, iV, iN)
									}
								} else {
									if haveUV {
										wi("f %d/%d %d/%d %d/%d", iV-1, iT-1, iV-2, iT-2, iV, iT)
									} else {
										wi("f %d %d %d", iV-1, iV-2, iV)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	_w.Write(facesBuff.Bytes())

	return nil
}

func (mesh *Mesh) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "obj":
		var buf bytes.Buffer
		log.Println("Error when exporting mesh: %v", mesh.ExportObj(&buf, nil, nil))
		webutils.WriteFile(w, bytes.NewReader(buf.Bytes()), wrsrc.Tag.Name+".obj")
	}
}
