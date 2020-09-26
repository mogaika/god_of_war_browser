package gxt

import (
	"encoding/binary"
	"image"
	"io"
	"log"
	"os"

	"github.com/mogaika/god_of_war_browser/psvita/textureformats"

	"github.com/pkg/errors"
)

// Based on github.com/xdanieldzd/GXTConvert

type Header struct {
	Magic         uint32
	Version       uint32
	TexturesCount uint32
	DataOffset    uint32
	DataSize      uint32
	PalettCountP4 uint32
	PalettCountP8 uint32
	Padding       uint32
}

type TextureInfo struct {
	Offset       uint32
	Size         uint32
	PaletteIndex uint32
	Flags        uint32

	Type   uint32
	Format uint32
	Width  uint16
	Height uint16

	MipMapsCount uint16
	Padding      uint16
}

type GXT struct {
	Header       Header
	TextureInfos []TextureInfo
}

func (g *GXT) parseTextureInfo(r io.Reader, ti *TextureInfo) error {
	if err := binary.Read(r, binary.LittleEndian, ti); err != nil {
		return errors.Wrapf(err, "Failed to read")
	}

	if ti.PaletteIndex != ^uint32(0) {
		return errors.Errorf("Unsupported pallete index %d", ti.PaletteIndex)
	}

	log.Printf("%+#v", ti)

	return nil
}

func (ti *TextureInfo) ToImage(r io.ReadSeeker) (image.Image, error) {
	if _, err := r.Seek(int64(ti.Offset), os.SEEK_SET); err != nil {
		return nil, errors.Wrapf(err, "Failed to seek")
	}

	data := make([]byte, ti.Size)

	if _, err := r.Read(data); err != nil {
		return nil, errors.Wrapf(err, "Failed to read")
	}

	var img *image.NRGBA

	width := int(ti.Width)
	height := int(ti.Height)

	switch ti.Format {
	case 0x87000000: // dxt5
		img = textureformats.DecompressImageDX5(data, width, height)
	case 0x85000000: // dxt1
		img = textureformats.DecompressImageDX1(data, width, height)
	default:
		return nil, errors.Errorf("Unsupported image format 0x%x", ti.Format)
	}

	switch ti.Type {
	case 0: // swizzled
		img = ImageUnSwizzle(img)
	default:
		return nil, errors.Errorf("Unsupported image type 0x%x", ti.Type)
	}

	img.SubImage(image.Rect(0, 0, width, height))

	return img, nil
}

func Open(r io.Reader) (*GXT, error) {
	g := &GXT{}

	if err := binary.Read(r, binary.LittleEndian, &g.Header); err != nil {
		return nil, errors.Wrapf(err, "Failed to read header")
	}

	if g.Header.Version != 0x10000003 {
		return nil, errors.Errorf("Unsupported version 0x%x", g.Header.Version)
	}

	if g.Header.PalettCountP4 != 0 || g.Header.PalettCountP8 != 0 {
		return nil, errors.Errorf("Unsupported palettes (%d,%d)", g.Header.PalettCountP4, g.Header.PalettCountP8)
	}

	log.Printf("%+#v", g.Header)

	g.TextureInfos = make([]TextureInfo, g.Header.TexturesCount)

	for i := range g.TextureInfos {
		if err := g.parseTextureInfo(r, &g.TextureInfos[i]); err != nil {
			return nil, errors.Wrapf(err, "Failed to parse texture %d info", i)
		}
	}

	return g, nil
}
