package uibackend

import (
	_ "embed" // using embed for the shader sources
	"fmt"
	"log"
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/r3d"
)

//go:embed shaders/main.vert
var textVertexShader string

//go:embed shaders/main.frag
var textFragmentShader string

// OpenGL4 implements a renderer based on github.com/go-gl/gl (v4.3-core).
type OpenGL4 struct {
	imguiIO imgui.IO

	program r3d.Program

	fontTexture            uint32
	attribLocationTex      int32
	attribLocationProjMtx  int32
	attribLocationPosition int32
	attribLocationUV       int32
	attribLocationColor    int32
	vboHandle              uint32
	elementsHandle         uint32
}

// NewOpenGL4 attempts to initialize a renderer.
// An OpenGL context has to be established before calling this function.
func NewOpenGL4(io imgui.IO) (*OpenGL4, error) {
	err := gl.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenGL: %w", err)
	}
	r := &OpenGL4{imguiIO: io}

	gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS)
	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(openglLogCallback, nil)

	r.createDeviceObjects()

	io.SetBackendFlags(io.GetBackendFlags() | imgui.BackendFlagsRendererHasVtxOffset)

	return r, nil
}

// Dispose cleans up the resources.
func (r *OpenGL4) Destroy() {
	r.invalidateDeviceObjects()
}

