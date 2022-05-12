package flp

import (
	"fmt"
	"log"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mesh "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
)

type ExportedFont struct {
	Meta   Font
	Meshes []file_mesh.Part
}

func (f *FLP) actionExportFont(wrsrc *wad.WadNodeRsrc) (*ExportedFont, error) {
	mesh, _, err := getMeshForFlp(wrsrc)
	if err != nil {
		return nil, fmt.Errorf("Cannot find mesh for flp: %v", err)
	}

	if len(f.Fonts) != 1 {
		return nil, fmt.Errorf("Multiple fonts (fonts count != 1) not supported")
	}

	ef := &ExportedFont{
		Meta:   f.Fonts[0],
		Meshes: make([]file_mesh.Part, 0),
	}

	materials := make(map[string]struct{})
	for i := range ef.Meta.MeshesRefs {
		mef := &ef.Meta.MeshesRefs[i]
		switch len(mef.Materials) {
		case 0:
		case 1:
			materials[mef.Materials[0].TextureName] = struct{}{}
		default:
			return nil, fmt.Errorf("Multiple materials (>1) for glyph")
		}
		if mef.MeshPartIndex != -1 {
			ef.Meshes = append(ef.Meshes, mesh.Parts[mef.MeshPartIndex])
			mef.MeshPartIndex = int16(len(ef.Meshes) - 1)
		}
	}

	if len(materials) != 1 {
		return nil, fmt.Errorf("Multiple materials (>1) for font")
	}

	return ef, nil
}

func (f *FLP) actionReplaceFont(wrsrc *wad.WadNodeRsrc, ef *ExportedFont) error {
	if f.Fonts == nil || len(f.Fonts) == 0 {
		return fmt.Errorf("No font to replace")
	}
	if len(f.Fonts) > 1 {
		return fmt.Errorf("Cannot replace font in multi-font flp file")
	}
	font := &f.Fonts[0]

	mesh, meshTagId, err := getMeshForFlp(wrsrc)
	if err != nil {
		return fmt.Errorf("Cannot find mesh for flp: %v", err)
	}

	log.Printf("Mesh parts count: %v", len(mesh.Parts))

	textureName := ""
	for _, ref := range font.MeshesRefs {
		if len(ref.Materials) != 0 && ref.Materials[0].TextureName != "" {
			textureName = ref.Materials[0].TextureName
			break
		}
	}
	if textureName == "" {
		return fmt.Errorf("Font required to use texture. No texture found.")
	}

	for char, iGlyph := range ef.Meta.CharNumberToSymbolIdMap {
		if iGlyph == -1 {
			continue
		}

		efMref := ef.Meta.MeshesRefs[iGlyph]

		if ogGlyph := font.CharNumberToSymbolIdMap[char]; ogGlyph == -1 {
			log.Printf("Inserting %d", char)
			// insert new
			newGlyphId := len(font.MeshesRefs)
			font.CharNumberToSymbolIdMap[char] = int16(newGlyphId)
			font.MeshesRefs = append(font.MeshesRefs, efMref)
			mref := &font.MeshesRefs[newGlyphId]

			for i := range mref.Materials {
				mref.Materials[i].TextureName = textureName
			}

			if mref.MeshPartIndex != -1 {
				newMeshPartIndex := len(mesh.Parts)
				mesh.Parts = append(mesh.Parts, ef.Meshes[efMref.MeshPartIndex])
				mref.MeshPartIndex = int16(newMeshPartIndex)

				// for part index 125 material id 129 joint mapper [127] part joint 127

				mPart := &mesh.Parts[mref.MeshPartIndex]
				mPart.JointId = uint16(newMeshPartIndex + 2)
				mObject := &mPart.Groups[0].Objects[0]
				mObject.JointMappers[0][0] = uint32(newMeshPartIndex + 2)
				mObject.MaterialId = uint16(newMeshPartIndex + 4)
			}

			font.SymbolWidths = append(font.SymbolWidths, ef.Meta.SymbolWidths[iGlyph])
		} else {
			log.Printf("Replacing %d", char)
			mref := &font.MeshesRefs[ogGlyph]
			if mref.MeshPartIndex != -1 {
				log.Printf("Replacing mesh %d with %d", mref.MeshPartIndex, efMref.MeshPartIndex)
				mPart := &mesh.Parts[mref.MeshPartIndex]
				mObject := &mPart.Groups[0].Objects[0]
				mObject.RawDmaAndJointsData = ef.Meshes[efMref.MeshPartIndex].Groups[0].Objects[0].RawDmaAndJointsData
			}

			font.SymbolWidths[ogGlyph] = ef.Meta.SymbolWidths[iGlyph]
		}
	}

	font.CharsCount = uint32(len(font.SymbolWidths))
	font.Size = ef.Meta.Size

	if err := wrsrc.Wad.UpdateTagsData(map[wad.TagId][]byte{
		meshTagId:    mesh.MarshalBuffer().Bytes(),
		wrsrc.Tag.Id: f.marshalBufferWithHeader().Bytes(),
	}); err != nil {
		return fmt.Errorf("Error when updating mesh and flp tags: %v", err)
	}

	return nil
}
