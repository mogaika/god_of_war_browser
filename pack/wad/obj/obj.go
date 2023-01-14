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
	"github.com/mogaika/god_of_war_browser/pack/wad/anm"
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
	ExternalId  int16
	UnkCoeef    float32

	// flags & 0x0001 != 0, ??? looks like require 0x0800
	// flags & 0x0002 != 0, ???
	// flags & 0x0004 != 0, ???
	// flags & 0x0008 != 0, is matrix external? or rotated additionaly?
	//                      then there is a ExternalId (mat2 index) you need to multiply matrix by mat2 rotation-only mat
	//                      after size calculation, but before translate
	//                      maybe replace rotation matrix?
	// flags & 0x0010 != 0, ??? looks like not used
	// flags & 0x0020 != 0, joint
	// flags & 0x0040 != 0, ignore parent scale??
	// flags & 0x0080 != 0, has inversebindposematrix
	// flags & 0x0100 != 0, ??? looks like not used
	// flags & 0x0200 != 0, rotation changes in animation
	// flags & 0x0400 != 0, position changes in animation
	// flags & 0x0800 != 0, scale changes in animation
	// flags & 0x1000 != 0, ??? probably particle, or maybe external usage.
	//                      example joint names: efxplane, emit, explosion, flash, flare, particleFXsplash,
	// flags & 0x2000 != 0, face camera. vertical rotation only probably.
	//                      example joint names: saveCamFace, camFaceJoint03
	// flags & 0x4000 != 0, halo, glow or flash effect
	// flags & 0x8000 != 0, then quaterion, else euler
	//
	// if flags & 0x6000 != 0  then make strange calculation with matrix invert
	Flags uint32

	IsSkinned   bool
	IsExternal  bool // then not used for rendering?
	IsQuaterion bool

	InvId int16

	BindToJointMat mgl32.Mat4 // bind world joint => local joint
	ParentToJoint  mgl32.Mat4 // idle parent local joint => local joint
	//BindWorldJoint    mgl32.Mat4 // bind world joint
	//OurJointToIdleMat mgl32.Mat4 // idle world joint
	//RenderMat         mgl32.Mat4 // bind world joint => idle world joint
}

const JOINT_CHILD_NONE = -1

type Object struct {
	Joints []Joint

	File0x20 uint32 // index of root joint ?
	File0x24 uint32 /*
		more flags
		0bit - need creation of array(joints count) of words? if dynamic model?
	*/

	dataOffset uint32

	Matrixes1 []mgl32.Mat4 // idle parent local joint => local joint
	Matrixes2 []mgl32.Mat4 // special mat4 for flag & 0x8 != 0 cases (external matrices). Used only as 3x3 matrix (rotation only?)

	// inverse bind pose matrix (only for joints that animated? or rendered? (skinned) or have ident mat for this)
	// Or only for joints which are not blended with other joints for any vertex so meshes can be saved in joint space instead of bind pose
	Matrixes3 []mgl32.Mat4 // bind world joint => local joint

	Vectors4 []mgl32.Vec4 // idle local joint pos xyz
	Vectors5 [][4]int32   // idle local joint pos rot quaterion Q.14fp
	Vectors6 []mgl32.Vec4 // idle local joint pose scale
	Vectors7 []mgl32.Vec4 // unknown (always zero?) unused???
}

func (obj *Object) StringJoint(id int, spaces string) string {
	j := obj.Joints[id]
	return fmt.Sprintf("%sjoint [%.4x <=%.4x %.4x->%.4x %t:%.4x : %v]  %s:\n%srot: %#v\n%spos: %#v\n%srot: %#v\n%ssiz: %#v\n%sv7 : %#v\n",
		spaces, j.Id, j.Parent, j.ChildsStart, j.ChildsEnd, j.IsSkinned, j.InvId, j.UnkCoeef, j.Name,
		spaces, obj.Matrixes1[j.Id], spaces, obj.Vectors4[j.Id],
		spaces, obj.Vectors5[j.Id], spaces, obj.Vectors6[j.Id],
		spaces, obj.Vectors7[j.Id])

	return fmt.Sprintf("[%.4x]%s %s\n", j.Id, spaces, j.Name)
}

func (obj *Object) StringTree() string {
	stack := make([]int, 0, 32)
	spaces := ""

	var buffer bytes.Buffer

	for i := range obj.Joints {
		j := obj.Joints[i]

		if j.Parent != JOINT_CHILD_NONE {
			for i == stack[len(stack)-1] {
				stack = stack[:len(stack)-1]
				spaces = spaces[:len(spaces)-2]
			}
		}

		buffer.WriteString(obj.StringJoint(i, spaces))

		if j.ChildsStart != JOINT_CHILD_NONE {
			if j.ChildsEnd == JOINT_CHILD_NONE && len(stack) > 0 {
				stack = append(stack, stack[len(stack)-1])
			} else {
				stack = append(stack, int(j.ChildsEnd))
			}
			spaces += " -"
		}
	}
	return buffer.String()
}

