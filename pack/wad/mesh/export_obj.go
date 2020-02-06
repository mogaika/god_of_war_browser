package mesh

import (
	"bytes"
	"fmt"
	"io"
	"math"

	"github.com/go-gl/mathgl/mgl32"
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
	// lastb := uint32(0)

	minimalTextureV := 0.0
	for _, part := range m.Parts {
		for _, group := range part.Groups {
			for _, object := range group.Objects {
				for iPacket := range object.Packets {
					for _, packet := range object.Packets[iPacket] {
						if packet.Uvs.U != nil {
							for _, v := range packet.Uvs.V {
								minimalTextureV = math.Min(minimalTextureV, float64(v))
							}
						}
					}
				}
			}
		}
	}

	_ = math.Floor(-minimalTextureV)

	for iPart, part := range m.Parts {
		for iGroup, group := range part.Groups {
			wi("o p%.2dg%.2d", iPart, iGroup)
			for iObject, object := range group.Objects {
				if materials != nil && int(object.MaterialId) < len(materials) {
					wi("usemtl %s", materials[object.MaterialId])
				}

				for i := range object.Packets {
					wi("g p%.2dg%.2do%.2dg.%2dp", iPart, iGroup, iObject, i)
					for _, packet := range object.Packets[i] {
						haveUV := packet.Uvs.U != nil
						haveNorm := packet.Norms.X != nil

						for iVertex := range packet.Trias.X {
							vertex := mgl32.Vec3{packet.Trias.X[iVertex], packet.Trias.Y[iVertex], packet.Trias.Z[iVertex]}

							/*
								if bones != nil && packet.Joints != nil && object.JointMapper != nil {
									jointId := int(packet.Joints[iVertex])
									bone := bones[object.JointMapper[jointId]]
									if lastb != object.JointMapper[jointId] {
										log.Println(jointId, object.JointMapper[jointId], part.JointId)
										log.Println(bone)
										lastb = object.JointMapper[jointId]
									}
									vertex = mgl32.TransformCoordinate(vertex, bone)
								}
							*/
							w("v %f %f %f", vertex[0], vertex[1], vertex[2])
							iV++
							if haveUV {
								w("vt %f %f", 3.0+packet.Uvs.U[iVertex], 4.0-packet.Uvs.V[iVertex])
								iT++
							}
							if haveNorm {
								w("vn %f %f %f", packet.Norms.X[iVertex], packet.Norms.Y[iVertex], packet.Norms.Z[iVertex])
								iN++
							}
							if !packet.Trias.Skip[iVertex] {
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
