package uibackend

import (
	"fmt"
	"log"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/inkyblackness/imgui-go/v4"
)

// GLFW implements a platform based on github.com/go-gl/glfw (v3.2).
type GLFW struct {
	imguiIO imgui.IO

	window *glfw.Window

	time             float64
	mouseJustPressed [3]bool
}

// NewGLFW attempts to initialize a GLFW context.
func NewGLFW(io imgui.IO) (*GLFW, error) {
	err := glfw.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize glfw: %w", err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)

	window, err := glfw.CreateWindow(1600, 900, "gow editor", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	p := &GLFW{
		imguiIO: io,
		window:  window,
	}
	p.setKeyMapping()
	p.installCallbacks()

	io.SetClipboard(&glfwClipboard{p: p})

	return p, nil
}

type glfwClipboard struct{ p *GLFW }

func (c *glfwClipboard) Text() (string, error) { return c.p.ClipboardText() }
func (c *glfwClipboard) SetText(value string)  { c.p.SetClipboardText(value) }

func (p *GLFW) Destroy() {
	p.window.Destroy()
	glfw.Terminate()
}

func (p *GLFW) ShouldStop() bool { return p.window.ShouldClose() }

func (p *GLFW) ProcessEvents() { glfw.PollEvents() }

func (p *GLFW) DisplaySize() [2]float32 {
	w, h := p.window.GetSize()
	return [2]float32{float32(w), float32(h)}
}

// FramebufferSize returns the dimension of the framebuffer.
func (platform *GLFW) FramebufferSize() [2]float32 {
	w, h := platform.window.GetFramebufferSize()
	return [2]float32{float32(w), float32(h)}
}

// NewFrame marks the begin of a render pass. It forwards all current state to imgui IO.
func (platform *GLFW) NewFrame() {
	// Setup display size (every frame to accommodate for window resizing)
	displaySize := platform.DisplaySize()
	platform.imguiIO.SetDisplaySize(imgui.Vec2{X: displaySize[0], Y: displaySize[1]})

	// Setup time step
	currentTime := glfw.GetTime()
	if platform.time > 0 {
		platform.imguiIO.SetDeltaTime(float32(currentTime - platform.time))
	}
	platform.time = currentTime

	// Setup inputs
	if platform.window.GetAttrib(glfw.Focused) != 0 {
		x, y := platform.window.GetCursorPos()
		platform.imguiIO.SetMousePosition(imgui.Vec2{X: float32(x), Y: float32(y)})
	} else {
		platform.imguiIO.SetMousePosition(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	for i := 0; i < len(platform.mouseJustPressed); i++ {
		down := platform.mouseJustPressed[i] || (platform.window.GetMouseButton(glfwButtonIDByIndex[i]) == glfw.Press)
		platform.imguiIO.SetMouseButtonDown(i, down)
		platform.mouseJustPressed[i] = false
	}
}

// PostRender performs a buffer swap.
func (platform *GLFW) PostRender() {
	platform.window.SwapBuffers()
}

func (platform *GLFW) setKeyMapping() {
	// Keyboard mapping. ImGui will use those indices to peek into the io.KeysDown[] array.
	platform.imguiIO.KeyMap(imgui.KeyTab, int(glfw.KeyTab))
	platform.imguiIO.KeyMap(imgui.KeyLeftArrow, int(glfw.KeyLeft))
	platform.imguiIO.KeyMap(imgui.KeyRightArrow, int(glfw.KeyRight))
	platform.imguiIO.KeyMap(imgui.KeyUpArrow, int(glfw.KeyUp))
	platform.imguiIO.KeyMap(imgui.KeyDownArrow, int(glfw.KeyDown))
	platform.imguiIO.KeyMap(imgui.KeyPageUp, int(glfw.KeyPageUp))
	platform.imguiIO.KeyMap(imgui.KeyPageDown, int(glfw.KeyPageDown))
	platform.imguiIO.KeyMap(imgui.KeyHome, int(glfw.KeyHome))
	platform.imguiIO.KeyMap(imgui.KeyEnd, int(glfw.KeyEnd))
	platform.imguiIO.KeyMap(imgui.KeyInsert, int(glfw.KeyInsert))
	platform.imguiIO.KeyMap(imgui.KeyDelete, int(glfw.KeyDelete))
	platform.imguiIO.KeyMap(imgui.KeyBackspace, int(glfw.KeyBackspace))
	platform.imguiIO.KeyMap(imgui.KeySpace, int(glfw.KeySpace))
	platform.imguiIO.KeyMap(imgui.KeyEnter, int(glfw.KeyEnter))
	platform.imguiIO.KeyMap(imgui.KeyEscape, int(glfw.KeyEscape))
	platform.imguiIO.KeyMap(imgui.KeyA, int(glfw.KeyA))
	platform.imguiIO.KeyMap(imgui.KeyC, int(glfw.KeyC))
	platform.imguiIO.KeyMap(imgui.KeyV, int(glfw.KeyV))
	platform.imguiIO.KeyMap(imgui.KeyX, int(glfw.KeyX))
	platform.imguiIO.KeyMap(imgui.KeyY, int(glfw.KeyY))
	platform.imguiIO.KeyMap(imgui.KeyZ, int(glfw.KeyZ))
}

func (platform *GLFW) installCallbacks() {
	platform.window.SetMouseButtonCallback(platform.mouseButtonChange)
	platform.window.SetScrollCallback(platform.mouseScrollChange)
	platform.window.SetKeyCallback(platform.keyChange)
	platform.window.SetCharCallback(platform.charChange)
}

const (
	mouseButtonPrimary   = 0
	mouseButtonSecondary = 1
	mouseButtonTertiary  = 2
	mouseButtonCount     = 3
)

var glfwButtonIndexByID = map[glfw.MouseButton]int{
	glfw.MouseButton1: mouseButtonPrimary,
	glfw.MouseButton2: mouseButtonSecondary,
	glfw.MouseButton3: mouseButtonTertiary,
}

var glfwButtonIDByIndex = map[int]glfw.MouseButton{
	mouseButtonPrimary:   glfw.MouseButton1,
	mouseButtonSecondary: glfw.MouseButton2,
	mouseButtonTertiary:  glfw.MouseButton3,
}

func (platform *GLFW) mouseButtonChange(window *glfw.Window, rawButton glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	buttonIndex, known := glfwButtonIndexByID[rawButton]

	if known && (action == glfw.Press) {
		platform.mouseJustPressed[buttonIndex] = true
	}
}

func (platform *GLFW) mouseScrollChange(window *glfw.Window, x, y float64) {
	platform.imguiIO.AddMouseWheelDelta(float32(x), float32(y))
}

func (platform *GLFW) keyChange(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		platform.imguiIO.KeyPress(int(key))
	}
	if action == glfw.Release {
		platform.imguiIO.KeyRelease(int(key))
	}

	// Modifiers are not reliable across systems
	platform.imguiIO.KeyCtrl(int(glfw.KeyLeftControl), int(glfw.KeyRightControl))
	platform.imguiIO.KeyShift(int(glfw.KeyLeftShift), int(glfw.KeyRightShift))
	platform.imguiIO.KeyAlt(int(glfw.KeyLeftAlt), int(glfw.KeyRightAlt))
	platform.imguiIO.KeySuper(int(glfw.KeyLeftSuper), int(glfw.KeyRightSuper))
}

func (platform *GLFW) charChange(window *glfw.Window, char rune) {
	platform.imguiIO.AddInputCharacters(string(char))
}

// ClipboardText returns the current clipboard text, if available.
func (platform *GLFW) ClipboardText() (string, error) {
	log.Printf("Get clipboard")
	return platform.window.GetClipboardString(), nil
}

// SetClipboardText sets the text as the current clipboard text.
func (platform *GLFW) SetClipboardText(text string) {
	log.Printf("Set clipboard %q", text)
	platform.window.SetClipboardString(text)
}
