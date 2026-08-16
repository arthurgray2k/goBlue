package player

import (
	"bytes"
	"encoding/binary"
	"math"
)

// GenerateTestChimeWAV creates a clean 1.2-second stereo acoustic chime (C5-E5-G5 triad) as a WAV byte slice.
func GenerateTestChimeWAV() []byte {
	const (
		sampleRate    = 44100
		numChannels   = 2
		bitsPerSample = 16
		durationSec   = 1.2
	)

	totalSamples := int(sampleRate * durationSec)
	var audioData bytes.Buffer

	// Frequencies for a pleasant major triad chime: C5 (523.25Hz), E5 (659.25Hz), G5 (783.99Hz)
	freqs := []float64{523.25, 659.25, 783.99}

	for i := 0; i < totalSamples; i++ {
		t := float64(i) / float64(sampleRate)

		// Exponential decay envelope
		envelope := math.Exp(-2.5 * t)
		// Fade-in during first 20ms to prevent click
		if t < 0.02 {
			envelope *= (t / 0.02)
		}

		var sampleVal float64
		for _, f := range freqs {
			sampleVal += math.Sin(2.0 * math.Pi * f * t)
		}
		sampleVal = (sampleVal / float64(len(freqs))) * envelope * 0.7

		// 16-bit PCM integer
		val16 := int16(sampleVal * 32767.0)

		// Left channel and Right channel
		_ = binary.Write(&audioData, binary.LittleEndian, val16)
		_ = binary.Write(&audioData, binary.LittleEndian, val16)
	}

	dataBytes := audioData.Bytes()
	dataSize := uint32(len(dataBytes))

	var wav bytes.Buffer
	// RIFF header
	wav.WriteString("RIFF")
	_ = binary.Write(&wav, binary.LittleEndian, uint32(36+dataSize))
	wav.WriteString("WAVE")

	// fmt subchunk
	wav.WriteString("fmt ")
	_ = binary.Write(&wav, binary.LittleEndian, uint32(16))          // Subchunk1Size for PCM
	_ = binary.Write(&wav, binary.LittleEndian, uint16(1))           // AudioFormat (1 = PCM)
	_ = binary.Write(&wav, binary.LittleEndian, uint16(numChannels)) // NumChannels
	_ = binary.Write(&wav, binary.LittleEndian, uint32(sampleRate))  // SampleRate
	byteRate := uint32(sampleRate * numChannels * bitsPerSample / 8)
	_ = binary.Write(&wav, binary.LittleEndian, byteRate) // ByteRate
	blockAlign := uint16(numChannels * bitsPerSample / 8)
	_ = binary.Write(&wav, binary.LittleEndian, blockAlign)            // BlockAlign
	_ = binary.Write(&wav, binary.LittleEndian, uint16(bitsPerSample)) // BitsPerSample

	// data subchunk
	wav.WriteString("data")
	_ = binary.Write(&wav, binary.LittleEndian, dataSize)
	wav.Write(dataBytes)

	return wav.Bytes()
}
