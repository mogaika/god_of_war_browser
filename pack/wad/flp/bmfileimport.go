package flp

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"image"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/mogaika/bmfont"
	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mesh "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	file_txr "github.com/mogaika/god_of_war_browser/pack/wad/txr"
)

func getMeshForFlp(wrsrc *wad.WadNodeRsrc) (*file_mesh.Mesh, wad.TagId, error) {
	meshName := strings.Replace(wrsrc.Name(), "FLP_", "", 1) + "_0"
	meshTag := wrsrc.Wad.GetTagByName(meshName, wrsrc.Tag.Id, false)
	if meshTag == nil {
		return nil, -1, fmt.Errorf("Cannot find mesh with name '%s'", meshName)
	}

	meshInstance, _, err := wrsrc.Wad.GetInstanceFromTag(meshTag.Id)
	if err != nil {
		return nil, -1, fmt.Errorf("Error when parsing mesh instance: %v", err)
	}

	return meshInstance.(*file_mesh.Mesh), meshTag.Id, nil
}

func getBmfFromArchive(zr *zip.Reader) (*bmfont.Font, error) {
	var fBmf *zip.File
	for _, f := range zr.File {
		if strings.ToLower(filepath.Ext(f.Name)) == ".fnt" {
			fBmf = f
			break
		}
	}
	if fBmf == nil {
		return nil, fmt.Errorf("Cannot find '*.fnt' file in archive")
	}
	f, err := fBmf.Open()
	if err != nil {
		return nil, fmt.Errorf("Cannot open bmf file: %v", err)
	}
	defer f.Close()
	raw, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Cannot read bmf file: %v", err)
	}

	return bmfont.NewFontFromBuf(raw)
}

func (f *FLP) actionImportBmFont(wrsrc *wad.WadNodeRsrc, zr *zip.Reader) error {
	if f.Fonts == nil || len(f.Fonts) == 0 {
		return nil
	}
	if len(f.Fonts) > 1 {
		return fmt.Errorf("Cannot import font in multi-font flp file")
	}
	font := &f.Fonts[0]

	mesh, meshTagId, err := getMeshForFlp(wrsrc)
	if err != nil {
		return fmt.Errorf("Cannot find mesh for flp: %v", err)
	}

	bmf, err := getBmfFromArchive(zr)
	if err != nil {
		return fmt.Errorf("Cannot get bmf file from archive: %v", err)
	}

	if err := f.ImportBmFont(font, mesh, bmf); err != nil {
		return fmt.Errorf("Error when importing bmf structs: %v", err)
	}

	if err := wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
		meshTagId:    mesh.MarshalBuffer().Bytes(),
		wrsrc.Tag.Id: f.marshalBufferWithHeader().Bytes(),
	}); err != nil {
		return fmt.Errorf("Error when updating mesh and flp tags: %v", err)
	}

	for _, texture_name := range bmf.Pages {
		ex := false
		var fTxr *zip.File
		for _, f := range zr.File {
			if f.Name == texture_name {
				ex = true
				fTxr = f
				break
			}
		}
		if !ex {
			return fmt.Errorf("Cannot find texture '%s' in archive", texture_name)
		}
		f, err := fTxr.Open()
		if err != nil {
			return fmt.Errorf("Cannot open texture '%s' from archive: %v", texture_name, err)
		}
		defer f.Close()

		if txrTag := wrsrc.Wad.GetTagByName("TXR_"+texture_name, meshTagId, false); txrTag != nil {
			txrInst, _, err := wrsrc.Wad.GetInstanceFromTag(txrTag.Id)
			if err != nil {
				return fmt.Errorf("Error when reuploading txr '%s': %v", texture_name, err)
			}
			txr := txrInst.(*file_txr.Texture)
			txr.ChangeTexture(wrsrc.Wad.GetNodeResourceByTagId(txrTag.Id), f)
		} else {
			img, _, err := image.Decode(f)
			if err != nil {
				return fmt.Errorf("Cannot decode image '%s': %v", texture_name, err)
			}
			file_txr.CreateNewTextureInWad(wrsrc.Wad, texture_name, meshTagId-4, img)
		}
	}

	return nil
}

