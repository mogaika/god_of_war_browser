package txr

import (
	"fmt"
	"log"

	"github.com/mogaika/god_of_war_browser/fbx"
	"github.com/mogaika/god_of_war_browser/fbx/cache"
	"github.com/mogaika/god_of_war_browser/pack/wad"
)

type FbxExporter struct {
	TextureId uint64
	VideoId   uint64
}

func (t *Texture) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbx.FBX, cache *cache.Cache) *FbxExporter {
	fe := &FbxExporter{
		TextureId: f.GenerateId(),
		VideoId:   f.GenerateId(),
	}
	defer cache.Add(wrsrc.Tag.Id, fe)

	ajaxI, err := t.Marshal(wrsrc)
	if err != nil {
		log.Panicf("Unable to marshal image %v: %v", wrsrc.Tag.Name, err)
	}
	ajax := ajaxI.(*Ajax)
	png := ajax.Images[0].Image

	name := fmt.Sprintf("%s_%d.png", wrsrc.Tag.Name, wrsrc.Tag.Id)
	// fileName := wrsrc.Wad.Name() + ".fbm\\" + name
	fileName := name

	f.AddExportFile(fileName, png)

	video := &fbx.Video{
		Id:               fe.VideoId,
		Name:             "Video::",
		Element:          "Clip",
		Type:             "Clip",
		UseMipMap:        0,
		Filename:         fileName,
		RelativeFilename: fileName,
		// 		Content:          base64.StdEncoding.EncodeToString(png),
		Properties70: fbx.Properties70{
			P: []*fbx.Propertie70{
				&fbx.Propertie70{Name: "Path", Type: "KString", Purpose: "XRefUrl", Value: fileName},
			},
		},
	}
	texture := &fbx.Texture{
		Id:                   fe.TextureId,
		Name:                 "Texture::",
		Type:                 "TextureVideoClip",
		Version:              202,
		TextureName:          "Texture::",
		FileName:             video.Filename,
		RelativeFilename:     video.RelativeFilename,
		Texture_Alpha_Source: "None",
		ModelUVTranslation:   []int{0, 0},
		ModelUVScaling:       []int{1, 1},
		Cropping:             []int{0, 0, 0, 0},
		Properties70: fbx.Properties70{
			P: []*fbx.Propertie70{
				&fbx.Propertie70{Name: "UseMaterial", Type: "bool", Value: 1},
				&fbx.Propertie70{Name: "CurrentTextureBlendMode", Type: "enum", Value: 0},
			},
		},
	}

	f.Objects.Texture = append(f.Objects.Texture, texture)
	f.Objects.Video = append(f.Objects.Video, video)
	f.Connections.C = append(f.Connections.C, fbx.Connection{
		Type: "OO", Child: video.Id, Parent: texture.Id,
	})

	return fe
}
