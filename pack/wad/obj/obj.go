package obj

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/mogaika/god_of_war_browser/pack/wad/anm"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/collision"
	"github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	"github.com/mogaika/god_of_war_browser/pack/wad/scr"
	"github.com/mogaika/god_of_war_browser/utils"
)

const OBJECT_MAGIC = 0x00040001
const HEADER_SIZE = 0x2C
const DATA_HEADER_SIZE = 0x30

type Joint struct {
	Id          int16
	Name        string
	ChildsStart int16
	ChildsEnd   int16
	Parent      int16
	UnkCoeef    float32

	IsSkinned bool
	InvId     int16

	BindToJointMat mgl32.Mat4 // bind space
	ParentToJoint  mgl32.Mat4 // idle space??

	OurJointToIdleMat mgl32.Mat4
	RenderMat         mgl32.Mat4
	TransformMat      mgl32.Mat4
}

const JOINT_CHILD_NONE = -1

type Object struct {
	Joints []Joint

	dataOffset  uint32
	jointsCount uint32

	Mat1count  uint32
	Vec2offset uint32
	Vec2count  uint32
	Mat3offset uint32
	Mat3count  uint32
	Vec4offset uint32
	Vec5offset uint32
	Vec6offset uint32
	Vec7offset uint32

	Matrixes1 []mgl32.Mat4 // bind pose
	Vectors2  [][4]uint32
	Matrixes3 []mgl32.Mat4 // inverse bind pose matrix (only for joints that animated? or rendered? (skinned))
	Vectors4  []mgl32.Vec4 // idle pos xyz
	Vectors5  [][4]int32   // idle pos rot Q.14fp
	Vectors6  []mgl32.Vec4 // idle pose scale
	Vectors7  []mgl32.Vec4
}

func (obj *Object) StringJoint(id int16, spaces string) string {
	j := obj.Joints[id]
	/*return fmt.Sprintf("%sjoint [%.4x <=%.4x %.4x->%.4x %t:%.4x : %v]  %s:\n%srot: %#v\n%spos: %#v\n%sv5 : %#v\n%ssiz: %#v\n%sv7 : %#v\n",
	spaces, j.Id, j.Parent, j.ChildsStart, j.ChildsEnd, j.IsSkinned, j.InvId, j.UnkCoeef, j.Name,
	spaces, obj.Matrixes1[j.Id], spaces, obj.Vectors4[j.Id],
	spaces, obj.Vectors5[j.Id], spaces, obj.Vectors6[j.Id],
	spaces, obj.Vectors7[j.Id])
	*/
	return fmt.Sprintf("[%.4x]%s %s\n", j.Id, spaces, j.Name)
}

func (obj *Object) StringTree() string {
	stack := make([]int16, 0, 32)
	spaces := string(make([]byte, 0, 64))
	spaces = ""

	var buffer bytes.Buffer

	for i := int16(0); i < int16(obj.jointsCount); i++ {
		j := obj.Joints[i]

		if j.Parent != JOINT_CHILD_NONE {
			for i == stack[len(stack)-1] {
				stack = stack[:len(stack)-1]
				spaces = spaces[:len(spaces)-2]
			}
		}

		buffer.WriteString(obj.StringJoint(i, spaces))

		if j.ChildsStart != JOINT_CHILD_NONE {
			if j.ChildsEnd == -1 && len(stack) > 0 {
				stack = append(stack, stack[len(stack)-1])
			} else {
				stack = append(stack, j.ChildsEnd)
			}
			spaces += " -"
		}
	}
	return buffer.String()
}

