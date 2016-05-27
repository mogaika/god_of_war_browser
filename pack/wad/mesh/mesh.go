package mesh

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

type MeshPacket struct {
	fileStruct uint32
	Rows       uint16
	Blocks     []*stBlock
}

type MeshObject struct {
	fileStruct uint32
	Type       uint16
	MaterialId uint8
	Packets    []*MeshPacket
}

type MeshGroup struct {
	fileStruct uint32
	Objects    []*MeshObject
}

type MeshPart struct {
	fileStruct uint32
	Groups     []*MeshGroup
	JointId    uint16 // parent joint
}

type Mesh struct {
	CommentStart uint32
	Parts        []*MeshPart
}

const MESH_MAGIC = 0x0001000f

func NewFromData(file []byte, exlog io.Writer) (*Mesh, error) {
	var err error

	u32 := func(idx uint32) uint32 {
		return binary.LittleEndian.Uint32(file[idx : idx+4])
	}
	u16 := func(idx uint32) uint16 {
		return binary.LittleEndian.Uint16(file[idx : idx+2])
	}
	u8 := func(idx uint32) uint8 {
		return file[idx]
	}

	if u32(0) != MESH_MAGIC {
		return nil, fmt.Errorf("Unknown mesh type")
	}

	mdlCommentStart := u32(4)
	if mdlCommentStart > uint32(len(file)) {
		mdlCommentStart = uint32(len(file))
	}

	partsCount := u32(8)
	parts := make([]*MeshPart, partsCount)
	for iPart := range parts {
		pPart := u32(0x50 + uint32(iPart)*4)
		groupsCount := u16(pPart + 2)

		part := &MeshPart{
			fileStruct: pPart,
			Groups:     make([]*MeshGroup, groupsCount),
			JointId:    u16(pPart + 8),
		}
		parts[iPart] = part

		for iGroup := range part.Groups {
			pGroup := pPart + u32(pPart+uint32(iGroup)*4+4)
			objectsCount := u32(pGroup + 4)

			group := &MeshGroup{
				fileStruct: pGroup,
				Objects:    make([]*MeshObject, objectsCount),
			}

			part.Groups[iGroup] = group

			for iObject := range group.Objects {
				pObject := pGroup + u32(0xc+pGroup+uint32(iObject)*4)

				objectType := u16(pObject)
				packetsCount := u32(pObject + 4) //u32(pObject+0xc) * uint32(u8(pObject+0x18))

				/*
					0x1d - static mesh (bridge, skybox)
					0x0e - dynamic? mesh (ship, hero, enemy)
				*/

				object := &MeshObject{
					fileStruct: pObject,
					Type:       objectType,
					Packets:    make([]*MeshPacket, 0),
				}

				group.Objects[iObject] = object

				if objectType == 0xe || objectType == 0x1d || objectType == 0x24 {
					object.MaterialId = u8(pObject + 8)

					for iPacket := uint32(0); iPacket < packetsCount; iPacket++ {
						pPacketInfo := pObject + 0x20 + iPacket*0x10
						pPacket := pObject + u32(pPacketInfo+4)

						packet := &MeshPacket{
							fileStruct: pPacket,
							Rows:       u16(pPacketInfo),
						}

						object.Packets = append(object.Packets, packet)

						packetSize := uint32(packet.Rows) * 0x10
						packetEnd := packetSize + packet.fileStruct

						fmt.Fprintf(exlog, "    packet: %d pos: 0x%.6x rows: 0x%.4x end: 0x%.6x\n",
							iPacket, packet.fileStruct, packet.Rows, packetEnd)

						err, packet.Blocks = VifRead1(file[packet.fileStruct:packetEnd], packet.fileStruct, exlog)
						if err != nil {
							return nil, err
						}
					}
				}

			}
		}
	}

	mesh := &Mesh{
		CommentStart: mdlCommentStart,
		Parts:        parts,
	}

	return mesh, nil
}

func init() {
	wad.SetHandler(MESH_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (interface{}, error) {
		fpath := filepath.Join("logs", w.Name, node.Name+".mesh.log")
		os.MkdirAll(filepath.Dir(fpath), 0777)
		f, _ := os.Create(fpath)

		file := make([]byte, node.Size)
		_, err := r.ReadAt(file, 0)
		if err != nil {
			return nil, err
		}

		return NewFromData(file, f)
	})
}
