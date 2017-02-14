package mesh

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

type stBlock struct {
	Uvs struct {
		U, V []float32
	}
	Trias struct {
		X, Y, Z []float32
		Skip    []bool
	}
	Norms struct {
		X, Y, Z []float32
	}
	Blend struct {
		R, G, B, A []uint16 // actually uint8, only for marshaling
	}
	Joints                 []uint16
	DebugPos               uint32
	HasTransparentBlending bool
}

type MeshObject struct {
	FileStruct  uint32
	Type        uint16
	MaterialId  uint8
	Blocks      [][]stBlock
	BonesUsed   uint16
	JointMapper []uint32
}

type MeshGroup struct {
	FileStruct uint32
	Objects    []*MeshObject
}

type MeshPart struct {
	FileStruct uint32
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
			FileStruct: pPart,
			Groups:     make([]*MeshGroup, groupsCount),
			JointId:    u16(pPart + 4 + uint32(groupsCount)*4),
		}
		parts[iPart] = part

		log.Printf("part %d: %.8x: %.8x  ...  %.8x", iPart, pPart, u32(pPart), u32(pPart+4+uint32(groupsCount)*4))

		for iGroup := range part.Groups {
			pGroup := pPart + u32(pPart+uint32(iGroup)*4+4)
			objectsCount := u32(pGroup + 4)

			group := &MeshGroup{
				FileStruct: pGroup,
				Objects:    make([]*MeshObject, objectsCount),
			}

			part.Groups[iGroup] = group

			log.Printf("- group %d: %.8x: %.8x %.8x[ocnt] %.8x", iGroup, pGroup, u32(pGroup), u32(pGroup+4), u32(pGroup+8))

			for iObject := range group.Objects {
				pObject := pGroup + u32(0xc+pGroup+uint32(iObject)*4)

				objectType := u16(pObject)
				packetsCount := u32(pObject + 4) //u32(pObject+0xc) * uint32(u8(pObject+0x18))

				/*
					0x1d - static mesh (bridge, skybox)
					0x0e - dynamic? mesh (ship, hero, enemy)
				*/

				object := &MeshObject{
					FileStruct: pObject,
					Type:       objectType,
				}

				group.Objects[iObject] = object

				log.Printf("- - object %d: %.8x:", iObject, pObject)
				log.Printf("         %.8x[otype,] %.8x[pckcnt] %.8x[matid,,,] %.8x", u32(pObject), u32(pObject+4), u32(pObject+8), u32(pObject+12))
				log.Printf("         %.8x         %.8x         %.8x           %.8x", u32(pObject+16), u32(pObject+20), u32(pObject+24), u32(pObject+28))

				fmt.Fprintf(exlog, "- - object %d: %.8x:\n", iObject, pObject)
				fmt.Fprintf(exlog, "         %.8x[otype,] %.8x[pckcnt] %.8x[matid,,,] %.8x\n", u32(pObject), u32(pObject+4), u32(pObject+8), u32(pObject+12))
				fmt.Fprintf(exlog, "         %.8x         %.8x         %.8x           %.8x\n", u32(pObject+16), u32(pObject+20), u32(pObject+24), u32(pObject+28))

				if objectType == 0xe || objectType == 0x1d || objectType == 0x24 {
					object.BonesUsed = u16(pObject + 10)
					object.MaterialId = u8(pObject + 8)

					//ds := NewMeshDataStream(file[pObject:], 0, pObject, exlog)
					dmaCalls := u32(pObject+0xc) * uint32(u8(pObject+0x18))
					fmt.Fprintf(exlog, "     --- DMAs: 0x%x * 0x%x = %d\n", u32(pObject+0xc), uint32(u8(pObject+0x18)), dmaCalls)
					object.Blocks = make([][]stBlock, dmaCalls)
					for iDmaChain := uint32(0); iDmaChain < dmaCalls; iDmaChain++ {
						offsetToObjet := 0x20 + iDmaChain*packetsCount*0x10
						pPacket := pObject + offsetToObjet
						fmt.Fprintf(exlog, "     --- DMA Chain --- %d >>>>>>>>>>>>>>\n", iDmaChain)
						ds := NewMeshDataStream(file[:], packetsCount, pPacket, pObject, exlog)

						err = ds.ParsePackets()

						if err != nil {
							return nil, err
						}

						object.Blocks[iDmaChain] = ds.Blocks()
					}

					/*
						rowsParsed := uint32(0)

						for iPacket := uint32(0); iPacket < packetsCount; iPacket++ {
							pPacketInfo := pObject + 0x20 + iPacket*0x10
							pPacket := pObject + u32(pPacketInfo+4)

							packetRows := u16(pPacketInfo)

							object.Packets = append(object.Packets, packet)

							packetSize := uint32(packet.Rows) * 0x10
							rowsParsed += uint32(packet.Rows)
							packetEnd := packetSize + packet.FileStruct

							log.Printf("- - - packet %d: %.8x: %.8x[rowscnt,] %.8x[packoff] %.8x %.8x packDat: %.8x", iPacket, pPacketInfo,
								u32(pPacketInfo), u32(pPacketInfo+4), u32(pPacketInfo+8), u32(pPacketInfo+12), packet.FileStruct)

							fmt.Fprintf(exlog, "    packet: %d pos: 0x%.6x rows: 0x%.4x end: 0x%.6x\n",
								iPacket, packet.FileStruct, packet.Rows, packetEnd)

							err, packet.Blocks = VifRead1(file[packet.FileStruct:packetEnd], packet.FileStruct, exlog)
							if err != nil {
								return nil, err
							}
						}
					*/
					if object.BonesUsed > 0 && len(object.Blocks) > 0 {
						object.JointMapper = make([]uint32, object.BonesUsed)
						pJointMapRaw := pObject + 0x20 + packetsCount*0x10*u32(pObject+0xc)*uint32(u8(pObject+0x18))
						for jointMapIndex := uint32(0); jointMapIndex < uint32(object.BonesUsed); jointMapIndex++ {
							object.JointMapper[jointMapIndex] = u32(pJointMapRaw + jointMapIndex*4)
							if object.JointMapper[jointMapIndex] > 0x1ff {
								return nil, fmt.Errorf("Probably incorrect JointMapper calculation. 0x%x is too large (pMapAddr:0x%x)", object.JointMapper[jointMapIndex], pJointMapRaw)
							}
						}
						fmt.Fprintf(exlog, "    >> jointmap: 0x%.8x => %#+v\n", pJointMapRaw, object.JointMapper)
					}
				} else {
					return nil, fmt.Errorf("Unknown mesh format %")
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

func (m *Mesh) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	return m, nil
}

func init() {
	wad.SetHandler(MESH_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		fpath := filepath.Join("logs", w.Name, fmt.Sprintf("%.4d-%s.mesh.log", node.Id, node.Name))
		os.MkdirAll(filepath.Dir(fpath), 0777)
		f, _ := os.Create(fpath)
		defer f.Close()

		file := make([]byte, node.Size)
		_, err := r.ReadAt(file, 0)
		if err != nil {
			return nil, err
		}

		return NewFromData(file, f)
	})
}
