package gow

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/mogaika/god_of_war_browser/editor/core"
	"github.com/mogaika/god_of_war_browser/editor/rendercontext"
	"github.com/mogaika/god_of_war_browser/pack/wad/txr"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Texture struct {
	RequireNoGroupMock
	OG  *txr.Texture
	GFX core.Ref[*GFX]
	PAL core.Ref[*GFX]
	LOD core.Ref[*Texture]

	glInited   bool
	glTextures []uint32
	glSize     imgui.Vec2
}

func (r *Texture) GetPixels(p *core.Project) (w, h int, pixels [][][4]byte) {
	gfx := r.GFX.Resolve(p)
	pal := r.PAL.Resolve(p)

	if gfx == nil || pal == nil {
		return 0, 0, nil
	}

	pixels = make([][][4]byte, 0, len(gfx.OG.Data)*len(pal.OG.Data))

	for iGfx := range gfx.OG.Data {
		for iPal := range pal.OG.Data {
			gfxData := gfx.OG.AsPaletteIndexes(iGfx)
			palData, err := pal.OG.AsRawPalette(iPal, true)
			if err != nil {
				panic(err)
			}

			w = int(gfx.OG.Width)
			h = int(gfx.OG.RealHeight)

			pcur := make([][4]byte, len(gfxData))
			for iPixel := range pcur {
				pcur[iPixel] = palData[gfxData[iPixel]]
			}

			pixels = append(pixels, pcur)
		}
	}

	return w, h, pixels
}

func (r *Texture) useGL(p *core.Project) bool {
	rendercontext.Use(r)
	if r.glInited {
		return true
	}
	r.glInited = true

	width, height, pixels := r.GetPixels(p)
	if len(pixels) == 0 {
		return false
	}
	r.glSize = imgui.Vec2{X: float32(width), Y: float32(height)}

	r.glTextures = make([]uint32, len(pixels))

	gl.GenTextures(int32(len(r.glTextures)), &r.glTextures[0])
	for iTexture := range r.glTextures {
		gl.BindTexture(gl.TEXTURE_2D, r.glTextures[iTexture])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(width), int32(height),
			0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels[iTexture][0]))
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}
	gl.BindTexture(gl.TEXTURE_2D, 0)
	runtime.KeepAlive(pixels)
	return true
}

func (r *Texture) ClearTempRenderData() {
	if !r.glInited {
		return
	}
	r.glInited = false

	if len(r.glTextures) != 0 {
		gl.DeleteTextures(int32(len(r.glTextures)), &r.glTextures[0])
		r.glTextures = nil
	}
}

func (r *Texture) RenderUI(p *core.Project) {
	core.ReflectView(p, r)
	imgui.Separator()

	imgui.Selectable("gfx")
	core.NewRefView(r.GFX).RenderUI(p)
	imgui.Selectable("pal")
	core.NewRefView(r.PAL).RenderUI(p)
	imgui.Selectable("lod")
	core.NewRefView(r.LOD).RenderUI(p)
	imgui.Selectable(fmt.Sprintf("LODParamK: %v", r.OG.LODParamK))
	imgui.Selectable(fmt.Sprintf("LODMultiplier: %v", r.OG.LODMultiplier))
	imgui.Selectable(fmt.Sprintf("Flags: 0x%.8x", r.OG.Flags))

	if r.useGL(p) {
		for _, tid := range r.glTextures {
			imgui.Image(imgui.TextureID(tid), r.glSize.Times(2))
		}
	}
}

func (r *Texture) RenderTooltip(p *core.Project) {
	const TEXTURE_PREVIEW_SIZE = 128
	mul := TEXTURE_PREVIEW_SIZE / utils.Max(r.glSize.X, r.glSize.Y)
	if r.useGL(p) && len(r.glTextures) != 0 {
		imgui.Image(imgui.TextureID(r.glTextures[0]), r.glSize.Times(mul))
	}
}