func NewFromData(buf []byte, objName string) (*Object, error) {
	obj := new(Object)

	obj.Joints = make([]Joint, binary.LittleEndian.Uint32(buf[0x1c:0x20]))
	dataOffset := binary.LittleEndian.Uint32(buf[0x28:0x2c])

	obj.File0x20 = binary.LittleEndian.Uint32(buf[0x20:])
	obj.File0x24 = binary.LittleEndian.Uint32(buf[0x24:])

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

	invid := int16(0)
	for i := range obj.Joints {
		jointBufStart := HEADER_SIZE + i*0x10
		jointBuf := buf[jointBufStart : jointBufStart+0x10]

		nameBufStart := HEADER_SIZE + len(obj.Joints)*0x10 + i*0x18
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
			IsExternal:  flags&0x8 != 0,
			IsSkinned:   flags&0x80 != 0, // || uint32(len(obj.Joints)) == mat3count
			IsQuaterion: flags&0x8000 != 0,
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

	utils.LogDump(obj)

	obj.FillJoints()

	s := ""
	for i, m := range obj.Matrixes3 {
		s += fmt.Sprintf("\n   m3[%.2x]: %f %f %f", i, m[12], m[13], m[14])
	}

	// log.Printf("%s\n%s", s, obj.StringTree())

	return obj, nil
}

func (obj *Object) FillJoints() {
	for i := range obj.Joints {
		j := &obj.Joints[i]
		j.ParentToJoint = obj.Matrixes1[i]

		if j.IsSkinned {
			j.BindToJointMat = obj.Matrixes3[j.InvId]
		} else {
			j.BindToJointMat = mgl32.Ident4()
		}
	}
}

type ObjMarshal struct {
	Data             *Object
	Model            *mdl.Ajax
	Collisions       []*collision.Collision
	Script           *scr.ScriptParams
	Animations       *anm.Animations
	Cameras          []interface{}
	ParticleEmitters []interface{}
	Particles        []interface{}
	SoundEmitters    []interface{}
}

func (obj *Object) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	mrshl := &ObjMarshal{Data: obj}
	for _, id := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id); err == nil {
			switch inst.(type) {
			case *mdl.Model, *scr.ScriptParams, *collision.Collision, *anm.Animations:
				if subFileMarshled, err := inst.Marshal(wrsrc.Wad.GetNodeResourceByNodeId(n.Id)); err != nil {
					log.Panicf("Obj %q marshal problem of subobject %q: %v", wrsrc.Tag.Name, n.Tag.Name, err)
					//continue
				} else {
					switch v := subFileMarshled.(type) {
					case *mdl.Ajax:
						mrshl.Model = v
					case *collision.Collision:
						mrshl.Collisions = append(mrshl.Collisions, v)
					case *scr.ScriptParams:
						mrshl.Script = v
					case *anm.Animations:
						mrshl.Animations = v
					}
				}
			}
		}
	}

	return mrshl, nil
}

const quat_to_float = 1.0 / (1 << 14)

// no conversion from euler
func (o *Object) getQuaterionLocalRotationForJoint(jointId int) mgl32.Quat {
	return mgl32.Quat{
		V: mgl32.Vec3{
			float32(o.Vectors5[jointId][0]),
			float32(o.Vectors5[jointId][1]),
			float32(o.Vectors5[jointId][2]),
		}.Mul(quat_to_float),
		W: float32(o.Vectors5[jointId][3]) * quat_to_float,
	}.Normalize()
}

// with conversion from euler
func (o *Object) GetQuaterionLocalRotationForJoint(jointId int) mgl32.Quat {
	if o.Joints[jointId].IsQuaterion {
		return o.getQuaterionLocalRotationForJoint(jointId).Normalize()
	} else {
		euler := o.getEulerLocalRotationForJoint(jointId)
		return utils.EulerToQuat(euler)
	}
}

// no conversion from quat, result in degrees
func (o *Object) getEulerLocalRotationForJoint(jointId int) mgl32.Vec3 {
	return mgl32.Vec3{
		float32(o.Vectors5[jointId][0]),
		float32(o.Vectors5[jointId][1]),
		float32(o.Vectors5[jointId][2])}.Mul(quat_to_float * 360.0)
}

// with conversion from quat, result in degrees
func (o *Object) GetEulerLocalRotationForJoint(jointId int) mgl32.Vec3 {
	if o.Joints[jointId].IsQuaterion {
		q := o.getQuaterionLocalRotationForJoint(jointId)
		return utils.QuatToEuler(q).Mul(180.0 / math.Pi)
	} else {
		return o.getEulerLocalRotationForJoint(jointId)
	}
}

func init() {
	wad.SetHandler(config.GOW1, OBJECT_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data, wrsrc.Wad.Name()+":"+wrsrc.Tag.Name)
	})
}
