package twktree

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type FileType int

const (
	FILE_TYPE_FLOAT FileType = iota
	FILE_TYPE_INT32
	FILE_TYPE_INT16
	FILE_TYPE_INT8
	FILE_TYPE_BOOL
	FILE_TYPE_STRING
	FILE_TYPE_BYTES
)

type VFSNode struct {
	Name   string
	Value  []byte
	Fields []*VFSNode
}

var _ yaml.Marshaler = (*VFSAbstractNode)(nil)

type yamlNodeTypeCommenter struct {
	convVal interface{}
}

func (tc *yamlNodeTypeCommenter) MarshalYAML() (interface{}, error) {
	log.Printf("tc %T", tc.convVal)
	switch val := tc.convVal.(type) {
	case int8, int16, int32, int, float32:
		return yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprint(val),
			LineComment: fmt.Sprintf("// original type %T", val),
		}, nil
	default:
		return val, nil
	}
}

type VFSAbstractNode struct {
	Name        string
	Declaration VFSDeclaration
	Value       interface{}
	Fields      []*VFSAbstractNode
}

type abstractNodeYAMLType struct {
	Name   yaml.Node
	Value  interface{}        `yaml:",omitempty"`
	Fields []*VFSAbstractNode `yaml:",omitempty"`
}

func (an *VFSAbstractNode) MarshalYAML() (interface{}, error) {
	value := an.Value
	if value != nil {
		switch convVal := value.(type) {
		case int8, uint8, int16, uint16, int32, uint32, int, uint, float32, float64:
			value = &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       fmt.Sprint(convVal),
				LineComment: fmt.Sprintf("%T", convVal),
			}
		case string:
			value = &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       convVal,
				LineComment: fmt.Sprintf("char[%d]", an.Declaration.(*VFSFieldDeclaration).Size),
			}
		}
	}
	return &abstractNodeYAMLType{
		Name: yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       an.Name,
			LineComment: fmt.Sprintf("hash 0x%.8x", utils.GameStringHashNodes(an.Name, 0)),
		},
		Value:  value,
		Fields: an.Fields,
	}, nil
}

func NewVFSNode(name string) *VFSNode {
	return &VFSNode{
		Name:   name,
		Fields: nil,
		Value:  nil,
	}
}

type Marshaler interface {
	MarshalTWK(*VFSAbstractNode) (*VFSNode, error)
}

type Unmarshaler interface {
	UnmarshalTWK(*VFSNode) (*VFSAbstractNode, error)
}

type VFSDeclaration interface {
	Marshaler
	Unmarshaler
}

type VFSFieldDeclaration struct {
	Type FileType
	Size int // for string or bytes
}

func NewFieldDeclaration(ft FileType, size int) *VFSFieldDeclaration {
	return &VFSFieldDeclaration{Type: ft, Size: size}
}

func (fd *VFSFieldDeclaration) MarshalTWK(an *VFSAbstractNode) (*VFSNode, error) {
	if len(an.Fields) != 0 {
		return nil, errors.Errorf("Field %q contains fields when expected value	", an.Name)
	}

	an.Declaration = fd
	n := NewVFSNode(an.Name)

	buf := make([]byte, 4)

	convVal := an.Value
	// normalize numbers

	convToNum := func(num int64) interface{} {
		if fd.Type == FILE_TYPE_FLOAT {
			return float32(num)
		} else {
			return int32(num)
		}
	}

	switch val := convVal.(type) {
	case int8:
		convVal = convToNum(int64(int8(val)))
	case uint8:
		convVal = convToNum(int64(int8(val)))
	case int16:
		convVal = convToNum(int64(int16(val)))
	case uint16:
		convVal = convToNum(int64(int16(val)))
	case int32:
		convVal = convToNum(int64(int32(val)))
	case uint32:
		convVal = convToNum(int64(int32(val)))
	case int:
		convVal = convToNum(int64(int(val)))
	case uint:
		convVal = convToNum(int64(int(val)))
	case float64:
		convVal = float32(val)
	}

	switch val := convVal.(type) {
	case bool:
		if fd.Type != FILE_TYPE_BOOL {
			return nil, errors.Errorf("Expected bool")
		}
		if val {
			buf[0] = 1
		}
	case int32:
		if fd.Type != FILE_TYPE_INT8 && fd.Type != FILE_TYPE_INT16 && fd.Type != FILE_TYPE_INT32 {
			return nil, errors.Errorf("Expected int, got %T", val)
		}
		binary.LittleEndian.PutUint32(buf, uint32(val))
	case float32:
		if fd.Type != FILE_TYPE_FLOAT {
			return nil, errors.Errorf("Expected float, got %T", val)
		}
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(val))
	case string:
		if fd.Type != FILE_TYPE_STRING {
			return nil, errors.Errorf("Expected string, got %T", val)
		}
		if len(val) > fd.Size {
			return nil, errors.Errorf("String too long (%d > %d)", len(val), fd.Size)
		}
		buf = utils.StringToBytes(val, true)
	case []byte:
		if fd.Type != FILE_TYPE_BYTES {
			return nil, errors.Errorf("Expected byte array, got %T", val)
		}
		if len(val) > fd.Size {
			return nil, errors.Errorf("Byte array too long (%d > %d)", len(val), fd.Size)
		}
		buf = val
	case nil:
		if fd.Type == FILE_TYPE_STRING {
			// empty string
			buf = buf[:1]
		} else {
			return nil, errors.Errorf("Got nil")
		}
	default:
		return nil, errors.Errorf("Unknown type %T for field declaration %d", val, fd.Type)
	}
	n.Value = buf
	return n, nil
}

