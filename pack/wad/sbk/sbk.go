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

type SmpdStream struct {
	B0  uint8
	B1  uint8
	B2  uint8
	B3  uint8
	U4  uint32
	B8  uint8
	B9  uint8
	B10 uint8
	B11 uint8
	B12 uint8
	B13 uint8
	B14 uint8

	AdpcmOffset uint32
	AdpcmSize   uint32
}

type Command struct {
	W0  uint16
	B2  uint8
	Cmd uint8
	U4  uint32
}

func (c *Command) Parse(b []byte) {
	c.W0 = binary.LittleEndian.Uint16(b[0:])
	c.B2 = b[2]
	c.Cmd = b[3]
	c.U4 = binary.LittleEndian.Uint32(b[4:])
}

type BankSound struct {
	Name          string
	Commands      []Command
	CommandOffset uint32

	B0 uint8
	B1 uint8 // == 1 if streamed from vpk?
	B5 uint8
	B6 uint8
}

func (d *BankSound) Parse(b []byte) {
	d.B0 = b[0]
	d.B1 = b[1]

	d.B5 = b[5]
	d.B6 = b[6]

	// if streamed from vpk file, then
	// B0 = 125
	// B1 = 7 / 6 / 2 - channels count or quality
	// B5 = 0
	// B6 = 0
	// singlecommand
	// B0 = 120
	// B1 = 0
	// B5 = 0
	// B6 = 0
	// multicommand
	// B0 = 110
	// B1 = 4
	// B5 = 1
	// B6 = 8

	d.Commands = make([]Command, b[4])
	d.CommandOffset = binary.LittleEndian.Uint32(b[8:])
}

func (d *BankSound) ParseCommands(c []byte) {

	for i := range d.Commands {
		d.Commands[i].Parse(c[d.CommandOffset+uint32(i)*8:])
	}
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
	AdpcmSize   uint32
	SomeInt2    uint32

	CommandsStart uint32
	BankSounds    []BankSound

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
	commandsStart := u32(0x20)
	b.AdpcmSize = u32(0x28)
	b.SomeInt2 = u32(0x2c)
	b.SmpdStart = u32(0x34)

	b.BankSounds = make([]BankSound, b.SoundsCount)
	for i := range b.BankSounds {
		doff := uint32(0x40 + i*0xc)
		b.BankSounds[i].Parse(h[doff:])
		b.BankSounds[i].ParseCommands(h[commandsStart:])
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
		for i := range sbk.Sounds {
			sbk.Bank.BankSounds[sbk.Sounds[i].StreamId].Name = sbk.Sounds[i].Name
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