// PreRender clears the framebuffer.
func (r *OpenGL4) PreRender(clearColor [3]float32) {
	gl.ClearColor(clearColor[0], clearColor[1], clearColor[2], 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

// Render translates the ImGui draw data to OpenGL3 commands.
func (r *OpenGL4) Render(displaySize [2]float32, framebufferSize [2]float32, drawData imgui.DrawData) {
	// Avoid rendering when minimized, scale coordinates for retina displays (screen coordinates != framebuffer coordinates)
	displayWidth, displayHeight := displaySize[0], displaySize[1]
	fbWidth, fbHeight := framebufferSize[0], framebufferSize[1]
	if (fbWidth <= 0) || (fbHeight <= 0) {
		return
	}
	drawData.ScaleClipRects(imgui.Vec2{
		X: fbWidth / displayWidth,
		Y: fbHeight / displayHeight,
	})

	// Backup GL state
	var lastActiveTexture int32
	gl.GetIntegerv(gl.ACTIVE_TEXTURE, &lastActiveTexture)
	gl.ActiveTexture(gl.TEXTURE0)
	var lastProgram int32
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &lastProgram)
	var lastTexture int32
	gl.GetIntegerv(gl.TEXTURE_BINDING_2D, &lastTexture)
	var lastSampler int32
	gl.GetIntegerv(gl.SAMPLER_BINDING, &lastSampler)
	var lastArrayBuffer int32
	gl.GetIntegerv(gl.ARRAY_BUFFER_BINDING, &lastArrayBuffer)
	var lastElementArrayBuffer int32
	gl.GetIntegerv(gl.ELEMENT_ARRAY_BUFFER_BINDING, &lastElementArrayBuffer)
	var lastVertexArray int32
	gl.GetIntegerv(gl.VERTEX_ARRAY_BINDING, &lastVertexArray)
	var lastPolygonMode [2]int32
	gl.GetIntegerv(gl.POLYGON_MODE, &lastPolygonMode[0])
	var lastViewport [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &lastViewport[0])
	var lastScissorBox [4]int32
	gl.GetIntegerv(gl.SCISSOR_BOX, &lastScissorBox[0])
	var lastBlendSrcRgb int32
	gl.GetIntegerv(gl.BLEND_SRC_RGB, &lastBlendSrcRgb)
	var lastBlendDstRgb int32
	gl.GetIntegerv(gl.BLEND_DST_RGB, &lastBlendDstRgb)
	var lastBlendSrcAlpha int32
	gl.GetIntegerv(gl.BLEND_SRC_ALPHA, &lastBlendSrcAlpha)
	var lastBlendDstAlpha int32
	gl.GetIntegerv(gl.BLEND_DST_ALPHA, &lastBlendDstAlpha)
	var lastBlendEquationRgb int32
	gl.GetIntegerv(gl.BLEND_EQUATION_RGB, &lastBlendEquationRgb)
	var lastBlendEquationAlpha int32
	gl.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &lastBlendEquationAlpha)
	lastEnableBlend := gl.IsEnabled(gl.BLEND)
	lastEnableCullFace := gl.IsEnabled(gl.CULL_FACE)
	lastEnableDepthTest := gl.IsEnabled(gl.DEPTH_TEST)
	lastEnableScissorTest := gl.IsEnabled(gl.SCISSOR_TEST)

	// Setup render state: alpha-blending enabled, no face culling, no depth testing, scissor enabled, polygon fill
	gl.Enable(gl.BLEND)
	gl.BlendEquation(gl.FUNC_ADD)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.SCISSOR_TEST)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

	// Setup viewport, orthographic projection matrix
	// Our visible imgui space lies from draw_data->DisplayPos (top left) to draw_data->DisplayPos+data_data->DisplaySize (bottom right).
	// DisplayMin is typically (0,0) for single viewport apps.
	gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))
	orthoProjection := [4][4]float32{
		{2.0 / displayWidth, 0.0, 0.0, 0.0},
		{0.0, 2.0 / -displayHeight, 0.0, 0.0},
		{0.0, 0.0, -1.0, 0.0},
		{-1.0, 1.0, 0.0, 1.0},
	}
	gl.UseProgram(r.program.Id)
	gl.Uniform1i(r.attribLocationTex, 0)
	gl.UniformMatrix4fv(r.attribLocationProjMtx, 1, false, &orthoProjection[0][0])
	gl.BindSampler(0, 0) // Rely on combined texture/sampler state.

	// Recreate the VAO every time
	// (This is to easily allow multiple GL contexts. VAO are not shared among GL contexts, and
	// we don't track creation/deletion of windows so we don't have an obvious key to use to cache them.)
	var vaoHandle uint32
	gl.GenVertexArrays(1, &vaoHandle)
	gl.BindVertexArray(vaoHandle)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vboHandle)
	gl.EnableVertexAttribArray(uint32(r.attribLocationPosition))
	gl.EnableVertexAttribArray(uint32(r.attribLocationUV))
	gl.EnableVertexAttribArray(uint32(r.attribLocationColor))
	vertexSize, vertexOffsetPos, vertexOffsetUv, vertexOffsetCol := imgui.VertexBufferLayout()
	gl.VertexAttribPointerWithOffset(uint32(r.attribLocationPosition), 2, gl.FLOAT, false, int32(vertexSize), uintptr(vertexOffsetPos))
	gl.VertexAttribPointerWithOffset(uint32(r.attribLocationUV), 2, gl.FLOAT, false, int32(vertexSize), uintptr(vertexOffsetUv))
	gl.VertexAttribPointerWithOffset(uint32(r.attribLocationColor), 4, gl.UNSIGNED_BYTE, true, int32(vertexSize), uintptr(vertexOffsetCol))
	indexSize := imgui.IndexBufferLayout()
	drawType := gl.UNSIGNED_SHORT
	const bytesPerUint32 = 4
	if indexSize == bytesPerUint32 {
		drawType = gl.UNSIGNED_INT
	}

	// Draw
	for _, list := range drawData.CommandLists() {
		vertexBuffer, vertexBufferSize := list.VertexBuffer()
		gl.BindBuffer(gl.ARRAY_BUFFER, r.vboHandle)
		gl.BufferData(gl.ARRAY_BUFFER, vertexBufferSize, vertexBuffer, gl.STREAM_DRAW)

		indexBuffer, indexBufferSize := list.IndexBuffer()
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.elementsHandle)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexBufferSize, indexBuffer, gl.STREAM_DRAW)

		for _, cmd := range list.Commands() {
			if cmd.HasUserCallback() {
				cmd.CallUserCallback(list)
			} else {
				gl.BindTexture(gl.TEXTURE_2D, uint32(cmd.TextureID()))
				clipRect := cmd.ClipRect()
				gl.Scissor(int32(clipRect.X), int32(fbHeight)-int32(clipRect.W), int32(clipRect.Z-clipRect.X), int32(clipRect.W-clipRect.Y))
				gl.DrawElementsBaseVertexWithOffset(gl.TRIANGLES, int32(cmd.ElementCount()), uint32(drawType),
					uintptr(cmd.IndexOffset()*indexSize), int32(cmd.VertexOffset()))
			}
		}
	}
	gl.DeleteVertexArrays(1, &vaoHandle)

	// Restore modified GL state
	gl.UseProgram(uint32(lastProgram))
	gl.BindTexture(gl.TEXTURE_2D, uint32(lastTexture))
	gl.BindSampler(0, uint32(lastSampler))
	gl.ActiveTexture(uint32(lastActiveTexture))
	gl.BindVertexArray(uint32(lastVertexArray))
	gl.BindBuffer(gl.ARRAY_BUFFER, uint32(lastArrayBuffer))
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, uint32(lastElementArrayBuffer))
	gl.BlendEquationSeparate(uint32(lastBlendEquationRgb), uint32(lastBlendEquationAlpha))
	gl.BlendFuncSeparate(uint32(lastBlendSrcRgb), uint32(lastBlendDstRgb), uint32(lastBlendSrcAlpha), uint32(lastBlendDstAlpha))
	if lastEnableBlend {
		gl.Enable(gl.BLEND)
	} else {
		gl.Disable(gl.BLEND)
	}
	if lastEnableCullFace {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}
	if lastEnableDepthTest {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
	if lastEnableScissorTest {
		gl.Enable(gl.SCISSOR_TEST)
	} else {
		gl.Disable(gl.SCISSOR_TEST)
	}
	gl.PolygonMode(gl.FRONT_AND_BACK, uint32(lastPolygonMode[0]))
	gl.Viewport(lastViewport[0], lastViewport[1], lastViewport[2], lastViewport[3])
	gl.Scissor(lastScissorBox[0], lastScissorBox[1], lastScissorBox[2], lastScissorBox[3])
}