func (fd *VFSFieldDeclaration) UnmarshalTWK(n *VFSNode) (*VFSAbstractNode, error) {
	if len(n.Fields) != 0 {
		return nil, errors.Errorf("Node should not contain childs")
	}
	an := &VFSAbstractNode{
		Name:        n.Name,
		Declaration: fd,
	}
	switch fd.Type {
	case FILE_TYPE_BOOL:
		an.Value = n.Value[0] != 0
	case FILE_TYPE_BYTES:
		data := make([]byte, len(n.Value))
		copy(data, n.Value)
		an.Value = data
	case FILE_TYPE_FLOAT:
		an.Value = math.Float32frombits(binary.LittleEndian.Uint32(n.Value))
	case FILE_TYPE_INT8:
		an.Value = int8(n.Value[0])
	case FILE_TYPE_INT16:
		an.Value = int16(binary.LittleEndian.Uint16(n.Value))
	case FILE_TYPE_INT32:
		an.Value = int32(binary.LittleEndian.Uint32(n.Value))
	case FILE_TYPE_STRING:
		s := utils.BytesToString(n.Value)
		if len(s) > fd.Size {
			return nil, errors.Errorf("String too long for this field (limit %d)", fd.Size)
		}
		an.Value = s
	default:
		return nil, errors.Errorf("Unknown parser type %d for field %q", fd.Type, an.Name)
	}

	return an, nil
}

type VFSDirectoryDeclarationField struct {
	Name        string
	Declaration VFSDeclaration
}

type VFSDirectoryDeclaration struct {
	Fields []VFSDirectoryDeclarationField
}

func NewDirectoryDeclaration() *VFSDirectoryDeclaration {
	return &VFSDirectoryDeclaration{Fields: make([]VFSDirectoryDeclarationField, 0, 8)}
}

func (dd *VFSDirectoryDeclaration) getDeclaration(name string) (VFSDeclaration, bool) {
	for i := range dd.Fields {
		if dd.Fields[i].Name == name {
			return dd.Fields[i].Declaration, true
		}
	}
	return nil, false
}

func (dd *VFSDirectoryDeclaration) MarshalTWK(an *VFSAbstractNode) (*VFSNode, error) {
	if an.Value != nil {
		return nil, errors.Errorf("Got non-empty value when marshaling directory")
	}

	an.Declaration = dd

	result := make([]*VFSNode, 0, len(an.Fields))

	for _, field := range an.Fields {
		fd, ok := dd.getDeclaration(field.Name)
		if !ok {
			return nil, errors.Errorf("Wasn't able to find declaration for field %q", field.Name)
		}

		n, err := fd.MarshalTWK(field)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse field %q", field.Name)
		}

		result = append(result, n)
	}

	return &VFSNode{Name: an.Name, Fields: result}, nil
}

func (dd *VFSDirectoryDeclaration) UnmarshalTWK(n *VFSNode) (*VFSAbstractNode, error) {
	if len(n.Value) != 0 {
		return nil, errors.Errorf("Got non-empty value when unmarshaling directory")
	}

	result := make([]*VFSAbstractNode, 0, len(n.Fields))

	for _, child := range n.Fields {
		fd, ok := dd.getDeclaration(child.Name)
		if !ok {
			return nil, errors.Errorf("Wasn't able to find declaration for field %q", child.Name)
		}

		field, err := fd.UnmarshalTWK(child)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse field %q", child.Name)
		}

		result = append(result, field)
	}

	return &VFSAbstractNode{Name: n.Name, Fields: result, Declaration: dd}, nil
}

func (dd *VFSDirectoryDeclaration) AddField(name string, field VFSDeclaration) {
	utils.GameStringHashRemember(name)
	dd.Fields = append(dd.Fields, VFSDirectoryDeclarationField{Name: name, Declaration: field})
}

// add string or bytes
func (dd *VFSDirectoryDeclaration) AddFieldA(name string, ft FileType, size int) {
	dd.AddField(name, &VFSFieldDeclaration{Type: ft, Size: size})
}

// add int, float, bool
func (dd *VFSDirectoryDeclaration) AddFieldN(name string, ft FileType) {
	dd.AddField(name, &VFSFieldDeclaration{Type: ft})
}

type VFSIndexedDirectoryDeclaration struct {
	dec VFSDeclaration
}

func NewIndexedDirectoryDeclaration(dec VFSDeclaration) *VFSIndexedDirectoryDeclaration {
	return &VFSIndexedDirectoryDeclaration{dec: dec}
}

func (id *VFSIndexedDirectoryDeclaration) MarshalTWK(an *VFSAbstractNode) (*VFSNode, error) {
	if an.Value != nil {
		return nil, errors.Errorf("Got non-empty value when marshaling directory")
	}

	result := make([]*VFSNode, 0, len(an.Fields))

	for _, field := range an.Fields {
		n, err := id.dec.MarshalTWK(field)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse field %q", field.Name)
		}

		result = append(result, n)
	}

	return &VFSNode{Name: an.Name, Fields: result}, nil
}

func (id *VFSIndexedDirectoryDeclaration) UnmarshalTWK(n *VFSNode) (*VFSAbstractNode, error) {
	if len(n.Value) != 0 {
		return nil, errors.Errorf("Got non-empty value when unmarshaling directory")
	}

	result := make([]*VFSAbstractNode, 0, len(n.Fields))
	for _, child := range n.Fields {
		field, err := id.dec.UnmarshalTWK(child)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse field %q", child.Name)
		}

		result = append(result, field)
	}

	return &VFSAbstractNode{Name: n.Name, Fields: result, Declaration: id}, nil
}
