package vagp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/mogaika/god_of_war_browser/ps2/adpcm"
	"github.com/mogaika/god_of_war_browser/utils"
)

type VAGP struct {
	WaveData   []byte `json:"-"`
	Channels   byte
	SampleRate uint32
}

func NewVAGPFromReader(r io.Reader) (*VAGP, error) {
	var buf [0x30]byte
	if _, err := r.Read(buf[:]); err != nil {
		return nil, err
	}

	if bytes.Compare([]byte{0x56, 0x41, 0x47, 0x70}, buf[:4]) != 0 {
		return nil, errors.New("Magic not matched")
	}

	vagp := &VAGP{
		Channels:   buf[0x1E],
		SampleRate: binary.BigEndian.Uint32(buf[0x10:0x14]),
		WaveData:   make([]byte, binary.BigEndian.Uint32(buf[0xC:0x10])),
	}

	if _, err := r.Read(vagp.WaveData); err != nil {
		return nil, err
	}

	return vagp, nil
}

func (vagp *VAGP) AsWave() (*bytes.Buffer, error) {
	if vagp.Channels > 1 {
		panic("Not mono not supported")
	}

	var buf bytes.Buffer

	//WaveWriteHeader(w io.Writer, channels int16, sampleRate uint32, dataSize uint32) error {
	if err := utils.WaveWriteHeader(&buf, 1, vagp.SampleRate, uint32((len(vagp.WaveData)/16)*28*2)); err != nil {
		return nil, err
	}

	adpcmstream := adpcm.NewAdpcmToWaveStream(&buf)
	_, err := adpcmstream.Write(vagp.WaveData)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}
