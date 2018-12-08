package sbk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/ps2/vagp"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/webutils"
)

const SBK_SBLK_MAGIC = 0x18
const SBK_VAG_MAGIC = 0x40018

type Sound struct {
	Name     string
	StreamId uint32 // file offset for vag
}

type BankDesc struct {
	F0  uint32
	F47 [4]uint8
	F8  uint32
}

type Bank struct {
	HeaderBlockStart uint32
	HeaderBlockSize  uint32
	StreamBlockStart uint32
	StreamBlockSize  uint32
	StreamBlock      []byte `json:"-"`
	Sounds           []byte `json:"-"`

	PseudoName  string
	SoundsCount uint16
	SomeInt1    uint32
	SomeInt2    uint32

	StartOfDecsSection uint32
	Descs              []BankDesc

	SmpdStart uint32
}

func (b *Bank) parseHeader(h []byte) error {
	u32 := func(pos uint32) uint32 {
		return binary.LittleEndian.Uint32(h[pos : pos+4])
	}
	u16 := func(pos uint32) uint16 {
		return binary.LittleEndian.Uint16(h[pos : pos+4])
	}

	b.PseudoName = utils.ReverseString(utils.BytesToString(h[0xc:0x10]))
	b.SoundsCount = u16(0x16)
	b.StartOfDecsSection = u32(0x20)
	b.SomeInt1 = u32(0x28)
	b.SomeInt2 = u32(0x2c)
	b.SmpdStart = u32(0x34)

	b.Descs = make([]BankDesc, b.SoundsCount)
	for i := range b.Descs {
		d := &b.Descs[i]
		doff := uint32(0x40 + i*12)
		d.F0 = u32(doff)
		copy(d.F47[:], h[doff+4:doff+8])
		d.F8 = u32(doff + 8)
	}

	return nil
}

type SBK struct {
	Sounds     []Sound
	IsVagFiles bool // if false - than Bank present
	Bank       *Bank
}

func (sbk *SBK) loadBank(r io.ReaderAt) error {
	var sbkInfoBuf [24]byte
	if _, err := r.ReadAt(sbkInfoBuf[:], 0); err != nil {
		return err
	}

	sbk.Bank = &Bank{
		HeaderBlockStart: binary.LittleEndian.Uint32(sbkInfoBuf[8:12]),
		HeaderBlockSize:  binary.LittleEndian.Uint32(sbkInfoBuf[12:16]),
		StreamBlockStart: binary.LittleEndian.Uint32(sbkInfoBuf[16:20]),
		StreamBlockSize:  binary.LittleEndian.Uint32(sbkInfoBuf[20:24]),
	}

	headBuf := make([]byte, sbk.Bank.HeaderBlockSize)
	if _, err := r.ReadAt(headBuf, int64(sbk.Bank.HeaderBlockStart)); err != nil {
		return err
	}

	if sbk.Bank.StreamBlockSize != 0 {
		sbk.Bank.StreamBlock = make([]byte, sbk.Bank.StreamBlockSize)
		if _, err := r.ReadAt(sbk.Bank.StreamBlock, int64(sbk.Bank.StreamBlockStart)); err != nil {
			return err
		}
	}

	return sbk.Bank.parseHeader(headBuf)
}

func NewFromData(f io.ReaderAt, isSblk bool, size uint32) (*SBK, error) {
	var soundsCount uint32
	if err := binary.Read(io.NewSectionReader(f, 4, 8), binary.LittleEndian, &soundsCount); err != nil {
		return nil, err
	}

	sbk := &SBK{
		Sounds:     make([]Sound, soundsCount),
		IsVagFiles: !isSblk,
	}

	for i := range sbk.Sounds {
		var buf [28]byte
		_, err := f.ReadAt(buf[:], 8+int64(i)*28)
		if err != nil {
			return nil, err
		}

		sbk.Sounds[i].Name = utils.BytesToString(buf[:24])
		sbk.Sounds[i].StreamId = binary.LittleEndian.Uint32(buf[24:])
	}

	if isSblk {
		bankLen := int64(8 + len(sbk.Sounds)*28)
		if err := sbk.loadBank(io.NewSectionReader(f, bankLen, int64(size)-bankLen)); err != nil {
			return sbk, err
		}
	}

	return sbk, nil
}

func (sbk *SBK) httpSendSound(w http.ResponseWriter, wrsrc *wad.WadNodeRsrc, sndName string, needWav bool) {
	if sbk.IsVagFiles {
		for iSnd, snd := range sbk.Sounds {
			if snd.Name == sndName {
				end := uint32(len(wrsrc.Tag.Data))
				if iSnd != len(sbk.Sounds)-1 {
					end = sbk.Sounds[iSnd+1].StreamId
				}
				log.Println(len(wrsrc.Tag.Data), snd.StreamId, end, end-snd.StreamId)
				vagpReader := bytes.NewReader(wrsrc.Tag.Data[snd.StreamId:end])
				if needWav {
					vag, err := vagp.NewVAGPFromReader(vagpReader)
					if err != nil {
						webutils.WriteError(w, err)
					} else {
						wav, err := vag.AsWave()
						if err != nil {
							webutils.WriteError(w, err)
						} else {
							webutils.WriteFile(w, wav, sndName+".WAV")
						}
					}
				} else {
					webutils.WriteFile(w, vagpReader, sndName+".VAG")
				}
				return
			}
		}
		webutils.WriteError(w, errors.New("Cannot find sound"))
	} else {
		start := int(8 + len(sbk.Sounds)*28)
		webutils.WriteFile(w, bytes.NewReader(wrsrc.Tag.Data[start:]), wrsrc.Name()+".SBK")
	}
}

func (sbk *SBK) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	sndName := r.URL.Query().Get("snd")

	switch action {
	case "wav":
		sbk.httpSendSound(w, wrsrc, sndName, true)
	case "vag":
		sbk.httpSendSound(w, wrsrc, sndName, false)
	default:
		log.Printf("Unknown action: %v", action)
	}

}

func (sbk *SBK) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return sbk, nil
}

func init() {
	wad.SetHandler(config.GOW1, SBK_SBLK_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(bytes.NewReader(wrsrc.Tag.Data), true, wrsrc.Tag.Size)
	})
	wad.SetHandler(config.GOW1, SBK_VAG_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(bytes.NewReader(wrsrc.Tag.Data), false, wrsrc.Tag.Size)
	})
}
