package obj

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/mdl"
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

	HaveInverse bool
	InvId       int16

	BindToJointMat mgl32.Mat4
	JointToIdleMat mgl32.Mat4
	BindToIdleMat  mgl32.Mat4
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

	Matrixes1 []mgl32.Mat4 // idle pose
	Vectors2  [][4]uint32
	Matrixes3 []mgl32.Mat4 // inverce matrices bind pose (not at all joints, if not present, use idle inverted pose)
	Vectors4  []mgl32.Vec4 // idle pose xyz
	Vectors5  [][4]int32
	Vectors6  []mgl32.Vec4 // idle pose scale
	Vectors7  []mgl32.Vec4
}

func (obj *Object) StringJoint(id int16, spaces string) string {
	j := obj.Joints[id]
	return fmt.Sprintf("%sjoint [%.4x <=%.4x %.4x->%.4x %t:%.4x : %v]  %s:\n%srot: %#v\n%spos: %#v\n%sv5 : %#v\n%ssiz: %#v\n%sv7 : %#v\n",
		spaces, j.Id, j.Parent, j.ChildsStart, j.ChildsEnd, j.HaveInverse, j.InvId, j.UnkCoeef, j.Name,
		spaces, obj.Matrixes1[j.Id], spaces, obj.Vectors4[j.Id],
		spaces, obj.Vectors5[j.Id], spaces, obj.Vectors6[j.Id],
		spaces, obj.Vectors7[j.Id])
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
			spaces += "  "
		}
	}
	return buffer.String()
}

func NewFromData(rdr io.ReaderAt) (*Object, error) {
	var file [HEADER_SIZE]byte
	_, err := rdr.ReadAt(file[:], 0)
	if err != nil {
		return nil, err
	}

	obj := new(Object)

	log.Printf(" OBJ: %.8x %.8x %.8x   %.8x %.8x %.8x",
		binary.LittleEndian.Uint32(file[0x4:0x8]),
		binary.LittleEndian.Uint32(file[0x8:0xc]),
		binary.LittleEndian.Uint32(file[0xc:0x10]),
		binary.LittleEndian.Uint32(file[0x10:0x14]),
		binary.LittleEndian.Uint32(file[0x14:0x18]),
		binary.LittleEndian.Uint32(file[0x18:0x1c]))

	obj.jointsCount = binary.LittleEndian.Uint32(file[0x1c:0x20])
	obj.dataOffset = binary.LittleEndian.Uint32(file[0x28:0x2c])

	obj.Joints = make([]Joint, obj.jointsCount)

	var matdata [DATA_HEADER_SIZE]byte
	_, err = rdr.ReadAt(matdata[:], int64(obj.dataOffset))
	if err != nil {
		return nil, err
	}

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
		var jointBuf [0x10]byte
		var nameBuf [0x18]byte

		_, err = rdr.ReadAt(jointBuf[:], int64(HEADER_SIZE+i*0x10))
		if err != nil {
			return nil, err
		}
		_, err = rdr.ReadAt(nameBuf[:], int64(HEADER_SIZE+int(obj.jointsCount)*0x10+i*0x18))
		if err != nil {
			return nil, err
		}

		flags := binary.LittleEndian.Uint32(jointBuf[0:4])

		isInvMat := flags&0xa0 == 0xa0 || obj.jointsCount == obj.Mat3count
		obj.Joints[i] = Joint{
			Name:        utils.BytesToString(nameBuf[:]),
			ChildsStart: int16(binary.LittleEndian.Uint16(jointBuf[0x4:0x6])),
			ChildsEnd:   int16(binary.LittleEndian.Uint16(jointBuf[0x6:0x8])),
			Parent:      int16(binary.LittleEndian.Uint16(jointBuf[0x8:0xa])),
			UnkCoeef:    math.Float32frombits(binary.LittleEndian.Uint32(jointBuf[0xc:0x10])),
			Id:          int16(i),
			HaveInverse: isInvMat,
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

	mat1buf := make([]byte, len(obj.Matrixes1)*0x40)
	vec2buf := make([]byte, len(obj.Vectors2)*0x10)
	mat3buf := make([]byte, len(obj.Matrixes3)*0x40)
	vec4buf := make([]byte, len(obj.Vectors4)*0x10)
	vec5buf := make([]byte, len(obj.Vectors5)*0x10)
	vec6buf := make([]byte, len(obj.Vectors6)*0x10)
	vec7buf := make([]byte, len(obj.Vectors7)*0x10)

	if _, err = rdr.ReadAt(mat1buf[:], int64(obj.dataOffset+DATA_HEADER_SIZE)); err != nil {
		return nil, err
	}
	if _, err = rdr.ReadAt(vec2buf[:], int64(obj.dataOffset+obj.Vec2offset)); err != nil {
		return nil, err
	}
	if _, err = rdr.ReadAt(mat3buf[:], int64(obj.dataOffset+obj.Mat3offset)); err != nil {
		return nil, err
	}
	if _, err = rdr.ReadAt(vec4buf[:], int64(obj.dataOffset+obj.Vec4offset)); err != nil {
		return nil, err
	}
	if _, err = rdr.ReadAt(vec5buf[:], int64(obj.dataOffset+obj.Vec5offset)); err != nil {
		return nil, err
	}
	if _, err = rdr.ReadAt(vec6buf[:], int64(obj.dataOffset+obj.Vec6offset)); err != nil {
		return nil, err
	}
	if _, err = rdr.ReadAt(vec7buf[:], int64(obj.dataOffset+obj.Vec7offset)); err != nil {
		return nil, err
	}

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

	s := ""
	for i, m := range obj.Matrixes3 {
		s += fmt.Sprintf("\n   m3[%.2x]: %f %f %f", i, m[12], m[13], m[14])
	}

	log.Printf("%s\n%s", s, obj.StringTree())

	return obj, nil
}

func (obj *Object) FeelJoints() {
	for i := range obj.Joints {
		j := &obj.Joints[i]
		if j.HaveInverse {
			j.BindToJointMat = obj.Matrixes3[j.InvId]
		} else {
			if j.Parent != JOINT_CHILD_NONE {
				//j.BindToJointMat = obj.Joints[j.Parent].BindToJointMat.Inv().Mul4(obj.Matrixes1[i]).Inv()
				j.BindToJointMat = obj.Joints[j.Parent].BindToJointMat
			} else {
				//j.BindToJointMat = obj.Matrixes1[i].Inv()
				j.BindToJointMat = mgl32.Ident4()
			}
		}

		j.JointToIdleMat = obj.Matrixes1[i]

		j.BindToIdleMat = j.BindToJointMat.Mul4(j.JointToIdleMat)
	}
}

type ObjMarshal struct {
	Data  *Object
	Model interface{}
}

func (obj *Object) Marshal(wd *wad.Wad, node *wad.WadNode) (interface{}, error) {
	var model interface{}

	for _, id := range node.SubNodes {
		nd := wd.Node(id).ResolveLink()
		if nd.Format == mdl.MODEL_MAGIC {
			modelFile, err := wd.Get(id)
			if err == nil {
				model, err = modelFile.Marshal(wd, nd)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	return &ObjMarshal{
		Data:  obj,
		Model: model,
	}, nil
}

func init() {
	wad.SetHandler(OBJECT_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		return NewFromData(r)
	})
}