func NewFromData(buf []byte) (*Object, error) {
	obj := new(Object)

	obj.jointsCount = binary.LittleEndian.Uint32(buf[0x1c:0x20])
	obj.dataOffset = binary.LittleEndian.Uint32(buf[0x28:0x2c])
	obj.Joints = make([]Joint, obj.jointsCount)

	matdata := buf[obj.dataOffset : obj.dataOffset+DATA_HEADER_SIZE]

	obj.Mat1count = binary.LittleEndian.Uint32(matdata[0:4])
	obj.Vec2offset = binary.LittleEndian.Uint32(matdata[4:8])
	obj.Vec2count = binary.LittleEndian.Uint32(matdata[8:12])
	obj.Mat3offset = binary.LittleEndian.Uint32(matdata[12:16])
	obj.Mat3count = binary.LittleEndian.Uint32(matdata[16:20])
	obj.Vec4offset = binary.LittleEndian.Uint32(matdata[32:36])
	obj.Vec5offset = binary.LittleEndian.Uint32(matdata[36:40])
	obj.Vec6offset = binary.LittleEndian.Uint32(matdata[40:44])
	obj.Vec7offset = binary.LittleEndian.Uint32(matdata[44:48])

	invid := int16(0)
	for i := range obj.Joints {
		jointBufStart := HEADER_SIZE + i*0x10
		jointBuf := buf[jointBufStart : jointBufStart+0x10]

		nameBufStart := HEADER_SIZE + int(obj.jointsCount)*0x10 + i*0x18
		nameBuf := buf[nameBufStart : nameBufStart+0x18]

		flags := binary.LittleEndian.Uint32(jointBuf[0:4])
		// if flags & 0x6000 != 0
		// then make strange calculation with matrix invert

		isInvMat := flags&0xa0 == 0xa0 || obj.jointsCount == obj.Mat3count
		obj.Joints[i] = Joint{
			Name:        utils.BytesToString(nameBuf[:]),
			ChildsStart: int16(binary.LittleEndian.Uint16(jointBuf[0x4:0x6])),
			ChildsEnd:   int16(binary.LittleEndian.Uint16(jointBuf[0x6:0x8])),
			Parent:      int16(binary.LittleEndian.Uint16(jointBuf[0x8:0xa])),
			UnkCoeef:    math.Float32frombits(binary.LittleEndian.Uint32(jointBuf[0xc:0x10])),
			Id:          int16(i),
			IsSkinned:   isInvMat,
			InvId:       invid,
		}

		if isInvMat {
			invid++
		}
	}

	obj.Matrixes1 = make([]mgl32.Mat4, obj.Mat1count)
	obj.Vectors2 = make([][4]uint32, obj.Vec2count+1)
	obj.Matrixes3 = make([]mgl32.Mat4, obj.Mat3count)
	obj.Vectors4 = make([]mgl32.Vec4, obj.Mat1count)
	obj.Vectors5 = make([][4]int32, obj.Mat1count)
	obj.Vectors6 = make([]mgl32.Vec4, obj.Mat1count)
	obj.Vectors7 = make([]mgl32.Vec4, obj.Mat1count)

	mat1buf := buf[obj.dataOffset+DATA_HEADER_SIZE : obj.dataOffset+DATA_HEADER_SIZE+uint32(len(obj.Matrixes1))*0x40]
	vec2buf := buf[obj.dataOffset+obj.Vec2offset : obj.dataOffset+obj.Vec2offset+uint32(len(obj.Vectors2))*0x10]
	mat3buf := buf[obj.dataOffset+obj.Mat3offset : obj.dataOffset+obj.Mat3offset+uint32(len(obj.Matrixes3))*0x40]
	vec4buf := buf[obj.dataOffset+obj.Vec4offset : obj.dataOffset+obj.Vec4offset+uint32(len(obj.Vectors4))*0x10]
	vec5buf := buf[obj.dataOffset+obj.Vec5offset : obj.dataOffset+obj.Vec5offset+uint32(len(obj.Vectors5))*0x10]
	vec6buf := buf[obj.dataOffset+obj.Vec6offset : obj.dataOffset+obj.Vec6offset+uint32(len(obj.Vectors6))*0x10]
	vec7buf := buf[obj.dataOffset+obj.Vec7offset : obj.dataOffset+obj.Vec7offset+uint32(len(obj.Vectors7))*0x10]

	for i := range obj.Matrixes1 {
		if err := binary.Read(bytes.NewReader(mat1buf[i*0x40:i*0x40+0x40]), binary.LittleEndian, &obj.Matrixes1[i]); err != nil {
			return nil, err
		}
	}
	for i := range obj.Vectors2 {
		if err := binary.Read(bytes.NewReader(vec2buf[i*0x10:i*0x10+0x10]), binary.LittleEndian, &obj.Vectors2[i]); err != nil {
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

	obj.FeelJoints()
	/*
		s := ""
		for i, m := range obj.Matrixes3 {
			s += fmt.Sprintf("\n   m3[%.2x]: %f %f %f", i, m[12], m[13], m[14])
		}

		log.Printf("%s\n%s", s, obj.StringTree())
	*/
	return obj, nil
}

func (obj *Object) FeelJoints() {
	for i := range obj.Joints {
		j := &obj.Joints[i]
		j.ParentToJoint = obj.Matrixes1[i]
		j.TransformMat = mgl32.Ident4()

		if j.IsSkinned {
			j.BindToJointMat = obj.Matrixes3[j.InvId]
		} else {
			j.BindToJointMat = mgl32.Ident4()
		}

		j.OurJointToIdleMat = j.ParentToJoint
		if j.Parent != JOINT_CHILD_NONE {
			j.OurJointToIdleMat = obj.Joints[j.Parent].OurJointToIdleMat.Mul4(j.ParentToJoint)
		}

		if j.IsSkinned {
			j.RenderMat = j.OurJointToIdleMat.Mul4(j.BindToJointMat)
		} else {
			j.RenderMat = j.OurJointToIdleMat
		}
	}
}

type ObjMarshal struct {
	Data       *Object
	Model      interface{}
	Collision  interface{}
	Script     interface{}
	Animations interface{}
}

func (obj *Object) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	mrshl := &ObjMarshal{Data: obj}
	for _, id := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id); err == nil {
			switch inst.(type) {
			case *mdl.Model, *scr.ScriptParams, *collision.Collision, *anm.Animations:
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
					case *anm.Animations:
						mrshl.Animations = subFileMarshled
					}
				}
			}
		}
	}

	return mrshl, nil
}

func init() {
	wad.SetHandler(config.GOW1ps2, OBJECT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
