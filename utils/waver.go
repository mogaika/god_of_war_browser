package utils

import (
	"encoding/binary"
	"io"
)

func WaveWriteHeader(w io.Writer, channels uint16, sampleRate uint32, dataSize uint32) error {
	var buf [0x2c]byte
	var pos = 0

	write16 := func(v uint16) {
		binary.LittleEndian.PutUint16(buf[pos:pos+2], v)
		pos += 2
	}

	write32 := func(v uint32) {
		binary.LittleEndian.PutUint32(buf[pos:pos+4], v)
		pos += 4
	}

	write32(0x46464952) // "RIFF"
	write32(36 + dataSize)
	write32(0x45564157)                        // "WAVE"
	write32(0x20746d66)                        // "fmt " chunk
	write32(16)                                // chunk size
	write16(1)                                 // PCM format
	write16(channels)                          // 1 - mono, 2 - stereo ,...
	write32(sampleRate)                        // sampleRate
	write32(sampleRate * uint32(channels) * 2) // byteRate (sampleRate * channels * bytesPerSample)
	write16(uint16(channels) * 2)              // blockAlign (channels * bytesPerSample)
	write16(16)                                // bits per sample
	write32(0x61746164)                        // "data"
	write32(dataSize)                          // data chunk size

	_, err := w.Write(buf[:])
	return err
}