func bmFontNewMeshPartReferenceFromChar(bmf *bmfont.Font, char *bmfont.Char, mesh *file_mesh.Mesh, prevRef *MeshPartReference) *MeshPartReference {
	meshPartIndex := int16(len(mesh.Parts))
	if prevRef != nil {
		meshPartIndex = prevRef.MeshPartIndex
	} else {
		mesh.Parts = append(mesh.Parts, file_mesh.Part{})
	}

	part := &mesh.Parts[meshPartIndex]
	part.Unk00 = 1
	part.JointId = uint16(meshPartIndex + 2)
	part.Groups = []file_mesh.Group{{
		Unk00: 0xd1000000,
		Unk08: 0,
	}}

	group := &part.Groups[0]
	group.Objects = []file_mesh.Object{{
		Type:                0xe,
		Unk02:               0,
		PacketsPerFilter:    2,
		MaterialId:          420, // I think noone cares about this one
		JointMapper:         []uint32{uint32(meshPartIndex + 2)},
		Unk0c:               1,
		Unk10:               0,
		Unk14:               0xffffff35,
		TextureLayersCount:  1,
		Unk19:               1,
		NextFreeVUBufferId:  1,
		Unk1c:               1,
		SourceVerticesCount: 4,
		RawDmaAndJointsData: generateMeshDmaPacketData(
			[2]float32{
				float32(char.Xoffset),
				float32(char.Xoffset) + float32(char.Width),
			}, [2]float32{
				-float32(bmf.Common.Base) + float32(char.Yoffset),
				-float32(bmf.Common.Base) + float32(char.Yoffset) + float32(char.Height),
			},
			[2]float32{
				float32(char.X) / float32(bmf.Common.ScaleW),
				float32(char.X+char.Width) / float32(bmf.Common.ScaleW),
			}, [2]float32{
				float32(char.Y) / float32(bmf.Common.ScaleH),
				float32(char.Y+char.Height) / float32(bmf.Common.ScaleH),
			}),
	}}

	return &MeshPartReference{
		MeshPartIndex: meshPartIndex,
		Materials: []MeshPartMaterialSlot{{
			Color:       0xffffffff,
			TextureName: "TXR_" + bmf.Pages[char.Page],
		}},
	}
}

func (f *FLP) ImportBmFont(font *Font, mesh *file_mesh.Mesh, bmf *bmfont.Font) error {
	fontAliases, err := config.GetFontAliases()
	if err != nil {
		return fmt.Errorf("Cannot load font aliases file: %v", err)
	}

	for iBmChar := range bmf.Chars {
		bmchar := &bmf.Chars[iBmChar]
		unicodeChar := rune(bmchar.Id)

		var ansiiCharId uint8 = 0
		if charAlias, charAliasExists := fontAliases[unicodeChar]; charAliasExists {
			ansiiCharId = charAlias
		} else {
			if unicodeChar < 0x100 {
				ansiiCharId = uint8(unicodeChar)
			} else {
				return fmt.Errorf("Cannot map char '%v' (%v). Please update font_aliases.cfg file", string(unicodeChar), unicodeChar)
			}
		}

		glyphId := font.CharNumberToSymbolIdMap[ansiiCharId]
		charWidth := int16(float32(bmchar.Xadvance) * file_mesh.GSFixedPoint8)
		if glyphId == -1 {
			// create new glyph
			newGlyphId := int16(font.CharsCount)
			newMeshRef := bmFontNewMeshPartReferenceFromChar(bmf, bmchar, mesh, nil)
			font.Flag4Datas2 = append(font.Flag4Datas2, *newMeshRef)
			font.SymbolWidths = append(font.SymbolWidths, charWidth)
			font.CharsCount++
			font.CharNumberToSymbolIdMap[ansiiCharId] = newGlyphId
		} else {
			// update exists glyph
			font.Flag4Datas2[glyphId] = *bmFontNewMeshPartReferenceFromChar(bmf, bmchar, mesh, &font.Flag4Datas2[glyphId])
			font.SymbolWidths[glyphId] = charWidth
		}
	}

	return nil
}

