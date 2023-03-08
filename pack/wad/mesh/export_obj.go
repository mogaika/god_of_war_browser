package mesh

import (
	"fmt"
	"io"
)

func (m *Mesh) ExportObj(_w io.Writer, materials []string) error {
	w := func(format string, args ...interface{}) {
		_w.Write(([]byte)(fmt.Sprintf(format+"\n", args...)))
	}

	cMesh := m.AsCommonMesh()

	for _, part := range cMesh.Parts {
		for _, group := range part.LodGroups {
			for _, object := range group.Objects {
				for _, vertex := range object.Vertices {
					w("v %f %f %f", vertex.Position[0], vertex.Position[1], vertex.Position[2])
				}

				if object.UVs != nil && len(object.UVs) > 0 {
					for _, uv := range object.UVs[0] {
						w("vt %f %f", uv[0], -uv[1])
					}
				}

				if object.Normals != nil {
					for _, normal := range object.Normals {
						w("vn %f %f %f", normal[0], normal[1], normal[2])
					}
				}
			}
		}
	}

	iV := uint32(1)
	iT := uint32(1)
	iN := uint32(1)

	for iPart, part := range cMesh.Parts {
		for iGroup, group := range part.LodGroups {
			for iObject, object := range group.Objects {
				w("o p%.2dlod%.2do%.2dm%.2d", iPart, iGroup, iObject, object.MaterialIndex)
				if materials != nil {
					w("usemtl %s", materials[object.MaterialIndex])
				}

				haveUV := object.UVs != nil && len(object.UVs) > 0
				haveNorm := object.Normals != nil

				for iIndex := 0; iIndex < len(object.Indexes); iIndex += 3 {
					indexes := object.Indexes[iIndex : iIndex+3]

					if haveNorm {
						if haveUV {
							w("f %v/%v/%v %v/%v/%v %v/%v/%v",
								iV+indexes[0], iT+indexes[0], iN+indexes[0],
								iV+indexes[1], iT+indexes[1], iN+indexes[1],
								iV+indexes[2], iT+indexes[2], iN+indexes[2])
						} else {
							w("f %v//%v %v//%v %v//%v",
								iV+indexes[0], iN+indexes[0],
								iV+indexes[1], iN+indexes[1],
								iV+indexes[2], iN+indexes[2])
						}
					} else {
						if haveUV {
							w("f %v/%v %v/%v %v/%v",
								iV+indexes[0], iT+indexes[0],
								iV+indexes[1], iT+indexes[1],
								iV+indexes[2], iT+indexes[2])
						} else {
							w("f %v %v %v",
								iV+indexes[0],
								iV+indexes[1],
								iV+indexes[2])
						}
					}
				}

				iV += uint32(len(object.Vertices))
				if haveUV {
					iT += uint32(len(object.BlendColors[0]))
				}
				if haveNorm {
					iN += uint32(len(object.Normals))
				}
			}
		}
	}

	return nil
}
