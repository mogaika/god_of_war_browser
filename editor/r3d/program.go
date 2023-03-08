package r3d

import (
	_ "embed"
	"unsafe"

	"log"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

//go:embed shaders/default.vert
var defaultVertexShader string

//go:embed shaders/default.frag
var defaultFragmentShader string

var defaultProgram *DefaultProgram

const default_program_bone_matrices_count = 200

type DefaultProgram struct {
	*Program

	uboBones uint32

	UProjectView int32
	UModel       int32
	UTexture     int32
	UColor       int32
	UUseTexture  int32
	UUseBones    int32

	APosition    int32
	AColor       int32
	ANormal      int32
	AUV          int32
	ABoneIndices int32
	ABoneWeights int32
}

func (dp *DefaultProgram) SetBoneMatrices(bones []mgl32.Mat4) {
	if len(bones) >= default_program_bone_matrices_count {
		panic(len(bones))
	} else if len(bones) == 0 {
		return
	}

	gl.BindBuffer(gl.UNIFORM_BUFFER, dp.uboBones)
	gl.BufferSubData(gl.UNIFORM_BUFFER, 0, len(bones)*int(unsafe.Sizeof(mgl32.Mat4{})), gl.Ptr(bones))
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
}

func GetDefaultProgram() *DefaultProgram {
	if defaultProgram == nil {
		dp := &DefaultProgram{}
		defaultProgram = dp

		dp.Program = MustLoadProgram(defaultVertexShader, defaultFragmentShader)

		bonesSize := int(default_program_bone_matrices_count * unsafe.Sizeof(mgl32.Mat4{}))

		// allocate memory for bones
		gl.GenBuffers(1, &dp.uboBones)
		gl.BindBuffer(gl.UNIFORM_BUFFER, dp.uboBones)
		gl.BufferData(gl.UNIFORM_BUFFER, bonesSize, nil, gl.STREAM_DRAW)
		gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
		// and connect them to shader
		gl.BindBufferRange(gl.UNIFORM_BUFFER, 0, dp.uboBones, 0, bonesSize)

		dp.SetBoneMatrices([]mgl32.Mat4{mgl32.Ident4()})

		dp.UProjectView = gl.GetUniformLocation(dp.Id, gl.Str("umProjectView\x00"))
		dp.UModel = gl.GetUniformLocation(dp.Id, gl.Str("umModel\x00"))
		dp.UTexture = gl.GetUniformLocation(dp.Id, gl.Str("uTexture\x00"))
		dp.UColor = gl.GetUniformLocation(dp.Id, gl.Str("uColor\x00"))
		dp.UUseTexture = gl.GetUniformLocation(dp.Id, gl.Str("uUseTexture\x00"))
		dp.UUseBones = gl.GetUniformLocation(dp.Id, gl.Str("uUseBones\x00"))

		dp.APosition = gl.GetAttribLocation(dp.Id, gl.Str("aPosition"+"\x00"))
		dp.AColor = gl.GetAttribLocation(dp.Id, gl.Str("aColor"+"\x00"))
		dp.ANormal = gl.GetAttribLocation(dp.Id, gl.Str("aNormal"+"\x00"))
		dp.AUV = gl.GetAttribLocation(dp.Id, gl.Str("aUV"+"\x00"))
		dp.ABoneIndices = gl.GetAttribLocation(dp.Id, gl.Str("aBoneIndices"+"\x00"))
		dp.ABoneWeights = gl.GetAttribLocation(dp.Id, gl.Str("aBoneWeights"+"\x00"))
	}
	return defaultProgram
}

type Program struct {
	Id                           uint32
	VertexShader, FragmentShader uint32
}

func (p *Program) Delete() {
	gl.DetachShader(p.Id, p.VertexShader)
	gl.DetachShader(p.Id, p.FragmentShader)
	gl.DeleteProgram(p.Id)
	gl.DeleteShader(p.VertexShader)
	gl.DeleteShader(p.FragmentShader)
}

func LoadProgram(vertexShaderText, fragmentShaderText string) (*Program, error) {
	p := &Program{}

	p.Id = gl.CreateProgram()

	if vs, err := LoadShader(gl.VERTEX_SHADER, vertexShaderText); err != nil {
		return nil, errors.Wrap(err, "vertex shader")
	} else {
		p.VertexShader = vs
	}

	if fs, err := LoadShader(gl.FRAGMENT_SHADER, fragmentShaderText); err != nil {
		gl.DeleteShader(p.VertexShader)
		return nil, errors.Wrap(err, "fragment shader")
	} else {
		p.FragmentShader = fs
	}

	gl.AttachShader(p.Id, p.VertexShader)
	gl.AttachShader(p.Id, p.FragmentShader)
	gl.LinkProgram(p.Id)

	var isLinked int32
	gl.GetProgramiv(p.Id, gl.LINK_STATUS, &isLinked)
	if isLinked == gl.FALSE {
		var logSize int32
		gl.GetProgramiv(p.Id, gl.INFO_LOG_LENGTH, &logSize)
		buf := make([]uint8, logSize+1)
		gl.GetProgramInfoLog(p.Id, int32(len(buf)), &logSize, &buf[0])
		errString := string(buf[:logSize])
		log.Printf("Failed to link program:\n%s", errString)

		p.Delete()
		return nil, errors.Errorf("failed to link program: %q", errString)
	} else {
		return p, nil
	}
}

func MustLoadProgram(vertexShaderText, fragmentShaderText string) *Program {
	program, err := LoadProgram(vertexShaderText, fragmentShaderText)
	if err != nil {
		panic(err)
	}
	return program
}

func LoadShader(xtype uint32, text string) (shader uint32, err error) {
	glShaderSource := func(handle uint32, source string) {
		csource, free := gl.Strs(source + "\x00")
		defer free()

		gl.ShaderSource(handle, 1, csource, nil)
	}

	shader = gl.CreateShader(xtype)
	glShaderSource(shader, text)
	gl.CompileShader(shader)

	var success int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		var logSize int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logSize)
		buf := make([]uint8, logSize+1)
		gl.GetShaderInfoLog(shader, int32(len(buf)), &logSize, &buf[0])
		errString := string(buf[:logSize])
		log.Printf("Failed to compile shader:\n%s", errString)

		gl.DeleteShader(shader)
		return gl.INVALID_INDEX, errors.Errorf("failed to compile shader: %q", errString)
	} else {
		return shader, nil
	}
}
