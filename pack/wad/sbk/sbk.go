package sbk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/ps2/adpcm"
	"github.com/mogaika/god_of_war_browser/ps2/vagp"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/webutils"
)

const SBK_SBLK_MAGIC = 0x18
const SBK_VAG_MAGIC = 0x40018
const GOW2_SBP_MAGIC = 0x00000015

type Sound struct {
	Name     string
	StreamId uint32 // file offset for vag
}

type UnkRef struct {
	StreamId uint32
}

func (u *UnkRef) Parse(bo binary.ByteOrder, b []byte) {
	u.StreamId = bo.Uint32(b[12:])
}

type VagRef struct {
	Name string
	B16  uint8
}

func (v *VagRef) Parse(b []byte) {
	v.Name = utils.BytesToString(b[8:16])
	v.B16 = b[16]
}

type SampleRef struct {
	// ADSR ?
	// https://github.com/PCSX2/pcsx2/blob/master/pcsx2/SPU2/defs.h#L115
	B0  uint8
	B1  uint8
	B2  uint8 // Pitch or speed related
	B3  uint8
	W4  uint16
	B6  uint8 // 0
	B7  uint8 // 0
	B8  uint8 // usually same as B9
	B9  uint8 // somewhere in range 0-5
	B10 uint8
	B11 uint8 // always 0xff
	W12 uint16
	B14 uint8
	B15 uint8

	AdpcmOffset uint32
	AdpcmSize   uint32
}

func (s *SampleRef) Parse(bo binary.ByteOrder, b []byte) {
	s.B0 = b[0]
	s.B1 = b[1]
	s.B2 = b[2]
	s.B3 = b[3]
	s.W4 = bo.Uint16(b[4:])
	s.B6 = b[6]
	s.B7 = b[7]
	s.B8 = b[8]
	s.B9 = b[9]
	s.B10 = b[10]
	s.B11 = b[11]
	s.W12 = bo.Uint16(b[12:])
	s.B14 = b[14]
	s.B15 = b[15]
	s.AdpcmOffset = bo.Uint32(b[16:])
	s.AdpcmSize = bo.Uint32(b[20:])
}

type Command struct {
	/*
		55f0 or 55e8

			Cmd:
			1 - play ? b2+cmd = addr SMPD
			5 - ?
			7 - play from external bank ? b2+cmd = addr of 0x20 [SMPD, u32, char[8], u32, pad[8]]
			8 - goto to other bank depending on result of cmd 34
			20 - U4 (50, 0)
			21 - B1 (0,100)
			22 - 0 - at end
			24 - 0
			25 - B1 (3,4,5,6) B2 (1) - random jump? B1 - rand max, B2 - always 1?
			26 - B3 (40), U4(200)
			27 - B1 (100)
			30 - B2 (115)
			31 - B2 (236) B3 (20), B1 (2), B2 (54), B3 (72)
			34 - set index for logical split B2=1 B3=stream index. paired with cmd 8
			35 - 0 at the end, maybe random choose mark?
			36 - 0
			39 - B1 (4) B2 (1)
			40 - B1 (236)

	*/
	Bytes [4]byte
	U0    uint32
	U4    uint32 // Delay ?

	Cmd       uint8
	SampleRef *SampleRef
	VagRef    *VagRef
	UnkRef    *UnkRef
}

func (c *Command) Parse(bo binary.ByteOrder, bsRefs *utils.BufStack, b []byte) {
	c.U0 = bo.Uint32(b[0:])
	c.U4 = bo.Uint32(b[4:])

	binary.BigEndian.PutUint32(c.Bytes[:], c.U0)

	c.Cmd = byte(c.U0 >> 24)
	addr := int(c.U0 & (uint32(1)<<24 - 1))

	switch c.Cmd {
	case 1:
		c.SampleRef = &SampleRef{}
		c.SampleRef.Parse(bo, bsRefs.Raw()[addr:])
		bsRefs.SubBuf("cmd_1_sam", addr)
	case 5:
		c.UnkRef = &UnkRef{}
		c.UnkRef.Parse(bo, bsRefs.Raw()[addr:])
		bsRefs.SubBuf("cmd_5_unk", addr)
	case 7:
		c.VagRef = &VagRef{}
		c.VagRef.Parse(bsRefs.Raw()[addr:])
		utils.LogDump(c.VagRef)
		bsRefs.SubBuf("cmd_7_vag", addr)
	case 8:
		c.UnkRef = &UnkRef{}
		c.UnkRef.Parse(bo, bsRefs.Raw()[addr:])
		bsRefs.SubBuf("cmd_8_unk", addr)
	}

}

type BankSound struct {
	name          string
	Commands      []Command
	commandOffset uint32

	B0 uint8
	B1 uint8
	B5 uint8
	B6 uint8
}

func (d *BankSound) Parse(bo binary.ByteOrder, b []byte) {
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
	d.commandOffset = bo.Uint32(b[8:])
}

func (d *BankSound) ParseCommands(bo binary.ByteOrder, bsCommands, bsSMPD *utils.BufStack) {
	c := bsCommands.Raw()
	for i := range d.Commands {
		d.Commands[i].Parse(bo, bsSMPD, c[d.commandOffset+uint32(i)*8:])
	}
}

type Bank struct {
	HeaderBlockStart uint32
	HeaderBlockSize  uint32
	StreamBlockStart uint32
	StreamBlockSize  uint32
	StreamBlock      *utils.BufStack `json:"-"`
	Sounds           []byte          `json:"-"`

	PseudoName  string
	SoundsCount uint16
	AdpcmSize   uint32
	SomeInt2    uint32

	CommandsStart uint32
	BankSounds    []BankSound

	SmpdStart uint32
}

