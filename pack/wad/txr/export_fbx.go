package txr

import (
	"fmt"
	"log"

	"github.com/mogaika/fbx/builders/bfbx73"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils/fbxbuilder"
)

type FbxExporter struct {
	TextureId int64
	VideoId   int64
}

func (t *Texture) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbxbuilder.FBXBuilder) *FbxExporter {
	fe := &FbxExporter{
		TextureId: f.GenerateId(),
		VideoId:   f.GenerateId(),
	}
	defer f.AddCache(wrsrc.Tag.Id, fe)

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

	video := bfbx73.Video(fe.VideoId, "\x00\x01Video", "Clip").AddNodes(
		bfbx73.Type("Clip"),
		bfbx73.Properties70().AddNodes(
			bfbx73.P("Path", "KString", "XRefUrl", "", fileName),
		),
		bfbx73.UseMipMap(0),
		bfbx73.Filename(fileName),
		bfbx73.RelativeFilename(fileName),
		bfbx73.Content(png),
	)

	texture := bfbx73.Texture(fe.TextureId, "\x00\x01Texture", "TextureVideoClip").AddNodes(
		bfbx73.Type("TextureVideoClip"),
		bfbx73.Version(202),
		bfbx73.Properties70().AddNodes(
			bfbx73.P("CurrentTextureBlendMode", "enum", "", "", int32(0)),
			bfbx73.P("UseMaterial", "bool", "", "", int32(1)),
		),
		bfbx73.Media(""),
		bfbx73.Filename(fileName),
		bfbx73.RelativeFilename(fileName),
		bfbx73.ModelUVTranslation(0, 0),
		bfbx73.ModelUVScaling(1, 1),
		bfbx73.Texture_Alpha_Source("None"), // TODO: alpha source not none?
		bfbx73.Cropping(0, 0, 0, 0),         // TODO: check that we do not need cropping on non-rectangle textures?
	)

	f.AddObjects(texture, video)
	f.AddConnections(bfbx73.C("OO", fe.VideoId, fe.TextureId))

	return fe
}
