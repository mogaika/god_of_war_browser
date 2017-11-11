package vpk

import (
	"encoding/binary"
	"io"

	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/ps2/adpcm"
	"github.com/mogaika/god_of_war_browser/utils"
)

type VPK struct {
	SampleRate uint32
	Channels   uint32
	DataSize   uint32 // of one channel
}

func NewVPKFromReader(r io.ReaderAt) (*VPK, error) {
	var header [0x20]byte
	if _, err := r.ReadAt(header[:], 0); err != nil {
		return nil, err
	}

	vpk := &VPK{
		DataSize:   binary.LittleEndian.Uint32(header[0x4:0x8]),
		SampleRate: binary.LittleEndian.Uint32(header[0x10:0x14]),
		Channels:   binary.LittleEndian.Uint32(header[0x14:0x18]), // probably incorrect
	}

	return vpk, nil
}

func (vpk *VPK) AsWave(r io.Reader, w io.Writer) (int, error) {
	if vpk.Channels > 4 {
		panic(vpk.Channels)
	}

	var inBlockSize = 0x1000 * vpk.Channels

	var in = make([]byte, inBlockSize)
	var out = make([]byte, adpcm.AdpcmSizeToWaveSize(int(inBlockSize)))

	// skip first sector
	if _, err := r.Read(in[:utils.SECTOR_SIZE]); err != nil {
		return 0, err
	}

	var adpcmStream = make([]*adpcm.AdpcmStream, vpk.Channels)
	for i := range adpcmStream {
		adpcmStream[i] = adpcm.NewAdpcmStream()
	}

	if err := utils.WaveWriteHeader(w, uint16(vpk.Channels), vpk.SampleRate, uint32(adpcm.AdpcmSizeToWaveSize(int(vpk.DataSize))*2)); err != nil {
		return 0, err
	}

	inLeft := int(vpk.DataSize)
	n := 0
	for inLeft > 0 {
		_, err := r.Read(in)
		if err != nil {
			return 0, err
		}

		datalen := 0x1000
		if inLeft < 0x1000 {
			datalen = inLeft
		}

		for i, s := range adpcmStream {
			buf, err := s.Unpack(in[i*0x1000 : i*0x1000+datalen])
			if err != nil {
				return 0, err
			}

			for k := 0; k < len(buf)/2; k++ {
				j := uint32(k * 2) // 16bit
				outchpos := j*vpk.Channels + uint32(i*2)
				out[outchpos] = buf[j]
				out[outchpos+1] = buf[j+1]
			}

		}

		if wn, err := w.Write(out); err != nil {
			return n, err
		} else {
			n += wn
		}

		inLeft -= 0x1000
	}

	return n, nil
}

func init() {
	h := func(p utils.ResourceSource, r *io.SectionReader) (interface{}, error) {
		return NewVPKFromReader(r)
	}
	pack.SetHandler(".VPK", h)
	pack.SetHandler(".VP1", h)
	pack.SetHandler(".VP2", h)
	pack.SetHandler(".VP3", h)
	pack.SetHandler(".VP4", h)
}
