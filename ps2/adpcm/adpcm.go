package adpcm

import (
	"errors"
	"io"
	"log"
)

var vag_f = [5][2]float64{
	{0.0, 0.0},
	{60.0 / 64.0, 0.0},
	{115.0 / 64.0, -52.0 / 64.0},
	{98.0 / 64.0, -55.0 / 64.0},
	{122.0 / 64.0, -60.0 / 64.0},
}

type AdpcmStream struct {
	hist1 float64
	hist2 float64
}

func AdpcmSizeToWaveSize(size int) int {
	// input:  4  bit adpcm size
	// output: 16 bit pcf   size
	return (size / 16) * 28 * 2
}

func (stream *AdpcmStream) Unpack(packs []byte) ([]byte, error) {
	if len(packs)%16 != 0 {
		return nil, errors.New("Support only 8 bytes blocks of stream")
	}

	var result = make([]byte, AdpcmSizeToWaveSize(len(packs)))
	var fsamples [28]float64

	iResultPos := 0
	for iBlock := 0; iBlock < len(packs)/16; iBlock++ {
		blockStart := iBlock * 16

		// used to fill space to end of vpk file
		if packs[blockStart] == 0xc0 {
			continue
		}

		predict_nr := uint32(packs[blockStart])
		shift_factor := predict_nr & 0xf
		predict_nr >>= 4

		if predict_nr > 5 {
			log.Printf("Strange sound. PredictNr > 5: %v. Block: %v", predict_nr, packs[blockStart:blockStart+16])
			predict_nr = 0
		}

		streampos := blockStart + 2

		for i := 0; i < 28; i += 2 {
			sample := uint32(packs[streampos])
			streampos++

			scale := int16(sample&0xf) << 12
			fsamples[i] = float64(scale >> shift_factor)

			scale = int16(sample&0xf0) << 8
			fsamples[i+1] = float64(scale >> shift_factor)
		}

		bsamples := result[iResultPos:]
		for i := range fsamples {
			fsamples[i] = fsamples[i] + stream.hist1*vag_f[predict_nr][0] + stream.hist2*vag_f[predict_nr][1]
			stream.hist2 = stream.hist1
			stream.hist1 = fsamples[i]
			d := int(fsamples[i] + 0.5)
			bsamples[i*2] = byte(d & 0xff)
			bsamples[i*2+1] = byte(d >> 8)
		}
		iResultPos += 56
	}
	return result[:iResultPos], nil
}

func NewAdpcmStream() *AdpcmStream {
	return &AdpcmStream{}
}

// in - adcmp
// out - 16 bit pcm
type AdpcmToWaveStream struct {
	outstream io.Writer
	AdpcmStream
}

func NewAdpcmToWaveStream(waveWriterOutput io.Writer) *AdpcmToWaveStream {
	return &AdpcmToWaveStream{outstream: waveWriterOutput}
}

func (stream *AdpcmToWaveStream) Write(p []byte) (n int, err error) {
	if len(p)%16 != 0 {
		return 0, errors.New("Support only 8 bytes blocks of stream")
	}

	packed, err := stream.Unpack(p)
	if err != nil {
		return 0, err
	}

	return stream.outstream.Write(packed)
}