func (b *Bank) parseHeader(bo binary.ByteOrder, bsHeader *utils.BufStack) error {
	b.PseudoName = utils.ReverseString(utils.BytesToString(bsHeader.Raw()[0xc:0x10]))
	b.SoundsCount = bsHeader.EU16(bo, 0x16)
	commandsStart := bsHeader.EU32(bo, 0x20)
	b.AdpcmSize = bsHeader.EU32(bo, 0x28)
	b.SomeInt2 = bsHeader.EU32(bo, 0x2c)
	b.SmpdStart = bsHeader.EU32(bo, 0x34)

	bsSoundsInfo := bsHeader.SubBuf("sounds_info", 0x40).SetSize(0xc * int(b.SoundsCount))
	bsCommands := bsHeader.SubBuf("commands", int(commandsStart)).SetSize(int(b.SmpdStart) - int(commandsStart))
	bsSMPD := bsHeader.SubBuf("smpd_streams", int(b.SmpdStart)).Expand()

	b.BankSounds = make([]BankSound, b.SoundsCount)
	for i := range b.BankSounds {
		b.BankSounds[i].Parse(bo, bsSoundsInfo.Read(0xc))
		b.BankSounds[i].ParseCommands(bo, bsCommands, bsSMPD)
	}

	return nil
}

type SBK struct {
	Sounds     []Sound
	IsVagFiles bool // if false - than Bank present
	Bank       *Bank
}

func (sbk *SBK) loadBank(bsBank *utils.BufStack) error {
	bsBankInfo := bsBank.SubBuf("bank_info", 0).SetSize(24)

	var bo binary.ByteOrder
	switch config.GetPlayStationVersion() {
	case config.PS3:
		bo = binary.BigEndian
	default:
		bo = binary.LittleEndian
	}

	sbk.Bank = &Bank{
		HeaderBlockStart: bsBankInfo.EU32(bo, 0x8),
		HeaderBlockSize:  bsBankInfo.EU32(bo, 0xc),
		StreamBlockStart: bsBankInfo.EU32(bo, 0x10),
		StreamBlockSize:  bsBankInfo.EU32(bo, 0x14),
	}

	bsBankHeader := bsBank.SubBuf("bank_header",
		int(sbk.Bank.HeaderBlockStart)).SetSize(int(sbk.Bank.HeaderBlockSize))

	if sbk.Bank.StreamBlockSize != 0 {
		sbk.Bank.StreamBlock = bsBank.SubBuf("bank_stream",
			int(sbk.Bank.StreamBlockStart)).SetSize(int(sbk.Bank.StreamBlockSize))
	}

	return sbk.Bank.parseHeader(bo, bsBankHeader)
}

func NewFromData(bs *utils.BufStack, isSblk bool) (*SBK, error) {
	bsHead := bs.SubBuf("head", 0).SetSize(8)

	defer func() { log.Println(bs.StringTree()) }()

	bsHead.Skip(4)
	soundsCount := bsHead.ReadLU32()

	sbk := &SBK{
		Sounds:     make([]Sound, soundsCount),
		IsVagFiles: !isSblk,
	}

	bsSoundInfo := bsHead.SubBufFollowing("sounds_info").SetSize(28 * int(soundsCount))

	for i := range sbk.Sounds {
		sbk.Sounds[i].Name = bsSoundInfo.ReadStringBuffer(24)
		sbk.Sounds[i].StreamId = bsSoundInfo.ReadLU32()
	}

	if isSblk {
		bsBank := bsSoundInfo.SubBufFollowing("banks").Expand()

		if err := sbk.loadBank(bsBank); err != nil {
			return sbk, errors.Wrapf(err, "Failed to load banks")
		}

		for i := range sbk.Sounds {
			sbk.Bank.BankSounds[sbk.Sounds[i].StreamId].name = sbk.Sounds[i].Name
		}
	}

	return sbk, nil
}

func (sbk *SBK) httpSendBankSMPD(w http.ResponseWriter, wrsrc *wad.WadNodeRsrc, offset, size int) {
	w.Header().Add("Content-Type", "audio/wav")

	webutils.WriteFileHeaders(w, fmt.Sprintf("%s_%d_%d.WAV", wrsrc.Tag.Name, offset, size))

	if err := utils.WaveWriteHeader(w, 1, 22050, uint32((size/16)*28*2)); err != nil {
		webutils.WriteError(w, err)
	}

	adpcmstream := adpcm.NewAdpcmToWaveStream(w)

	if _, err := adpcmstream.Write(sbk.Bank.StreamBlock.Raw()[offset : offset+size]); err != nil {
		webutils.WriteError(w, err)
	}
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
							w.Header().Add("Content-Type", "audio/wav")
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
	case "smpd":
		var offset, size int
		fmt.Sscan(r.URL.Query().Get("offset"), &offset)
		fmt.Sscan(r.URL.Query().Get("size"), &size)

		sbk.httpSendBankSMPD(w, wrsrc, offset, size)
	default:
		log.Printf("Unknown action: %v", action)
	}

}

func (sbk *SBK) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return sbk, nil
}

func init() {
	wad.SetServerHandler(config.GOW1, SBK_SBLK_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(utils.NewBufStack("sblk", wrsrc.Tag.Data), true)
	})
	wad.SetServerHandler(config.GOW1, SBK_VAG_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(utils.NewBufStack("sbk_vag", wrsrc.Tag.Data), false)
	})

	wad.SetServerHandler(config.GOW2, GOW2_SBP_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(utils.NewBufStack("sbp_vag", wrsrc.Tag.Data), true)
	})
}
