package obj

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const OBJECT_MAGIC_GOW2 = 0x00010001

func NewFromDataGow2(buf []byte, objName string) (*Object, error) {
	obj := new(Object)

	obj.Joints = make([]Joint, binary.LittleEndian.Uint32(buf[4:8]))

	dataOffset := binary.LittleEndian.Uint32(buf[0x10:0x14])

	matdata := buf[dataOffset : dataOffset+DATA_HEADER_SIZE]

	mat1count := binary.LittleEndian.Uint32(matdata[0:4])
	mat2offset := binary.LittleEndian.Uint32(matdata[4:8])
	mat2count := binary.LittleEndian.Uint32(matdata[8:12])
	mat3offset := binary.LittleEndian.Uint32(matdata[12:16])
	mat3count := binary.LittleEndian.Uint32(matdata[16:20])
	vec4offset := binary.LittleEndian.Uint32(matdata[32:36])
	vec5offset := binary.LittleEndian.Uint32(matdata[36:40])
	vec6offset := binary.LittleEndian.Uint32(matdata[40:44])
	vec7offset := binary.LittleEndian.Uint32(matdata[44:48])

	// called := false

	log.Printf("mat1 %d mat2 %d mat3 %d", mat1count, mat2count, mat3count)

	invid := int16(0)
	for i := range obj.Joints {
		jointBufStart := 0x14 + i*0x10
		jointBuf := buf[jointBufStart : jointBufStart+0x10]

		nameBufStart := 0x14 + len(obj.Joints)*0x10 + i*0x18
		nameBuf := buf[nameBufStart : nameBufStart+0x18]

		flags := binary.LittleEndian.Uint32(jointBuf[0:4])

		obj.Joints[i] = Joint{
			Name:        utils.BytesToString(nameBuf[:]),
			Flags:       flags,
			ChildsStart: int16(binary.LittleEndian.Uint16(jointBuf[0x4:0x6])),
			ChildsEnd:   int16(binary.LittleEndian.Uint16(jointBuf[0x6:0x8])),
			Parent:      int16(binary.LittleEndian.Uint16(jointBuf[0x8:0xa])),
			ExternalId:  int16(binary.LittleEndian.Uint16(jointBuf[0xa:0xc])),
			UnkCoeef:    math.Float32frombits(binary.LittleEndian.Uint32(jointBuf[0xc:0x10])),
			Id:          int16(i),
			IsSkinned:   flags&0x80 != 0, // || uint32(len(obj.Joints)) == mat3count
			IsExternal:  flags&0x8 != 0,
			InvId:       invid,
		}

		joint := &obj.Joints[i]

		/*
			fti := func(flag uint32) int {
				if flags&flag != 0 {
					return 1
				} else {
					return 0
				}
			}

			//if fti(8) != 0 {
			if !called {
				called = true
				log.Printf("loading object %q", objName)
			}
			log.Printf("[%03d] p %03d ce %03d 0x%.8x ext %v joint %v ignParScale %v hasIPB %v strngInv %v quat %v %q r %v",
				i,
				joint.Parent, joint.ChildsEnd,
				joint.Flags,
				fti(0x8), fti(0x20), fti(0x40), fti(0x80), fti(0x4000),
				fti(0x8000),
				joint.Name,
			)
			//}
		*/
		_ = joint
		if joint.IsSkinned {
			invid++
		}
	}
	if invid != int16(mat3count) {
		return nil, fmt.Errorf("Invalid inv mat id calculation %v != %v", invid, mat3count)
	}

	// log.Println(obj.File0x20, obj.File0x24, obj.jointsCount)
	if obj.File0x20 != 0 {
		return nil, fmt.Errorf("Invalid File0x20 == 0x%x", obj.File0x20)
	}

	obj.Matrixes1 = make([]mgl32.Mat4, mat1count)
	obj.Matrixes2 = make([]mgl32.Mat4, mat2count)
	obj.Matrixes3 = make([]mgl32.Mat4, mat3count)
	obj.Vectors4 = make([]mgl32.Vec4, mat1count)
	obj.Vectors5 = make([][4]int32, mat1count)
	obj.Vectors6 = make([]mgl32.Vec4, mat1count)
	obj.Vectors7 = make([]mgl32.Vec4, mat1count)

	mat1buf := buf[dataOffset+DATA_HEADER_SIZE : dataOffset+DATA_HEADER_SIZE+uint32(len(obj.Matrixes1))*0x40]
	mat2buf := buf[dataOffset+mat2offset : dataOffset+mat2offset+uint32(len(obj.Matrixes2))*0x40]
	mat3buf := buf[dataOffset+mat3offset : dataOffset+mat3offset+uint32(len(obj.Matrixes3))*0x40]
	vec4buf := buf[dataOffset+vec4offset : dataOffset+vec4offset+uint32(len(obj.Vectors4))*0x10]
	vec5buf := buf[dataOffset+vec5offset : dataOffset+vec5offset+uint32(len(obj.Vectors5))*0x10]
	vec6buf := buf[dataOffset+vec6offset : dataOffset+vec6offset+uint32(len(obj.Vectors6))*0x10]
	vec7buf := buf[dataOffset+vec7offset : dataOffset+vec7offset+uint32(len(obj.Vectors7))*0x10]

	for i := range obj.Matrixes1 {
		if err := binary.Read(bytes.NewReader(mat1buf[i*0x40:i*0x40+0x40]), binary.LittleEndian, &obj.Matrixes1[i]); err != nil {
			return nil, err
		}
	}
	for i := range obj.Matrixes2 {
		if err := binary.Read(bytes.NewReader(mat2buf[i*0x40:i*0x40+0x40]), binary.LittleEndian, &obj.Matrixes2[i]); err != nil {
			return nil, err
		}
	}
	for i := range obj.Matrixes3 {
		if err := binary.Read(bytes.NewReader(mat3buf[i*0x40:i*0x40+0x40]), binary.LittleEndian, &obj.Matrixes3[i]); err != nil {
			return nil, err
		}
	}
	for i := range obj.Vectors4 {
		if err := binary.Read(bytes.NewReader(vec4buf[i*0x10:i*0x10+0x10]), binary.LittleEndian, &obj.Vectors4[i]); err != nil {
			return nil, err
		}
		if err := binary.Read(bytes.NewReader(vec5buf[i*0x10:i*0x10+0x10]), binary.LittleEndian, &obj.Vectors5[i]); err != nil {
			return nil, err
		}
		if err := binary.Read(bytes.NewReader(vec6buf[i*0x10:i*0x10+0x10]), binary.LittleEndian, &obj.Vectors6[i]); err != nil {
			return nil, err
		}
		if err := binary.Read(bytes.NewReader(vec7buf[i*0x10:i*0x10+0x10]), binary.LittleEndian, &obj.Vectors7[i]); err != nil {
			return nil, err
		}
	}

	//utils.LogDump(obj)

	obj.FillJoints()
	/*
		s := ""
		for i, m := range obj.Matrixes3 {
			s += fmt.Sprintf("\n   m3[%.2x]: %f %f %f", i, m[12], m[13], m[14])
		}
	*/
	// log.Printf("%s\n%s", s, obj.StringTree())

	return obj, nil
}

func init() {
	wad.SetServerHandler(config.GOW2, OBJECT_MAGIC_GOW2, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromDataGow2(wrsrc.Tag.Data, wrsrc.Wad.Name()+":"+wrsrc.Tag.Name)
	})
}
