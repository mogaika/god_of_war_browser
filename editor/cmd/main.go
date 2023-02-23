package main

import (
	"log"
	"runtime"

	"github.com/mogaika/god_of_war_browser/editor/app"
	"github.com/mogaika/god_of_war_browser/editor/uibackend"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/inkyblackness/imgui-go/v4"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	context := imgui.CreateContext(nil)
	defer context.Destroy()

	io := imgui.CurrentIO()
	io.Fonts().AddFontFromFileTTF("editor/fonts/Cousine-Regular.ttf", 20)

	back, err := uibackend.NewGLFW(io)
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
	defer back.Destroy()

	ren, err := uibackend.NewOpenGL3(io)
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
	defer ren.Destroy()

	log.Printf("Version: %q", gl.GoStr(gl.GetString(gl.VERSION)))

	clearColor := [3]float32{0.15, 0.15, 0.15}

	for !back.ShouldStop() {
		back.ProcessEvents()

		back.NewFrame()
		imgui.NewFrame()

		app.SetFramebufferSize(back.FramebufferSize())

		app.RenderUI()

		imgui.Render()
		ren.PreRender(clearColor)

		// here we should render game?

		ren.Render(back.DisplaySize(), back.FramebufferSize(), imgui.RenderedDrawData())
		back.PostRender()
	}
}