func (r *OpenGL4) createDeviceObjects() {
	// Backup GL state
	var lastTexture int32
	var lastArrayBuffer int32
	var lastVertexArray int32
	gl.GetIntegerv(gl.TEXTURE_BINDING_2D, &lastTexture)
	gl.GetIntegerv(gl.ARRAY_BUFFER_BINDING, &lastArrayBuffer)
	gl.GetIntegerv(gl.VERTEX_ARRAY_BINDING, &lastVertexArray)

	r.program = *r3d.MustLoadProgram(textVertexShader, textFragmentShader)

	r.attribLocationTex = gl.GetUniformLocation(r.program.Id, gl.Str("Texture"+"\x00"))
	r.attribLocationProjMtx = gl.GetUniformLocation(r.program.Id, gl.Str("ProjMtx"+"\x00"))
	r.attribLocationPosition = gl.GetAttribLocation(r.program.Id, gl.Str("Position"+"\x00"))
	r.attribLocationUV = gl.GetAttribLocation(r.program.Id, gl.Str("UV"+"\x00"))
	r.attribLocationColor = gl.GetAttribLocation(r.program.Id, gl.Str("Color"+"\x00"))

	gl.GenBuffers(1, &r.vboHandle)
	gl.GenBuffers(1, &r.elementsHandle)

	r.createFontsTexture()

	// Restore modified GL state
	gl.BindTexture(gl.TEXTURE_2D, uint32(lastTexture))
	gl.BindBuffer(gl.ARRAY_BUFFER, uint32(lastArrayBuffer))
	gl.BindVertexArray(uint32(lastVertexArray))
}

func (r *OpenGL4) createFontsTexture() {
	// Build texture atlas
	io := imgui.CurrentIO()
	image := io.Fonts().TextureDataRGBA32()

	// Upload texture to graphics system
	var lastTexture int32
	gl.GetIntegerv(gl.TEXTURE_BINDING_2D, &lastTexture)
	gl.GenTextures(1, &r.fontTexture)
	gl.BindTexture(gl.TEXTURE_2D, r.fontTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(image.Width), int32(image.Height),
		0, gl.RGBA, gl.UNSIGNED_BYTE, image.Pixels)

	// Store our identifier
	io.Fonts().SetTextureID(imgui.TextureID(r.fontTexture))

	// Restore state
	gl.BindTexture(gl.TEXTURE_2D, uint32(lastTexture))
}

func (r *OpenGL4) invalidateDeviceObjects() {
	if r.vboHandle != 0 {
		gl.DeleteBuffers(1, &r.vboHandle)
	}
	r.vboHandle = 0
	if r.elementsHandle != 0 {
		gl.DeleteBuffers(1, &r.elementsHandle)
	}
	r.elementsHandle = 0

	r.program.Delete()

	if r.fontTexture != 0 {
		gl.DeleteTextures(1, &r.fontTexture)
		imgui.CurrentIO().Fonts().SetTextureID(0)
		r.fontTexture = 0
	}
}

var glConstToString = map[uint32]string{
	gl.DEBUG_SOURCE_API:           "API",
	gl.DEBUG_SOURCE_WINDOW_SYSTEM: "WINDOW SYSTEM",
	gl.DEBUG_SOURCE_THIRD_PARTY:   "THIRD PARTY",
	gl.DEBUG_SOURCE_APPLICATION:   "APPLICATION",
	gl.DEBUG_SOURCE_OTHER:         "OTHER",

	gl.DEBUG_TYPE_ERROR:               "ERROR",
	gl.DEBUG_TYPE_DEPRECATED_BEHAVIOR: "DEPRECATED BEHAVIOR",
	gl.DEBUG_TYPE_UNDEFINED_BEHAVIOR:  "UNDEFINED BEHAVIOR",
	gl.DEBUG_TYPE_PORTABILITY:         "PORTABILITY",
	gl.DEBUG_TYPE_PERFORMANCE:         "PERFORMANCE",
	gl.DEBUG_TYPE_OTHER:               "OTHER",
	gl.DEBUG_TYPE_MARKER:              "MARKER",

	gl.DEBUG_SEVERITY_HIGH:         "HIGH",
	gl.DEBUG_SEVERITY_MEDIUM:       "MEDIUM",
	gl.DEBUG_SEVERITY_LOW:          "LOW",
	gl.DEBUG_SEVERITY_NOTIFICATION: "NOTIFICATION",
}

func openglLogCallback(source uint32, gltype uint32, id uint32,
	severity uint32, length int32, message string, userParam unsafe.Pointer) {

	log.Printf("[gl] id:%v severity:%v src:%v type:%v %q",
		id, glConstToString[severity], glConstToString[source], glConstToString[gltype], message)
	if gltype == gl.DEBUG_TYPE_ERROR {
		panic(message)
	}
}
