package sounds

import (
	"encoding/binary"
	"math"
)

const sampleRate = 44100

// WAV creates mono 16-bit PCM audio for the built-in tone.
func WAV(id string) []byte {
	tone := find(id)
	duration := 0.0
	for _, n := range tone.notes {
		duration = math.Max(duration, n.start+n.length)
	}
	count := int(math.Ceil(duration * sampleRate))
	data := make([]byte, 44+count*2)
	copy(data, "RIFF")
	binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
	copy(data[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(data[16:], 16)
	binary.LittleEndian.PutUint16(data[20:], 1)
	binary.LittleEndian.PutUint16(data[22:], 1)
	binary.LittleEndian.PutUint32(data[24:], sampleRate)
	binary.LittleEndian.PutUint32(data[28:], sampleRate*2)
	binary.LittleEndian.PutUint16(data[32:], 2)
	binary.LittleEndian.PutUint16(data[34:], 16)
	copy(data[36:], "data")
	binary.LittleEndian.PutUint32(data[40:], uint32(count*2))
	for i := 0; i < count; i++ {
		value, time := 0.0, float64(i)/sampleRate
		for _, n := range tone.notes {
			t := time - n.start
			if t < 0 || t >= n.length {
				continue
			}
			envelope := math.Min(1, t/.008) * math.Pow(1-t/n.length, 2)
			wave := math.Sin(2 * math.Pi * n.frequency * t)
			if tone.wave == "triangle" {
				wave = 2 / math.Pi * math.Asin(wave)
			}
			value += wave * envelope * .25
		}
		value = math.Max(-1, math.Min(1, value))
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(int16(value*32767)))
	}
	return data
}
