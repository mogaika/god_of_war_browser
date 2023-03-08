package r3d

import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/rendercontext"
)

type View3D struct {
	Nodes []Node

	glInited          bool
	glWidth, glHeight int32
	glFramebuffer     uint32
	glTexture         uint32
	glRBO             uint32
}

func (v *View3D) useGL(w, h int32) {
	if w != v.glWidth || h != v.glHeight {
		v.ClearTempRenderData()
	}
	v.glWidth = w
	v.glHeight = h

	rendercontext.Use(v)
	if v.glInited {
		return
	}
	v.glInited = true

	gl.GenFramebuffers(1, &v.glFramebuffer)
	gl.GenTextures(1, &v.glTexture)
	gl.GenRenderbuffers(1, &v.glRBO)

	gl.BindTexture(gl.TEXTURE_2D, v.glTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, v.glWidth, v.glHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.BindRenderbuffer(gl.RENDERBUFFER, v.glRBO)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, v.glWidth, v.glHeight)
	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)

	gl.BindFramebuffer(gl.FRAMEBUFFER, v.glFramebuffer)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, v.glTexture, 0)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, v.glRBO)
	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	if status != gl.FRAMEBUFFER_COMPLETE {
		panic(status)
	}
}

func (v *View3D) ClearTempRenderData() {
	if !v.glInited {
		return
	}
	v.glInited = false

	gl.DeleteTextures(1, &v.glTexture)
	gl.DeleteFramebuffers(1, &v.glFramebuffer)
	gl.DeleteRenderbuffers(1, &v.glRBO)
}

func (v *View3D) BeforeRender3D(w, h int32) {
	v.useGL(w, h)

	gl.BindFramebuffer(gl.FRAMEBUFFER, v.glFramebuffer)

	gl.Viewport(0, 0, v.glWidth, v.glHeight)
	gl.ClearColor(0.15, 0.15, 0.2, 1.0)
	gl.ClearDepth(1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (v *View3D) AfterRender3D() imgui.TextureID {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	return imgui.TextureID(v.glTexture)
}