func generateMeshDmaPacketData(x [2]float32, y [2]float32, texture_u [2]float32, texture_v [2]float32) []byte {
	dmaPacketData := make([]byte, len(preparedMeshDmaPacketData))
	copy(dmaPacketData, preparedMeshDmaPacketData)

	setXYLine := func(idx int, x float32, y float32) {
		binary.LittleEndian.PutUint16(dmaPacketData[0x4c+idx*8:], uint16(int16(x*file_mesh.GSFixedPoint8)))
		binary.LittleEndian.PutUint16(dmaPacketData[0x4c+idx*8+2:], uint16(int16(y*file_mesh.GSFixedPoint8)))
	}
	setXYLine(0, x[1], y[1])
	setXYLine(1, x[0], y[1])
	setXYLine(2, x[1], y[0])
	setXYLine(3, x[0], y[0])

	setUvLine := func(idx int, u float32, v float32) {
		binary.LittleEndian.PutUint16(dmaPacketData[0x38+idx*4:], uint16(int16(u*file_mesh.GSFixedPoint24)))
		binary.LittleEndian.PutUint16(dmaPacketData[0x38+idx*4+2:], uint16(int16(v*file_mesh.GSFixedPoint24)))
	}
	setUvLine(0, texture_u[1], texture_v[1])
	setUvLine(1, texture_u[0], texture_v[1])
	setUvLine(2, texture_u[1], texture_v[0])
	setUvLine(3, texture_u[0], texture_v[0])

	return dmaPacketData
}

var preparedMeshDmaPacketData = []byte{0x7, 0x0, 0x0, 0x30, 0x50, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14, 0x0, 0x0, 0x0, 0x0, 0x1a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x1, 0x0, 0x1, 0x1, 0x80, 0x4, 0x65, 0xe0, 0xa, 0x5f, 0xb, 0xe0, 0x9, 0x5f, 0xb, 0xe0, 0xa, 0x1f, 0x9, 0xe0, 0x9, 0x1f, 0x9, 0x2, 0x80, 0x4, 0x6d, 0x45, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80, 0x35, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80, 0x45, 0x2, 0x82, 0xfd, 0x0, 0x0, 0x0, 0x0, 0x35, 0x0, 0x82, 0xfd, 0x0, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x1, 0x0, 0x80, 0x1, 0x6c, 0x4, 0x80, 0x0, 0x0, 0x2, 0x40, 0x2e, 0x20, 0x52, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80, 0xf4, 0x80, 0x1, 0x6c, 0x0, 0x80, 0x9e, 0x41, 0x0, 0x80, 0x9f, 0xc1, 0x0, 0x0, 0x0, 0x0, 0x1d, 0x86, 0xd0, 0x41, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

/*
([]uint8) (len=160) {
 00000000  07 00 00 30 50 00 00 00  00 00 00 00 00 00 00 00  |...0P...........|
 00000010  00 00 00 60 00 00 00 00  00 00 00 14 00 00 00 00  |...`............|
 00000020  1a 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
 00000030  02 01 00 01 01 80 04 65  e0 0a 5f 0b e0 09 5f 0b  |.......e.._..._.|
 00000040  e0 0a 1f 09 e0 09 1f 09  02 80 04 6d 45 02 00 00  |...........mE...|
 00000050  00 00 00 80 35 00 00 00  00 00 00 80 45 02 82 fd  |....5.......E...|
 00000060  00 00 00 00 35 00 82 fd  00 00 00 00 01 01 00 01  |....5...........|
 00000070  00 80 01 6c 04 80 00 00  02 40 2e 20 52 00 00 00  |...l.....@. R...|
 00000080  00 00 00 80 f4 80 01 6c  00 80 9e 41 00 80 9f c1  |.......l...A....|
 00000090  00 00 00 00 1d 86 d0 41  00 00 00 00 00 00 00 00  |.......A........|
}
[ uv2] cmd: 0x65 elements: 0x04 components: 2 width: 16 target: 0x001 sign: true tops: true size: 000010
[xyzw] cmd: 0x6d elements: 0x04 components: 4 width: 16 target: 0x002 sign: true tops: true size: 000020
[bndr] cmd: 0x6c elements: 0x01 components: 4 width: 32 target: 0x0f4 sign: true tops: true size: 000010
*/
