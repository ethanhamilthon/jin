package voice

import (
	"bytes"
	"encoding/binary"
	"math"
)

// SampleRate is the rate of the recorded audio: 16-bit mono PCM.
const SampleRate = 16000

const (
	bytesPerSecond = SampleRate * 2
	minSeconds     = 0.3
	silentRMS      = 60.0
)

// WAV wraps raw 16-bit mono PCM in a WAV container.
func WAV(pcm []byte) []byte {
	var b bytes.Buffer
	b.WriteString("RIFF")
	_ = binary.Write(&b, binary.LittleEndian, uint32(36+len(pcm)))
	b.WriteString("WAVEfmt ")
	for _, field := range []any{uint32(16), uint16(1), uint16(1), uint32(SampleRate), uint32(bytesPerSecond), uint16(2), uint16(16)} {
		_ = binary.Write(&b, binary.LittleEndian, field)
	}
	b.WriteString("data")
	_ = binary.Write(&b, binary.LittleEndian, uint32(len(pcm)))
	b.Write(pcm)
	return b.Bytes()
}

// Seconds is the length of the PCM audio.
func Seconds(pcm []byte) float64 { return float64(len(pcm)) / bytesPerSecond }

// Speechless reports audio too short or too quiet to hold speech. Models
// invent words for silence, so such audio is not sent.
func Speechless(pcm []byte) bool {
	if Seconds(pcm) < minSeconds {
		return true
	}
	var sum float64
	samples := len(pcm) / 2
	for i := range samples {
		v := float64(int16(binary.LittleEndian.Uint16(pcm[2*i:])))
		sum += v * v
	}
	return math.Sqrt(sum/float64(samples)) < silentRMS
}
