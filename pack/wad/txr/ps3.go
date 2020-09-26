package txr

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/bits"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

type Ps3Texture struct {
	Unk00           uint32
	DataTotalSize   uint32
	Unk08           uint32
	Zero0c          uint32
	DataOffset      uint32
	DataPayloadSize uint32
	Unk18           uint8
	MipMapCounts    uint8
	Unk1a           uint8
	Zero1b          uint8
	Unk1c           uint32
	Width           uint16
	Height          uint16
	Zero24          uint8
	Unk25           uint8

	images []image.Image
}

func (t *Ps3Texture) Images() []image.Image {
	return t.images
}

type Ps3TextureAjax struct {
	Ps3Texture
	Images [][]byte
}

func (t *Ps3Texture) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	a := &Ps3TextureAjax{
		Ps3Texture: *t,
		Images:     make([][]byte, len(t.images)),
	}

	for i, img := range t.images {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
		a.Images[i] = buf.Bytes()
	}

	return a, nil
}

func (t *Ps3Texture) checkUnksAndZeros() error {
	if t.Unk00 != 0x2000000 {
		return fmt.Errorf("Incorrect Unk00: %v", t.Unk00)
	}
	if t.Unk08 != 1 {
		return fmt.Errorf("Incorrect Unk08: %v", t.Unk08)
	}
	if t.Zero0c != 0 {
		return fmt.Errorf("Incorrect Zero0c: %v", t.Zero0c)
	}
	if t.Unk18 != 0x85 && t.Unk18 != 0x9e {
		return fmt.Errorf("Incorrect Unk18: %v", t.Unk18)
	}
	if t.Unk1a != 2 {
		return fmt.Errorf("Incorrect Unk1a: %v", t.Unk1a)
	}
	// 0xA9E4 - fog~pal_fog
	if t.Unk1c != 0xAAE4 && t.Unk1c != 0xA9E4 {
		return fmt.Errorf("Incorrect Unk1c: %v", t.Unk1c)
	}
	if t.Zero24 != 0 {
		return fmt.Errorf("Incorrect Zero24: %v", t.Zero24)
	}
	if t.Unk25 != 1 {
		return fmt.Errorf("Incorrect Unk25: %v", t.Unk25)
	}

	return nil
}

func ps3SwizzleIndex(x, y, width, height uint32) int {
	offset := uint32(0)
	shift := uint32(0)

	log2w := bits.TrailingZeros32(width)
	log2h := bits.TrailingZeros32(height)

	for {
		// log.Printf("before step  %v,%v,%v  off %v  shift %v  log %v %v %v",
		//	x, y, z, offset, shift, log2w, log2h, log2d)
		if log2w > 0 {
			offset |= (x & 1) << shift
			shift++
			x >>= 1
			log2w--
		}
		if log2h > 0 {
			offset |= (y & 1) << shift
			shift++
			y >>= 1
			log2h--
		}
		// log.Printf("after step   %v,%v,%v  off %v  shift %v  log %v %v %v",
		//	x, y, z, offset, shift, log2w, log2h, log2d)
		if x > 0 || y > 0 {
			continue
		} else {
			//log.Printf("RETURN    %v,%v => %v", in_x, in_y, offset)
			return int(offset)
		}
	}
}

func (t *Ps3Texture) imageFromBs(bs *utils.BufStack, width, height int, unswizzle bool) image.Image {
	i := image.NewNRGBA(image.Rect(0, 0, width, height))

	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := index * 4
			if unswizzle {
				pos = ps3SwizzleIndex(uint32(x), uint32(y), uint32(width), uint32(height)) * 4
			}
			i.SetNRGBA(x, y, color.NRGBA{
				A: bs.Byte(pos),
				R: bs.Byte(pos + 1),
				G: bs.Byte(pos + 2),
				B: bs.Byte(pos + 3),
			})
			index++
		}
	}
	bs.VerifySize(index * 4)

	return i
}

func (t *Ps3Texture) loadImages(dataBs *utils.BufStack) error {
	t.images = make([]image.Image, 0)
	dataOffset := uint32(0)
	curW := t.Width
	curH := t.Height
	for mipmapId := 0; ; mipmapId++ {
		if curW == 0 && curH == 0 {
			if mipmapId != int(t.MipMapCounts) {
				return fmt.Errorf("Mipmap count and detected count do not match (%v != %v)", mipmapId, t.MipMapCounts)
			}
			break
		}

		if curW == 0 {
			curW = 1
		}
		if curH == 0 {
			curH = 1
		}

		imageRealSize := uint32(curW) * uint32(curH) * 4

		mipmapBs := dataBs.SubBuf(fmt.Sprintf("mipmap%d", mipmapId), int(dataOffset)).SetSize(int(imageRealSize))

		t.images = append(t.images, t.imageFromBs(mipmapBs, int(curW), int(curH), true))

		dataOffset += imageRealSize
		curW /= 2
		curH /= 2
	}
	return nil
}

func NewPs3TextureFromData(bs *utils.BufStack) (*Ps3Texture, error) {
	bs.SubBuf("serverId", 0).SetSize(4)
	texBs := bs.SubBuf("ps3texture", 4)
	headerBs := texBs.SubBuf("header", 0).SetSize(0x80)

	t := &Ps3Texture{
		Unk00:           headerBs.BU32(0),
		DataTotalSize:   headerBs.BU32(0x04),
		Unk08:           headerBs.BU32(0x08),
		Zero0c:          headerBs.BU32(0x0c),
		DataOffset:      headerBs.BU32(0x10),
		DataPayloadSize: headerBs.BU32(0x14),
		Unk18:           headerBs.Byte(0x18),
		MipMapCounts:    headerBs.Byte(0x19),
		Unk1a:           headerBs.Byte(0x1a),
		Zero1b:          headerBs.Byte(0x1b),
		Unk1c:           headerBs.BU32(0x1c),
		Width:           headerBs.BU16(0x20),
		Height:          headerBs.BU16(0x22),
		Zero24:          headerBs.Byte(0x24),
		Unk25:           headerBs.Byte(0x25),
	}

	if err := t.checkUnksAndZeros(); err != nil {
		return nil, fmt.Errorf("Invalid unk or zero val: %v", err)
	}

	dataBs := texBs.SubBuf("data", int(t.DataOffset)).SetSize(int(t.DataTotalSize))
	payloadDataBs := dataBs.SubBuf("payload", 0).SetSize(int(t.DataPayloadSize))
	dataBs.SubBuf("padding", int(t.DataPayloadSize)).SetSize(int(t.DataTotalSize - t.DataPayloadSize))

	if err := t.loadImages(payloadDataBs); err != nil {
		return nil, fmt.Errorf("Error loading images: %v", err)
	}

	log.Printf("\n%v", bs.StringTree())

	return t, nil
}

func (txr *Texture) findPSNextGenTexture(wrsrc *wad.WadNodeRsrc) (*wad.Node, wad.File, error) {
	node := wrsrc.Wad.GetNodeByName(txr.SubTxrName, wrsrc.Node.Id, false)
	if node == nil {
		return nil, nil, fmt.Errorf("Cannot find next gen texture: %s", txr.SubTxrName)
	}

	texture, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id)
	if err != nil {
		return node, nil, fmt.Errorf("Error getting next gen texture %s: %v", txr.SubTxrName, err)
	}

	return node, texture, nil
}

func (txr *Texture) changeTexturePS3(wrsrc *wad.WadNodeRsrc, img image.Image) error {
	_, _, err := txr.findPSNextGenTexture(wrsrc)
	if err != nil {
		return err
	}

	return fmt.Errorf("Not implemented")
}
